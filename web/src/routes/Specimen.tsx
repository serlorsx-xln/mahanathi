// /specimen - the Phase 3 type decision page. Every string is real
// output of the Phase 2 pipeline, captured 2026-10-29 (the fixture in
// web/public/fixture/). Three pairings, each with the sentence face
// switchable, so the UI decision and the serif decision stay
// independent (docs/research/typography.md, section 4).
import '../design/fonts.css';
import './specimen.css';

const NATIONAL = 'ตอนนี้ 18 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ที่นครปฐม กาญจนบุรี และสุพรรณบุรี';

// Real station sentences from the same snapshot: station, then sentence.
const STATION_SENTENCES: [string, string][] = [
  ['BPK003 บางน้ำเปรี้ยว', 'สูงกว่าตลิ่ง 0.39 เมตร (น้ำล้นตลิ่ง) ลดลง 5 ซม. ใน 6 ชั่วโมงที่ผ่านมา'],
  [
    'BKK003 คลองมหาสวัสดิ บางกรวย-สวนผัก',
    'ต่ำกว่าตลิ่ง 0.30 เมตร สูงขึ้น 8 ซม. ใน 6 ชั่วโมงที่ผ่านมา',
  ],
  [
    'NPNPU01 เขื่อนน้ำพุง บ้านคำเพิ่ม',
    'สถานีนี้ยังไม่มีข้อมูลระดับตลิ่ง ระดับน้ำ 279.67 ม.รทก. ทรงตัว',
  ],
];

// AY has the longest province name and every stacked-mark test the
// brief names, plus the clipping-risk words from the research.
const STACK_CHECK = 'จังหวัดพระนครศรีอยุธยา · น้ำล้นตลิ่ง · ผู้ใหญ่บ้าน · ปิ่นเกล้า · ปั๊มน้ำ';

// Real numbers from the snapshot: signed distance to bank (minus sign
// U+2212), percent of bank, measured time. Tests figure alignment.
const READOUT: [string, string, string, string, string][] = [
  ['BPK003', '2.048', '+0.39', '121.63', '14:40'],
  ['BKK003', '1.769', '−0.30', '94.20', '14:40'],
  ['THA010', '8.297', '+1.79', '121.63', '13:10'],
  ['NPNPU01', '279.670', '·', '·', '11:10'],
];

const PAIRINGS = [
  {
    id: 'a',
    label: 'ก. IBM Plex Sans Thai',
    note: 'เซตในชื่อ Mahanathi Sans ตามเงื่อนไข OFL เรื่องชื่อสงวน',
    family: "'Mahanathi Sans', sans-serif",
  },
  {
    id: 'b',
    label: 'ข. Anuphan',
    note: 'ไทยวาดใหม่บนละตินของ Plex Sans ตัวเดียวกัน',
    family: "'Anuphan', sans-serif",
  },
  {
    id: 'c',
    label: 'ค. Noto Sans Thai',
    note: 'ตัวเทียบ (control)',
    family: "'Noto Sans Thai', sans-serif",
  },
];

const SENTENCE_FACES = [
  { label: 'sans ตัวหนา', css: { fontFamily: 'inherit', fontWeight: 600 } },
  { label: 'Noto Serif Thai', css: { fontFamily: "'Noto Serif Thai', serif" } },
  { label: 'Pridi', css: { fontFamily: "'Pridi', serif" } },
  { label: 'Trirong', css: { fontFamily: "'Trirong', serif" } },
];

export function Specimen() {
  return (
    <article className="specimen">
      <header className="specimen__head">
        <h1>ตัวอย่างตัวอักษร</h1>
        <p>
          หน้าพัฒนาสำหรับเลือกฟอนต์ ข้อความทุกบรรทัดเป็นผลลัพธ์จริงของขั้นที่ 2 จากภาพรวมเวลา 21:53
          น. วันที่ 29 ตุลาคม 2026
        </p>
        <p className="specimen__note">
          ตัดสินแล้วตาม ADR 0005 (IBM Plex Sans Thai + Noto Serif Thai)
          ต่อมาถอน Noto Serif Thai ตามกฎไร้หัวตัวอักษรของเจ้าของ (ภาคผนวกของ ADR 0005)
          หลังจากนี้หน้านี้คงไว้เป็นหลักฐาน ฟอนต์ที่ไม่ถูกเลือกถูกลบออกจาก build
          ถ้าต้องการเปิดเปรียบเทียบสด ให้รัน scripts/research/promote_fonts.py แล้วรีเฟรชหน้านี้
        </p>
      </header>

      {PAIRINGS.map((p) => (
        <section key={p.id} className="specimen__pairing" style={{ fontFamily: p.family }}>
          <h2>{p.label}</h2>
          <p className="specimen__note">{p.note}</p>

          <div className="specimen__sentences">
            {SENTENCE_FACES.map((f) => (
              <figure key={f.label} className="specimen__sentence">
                <figcaption>{f.label}</figcaption>
                <blockquote style={f.css}>{NATIONAL}</blockquote>
              </figure>
            ))}
          </div>

          <div className="specimen__body">
            {STATION_SENTENCES.map(([station, sentence]) => (
              <p key={station}>
                <strong>{station}</strong> {sentence}
              </p>
            ))}
            <p className="specimen__stack">{STACK_CHECK}</p>
            <p className="specimen__province">จังหวัดพระนครศรีอยุธยา</p>
          </div>

          <table className="specimen__readout">
            <thead>
              <tr>
                <th>สถานี</th>
                <th>ระดับน้ำ ม.รทก.</th>
                <th>ห่างตลิ่ง ม.</th>
                <th>% ของตลิ่ง</th>
                <th>เวลาวัด</th>
              </tr>
            </thead>
            <tbody>
              {READOUT.map((row) => (
                <tr key={row[0]}>
                  {row.map((cell, i) => (
                    <td key={i}>{cell}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      ))}
    </article>
  );
}
