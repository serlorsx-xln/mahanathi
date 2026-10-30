// The /lookdev map on MapLibre (ADR 0007): globe projection, our
// terrarium DEM as terrain and hillshade, the self-built Protomaps
// dark basemap when it is present, our borders and rivers, and one
// continuous night fog (the style's sky block). The DOM contracts the
// e2e relies on - data-scene-ready, the preset labels, the
// exaggeration slider, the flight - are the same as the hand-built
// relief's, so the evidence approach carries over unchanged.
import { useEffect, useRef } from 'react';
import {
  Map as MLMap,
  addProtocol,
  setWorkerUrl,
  type StyleSpecification,
  type LngLatBoundsLike,
} from 'maplibre-gl';
import { Protocol } from 'pmtiles';
import * as topojson from 'topojson-client';
import 'maplibre-gl/dist/maplibre-gl.css';
import { buildStyle } from './style';
import { NATIONAL, STATION, type Preset } from './presets';
import { GaugesLayer } from './gauges';
import { DISC_LAYER, addDiscLayer, setDiscData } from './discs';
import type { StationState } from './stations';

// MapLibre 6 loads its worker by resolving `./maplibre-gl-worker.mjs`
// against import.meta.url, which points into our chunk once Vite has
// bundled us - the file is not there and every worker fails ("Worker
// failed to load", and the map never fires `load`). Point it at the
// worker emitted as its own asset instead. The worker bundle is
// self-contained, so this is the whole fix.
import workerUrl from 'maplibre-gl/dist/maplibre-gl-worker.mjs?url';
setWorkerUrl(workerUrl);

export interface MapProps {
  bordersTopo: unknown;
  riversGeo: GeoJSON.FeatureCollection;
  stations: StationState[];
  /** poles: the 3D custom layer (default). discs: the flat fallback
   *  behind ?gauges=discs, for the lead's visual comparison. */
  gaugeMode: 'poles' | 'discs';
  preset: Preset;
  exaggeration: number;
  flightRequest: number;
  onReady: () => void;
  onFlightDone: () => void;
  onStationSelect: (station: StationState) => void;
}

// One shared PMTiles protocol instance for the page (it caches tiles).
const pmtiles = new Protocol();
addProtocol('pmtiles', pmtiles.tile);

function bordersToGeoJson(topo: unknown): GeoJSON.FeatureCollection {
  // The shipped borders.topo.json has one object whose geometries are
  // the border segments; every line becomes one feature.
  const t = topo as { objects: Record<string, unknown> };
  const key = Object.keys(t.objects)[0];
  const obj = key ? t.objects[key] : null;
  if (!key || !obj) return { type: 'FeatureCollection', features: [] };
  const geo = topojson.feature(
    t as Parameters<typeof topojson.feature>[0],
    obj as Parameters<typeof topojson.feature>[1],
  ) as unknown as GeoJSON.FeatureCollection;
  return geo;
}

