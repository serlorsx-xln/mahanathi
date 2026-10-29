package thaiwater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the public ThaiWater API host.
const DefaultBaseURL = "https://api-v3.thaiwater.net"

const (
	loadPath  = "/api/v1/thaiwater30/public/waterlevel_load"
	graphPath = "/api/v1/thaiwater30/public/waterlevel_graph"

	// MaxGraphConcurrency is the politeness limit agreed for history calls.
	MaxGraphConcurrency = 4
)

// ict is Indochina Time. Thailand has no daylight saving, so a fixed zone is
// exact and needs no tzdata in the container.
var ict = time.FixedZone("ICT", 7*3600)

// Options configures a Client. Zero values take the defaults noted.
type Options struct {
	BaseURL    string       // required, http or https
	UserAgent  string       // required; names the project and its repository
	HTTPClient *http.Client // default: a client with no overall timeout (attempts have their own)

	LoadTimeout  time.Duration // per attempt; default 60 s (the snapshot is about 1.4 MB)
	GraphTimeout time.Duration // per attempt; default 20 s
	MaxAttempts  int           // default 4
	BaseDelay    time.Duration // first backoff ceiling; default 1 s
	MaxDelay     time.Duration // backoff cap, also the longest Retry-After honoured; default 30 s

	GraphConcurrency int   // history calls in flight; default and maximum 4
	MaxBodyBytes     int64 // default 16 MiB

	// Test hooks. Sleep must return ctx.Err() if ctx ends first; Jitter
	// returns a value in [0, 1).
	Sleep  func(ctx context.Context, d time.Duration) error
	Jitter func() float64
}

// Client calls the ThaiWater API. It is safe for concurrent use.
type Client struct {
	opts  Options
	graph chan struct{} // semaphore for history calls
}

// HTTPError is a non-2xx response.
type HTTPError struct {
	Status int
	URL    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("thaiwater: %s: HTTP %d", e.URL, e.Status)
}

// retryable is an error worth another attempt, optionally after a delay the
// server asked for.
type retryable struct {
	err   error
	after time.Duration
}

func (r *retryable) Error() string { return r.err.Error() }
func (r *retryable) Unwrap() error { return r.err }

// New validates opts and fills in defaults.
func New(opts Options) (*Client, error) {
	u, err := url.Parse(opts.BaseURL)
	if opts.BaseURL == "" || err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("thaiwater: BaseURL %q is not an http(s) URL", opts.BaseURL)
	}
	opts.BaseURL = strings.TrimRight(opts.BaseURL, "/")
	if strings.TrimSpace(opts.UserAgent) == "" {
		return nil, errors.New("thaiwater: UserAgent is required (name the project and its repository)")
	}
	if opts.GraphConcurrency < 0 || opts.GraphConcurrency > MaxGraphConcurrency {
		return nil, fmt.Errorf("thaiwater: GraphConcurrency %d is outside 1..%d", opts.GraphConcurrency, MaxGraphConcurrency)
	}
	if opts.GraphConcurrency == 0 {
		opts.GraphConcurrency = MaxGraphConcurrency
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{}
	}
	if opts.LoadTimeout <= 0 {
		opts.LoadTimeout = 60 * time.Second
	}
	if opts.GraphTimeout <= 0 {
		opts.GraphTimeout = 20 * time.Second
	}
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 4
	}
	if opts.BaseDelay <= 0 {
		opts.BaseDelay = time.Second
	}
	if opts.MaxDelay <= 0 {
		opts.MaxDelay = 30 * time.Second
	}
	if opts.MaxBodyBytes <= 0 {
		opts.MaxBodyBytes = 16 << 20
	}
	if opts.Sleep == nil {
		opts.Sleep = sleep
	}
	if opts.Jitter == nil {
		opts.Jitter = rand.Float64
	}
	return &Client{opts: opts, graph: make(chan struct{}, opts.GraphConcurrency)}, nil
}

