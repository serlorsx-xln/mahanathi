-- 0001: stations, readings, ingest runs and a small key-value table.
-- Times are Unix seconds (UTC). Water levels are whole millimetres above
-- mean sea level, so stored values are exact.

CREATE TABLE stations (
	id              INTEGER PRIMARY KEY, -- ThaiWater tele_station id
	code            TEXT    NOT NULL,    -- tele_station_oldcode
	name            TEXT    NOT NULL,
	lat             REAL    NOT NULL,
	lon             REAL    NOT NULL,
	province_code   TEXT    NOT NULL,
	province        TEXT    NOT NULL,
	amphoe          TEXT    NOT NULL,
	tumbon          TEXT    NOT NULL,
	basin           TEXT    NOT NULL,
	river           TEXT    NOT NULL,
	agency          TEXT    NOT NULL,
	-- The references in use (ADR 0003): domain.Refs, and metres MSL. The two
	-- levels are stored as normalize gave them and mean something only as
	-- far as refs allows.
	refs            INTEGER NOT NULL,
	bank_full       REAL    NOT NULL,
	ground          REAL    NOT NULL,
	-- As waterlevel_load published them (2 decimals), for the data-quality
	-- report and for re-resolving when graph references arrive.
	load_bank_full  REAL,
	load_ground     REAL,
	-- Full precision from waterlevel_graph, and when they were fetched.
	graph_bank_full REAL,
	graph_ground    REAL,
	history_at      INTEGER, -- last successful 72 h backfill; NULL means needed
	first_seen      INTEGER NOT NULL,
	last_seen       INTEGER NOT NULL
) STRICT;

CREATE TABLE readings (
	station_id  INTEGER NOT NULL REFERENCES stations (id),
	measured_at INTEGER NOT NULL,
	wl_mm       INTEGER NOT NULL,
	source      INTEGER NOT NULL, -- 1 waterlevel_load (2 decimals), 2 waterlevel_graph (3)
	PRIMARY KEY (station_id, measured_at)
) STRICT, WITHOUT ROWID;

-- The timeline and the trend read every station's recent hours at once.
CREATE INDEX readings_measured_at ON readings (measured_at);

CREATE TABLE ingest_runs (
	id                INTEGER PRIMARY KEY,
	started_at        INTEGER NOT NULL,
	finished_at       INTEGER NOT NULL,
	rows              INTEGER NOT NULL, -- rows in waterlevel_load
	observations      INTEGER NOT NULL, -- usable rows
	dropped           INTEGER NOT NULL,
	warnings          INTEGER NOT NULL,
	upserted          INTEGER NOT NULL, -- readings inserted or changed
	history_ok        INTEGER NOT NULL, -- stations backfilled
	history_failed    INTEGER NOT NULL,
	rule_web_ok       INTEGER NOT NULL, -- 1 if rule_web matched the known thresholds
	oracle_checked    INTEGER NOT NULL, -- rows with a published level
	oracle_mismatches INTEGER NOT NULL, -- of those, levels we compute differently
	error             TEXT    NOT NULL  -- empty when the run succeeded
) STRICT;

CREATE TABLE meta (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
) STRICT;
