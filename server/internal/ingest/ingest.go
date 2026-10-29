package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/normalize"
	"github.com/serlorsx-xln/mahanathi/server/internal/store"
	"github.com/serlorsx-xln/mahanathi/server/internal/thaiwater"
)

// How much history a backfill asks for, and how long a reference
// disagreement is allowed to cost a request.
const (
	historyWindow = 72 * time.Hour
	refetchAfter  = 24 * time.Hour
)

// The schedule defaults from the engineering brief: hourly, at :07, after
// the hourly stations have published (the 19:00 readings were out by 19:10
// but the 17:00 ones only by 17:54, so seven minutes past is the compromise).
const (
	DefaultEvery  = time.Hour
	DefaultOffset = 7 * time.Minute
	minEvery      = time.Minute
)

// ruleWebKey is where the last seen rule_web is kept between runs.
const ruleWebKey = "rule_web"

// Source is the part of the ThaiWater client the ingester needs. The client
// itself applies retries and the concurrency limit to its own calls.
type Source interface {
	Load(ctx context.Context) (*thaiwater.LoadResponse, error)
	Graph(ctx context.Context, stationID int64, from, to time.Time) (*thaiwater.GraphResponse, error)
}

// Options configures an Ingester. Now, Wait and Logger are for tests; every
// field is optional except Source and Store.
type Options struct {
	Source      Source
	Store       *store.Store
	Logger      *slog.Logger
	Now         func() time.Time
	OnChange    func()                                           // called once per run that stored changes
	Concurrency int                                              // history requests in flight; default and cap 4
	Every       time.Duration                                    // default hourly
	Offset      time.Duration                                    // default 7 minutes past
	Wait        func(ctx context.Context, d time.Duration) error // default: sleep
}

// Ingester runs the pipeline on a schedule.
type Ingester struct {
	src         Source
	st          *store.Store
	log         *slog.Logger
	now         func() time.Time
	onChange    func()
	concurrency int
	every       time.Duration
	offset      time.Duration
	wait        func(ctx context.Context, d time.Duration) error
}

// New validates the options and returns an Ingester.
func New(o Options) (*Ingester, error) {
	if o.Source == nil {
		return nil, fmt.Errorf("ingest: no Source")
	}
	if o.Store == nil {
		return nil, fmt.Errorf("ingest: no Store")
	}
	every := o.Every
	if every == 0 {
		every = DefaultEvery
	}
	if every < minEvery {
		return nil, fmt.Errorf("ingest: Every %v is shorter than %v", every, minEvery)
	}
	if o.Offset < 0 || o.Offset >= every {
		return nil, fmt.Errorf("ingest: Offset %v is outside 0..<%v", o.Offset, every)
	}
	if o.Concurrency < 0 {
		return nil, fmt.Errorf("ingest: Concurrency %d is negative", o.Concurrency)
	}
	concurrency := thaiwater.MaxGraphConcurrency
	if o.Concurrency > 0 && o.Concurrency < concurrency {
		concurrency = o.Concurrency
	}
	in := &Ingester{
		src:         o.Source,
		st:          o.Store,
		log:         o.Logger,
		now:         o.Now,
		onChange:    o.OnChange,
		concurrency: concurrency,
		every:       every,
		offset:      o.Offset,
		wait:        o.Wait,
	}
	if in.log == nil {
		in.log = slog.Default()
	}
	if in.now == nil {
		in.now = time.Now
	}
	if in.onChange == nil {
		in.onChange = func() {}
	}
	if in.wait == nil {
		in.wait = func(ctx context.Context, d time.Duration) error {
			t := time.NewTimer(d)
			defer t.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-t.C:
				return nil
			}
		}
	}
	return in, nil
}