// Load fetches the national snapshot of latest readings.
func (c *Client) Load(ctx context.Context) (*LoadResponse, error) {
	var out LoadResponse
	err := c.get(ctx, c.opts.BaseURL+loadPath, c.opts.LoadTimeout, &out, func() error {
		if out.WaterlevelData.Result != "OK" {
			return fmt.Errorf("waterlevel_data.result = %q", out.WaterlevelData.Result)
		}
		if len(out.WaterlevelData.Data) == 0 {
			return errors.New("waterlevel_data.data is empty")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// Graph fetches a station's history between two instants. The endpoint takes
// inclusive ICT calendar dates, so the range is widened to whole Bangkok days.
// At most GraphConcurrency calls run at once; the rest wait their turn.
func (c *Client) Graph(ctx context.Context, stationID int64, from, to time.Time) (*GraphResponse, error) {
	select {
	case c.graph <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-c.graph }()

	q := url.Values{}
	q.Set("station_type", "tele_waterlevel")
	q.Set("station_id", strconv.FormatInt(stationID, 10))
	q.Set("start_date", from.In(ict).Format(time.DateOnly))
	q.Set("end_date", to.In(ict).Format(time.DateOnly))

	var out GraphResponse
	err := c.get(ctx, c.opts.BaseURL+graphPath+"?"+q.Encode(), c.opts.GraphTimeout, &out, func() error {
		if out.Result != "OK" {
			return fmt.Errorf("result = %q", out.Result)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// get performs one logical request with retries. check validates the decoded
// body; a failed check is retried like a 5xx, because a proxy or maintenance
// page can answer 200 with something other than the API's JSON.
func (c *Client) get(ctx context.Context, rawURL string, timeout time.Duration, out any, check func() error) error {
	var last error
	for attempt := 0; attempt < c.opts.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := c.backoff(attempt - 1)
			var r *retryable
			if errors.As(last, &r) && r.after > 0 {
				delay = r.after
			}
			if err := c.opts.Sleep(ctx, delay); err != nil {
				return err
			}
		}
		err := c.attempt(ctx, rawURL, timeout, out, check)
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var r *retryable
		if !errors.As(err, &r) {
			return err
		}
		last = err
	}
	return fmt.Errorf("thaiwater: gave up after %d attempts: %w", c.opts.MaxAttempts, last)
}

func (c *Client) attempt(ctx context.Context, rawURL string, timeout time.Duration, out any, check func() error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.opts.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.opts.HTTPClient.Do(req)
	if err != nil {
		return &retryable{err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		herr := &HTTPError{Status: resp.StatusCode, URL: rawURL}
		switch {
		case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusRequestTimeout || resp.StatusCode >= 500:
			return &retryable{err: herr, after: c.retryAfter(resp.Header.Get("Retry-After"))}
		default:
			return herr
		}
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, c.opts.MaxBodyBytes+1))
	if err != nil {
		return &retryable{err: fmt.Errorf("thaiwater: reading %s: %w", rawURL, err)}
	}
	if int64(len(body)) > c.opts.MaxBodyBytes {
		return fmt.Errorf("thaiwater: %s: body larger than %d bytes", rawURL, c.opts.MaxBodyBytes)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return &retryable{err: fmt.Errorf("thaiwater: decoding %s: %w", rawURL, err)}
	}
	if err := check(); err != nil {
		return &retryable{err: fmt.Errorf("thaiwater: %s: %w", rawURL, err)}
	}
	return nil
}

// backoff is "full jitter": a uniform delay in [0, min(MaxDelay,
// BaseDelay·2^attempt)). It spreads retries from many clients and never waits
// longer than the cap.
func (c *Client) backoff(attempt int) time.Duration {
	ceiling := float64(c.opts.BaseDelay) * math.Pow(2, float64(attempt))
	if ceiling > float64(c.opts.MaxDelay) {
		ceiling = float64(c.opts.MaxDelay)
	}
	return time.Duration(c.opts.Jitter() * ceiling)
}

// retryAfter reads a delay in seconds, honoured only up to MaxDelay. The
// HTTP-date form is not used by this API and is ignored.
func (c *Client) retryAfter(v string) time.Duration {
	secs, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || secs <= 0 {
		return 0
	}
	d := time.Duration(secs) * time.Second
	if d > c.opts.MaxDelay {
		return c.opts.MaxDelay
	}
	return d
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
