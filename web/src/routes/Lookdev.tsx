// /lookdev - the Phase 3b relief on MapLibre (ADR 0007): the terrain,
// hillshade, basemap and fog come from the engine; our data (borders,
// rivers, and from tomorrow the gauges) renders on top. The relief
// fills the frame; the national sentence and the answer panel float
// over it; the primary action ใกล้ฉัน flies to the selected river. Dev
// controls (presets, exaggeration) sit at the bottom and do not ship
// with the product.
import { lazy, Suspense, useEffect, useMemo, useState } from 'react';
import { situations, stale } from '../design/tokens';
import type { StationState } from '../map/stations';
import { NATIONAL, PRESETS, STATION, type Preset } from '../map/presets';
import './lookdev.css';

const LookdevMap = lazy(async () => {
  const m = await import('../map/LookdevMap');
  return { default: m.LookdevMap };
});

export interface Fixture {
  generatedAt: string;
  summary: {
    national: string;
    levels: number[];
    noLevel: number;
    staleCount: number;
  };
  stations: StationState[];
}

const SOURCE_LINE = 'แหล่งที่มา: ThaiWater (กรมชลประทาน) · แผนที่ฐาน: OpenStreetMap (ODbL) · Protomaps';

/** The gauge rendering for this page load: poles (the 3D layer) by
 *  default, discs behind ?gauges=discs for the lead's comparison. */
function gaugeMode(): 'poles' | 'discs' {
  if (typeof window === 'undefined') return 'poles';
  return new URLSearchParams(window.location.search).get('gauges') === 'discs'
    ? 'discs'
    : 'poles';
}

/** The relief needs WebGL. Without it the page keeps its shell and the
 *  sentence, and says so plainly (the product's 2D map fallback is a
 *  later phase; this is the look-dev prototype's honest floor). */
function webglAvailable(): boolean {
  if (typeof window === 'undefined') return true; // the SSR render
  const c = document.createElement('canvas');
  return !!(c.getContext('webgl2') ?? c.getContext('webgl'));
}

/** LINE's in-app browser (J6): detect it and offer to open externally -
 *  the terrain needs a real browser's WebGL and memory headroom. The
 *  offer is a link, because only a navigation LINE intercepts can leave
 *  the webview; `?openExternalBrowser=1` is LINE's documented escape
 *  hatch. */
function inLineBrowser(): boolean {
  if (typeof window === 'undefined') return false;
  return /Line\//i.test(window.navigator.userAgent);
}

function formatTime(iso: string): string {
  const d = new Date(iso);
  return (
    d.toLocaleTimeString('th-TH', {
      hour: '2-digit',
      minute: '2-digit',
      timeZone: 'Asia/Bangkok',
    }) + ' น.'
  );
}

