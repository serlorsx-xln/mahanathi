// /color - the situation-scale evidence page for the Phase 3 gate:
// the swatch strip with the official labels, the same strip under
// deuteranopia and protanopia simulation (Machado et al. 2009, applied
// in the page by an SVG colour matrix so the screenshot shows exactly
// what a colour-blind visitor sees), and the computed WCAG contrast of
// every text role on every surface. The numbers come from the same
// tokens module the tests use.
import { colors, contrast, situations, stale } from '../design/tokens';
import './color.css';

// Machado, Oliveira & Fernandes 2009, severity 1.0. Applied to linear
// RGB by the browser's feColorMatrix, which is what the SVG filter
// pipeline does before this matrix.
const DEUTAN =
  '0.367322 0.860646 -0.227968 0 0  0.280085 0.672501 0.047413 0 0  -0.011820 0.042940 0.968881 0 0  0 0 0 1 0';
const PROTAN =
  '0.152286 1.052583 -0.204868 0 0  0.114503 0.786281 0.099216 0 0  -0.003882 -0.048116 1.051998 0 0  0 0 0 1 0';

const SURFACES = [
  ['พื้นหลัง (ground)', colors.ground],
  ['แผงคำตอบ (panel)', colors.panel],
  ['แผงยกสูง (elevated)', colors.elevated],
] as const;

const TEXT_ROLES = [
  ['ข้อความหลัก', colors.textPrimary],
  ['ข้อความรอง', colors.textSecondary],
  ['ข้อความจาง (เวลาวัด แหล่งที่มา)', colors.textFaint],
  ['สีน้ำ (ลิงก์ โฟกัส)', colors.accentWater],
] as const;

function Swatch({
  hex,
  label,
  sub,
  hollow = false,
}: {
  hex: string;
  label: string;
  sub: string;
  /** Stale is hollow by design: a stroke, never a fill, so it cannot
   *  be confused with the near-grey level 2 on any vision. */
  hollow?: boolean;
}) {
  return (
    <div className="color__swatch">
      <span
        className={hollow ? 'color__chip color__chip--hollow' : 'color__chip'}
        style={
          hollow
            ? { background: 'transparent', border: `2px solid ${hex}` }
            : { background: hex }
        }
      />
      <span className="color__name">{label}</span>
      <span className="color__sub">{sub}</span>
    </div>
  );
}

export function ColorEvidence() {
  return (
    <article className="color">
      <header className="color__head">
        <h1>มาตราส่วนสถานการณ์น้ำ</h1>
        <p>
          หน้าพัฒนาเพื่อตรวจหลักฐานการออกแบบสี ระดับ 4 และ 5 เท่านั้นที่ได้สีเข้ม ระดับ 1
          เป็นสีคนละตระกูลกับน้ำท่วม ทุกระดับผ่าน WCAG AA บนทุกพื้นผิว
        </p>
      </header>

      <svg width="0" height="0" aria-hidden="true" focusable="false">
        <defs>
          <filter id="cvd-deutan" colorInterpolationFilters="linearRGB">
            <feColorMatrix type="matrix" values={DEUTAN} />
          </filter>
          <filter id="cvd-protan" colorInterpolationFilters="linearRGB">
            <feColorMatrix type="matrix" values={PROTAN} />
          </filter>
        </defs>
      </svg>

      <section>
        <h2>แถบสีตามสายตาปกติ</h2>
        <div className="color__strip">
          {situations.map((s) => (
            <Swatch key={s.level} hex={s.hex} label={s.label} sub={`ระดับ ${s.level}`} />
          ))}
          <Swatch hex={stale.hex} label={stale.label} sub="วงกลมว่าง เส้นขอบอย่างเดียว" hollow />
        </div>
      </section>

      <section>
        <h2>ภาพจำลองตาบอดสีเขียว-แดง</h2>
        <p className="color__note">
          จำลองด้วยเมทริกซ์ของ Machado et al. 2009 ที่ระดับรุนแรงเต็ม ทับลงบนหน้าจริง
        </p>
        <h3>deuteranopia</h3>
        <div className="color__strip" style={{ filter: 'url(#cvd-deutan)' }}>
          {situations.map((s) => (
            <Swatch key={s.level} hex={s.hex} label={s.label} sub={`ระดับ ${s.level}`} />
          ))}
        </div>
        <h3>protanopia</h3>
        <div className="color__strip" style={{ filter: 'url(#cvd-protan)' }}>
          {situations.map((s) => (
            <Swatch key={s.level} hex={s.hex} label={s.label} sub={`ระดับ ${s.level}`} />
          ))}
        </div>
        <p className="color__note">
          สีแดงและสีเขียวเข้าใกล้กันเป็นสีเหลืองมะกอกในภาพจำลองทั้งสอง ระดับจึงแยกกันด้วยความสว่าง
          ความอิ่มสี และแกนน้ำเงินของระดับ 1 ที่อยู่รอดจากการจำลองทั้งสองแบบ
          คู่ที่ใกล้กันที่สุดห่างกันราว 14-15 หน่วย delta-E
        </p>
      </section>

      <section>
        <h2>ค่าความคมชัดที่คำนวณได้</h2>
        <table className="color__table">
          <thead>
            <tr>
              <th>บทบาทข้อความ</th>
              {SURFACES.map(([name]) => (
                <th key={name}>{name}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {TEXT_ROLES.map(([role, hex]) => (
              <tr key={role}>
                <th scope="row">
                  <span className="color__inline-chip" style={{ background: hex }} />
                  {role}
                </th>
                {SURFACES.map(([, bg]) => (
                  <td key={bg}>{contrast(hex, bg).toFixed(2)}</td>
                ))}
              </tr>
            ))}
            {situations.map((s) => (
              <tr key={s.level}>
                <th scope="row">
                  <span className="color__inline-chip" style={{ background: s.hex }} />
                  {s.label}
                </th>
                {SURFACES.map(([, bg]) => (
                  <td key={bg}>{contrast(s.hex, bg).toFixed(2)}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
        <p className="color__note">
          เกณฑ์ WCAG AA: 4.5:1 สำหรับข้อความขนาดปกติ ทุกช่องในตารางต้องเกินเกณฑ์นี้
          (ตรวจโดยทดสอบอัตโนมัติใน tokens.test.ts)
        </p>
      </section>
    </article>
  );
}
