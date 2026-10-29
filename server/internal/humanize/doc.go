// Package humanize turns a normalised reading and its recent history into one
// plain Thai sentence plus the structured facts behind it.
//
// It is pure: no I/O, no clock reads, no randomness. It never predicts and
// never reassures; a banned-phrase test fuzzes every output against words
// such as ปลอดภัย, จะท่วม and ไม่ท่วม. The rules it implements are the humanize
// spec in docs/brief/02-ENGINEERING.md §3, and every branch is covered by a
// table or golden test.
package humanize