// RunOnce fetches the snapshot, backfills what is missing and records the
// run. A failed snapshot is recorded and returned as an error; a failed
// history fetch only lowers that run's counts.
func (in *Ingester) RunOnce(ctx context.Context) (store.Run, error) {
	now := in.now()
	var run store.Run
	run.Started = now

	resp, err := in.src.Load(ctx)
	if err != nil {
		run.Finished = in.now()
		run.Err = err.Error()
		in.record(ctx, &run)
		in.log.Error("ingest run failed", "error", err)
		return run, err
	}

	// The thresholds are the product's, so a silent change on ThaiWater's
	// side must be loud here, twice: against what we hard-code and against
	// what we saw last time.
	rules := resp.Scale.Data.RuleWeb
	ruleWeb := formatRuleWeb(rules)
	if err := normalize.CheckRuleWeb(rules); err != nil {
		run.RuleWebOK = false
		in.log.Error("rule_web differs from the thresholds the product uses", "error", err)
	} else {
		run.RuleWebOK = true
	}
	if old, ok, err := in.st.Meta(ctx, ruleWebKey); err != nil {
		return in.fail(ctx, &run, err)
	} else if ok && old != ruleWeb {
		in.log.Error("rule_web changed since the last run", "old", old, "new", ruleWeb)
	}

	refs, err := in.st.GraphRefs(ctx)
	if err != nil {
		return in.fail(ctx, &run, err)
	}
	res := normalize.Load(resp, refs)
	run.Rows, run.Observations, run.Dropped = res.Rows, len(res.Observations), res.Dropped

	upserted, err := in.st.SaveObservations(ctx, res.Observations, now)
	if err != nil {
		return in.fail(ctx, &run, err)
	}

	// A reference that disagrees beyond rounding is re-fetched at most once
	// a day; the current references stay in use until then.
	for _, w := range res.Warnings {
		if w.Kind == normalize.WarnRefMismatch {
			if marked, err := in.st.ForgetHistory(ctx, w.StationID, now.Add(-refetchAfter)); err != nil {
				return in.fail(ctx, &run, err)
			} else if marked {
				in.log.Warn("reference disagreement; refetching history", "station", w.StationID, "detail", w.Detail)
			}
		}
	}

	need, err := in.st.NeedsHistory(ctx)
	if err != nil {
		return in.fail(ctx, &run, err)
	}
	histUp, graphWarns, ok, failed, err := in.backfill(ctx, need, now)
	if err != nil {
		return in.fail(ctx, &run, err)
	}
	upserted += histUp
	run.HistoryOK, run.HistoryFailed = ok, failed

	// History just stored full-precision references, so the snapshot is
	// normalised once more before it is scored: the stored stations and the
	// oracle must see the same numbers the API will serve.
	final := res
	if ok > 0 {
		if refs, err = in.st.GraphRefs(ctx); err != nil {
			return in.fail(ctx, &run, err)
		}
		final = normalize.Load(resp, refs)
		if n, err := in.st.SaveObservations(ctx, final.Observations, now); err != nil {
			return in.fail(ctx, &run, err)
		} else {
			upserted += n
		}
	}
	run.Warnings = len(final.Warnings) + graphWarns
	in.oracle(&run, final)

	if pruned, err := in.st.Prune(ctx, now); err != nil {
		return in.fail(ctx, &run, err)
	} else if pruned > 0 {
		in.log.Debug("pruned old readings", "count", pruned)
	}

	run.Upserted = upserted
	run.Finished = in.now()
	in.record(ctx, &run)
	if err := in.st.SetMeta(ctx, ruleWebKey, ruleWeb); err != nil {
		in.log.Error("recording rule_web failed", "error", err)
	}
	in.log.Info("ingest run",
		"rows", run.Rows, "observations", run.Observations, "dropped", run.Dropped,
		"upserted", run.Upserted, "history_ok", run.HistoryOK, "history_failed", run.HistoryFailed,
		"warnings", run.Warnings, "oracle_checked", run.OracleChecked, "oracle_mismatches", run.OracleMismatches,
		"took", run.Finished.Sub(run.Started).Round(time.Millisecond))
	if upserted > 0 {
		in.onChange()
	}
	return run, nil
}

