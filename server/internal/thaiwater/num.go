package thaiwater

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

// Num is a number as ThaiWater sends it. The load endpoint sends water levels
// and percents as JSON strings ("3.17"), the station block sends references
// as JSON numbers, and any of them can be null or missing entirely. Num keeps
// all of that visible instead of collapsing it to 0:
//
//   - Present is false only when the key was absent.
//   - Valid is true only for a finite number, from either a string or a
//     number. Null, "" and unparsable text are not valid.
//   - Raw keeps the text of an unparsable value so normalize can report it.
type Num struct {
	Value   float64
	Valid   bool
	Present bool
	Raw     string
}

// Junk reports a value that was present and non-empty but not a number, such
// as "-" or "n/a". Null and "" are plain absence, not junk.
func (n Num) Junk() bool { return n.Present && !n.Valid && n.Raw != "" }

// Ptr returns the value, or nil when it is not valid.
func (n Num) Ptr() *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Value
	return &v
}

// UnmarshalJSON accepts a number, a numeric string, or null. Anything else is
// recorded as junk rather than failing the whole response, because one bad
// station must not cost the other 800.
func (n *Num) UnmarshalJSON(b []byte) error {
	*n = Num{Present: true}
	b = bytes.TrimSpace(b)
	switch {
	case bytes.Equal(b, []byte("null")):
		return nil
	case len(b) > 0 && b[0] == '"':
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		n.parse(s)
		return nil
	default:
		n.parse(string(b))
		return nil
	}
}

func (n *Num) parse(s string) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		n.Raw = s
		return
	}
	n.Value, n.Valid = f, true
}
