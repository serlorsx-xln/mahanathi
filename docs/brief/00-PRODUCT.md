# Product brief: มหานที (Mahanathi)

> Owner: ชานน · Lead brief written 2026-10-29 · Deadline: live and rehearsed by 2026-10-31 12:00 ICT

## One sentence
Every river gauge in Thailand, live, on a 3D relief of the country. Tap **ใกล้ฉัน** and get one plain Thai sentence about the river nearest you.

> *"แม่น้ำเจ้าพระยาใกล้คุณที่สุด (สถานีบางไทร, 3.2 กม.) ตอนนี้ต่ำกว่าตลิ่ง 1.40 เมตร สูงขึ้น 12 ซม. ใน 6 ชั่วโมงที่ผ่านมา · วัดเมื่อ 14:00 น."*

## Why it exists
- Thailand publishes hourly water levels for about 800 river stations through the National Hydroinformatics Data Center (ThaiWater). The data is excellent, but the presentation is built for hydrologists: tables of metres above mean sea level (ม.รทก.), station codes and five garish colours. Ordinary people cannot answer the only question they have: *what is the river near me doing right now?*
- It is flood season. At 14:00 on 2026-10-29, 22 stations were above bank-full and 145 were at "high water". The most affected provinces were Ayutthaya (17), Surat Thani (11), Chumphon (10) and Bangkok (6).
- Context: this is the owner's homework for the YWC21 Web Programming interview. The camp theme is *"Re:Humanize the Web — จากมหานทีแห่งอักขรา สู่ขอบฟ้าแห่งปัญญาประดิษฐ์"*. The project's name and purpose come straight from it: turning expert numbers into human language. The theme is never preached in the UI.

## Users and the moment of use
1. **Primary:** a person in a flood-prone area such as Ayutthaya, Pathum Thani, Nonthaburi or the southern provinces. They are checking at night on a phone, with rain outside, worried about the river or canal behind their house. They need a one-glance answer, trustworthy numbers and the time of measurement.
2. **Secondary:** curious people and journalists who want the national picture: where water is high right now and how it changed over the last 72 hours.
3. **Demo audience:** YWC judges, who are web programmers and designers. They will see it on a laptop and hold it on their own phone for 20 seconds. They are experts at spotting template or AI-generic work.

## Core jobs (must ship)
| # | Job | Done when |
|---|---|---|
| J1 | See the country now | The first frame shows the relief with every live station encoded and one computed national summary sentence, in under 2 s on a mid-range phone. |
| J2 | ใกล้ฉัน | One tap leads to geolocation (computed on the device), the nearest station with valid data, a camera flight to it and the human sentence. If permission is denied, search is offered with helpful copy. |
| J3 | Search | Type a province, district (อำเภอ), river or station in Thai, tolerant of tone marks and spacing, to get the same answer panel. |
| J4 | Trust | Every number shows its measured time and the ThaiWater source. Stale or suspect data is labelled, never shown as current. |
| J5 | See change | A 72-hour timeline scrubber animates every gauge. The station panel shows a small history chart of the station against bank-full. |
| J6 | Works everywhere | Usable on phone and desktop, inside the LINE in-app browser (or with a prompt to open externally), and without WebGL through a designed 2D fallback. It honours reduced motion. |

## Explicit non-goals
- **No prediction, no warnings and no safety claims.** Forecasting is the job of the authorities. We show measured values and their trend.
- No accounts, no login, no personal data. Location never leaves the device.
- No AI inside the product. Numbers come from tested code. This is a deliberate decision and the owner can explain it.
- No official look. Use our own brand, credit the source clearly, and never imitate a government site.

## Success criteria for the demo
- A judge scans a QR code, taps ใกล้ฉัน and reads a correct sentence about a real river within 20 seconds.
- On the laptop, the first frame makes the room go quiet: a premium, serious, beautiful instrument, not a template.
- Every "why" question has an answer in the docs: why this data, why no prediction, why SQLite, how stale data is handled, what the AI did and what must be reviewed by a human.
