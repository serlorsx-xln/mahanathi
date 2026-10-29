package api

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/serlorsx-xln/mahanathi/server/internal/domain"
	"github.com/serlorsx-xln/mahanathi/server/internal/normalize"
	"github.com/serlorsx-xln/mahanathi/server/internal/store"
)

// The timeline is a binary body, not JSON: 800 stations × 72 hourly values
// as JSON numbers are 96 KB compressed however smooth the water is, and the
// brief's budget is 80 KB. The values are delta-coded varints over a
// presence bitmap, which keeps the full 0.01 % precision of /now and still
// leaves headroom. The format is documented in docs/api.md and versioned in
// its header.
const (
	timelineMagic   = "MNTL"
	timelineVersion = 1
	// timelineContentType is the body's type: an opaque, versioned blob.
	timelineContentType = "application/octet-stream"
	// maxTimelineValue is the clamp: percent of bank-full in hundredths.
	// Anything beyond it (the plausibility bound is 400 %) is pinned.
	maxTimelineValue = 65535
)

// timelineData is the timeline before encoding.
type timelineData struct {
	From     time.Time // the first bucket end
	Until    time.Time // the last bucket end
	Hours    int
	Stations []int64     // ascending, with at least one value
	Values   [][]*uint16 // parallel to Stations; nil means no reading
}

// buildTimeline assembles the timeline over the hours ending at the hour of
// now. Only stations with usable references (ADR 0003) and at least one
// value are listed.
func buildTimeline(stations []store.StationState, readings map[int64][]domain.Reading, now time.Time, hours int) timelineData {
	until := domain.HourEnd(now)
	t := timelineData{
		From:  until.Add(-time.Duration(hours-1) * time.Hour),
		Until: until,
		Hours: hours,
	}
	for _, st := range stations {
		if st.Station.Refs != domain.RefsComplete {
			continue
		}
		pts := normalize.Hourly(readings[st.Station.ID], until, hours)
		vals := make([]*uint16, hours)
		any := false
		for i, p := range pts {
			if !p.OK {
				continue
			}
			h, ok := st.Station.PctHundredths(p.WL)
			if !ok {
				continue
			}
			if h < 0 {
				h = 0
			}
			if h > maxTimelineValue {
				h = maxTimelineValue
			}
			v := uint16(h)
			vals[i] = &v
			any = true
		}
		if !any {
			continue
		}
		t.Stations = append(t.Stations, st.Station.ID)
		t.Values = append(t.Values, vals)
	}
	return t
}

// encodeTimeline writes the MNTL format:
//
//	 0  magic "MNTL"                    4 bytes
//	 4  version = 1                     1 byte
//	 5  hours                           1 byte
//	 6  reserved = 0                    2 bytes
//	 8  fromUnix (first bucket end)     8 bytes, LE
//	16  stationCount                    4 bytes, LE
//	20  station ids, ascending          8 bytes each, LE
//	then per station, in the same order:
//	   presence bitmap                  ceil(hours/8) bytes; bit 7−i%8 of
//	                                    byte i/8 is set when bucket i has
//	                                    a value
//	   one varint per present bucket    the first of each run (the start,
//	                                    or the bucket after a gap) is the
//	                                    absolute value; the rest are
//	                                    zigzag varints of the difference
//	                                    from the previous present bucket
//
// A value is the percent of bank-full in hundredths, clamped to
// [0, 65535]: 10457 is 104.57 %.
func encodeTimeline(t timelineData) []byte {
	buf := &bytes.Buffer{}
	buf.WriteString(timelineMagic)
	buf.WriteByte(timelineVersion)
	buf.WriteByte(byte(t.Hours))
	buf.Write([]byte{0, 0})
	var num [8]byte
	binary.LittleEndian.PutUint64(num[:], uint64(t.From.Unix()))
	buf.Write(num[:])
	binary.LittleEndian.PutUint32(num[:4], uint32(len(t.Stations)))
	buf.Write(num[:4])
	for _, id := range t.Stations {
		binary.LittleEndian.PutUint64(num[:], uint64(id))
		buf.Write(num[:])
	}

	bitmapLen := (t.Hours + 7) / 8
	for _, vals := range t.Values {
		bitmap := make([]byte, bitmapLen)
		for i, v := range vals {
			if v != nil {
				bitmap[i/8] |= 1 << (7 - i%8)
			}
		}
		buf.Write(bitmap)

		prev := int64(0)
		run := false // whether the previous present bucket is adjacent
		for _, v := range vals {
			if v == nil {
				run = false
				continue
			}
			x := int64(*v)
			if run {
				putVarint(buf, zigzag(x-prev))
			} else {
				putVarint(buf, uint64(x))
			}
			prev, run = x, true
		}
	}
	return buf.Bytes()
}

// putVarint writes v as an unsigned LEB128 varint.
func putVarint(buf *bytes.Buffer, v uint64) {
	for v >= 0x80 {
		buf.WriteByte(byte(v) | 0x80)
		v >>= 7
	}
	buf.WriteByte(byte(v))
}

// zigzag maps a signed difference to an unsigned varint, so small deltas of
// either sign stay one byte: −1 → 1, 1 → 2, −63 → 125, 63 → 126.
func zigzag(d int64) uint64 {
	return uint64(d<<1) ^ uint64(d>>63)
}