export function Lookdev() {
  const [data, setData] = useState<{
    fixture: Fixture;
    borders: unknown;
    rivers: GeoJSON.FeatureCollection;
  } | null>(null);
  const [preset, setPreset] = useState<Preset>(NATIONAL);
  const [exaggeration, setExaggeration] = useState(1.5);
  const [flightRequest, setFlightRequest] = useState(0);
  const [flightDone, setFlightDone] = useState(true);
  // Surfaced as data-scene-ready on the <main>: the panel can appear
  // before the lazy scene chunk has rendered its first frame, and the
  // e2e evidence must not screenshot an empty canvas.
  const [sceneReady, setSceneReady] = useState(false);
  const [webgl] = useState(webglAvailable);
  const [failed, setFailed] = useState(false);
  const [attempt, setAttempt] = useState(0);
  const [mode] = useState(gaugeMode);
  const [lineHintDismissed, setLineHintDismissed] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    let cancelled = false;
    const markFailed = () => {
      if (!cancelled) setFailed(true);
    };
    void (async () => {
      try {
        const get = (url: string) =>
          fetch(url, { signal: controller.signal }).then((r) => r.json());
        const [fixture, borders, rivers] = (await Promise.all([
          get('/fixture/now-2026-10-29.json'),
          get('/geodata/borders.topo.json'),
          get('/geodata/rivers.geojson'),
        ])) as unknown as [Fixture, unknown, GeoJSON.FeatureCollection];
        if (!controller.signal.aborted) {
          setData({ fixture, borders, rivers });
          setFailed(false);
        }
      } catch (e) {
        if (!controller.signal.aborted && !(e instanceof DOMException && e.name === 'AbortError')) {
          markFailed();
        }
      }
    })();
    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [attempt]);

  const selected = useMemo(() => {
    if (!data || preset.station === undefined) return null;
    return data.fixture.stations.find((s) => s.id === preset.station) ?? null;
  }, [data, preset]);

  const summary = data?.fixture.summary;

  return (
    <main className="lookdev" data-scene-ready={sceneReady ? '1' : undefined}>
      <div
        className="lookdev__canvas"
        role="img"
        aria-label="แผนที่ภูมิประเทศประเทศไทย แสดงเกจวัดระดับน้ำทุกสถานี"
      >
        {data && webgl ? (
          <Suspense fallback={<div className="lookdev__skeleton" />}>
            <LookdevMap
              bordersTopo={data.borders}
              riversGeo={data.rivers}
              stations={data.fixture.stations}
              gaugeMode={mode}
              preset={preset}
              exaggeration={exaggeration}
              flightRequest={flightRequest}
              onFlightDone={() => {
                setFlightDone(true);
              }}
              onReady={() => {
                setSceneReady(true);
              }}
              onStationSelect={(s) => {
                setPreset({
                  ...STATION,
                  station: s.id,
                  view: { center: [s.lon, s.lat], zoom: 12.2, pitch: 60, bearing: -24 },
                });
              }}
            />
          </Suspense>
        ) : !data && !failed ? (
          <div className="lookdev__skeleton" role="status">
            <span className="lookdev__sr">กำลังโหลดแผนที่และข้อมูลระดับน้ำ</span>
          </div>
        ) : null}
      </div>

      {failed && (
        <section className="lookdev__error" aria-label="โหลดข้อมูลไม่สำเร็จ">
          <p>โหลดข้อมูลระดับน้ำไม่ได้ ตรวจการเชื่อมต่อแล้วลองใหม่ได้เลย</p>
          <button
            type="button"
            onClick={() => {
              setFailed(false); // back to the loading status while it retries
              setAttempt((n) => n + 1);
            }}
          >
            ลองใหม่
          </button>
        </section>
      )}

      {!webgl && (
        <p className="lookdev__nogl" role="status">
          อุปกรณ์นี้ไม่รองรับกราฟิก 3D จึงแสดงเฉพาะข้อความสรุป ข้อมูลระดับน้ำยังอ่านได้ตามปกติ
        </p>
      )}

      <header className="lookdev__chrome lookdev__chrome--top">
        <div className="lookdev__wordmark">มหานที</div>
        <button
          type="button"
          className="lookdev__nearme"
          onClick={() => {
            setFlightDone(false);
            setPreset(STATION);
            setFlightRequest((n) => n + 1);
          }}
        >
          ใกล้ฉัน<span className="lookdev__nearme-note">· ตัวอย่างสถานี</span>
        </button>
      </header>

      {inLineBrowser() && !lineHintDismissed && (
        <aside className="lookdev__linehint" aria-label="เปิดในเบราว์เซอร์">
          <p>
            แผนที่ภูมิประเทศ 3 มิติทำงานได้ดีที่สุดในเบราว์เซอร์
            <a href="?openExternalBrowser=1">เปิดในเบราว์เซอร์ภายนอก</a>
          </p>
          <button
            type="button"
            className="lookdev__linehint-close"
            aria-label="ปิดคำแนะนำ"
            onClick={() => {
              setLineHintDismissed(true);
            }}
          >
            <svg viewBox="0 0 16 16" width="12" height="12" aria-hidden="true">
              <path
                d="M3 3 L13 13 M13 3 L3 13"
                stroke="currentColor"
                stroke-width="1.75"
                stroke-linecap="round"
              />
            </svg>
          </button>
        </aside>
      )}

      {summary && (
        <section className="lookdev__panel" aria-label="สรุปสถานการณ์น้ำ">
          <p className="lookdev__sentence">{summary.national}</p>
          <dl className="lookdev__numbers">
            {situations.map((s, i) => (
              <div key={s.level} className="lookdev__number">
                <dt>
                  <span className="lookdev__dot" style={{ background: s.hex }} />
                  {s.label}
                </dt>
                <dd>{summary.levels[i] ?? 0}</dd>
              </div>
            ))}
            <div className="lookdev__number">
              <dt>
                <span className="lookdev__dot lookdev__dot--hollow" />
                {stale.label}
              </dt>
              <dd>{summary.staleCount}</dd>
            </div>
          </dl>
          <p className="lookdev__meta">
            ภาพรวมเวลา {formatTime(data.fixture.generatedAt)} · {SOURCE_LINE}
          </p>
        </section>
      )}

      {selected && (
        <section className="lookdev__station" aria-label="สถานีที่เลือก">
          <h2>{selected.name}</h2>
          <p className="lookdev__sentence lookdev__sentence--station">{selected.sentence}</p>
          <dl className="lookdev__numbers lookdev__numbers--station">
            <div>
              <dt>ระดับน้ำ</dt>
              <dd>{selected.wl !== null ? `${selected.wl.toFixed(3)} ม.รทก.` : '·'}</dd>
            </div>
            <div>
              <dt>ระดับตลิ่ง</dt>
              <dd>{selected.bankFull !== null ? `${selected.bankFull.toFixed(3)} ม.รทก.` : '·'}</dd>
            </div>
            <div>
              <dt>% ของตลิ่ง</dt>
              <dd>{selected.pct !== null ? `${selected.pct.toFixed(2)}%` : '·'}</dd>
            </div>
            <div>
              <dt>วัดเมื่อ</dt>
              <dd>{selected.measuredAt ? formatTime(selected.measuredAt) : '·'}</dd>
            </div>
          </dl>
        </section>
      )}

      <footer className="lookdev__chrome lookdev__chrome--bottom">
        <div className="lookdev__presets" role="group" aria-label="มุมมอง">
          {PRESETS.map((p) => (
            <button
              key={p.key}
              type="button"
              aria-pressed={preset.key === p.key}
              className={
                preset.key === p.key ? 'lookdev__preset lookdev__preset--active' : 'lookdev__preset'
              }
              onClick={() => {
                setPreset(p);
              }}
            >
              {p.label}
            </button>
          ))}
        </div>
        <label className="lookdev__exagg">
          อัตรายกสูง {exaggeration}×
          <input
            type="range"
            min={1}
            max={40}
            step={0.5}
            value={exaggeration}
            onChange={(e) => {
              setExaggeration(Number(e.target.value));
            }}
          />
        </label>
        {flightDone ? null : (
          <span className="lookdev__flying" role="status">
            กำลังบิน…
          </span>
        )}
      </footer>
    </main>
  );
}
