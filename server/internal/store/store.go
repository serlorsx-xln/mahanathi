package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/normalize"

	_ "modernc.org/sqlite" // the pure-Go driver, registered as "sqlite"
)

// Retention is how long readings are kept.
const Retention = 30 * 24 * time.Hour

// ErrNewerSchema means the database was migrated by a newer binary. Running
// an older binary on it could misread columns, so Open refuses.
var ErrNewerSchema = errors.New("store: the database schema is newer than this binary")

// Where a reading came from. The load rounds water levels to 2 decimals and
// the graph gives 3, so for the same minute the higher source wins.
const (
	sourceLoad  = 1
	sourceGraph = 2
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// migrations are applied in file-name order; migration i brings the schema to
// user_version i+1.
var migrations = func() []string {
	names, err := fs.Glob(migrationFiles, "migrations/*.sql")
	if err != nil {
		panic(err)
	}
	sort.Strings(names)
	out := make([]string, len(names))
	for i, n := range names {
		b, err := migrationFiles.ReadFile(n)
		if err != nil {
			panic(err)
		}
		out[i] = string(b)
	}
	return out
}()

// Store is the SQLite database: one writer connection, so writes never
// contend, and a small pool of read-only connections, which WAL lets run
// alongside it.
type Store struct {
	w, r *sql.DB
}

// Open opens or creates the database at path and applies any migrations.
func Open(ctx context.Context, path string) (*Store, error) {
	if strings.ContainsAny(path, "?#") {
		return nil, fmt.Errorf("store: path %q may not contain ? or #", path)
	}
	w, err := sql.Open("sqlite", dsn(path, true))
	if err != nil {
		return nil, err
	}
	w.SetMaxOpenConns(1)
	w.SetConnMaxIdleTime(0)
	if err := migrate(ctx, w); err != nil {
		w.Close()
		return nil, err
	}
	r, err := sql.Open("sqlite", dsn(path, false))
	if err != nil {
		w.Close()
		return nil, err
	}
	r.SetMaxOpenConns(4)
	if err := r.PingContext(ctx); err != nil {
		w.Close()
		r.Close()
		return nil, err
	}
	return &Store{w: w, r: r}, nil
}

func dsn(path string, writer bool) string {
	q := url.Values{}
	q.Add("_pragma", "busy_timeout(5000)")
	q.Add("_pragma", "foreign_keys(1)")
	if writer {
		q.Add("_pragma", "journal_mode(WAL)")
		q.Add("_pragma", "synchronous(NORMAL)")
		q.Set("_txlock", "immediate")
	} else {
		q.Add("_pragma", "query_only(1)")
	}
	return "file:" + path + "?" + q.Encode()
}

func migrate(ctx context.Context, db *sql.DB) error {
	var v int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&v); err != nil {
		return fmt.Errorf("store: reading schema version: %w", err)
	}
	if v > len(migrations) {
		return fmt.Errorf("%w (database %d, binary %d)", ErrNewerSchema, v, len(migrations))
	}
	for i := v; i < len(migrations); i++ {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, migrations[i]); err != nil {
			tx.Rollback()
			return fmt.Errorf("store: migration %d: %w", i+1, err)
		}
		// PRAGMA takes no parameters; i is an int we control.
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", i+1)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// Close closes both pools.
func (s *Store) Close() error { return errors.Join(s.r.Close(), s.w.Close()) }

// StationState is a station as stored, with its newest reading.
type StationState struct {
	Station   domain.Station
	LoadRefs  normalize.LoadRefs
	Latest    *domain.Reading // nil if no reading is stored
	FirstSeen time.Time
	LastSeen  time.Time
}

const upsertStation = `
INSERT INTO stations (id, code, name, lat, lon, province_code, province, amphoe, tumbon, basin, river, agency,
	refs, bank_full, ground, load_bank_full, load_ground, first_seen, last_seen)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (id) DO UPDATE SET
	code = excluded.code, name = excluded.name, lat = excluded.lat, lon = excluded.lon,
	province_code = excluded.province_code, province = excluded.province, amphoe = excluded.amphoe,
	tumbon = excluded.tumbon, basin = excluded.basin, river = excluded.river, agency = excluded.agency,
	refs = excluded.refs, bank_full = excluded.bank_full, ground = excluded.ground,
	load_bank_full = excluded.load_bank_full, load_ground = excluded.load_ground,
	last_seen = excluded.last_seen`

// upsertReading inserts a reading, or replaces one for the same minute when
// the new value comes from an equal or better source and differs. Rows
// affected is therefore 1 only for a real change.
const upsertReading = `
INSERT INTO readings (station_id, measured_at, wl_mm, source) VALUES (?, ?, ?, ?)
ON CONFLICT (station_id, measured_at) DO UPDATE SET wl_mm = excluded.wl_mm, source = excluded.source
WHERE excluded.source >= readings.source
	AND (excluded.wl_mm != readings.wl_mm OR excluded.source != readings.source)`

// SaveObservations upserts the stations and readings of one snapshot, seen
// at `seen`, in one transaction. It returns how many readings were inserted
// or changed, so a repeated snapshot reports 0. A reading older than
// Retention is not stored: the load keeps publishing the last reading of a
// dead station, and storing it only for Prune to delete it would count as a
// change every hour.
func (s *Store) SaveObservations(ctx context.Context, obs []normalize.Observation, seen time.Time) (int, error) {
	tx, err := s.w.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	st, err := tx.PrepareContext(ctx, upsertStation)
	if err != nil {
		return 0, err
	}
	rd, err := tx.PrepareContext(ctx, upsertReading)
	if err != nil {
		return 0, err
	}
	changed := 0
	for _, o := range obs {
		x := o.Station
		if _, err := st.ExecContext(ctx, x.ID, x.Code, x.Name, x.Lat, x.Lon, x.ProvinceCode, x.Province, x.Amphoe,
			x.Tumbon, x.Basin, x.River, x.Agency, int(x.Refs), x.BankFull, x.Ground,
			o.LoadRefs.BankFull, o.LoadRefs.Ground, seen.Unix(), seen.Unix()); err != nil {
			return 0, fmt.Errorf("store: station %d: %w", x.ID, err)
		}
		if o.Reading.At.Before(seen.Add(-Retention)) {
			continue
		}
		n, err := affected(rd.ExecContext(ctx, x.ID, o.Reading.At.Unix(), domain.MM(o.Reading.WL), sourceLoad))
		if err != nil {
			return 0, fmt.Errorf("store: reading for %d: %w", x.ID, err)
		}
		changed += n
	}
	return changed, tx.Commit()
}

// SaveHistory stores a station's backfilled history and full-precision
// references, and marks the station as having history as of `at`. The
// station must already exist.
func (s *Store) SaveHistory(ctx context.Context, id int64, refs normalize.GraphRefs, readings []domain.Reading, at time.Time) (int, error) {
	tx, err := s.w.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	n, err := affected(tx.ExecContext(ctx, `UPDATE stations SET graph_bank_full = ?, graph_ground = ?, history_at = ? WHERE id = ?`,
		refs.BankFull, refs.Ground, at.Unix(), id))
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, fmt.Errorf("store: history for unknown station %d", id)
	}
	rd, err := tx.PrepareContext(ctx, upsertReading)
	if err != nil {
		return 0, err
	}
	changed := 0
	for _, r := range readings {
		n, err := affected(rd.ExecContext(ctx, id, r.At.Unix(), domain.MM(r.WL), sourceGraph))
		if err != nil {
			return 0, fmt.Errorf("store: history for %d: %w", id, err)
		}
		changed += n
	}
	return changed, tx.Commit()
}

