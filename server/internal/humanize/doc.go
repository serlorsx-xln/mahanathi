// Package humanize turns a normalised reading and its recent history into one
// plain Thai sentence plus the structured facts behind it.
//
// It is pure: no I/O, no clock reads, no randomness. It never predicts and
// never reassures; a banned-phrase test runs every output against words such
// as ปลอดภัย, จะท่วม and ไม่ท่วม, across all recorded fixtures and generated
// variations. It implements the humanize spec in docs/brief/02-ENGINEERING.md
// §3, as amended by the lead after Phase 1. The table below is the contract;
// each row names the test cases that pin it down (humanize_test.go).
//
// # Station sentence
//
// Distances are wl − bank-full, computed in whole millimetres and shown in
// metres with two decimals, rounded half away from zero (thaifmt).
//
//	Case                      Sentence                                                     Level      Tests
//	------------------------  -----------------------------------------------------------  ---------  -----------------------
//	below bank-full           ต่ำกว่าตลิ่ง 1.40 เมตร                                       computed   below*, half-cm
//	above bank-full           สูงกว่าตลิ่ง 0.35 เมตร (น้ำล้นตลิ่ง)                         5          above*
//	exactly at bank-full      ระดับน้ำเท่ากับตลิ่ง                                          4          at-bank
//	within 5 mm of the bank   สูงกว่าตลิ่งไม่ถึง 1 ซม. (น้ำล้นตลิ่ง) / ต่ำกว่าตลิ่งไม่ถึง 1 ซม.  computed   tiny-*
//	ground absent (bank-only) the distance, without "(น้ำล้นตลิ่ง)"                         none       bank-only-*
//	no usable bank-full       สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง ระดับน้ำ 331.73 ม.รทก.          none       no-bank-*
//	stale (older than 3 h)    leads with ข้อมูลล่าสุดเมื่อ 09:00 น. (ยังไม่อัปเดต), then the   computed   stale-*
//	                          last known state prefixed ตอนนั้น; the date is added when it
//	                          was not today in Bangkok; no trend
//	suspect                   ค่าที่รายงานล่าสุดผิดปกติ (reasons) จึงไม่นำมาแสดงเป็น          none       suspect-*
//	                          ระดับน้ำปัจจุบัน; no distance, no trend; after the stale lead
//	                          when both apply
//	no reading at all         ยังไม่มีข้อมูลระดับน้ำจากสถานีนี้                               none       no-reading
//
// Notes on the cases above:
//   - "Use the situation label if present" from the spec never applies: per
//     ADR 0003 the published situation_level is an oracle only, and ThaiWater
//     publishes no level for exactly the rows that lack references.
//   - Bank-only follows the lead's answer to Phase 1 question 2: state the
//     distance (wl − min_bank), but no percent and no level. The "(น้ำล้นตลิ่ง)"
//     label is a level label, so it is left out too.
//   - A suspect reading gets no level, so a client that forgets the suspect
//     flag shows it as unknown rather than as an overflow.
//
// # Trend
//
// The trend compares the latest reading with the earliest hourly bucket in the
// six hours before it, and is appended after the distance (or the water level)
// for current, believable readings:
//
//	|Δ| < 5 cm                ทรงตัว                                                         steady*
//	rising, under 1 m         สูงขึ้น 12 ซม. ใน 6 ชั่วโมงที่ผ่านมา                             rising*
//	falling, 1 m or more      ลดลง 1.20 เมตร ใน 6 ชั่วโมงที่ผ่านมา                             falling*
//	fewer than 4 points       omitted rather than guessed                                    sparse*
//
// Points are hourly buckets (the last reading in (HH−1:00, HH:00]) in the
// window H−6 h to H−1 h, plus the latest reading itself as the point for its
// own hour H. Four points in total are needed. When the earliest available
// bucket is not 6 h back, the sentence states the real span ("ใน 4 ชั่วโมง
// ที่ผ่านมา") instead of claiming six. Centimetres are used below 1 m, after
// rounding: 99.6 cm is shown as 1.00 เมตร.
//
// # National sentence
//
// Built from current readings only (not stale, not suspect):
//
//	no current readings       ยังไม่มีข้อมูลล่าสุดจากสถานีวัดระดับน้ำ                         national-none
//	0 overflowing             ตอนนี้ไม่มีสถานีที่น้ำล้นตลิ่ง [มี 145 สถานีที่น้ำมาก]          national-zero*
//	1 overflowing             ตอนนี้ 1 สถานีน้ำล้นตลิ่ง อยู่ที่ชุมพร                          national-one
//	all in one province       ตอนนี้ 3 สถานีน้ำล้นตลิ่ง ทั้งหมดอยู่ที่ชุมพร                   national-one-province
//	≤ 3 provinces hold all    … ทั้งหมดอยู่ที่ A และB                                         national-all-two
//	≤ 3 provinces hold > half … ส่วนใหญ่อยู่ที่ A B และC                                     national-majority*, national-tie-*
//	otherwise                 … กระจายอยู่ใน 10 จังหวัด                                      national-spread, national-tie-at-the-cutoff
//	unknown province, no most … the count alone                                              national-*-with-unknown
//
// "ส่วนใหญ่" (most) is only said when it is true and the names are not an
// arbitrary pick: the fewest top provinces, at most three, that together
// hold more than half of the overflowing stations, where the last one named
// has more stations than the first one left out. With 2, 1, 1, 1, 1 the
// first three hold 4 of 6, but naming two of the four tied provinces would
// be a coin toss, so the sentence says กระจายอยู่ใน 5 จังหวัด instead.
// Provinces are ranked by count; ties among names that are all shown are
// ordered by Unicode code point, which keeps the output deterministic but is
// not Thai dictionary order (เชียงใหม่ sorts after ลำปาง).
package humanize