function reducedMotion(): boolean {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

export function LookdevMap({
  bordersTopo,
  riversGeo,
  stations,
  gaugeMode,
  preset,
  exaggeration,
  flightRequest,
  onReady,
  onFlightDone,
  onStationSelect,
}: MapProps) {
  const container = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<MLMap | null>(null);
  const gaugesRef = useRef<GaugesLayer | null>(null);
  const readyRef = useRef(false);
  const onReadyRef = useRef(onReady);
  const onFlightDoneRef = useRef(onFlightDone);
  const onStationSelectRef = useRef(onStationSelect);
  const presetRef = useRef(preset);
  const stationsRef = useRef(stations);

  // Latest-ref wiring happens in effects, not during render.
  useEffect(() => {
    onReadyRef.current = onReady;
  });
  useEffect(() => {
    onFlightDoneRef.current = onFlightDone;
  });
  useEffect(() => {
    onStationSelectRef.current = onStationSelect;
  });
  useEffect(() => {
    presetRef.current = preset;
  });

  useEffect(() => {
    if (!container.current) return;
    const ctrl = new AbortController();
    let map: MLMap | null = null;

    // The style is built for what exists: the page never references a
    // basemap file that is not there (it is built, not shipped, so its
    // presence is a local fact to check, not an assumption). A bare
    // HEAD is not enough - dev and preview servers fall back to
    // index.html with 200 for missing paths, which the PMTiles reader
    // then chokes on ("wrong magic number") - so the content type must
    // be the archive's, not a document's.
    void fetch('/geodata/thailand.pmtiles', {
      method: 'HEAD',
      signal: ctrl.signal,
    })
      .then((r) => {
        const type = r.headers.get('content-type') ?? '';
        return r.ok && type !== '' && !type.includes('text/html');
      })
      .catch(() => false)
      .then((hasPmtiles) => {
        if (ctrl.signal.aborted || !container.current) return;

        const style: StyleSpecification = buildStyle({ basemap: hasPmtiles });
        const borders = bordersToGeoJson(bordersTopo);
        const riversSource = style.sources.rivers;
        const bordersSource = style.sources.borders;
        if (riversSource?.type === 'geojson') riversSource.data = riversGeo;
        if (bordersSource?.type === 'geojson') bordersSource.data = borders;

        map = new MLMap({
          container: container.current,
          style,
          // Start on the national bounds. The first view cuts (bug 7's
          // lesson); later preset changes are the ones that move.
          bounds: NATIONAL.bounds as LngLatBoundsLike,
          fitBoundsOptions: { padding: 72, pitch: 0, bearing: 0, animate: false },
          attributionControl: false,
          maxPitch: 70,
          // The e2e evidence pipeline samples the canvas (pixel checks
          // against the tokens); without this the readback races the
          // compositor and returns the clear colour.
          canvasContextAttributes: { preserveDrawingBuffer: true },
          maxBounds: [
            [90, 0],
            [113, 27],
          ],
        });
        mapRef.current = map;

        map.on('load', () => {
          if (ctrl.signal.aborted || !map) return;
          const m = map;
          void m.setProjection({ type: 'globe' });
          m.setTerrain({ source: 'dem', exaggeration });
          // The gauges: poles by default, discs behind ?gauges=discs.
          if (gaugeMode === 'discs') {
            addDiscLayer(m, stationsRef.current);
          } else {
            const gauges = new GaugesLayer();
            gaugesRef.current = gauges;
            m.addLayer(gauges);
            gauges.setStations(stationsRef.current);
          }
          // Station selection: click a disc, or click near a pole.
          m.on('click', (e) => {
            const nearest = nearestStation(m, stationsRef.current, e.point);
            if (nearest) onStationSelectRef.current(nearest);
          });
          if (gaugeMode === 'discs') {
            m.on('mouseenter', DISC_LAYER, () => {
              m.getCanvas().style.cursor = 'pointer';
            });
            m.on('mouseleave', DISC_LAYER, () => {
              m.getCanvas().style.cursor = '';
            });
          }
          readyRef.current = true;
          onReadyRef.current();
        });
      });

    return () => {
      ctrl.abort();
      mapRef.current?.remove();
      mapRef.current = null;
      readyRef.current = false;
    };
    // The map is built once; preset, exaggeration and flight flow
    // through the effects below.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Presets: cut under reduced motion, fly otherwise.
  useEffect(() => {
    const map = mapRef.current;
    if (!map || !readyRef.current) return;
    applyPreset(map, preset, !reducedMotion());
  }, [preset]);

  // Reduced motion is watched live, not read once (bug 7): if the
  // visitor turns it on while the page is open, the current view cuts
  // instead of waiting for the next flight to respect it.
  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
    const onChange = () => {
      const map = mapRef.current;
      if (mq.matches && map && readyRef.current) {
        applyPreset(map, presetRef.current, false);
      }
    };
    mq.addEventListener('change', onChange);
    return () => {
      mq.removeEventListener('change', onChange);
    };
  }, []);

  // Exaggeration: retarget the terrain in place; the gauges re-seat
  // themselves because their ground contact reads the terrain.
  useEffect(() => {
    const map = mapRef.current;
    if (!map || !readyRef.current) return;
    map.setTerrain({ source: 'dem', exaggeration });
    gaugesRef.current?.setStations(stationsRef.current);
  }, [exaggeration]);

  // Stations: pushed to whichever gauge layer is mounted, once per
  // change - never per frame (bug 5).
  useEffect(() => {
    stationsRef.current = stations;
    const map = mapRef.current;
    if (!map || !readyRef.current) return;
    if (gaugeMode === 'discs') {
      setDiscData(map, stations);
    } else {
      gaugesRef.current?.setStations(stations);
    }
  }, [stations, gaugeMode]);

  // The one authored flight: ใกล้ฉัน. Cut to national, then fly to the
  // sample station (its geolocation twin lands in Phase 4).
  useEffect(() => {
    const map = mapRef.current;
    if (!map || !readyRef.current || flightRequest === 0) return;
    if (reducedMotion()) {
      applyPreset(map, STATION, false);
      onFlightDoneRef.current();
      return;
    }
    applyPreset(map, NATIONAL, false);
    const t = window.setTimeout(() => {
      const m = mapRef.current;
      if (!m) return;
      m.flyTo({ ...viewOf(STATION), curve: 1.5, speed: 0.9, essential: true });
      m.once('moveend', () => {
        onFlightDoneRef.current();
      });
    }, 120);
    return () => {
      window.clearTimeout(t);
    };
  }, [flightRequest]);

  return <div ref={container} className="lookdev__map" />;
}

function viewOf(preset: Preset): {
  center: [number, number];
  zoom: number;
  pitch: number;
  bearing: number;
} {
  return preset.view ?? { center: [100.9, 13.8], zoom: 4.6, pitch: 0, bearing: 0 };
}

/** The station whose ground point is nearest the click, within a
 *  finger-sized radius. A screen-space hit test - the honest floor for
 *  the prototype (pitch occlusion is not modelled; the sheet names the
 *  station, so a miss behind a ridge is visible, not silent). */
function nearestStation(
  map: MLMap,
  stations: StationState[],
  point: { x: number; y: number },
): StationState | null {
  let best: StationState | null = null;
  let bestDist = 16; // px
  for (const s of stations) {
    const p = map.project([s.lon, s.lat]);
    const d = Math.hypot(p.x - point.x, p.y - point.y);
    if (d < bestDist) {
      bestDist = d;
      best = s;
    }
  }
  return best;
}

function applyPreset(map: MLMap, preset: Preset, animate: boolean): void {
  if (preset.bounds) {
    map.fitBounds(preset.bounds, { padding: 72, pitch: 0, bearing: 0, animate, duration: 1600 });
    return;
  }
  const v = viewOf(preset);
  if (animate) {
    map.flyTo({ ...v, curve: 1.4, speed: 0.9, essential: true });
  } else {
    map.jumpTo(v);
  }
}