// affected returns the rows a statement changed.
func affected(res sql.Result, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	return int(n), err
}

// NeedsHistory lists stations whose 72 h history has not been fetched, by id.
func (s *Store) NeedsHistory(ctx context.Context) ([]int64, error) {
	rows, err := s.r.QueryContext(ctx, `SELECT id FROM stations WHERE history_at IS NULL ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// ForgetHistory asks for a station's history and references to be fetched
// again on the next run, if they were last fetched before `before`. The
// cut-off stops a lasting disagreement between ThaiWater's endpoints from
// costing a request every hour. The current references stay in use until
// the refetch. It reports whether the station was marked.
func (s *Store) ForgetHistory(ctx context.Context, id int64, before time.Time) (bool, error) {
	n, err := affected(s.w.ExecContext(ctx, `UPDATE stations SET history_at = NULL WHERE id = ? AND history_at < ?`,
		id, before.Unix()))
	return n == 1, err
}

// GraphRefs returns the full-precision references of every station that has
// them, for normalize.Load.
func (s *Store) GraphRefs(ctx context.Context) (map[int64]normalize.GraphRefs, error) {
	rows, err := s.r.QueryContext(ctx, `SELECT id, graph_bank_full, graph_ground FROM stations
		WHERE graph_bank_full IS NOT NULL OR graph_ground IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]normalize.GraphRefs{}
	for rows.Next() {
		var id int64
		var bank, ground sql.NullFloat64
		if err := rows.Scan(&id, &bank, &ground); err != nil {
			return nil, err
		}
		out[id] = normalize.GraphRefs{BankFull: ptr(bank), Ground: ptr(ground)}
	}
	return out, rows.Err()
}

