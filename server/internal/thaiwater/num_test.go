package thaiwater

import (
	"encoding/json"
	"testing"
)

func TestNumAcceptsEveryShapeTheAPIUses(t *testing.T) {
	type holder struct {
		N Num `json:"n"`
	}
	cases := []struct {
		name                 string
		json                 string
		value                float64
		valid, present, junk bool
	}{
		{"string as the load endpoint sends levels", `{"n":"3.17"}`, 3.17, true, true, false},
		{"string with spaces", `{"n":" 40.60 "}`, 40.6, true, true, false},
		{"negative string", `{"n":"-18.93"}`, -18.93, true, true, false},
		{"float as the station block sends references", `{"n":-2.065}`, -2.065, true, true, false},
		{"integer", `{"n":0}`, 0, true, true, false},
		{"float32 noise from the graph endpoint", `{"n":40.599998}`, 40.599998, true, true, false},
		{"null", `{"n":null}`, 0, false, true, false},
		{"absent key", `{}`, 0, false, false, false},
		{"empty string", `{"n":""}`, 0, false, true, false},
		{"dash placeholder", `{"n":"-"}`, 0, false, true, true},
		{"text", `{"n":"n/a"}`, 0, false, true, true},
		{"NaN string is not a number we accept", `{"n":"NaN"}`, 0, false, true, true},
		{"Inf string", `{"n":"+Inf"}`, 0, false, true, true},
		{"boolean", `{"n":true}`, 0, false, true, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var h holder
			if err := json.Unmarshal([]byte(c.json), &h); err != nil {
				t.Fatalf("Unmarshal(%s): %v", c.json, err)
			}
			if h.N.Valid != c.valid || h.N.Present != c.present || h.N.Junk() != c.junk {
				t.Fatalf("Num = %+v (junk %v), want valid=%v present=%v junk=%v", h.N, h.N.Junk(), c.valid, c.present, c.junk)
			}
			if c.valid && h.N.Value != c.value {
				t.Errorf("Value = %v, want %v", h.N.Value, c.value)
			}
		})
	}
}

func TestNumPtr(t *testing.T) {
	if (Num{}).Ptr() != nil {
		t.Error("invalid Num gave a non-nil pointer")
	}
	if p := (Num{Value: 1.5, Valid: true, Present: true}).Ptr(); p == nil || *p != 1.5 {
		t.Errorf("Ptr() = %v", p)
	}
}