// backfill fetches the 72 h history of every station in need, at most
// in.concurrency requests in flight. Each station gets its own slot in the
// results, so the workers never share state.
func (in *Ingester) backfill(ctx context.Context, need []int64, now time.Time) (upserted, warns, ok, failed int, err error) {
	type outcome struct {
		changed int
		warns   int
		err     error
	}
	outcomes := make([]outcome, len(need))
	sem := make(chan struct{}, in.concurrency)
	var wg sync.WaitGroup
	for i, id := range need {
		wg.Add(1)
		go func(i int, id int64) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			resp, err := in.src.Graph(ctx, id, now.Add(-historyWindow), now)
			if err != nil {
				outcomes[i] = outcome{err: err}
				return
			}
			refs, readings, warns := normalize.Graph(id, resp)
			changed, err := in.st.SaveHistory(ctx, id, refs, readings, now)
			outcomes[i] = outcome{changed: changed, warns: len(warns), err: err}
		}(i, id)
	}
	wg.Wait()
	for i, o := range outcomes {
		if o.err != nil {
			failed++
			in.log.Warn("history fetch failed", "station", need[i], "error", o.err)
			continue
		}
		ok++
		upserted += o.changed
		warns += o.warns
	}
	return upserted, warns, ok, failed, nil
}

// oracle compares our levels with ThaiWater's published ones, on the rows
// where both exist. Differences are a warning, never a correction: our
// numbers are the product's, and the comparison is monitoring.
func (in *Ingester) oracle(run *store.Run, res normalize.Result) {
	var differing []string
	for _, o := range res.Observations {
		if o.Published.Level == nil || o.Station.Refs != domain.RefsComplete {
			continue
		}
		run.OracleChecked++
		if got := o.Station.Level(o.Reading.WL); int(got) != *o.Published.Level {
			run.OracleMismatches++
			differing = append(differing, fmt.Sprintf("%s: ours %d, published %d", o.Station.Code, got, *o.Published.Level))
		}
	}
	if len(differing) > 0 {
		in.log.Warn("levels differ from ThaiWater's published levels", "count", len(differing), "stations", differing)
	}
}

// record writes the run. Failing to record is logged; the run it describes
// still happened.
func (in *Ingester) record(ctx context.Context, run *store.Run) {
	if err := in.st.RecordRun(ctx, *run); err != nil {
		in.log.Error("recording the run failed", "error", err)
	}
}

func (in *Ingester) fail(ctx context.Context, run *store.Run, err error) (store.Run, error) {
	run.Finished = in.now()
	run.Err = err.Error()
	in.record(ctx, run)
	in.log.Error("ingest run failed", "error", err)
	return *run, err
}

// Run runs once at start and then on the schedule, forever, until the
// context is cancelled. A failed run is logged and recorded but never stops
// the loop. The returned error is the context's.
func (in *Ingester) Run(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := in.RunOnce(ctx); err != nil {
		// already logged and recorded
		_ = err
	}
	for {
		d := NextRun(in.now(), in.every, in.offset).Sub(in.now())
		if err := in.wait(ctx, d); err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, err := in.RunOnce(ctx); err != nil {
			_ = err
		}
	}
}

// NextRun is the first instant strictly after now at offset past each every.
// Alignment is to the Unix epoch, which for ICT (UTC+7, no daylight saving)
// is wall-clock alignment for any every that divides an hour.
func NextRun(now time.Time, every, offset time.Duration) time.Time {
	next := now.Truncate(every).Add(offset)
	if !next.After(now) {
		next = next.Add(every)
	}
	return next
}

// formatRuleWeb renders rule_web as a short, stable, comparable string.
func formatRuleWeb(rules []thaiwater.Rule) string {
	parts := make([]string, len(rules))
	for i, r := range rules {
		parts[i] = fmt.Sprintf("%s %s=%d", r.Operator, r.Term, r.Level)
	}
	return strings.Join(parts, "; ")
}