const selectStation = `
SELECT s.id, s.code, s.name, s.lat, s.lon, s.province_code, s.province, s.amphoe, s.tumbon, s.basin, s.river,
	s.agency, s.refs, s.bank_full, s.ground, s.load_bank_full, s.load_ground, s.first_seen, s.last_seen,
	r.measured_at, r.wl_mm
FROM stations s
LEFT JOIN readings r ON r.station_id = s.id
	AND r.measured_at = (SELECT max(measured_at) FROM readings WHERE station_id = s.id)`

// Stations returns every station with its newest reading, by id.
func (s *Store) Stations(ctx context.Context) ([]StationState, error) {
	rows, err := s.r.QueryContext(ctx, selectStation+` ORDER BY s.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []StationState
	for rows.Next() {
		st, err := scanStation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

// Station returns one station; ok is false if it is unknown.
func (s *Store) Station(ctx context.Context, id int64) (st StationState, ok bool, err error) {
	rows, err := s.r.QueryContext(ctx, selectStation+` WHERE s.id = ?`, id)
	if err != nil {
		return st, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return st, false, rows.Err()
	}
	st, err = scanStation(rows)
	return st, err == nil, err
}

func scanStation(rows *sql.Rows) (StationState, error) {
	var st StationState
	x := &st.Station
	var refs int
	var loadBank, loadGround sql.NullFloat64
	var first, last int64
	var at, wl sql.NullInt64
	if err := rows.Scan(&x.ID, &x.Code, &x.Name, &x.Lat, &x.Lon, &x.ProvinceCode, &x.Province, &x.Amphoe,
		&x.Tumbon, &x.Basin, &x.River, &x.Agency, &refs, &x.BankFull, &x.Ground, &loadBank, &loadGround,
		&first, &last, &at, &wl); err != nil {
		return st, err
	}
	x.Refs = domain.Refs(refs)
	st.LoadRefs = normalize.LoadRefs{BankFull: ptr(loadBank), Ground: ptr(loadGround)}
	st.FirstSeen, st.LastSeen = unix(first), unix(last)
	if at.Valid {
		st.Latest = &domain.Reading{StationID: x.ID, At: unix(at.Int64), WL: metres(wl.Int64)}
	}
	return st, nil
}

// History returns one station's readings in (since, until], oldest first.
func (s *Store) History(ctx context.Context, id int64, since, until time.Time) ([]domain.Reading, error) {
	rows, err := s.r.QueryContext(ctx, `SELECT station_id, measured_at, wl_mm FROM readings
		WHERE station_id = ? AND measured_at > ? AND measured_at <= ? ORDER BY measured_at`,
		id, since.Unix(), until.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Reading
	err = scanReadings(rows, func(r domain.Reading) { out = append(out, r) })
	return out, err
}

// ReadingsSince returns every station's readings after `since`, oldest first,
// keyed by station: the input for the timeline and the trends.
func (s *Store) ReadingsSince(ctx context.Context, since time.Time) (map[int64][]domain.Reading, error) {
	rows, err := s.r.QueryContext(ctx, `SELECT station_id, measured_at, wl_mm FROM readings
		WHERE measured_at > ? ORDER BY station_id, measured_at`, since.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64][]domain.Reading{}
	err = scanReadings(rows, func(r domain.Reading) { out[r.StationID] = append(out[r.StationID], r) })
	return out, err
}

func scanReadings(rows *sql.Rows, add func(domain.Reading)) error {
	for rows.Next() {
		var id, at, wl int64
		if err := rows.Scan(&id, &at, &wl); err != nil {
			return err
		}
		add(domain.Reading{StationID: id, At: unix(at), WL: metres(wl)})
	}
	return rows.Err()
}

// Prune deletes readings older than Retention at `now` and returns how many.
func (s *Store) Prune(ctx context.Context, now time.Time) (int64, error) {
	res, err := s.w.ExecContext(ctx, `DELETE FROM readings WHERE measured_at < ?`, now.Add(-Retention).Unix())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Run is the record of one ingest run.
type Run struct {
	Started, Finished time.Time
	Rows              int // rows in the snapshot
	Observations      int // usable rows
	Dropped           int
	Warnings          int
	Upserted          int // readings inserted or changed
	HistoryOK         int // stations backfilled
	HistoryFailed     int
	RuleWebOK         bool
	OracleChecked     int // rows with a published level
	OracleMismatches  int // of those, rows where our level differs
	Err               string
}

const runColumns = `started_at, finished_at, rows, observations, dropped, warnings, upserted, history_ok,
	history_failed, rule_web_ok, oracle_checked, oracle_mismatches, error`

// RecordRun stores a run.
func (s *Store) RecordRun(ctx context.Context, r Run) error {
	_, err := s.w.ExecContext(ctx, `INSERT INTO ingest_runs (`+runColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.Started.Unix(), r.Finished.Unix(), r.Rows, r.Observations, r.Dropped, r.Warnings, r.Upserted,
		r.HistoryOK, r.HistoryFailed, r.RuleWebOK, r.OracleChecked, r.OracleMismatches, r.Err)
	return err
}

// RunCount is how many runs are recorded.
func (s *Store) RunCount(ctx context.Context) (int, error) {
	var n int
	err := s.r.QueryRowContext(ctx, `SELECT count(*) FROM ingest_runs`).Scan(&n)
	return n, err
}

// LastRun is the most recent run, successful or not.
func (s *Store) LastRun(ctx context.Context) (Run, bool, error) {
	return s.run(ctx, `ORDER BY id DESC LIMIT 1`)
}

// LastGoodRun is the most recent run that finished without an error.
func (s *Store) LastGoodRun(ctx context.Context) (Run, bool, error) {
	return s.run(ctx, `WHERE error = '' ORDER BY id DESC LIMIT 1`)
}

func (s *Store) run(ctx context.Context, tail string) (Run, bool, error) {
	var r Run
	var started, finished int64
	err := s.r.QueryRowContext(ctx, `SELECT `+runColumns+` FROM ingest_runs `+tail).Scan(
		&started, &finished, &r.Rows, &r.Observations, &r.Dropped, &r.Warnings, &r.Upserted,
		&r.HistoryOK, &r.HistoryFailed, &r.RuleWebOK, &r.OracleChecked, &r.OracleMismatches, &r.Err)
	if errors.Is(err, sql.ErrNoRows) {
		return r, false, nil
	}
	if err != nil {
		return r, false, err
	}
	r.Started, r.Finished = unix(started), unix(finished)
	return r, true, nil
}

// Meta reads a value from the key-value table.
func (s *Store) Meta(ctx context.Context, key string) (string, bool, error) {
	var v string
	err := s.r.QueryRowContext(ctx, `SELECT value FROM meta WHERE key = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return v, err == nil, err
}

// SetMeta writes a value to the key-value table.
func (s *Store) SetMeta(ctx context.Context, key, value string) error {
	_, err := s.w.ExecContext(ctx, `INSERT INTO meta (key, value) VALUES (?, ?)
		ON CONFLICT (key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}

func unix(s int64) time.Time { return time.Unix(s, 0).UTC() }

func metres(mm int64) float64 { return float64(mm) / 1000 }

func ptr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	return &v.Float64
}
