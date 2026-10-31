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
import { buildStyle, hillshadeExaggeration, skyFor, type ViewMode } from './style';
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
  /** The view mode (Phase 4): 2D flattens to mercator + full fog, 3D
   *  is the globe relief. Switching is sequenced, never same-frame. */
  view: ViewMode;
  onReady: () => void;
  onFlightDone: () => void;
  onStationSelect: (station: StationState) => void;
  /** The map is mid view-switch (the route disables the toggle). */
  onViewSwitching?: (switching: boolean) => void;
  /** The map itself failed to initialise; the route falls to its
   *  no-WebGL floor (the same honest shell). */
  onError: () => void;
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
  view,
  onReady,
  onFlightDone,
  onStationSelect,
  onError,
  onViewSwitching,
}: MapProps) {
  const container = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<MLMap | null>(null);
  const gaugesRef = useRef<GaugesLayer | null>(null);
  const readyRef = useRef(false);
  const onFlightDoneRef = useRef(onFlightDone);
  const onStationSelectRef = useRef(onStationSelect);
  const presetRef = useRef(preset);
  const stationsRef = useRef(stations);
  const exaggerationRef = useRef(exaggeration);
  const flightRequestRef = useRef(flightRequest);
  const onReadyRef = useRef(onReady);
  const onErrorRef = useRef(onError);
  const onViewSwitchingRef = useRef(onViewSwitching);
  // The view the map is actually showing (set at load and at each
  // switch's end) - the [view] effect no-ops until it differs.
  const appliedViewRef = useRef<ViewMode>(view);
  const viewRef = useRef(view);
  const switchingRef = useRef(false);

  // Latest-ref wiring happens in effects, not during render.
  useEffect(() => {
    onReadyRef.current = onReady;
  });
  useEffect(() => {
    onErrorRef.current = onError;
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
    exaggerationRef.current = exaggeration;
  });
  useEffect(() => {
    flightRequestRef.current = flightRequest;
  });
  useEffect(() => {
    viewRef.current = view;
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

        const style: StyleSpecification = buildStyle({
          basemap: hasPmtiles,
          view: viewRef.current,
        });
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
          // The initial view (a deep link or localStorage can ask for
          // 2D): the globe is the product's signature, mercator the
          // flat map with the full fog the engine supports there.
          if (viewRef.current === '2d') {
            void m.setProjection({ type: 'mercator' });
          } else {
            void m.setProjection({ type: 'globe' });
            // The ref, not the first-render closure: a slider move
            // during load must not be overwritten by a stale value.
            m.setTerrain({ source: 'dem', exaggeration: exaggerationRef.current });
          }
          appliedViewRef.current = viewRef.current;
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
          // The e2e evidence (and Phase 4 deep-linking) drives the
          // camera through this handle; the user's paths are the
          // presets, the flight and the click.
          (window as unknown as { __mahanathiMap?: MLMap }).__mahanathiMap = m;
          readyRef.current = true;
          // Anything asked for while the map was loading is applied
          // now, from the refs, not dropped on the floor: a preset
          // click, a slider move, or ใกล้ฉัน during load used to be
          // swallowed (the effects no-op until ready) and the flight
          // spinner never resolved.
          const pendingPreset = presetRef.current;
          if (pendingPreset.key !== preset.key) {
            applyPreset(m, pendingPreset, !reducedMotion(), appliedViewRef.current);
          }
          if (flightRequestRef.current > 0) {
            applyPreset(m, NATIONAL, false, appliedViewRef.current);
            window.setTimeout(() => {
              const mm = mapRef.current;
              if (!mm) return;
              mm.flyTo({ ...viewOf(STATION), duration: 2000, curve: 1.5, essential: true });
              mm.once('moveend', () => {
                onFlightDoneRef.current();
              });
            }, 120);
          }
          onReadyRef.current();
        });
      })
      .catch((e: unknown) => {
        if (ctrl.signal.aborted) return;
        if (e instanceof DOMException && e.name === 'AbortError') return;
        console.error('map initialisation failed', e);
        onErrorRef.current();
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

  // Presets: cut under reduced motion, fly otherwise. Flattened while
  // a view switch is in flight (the pill and ใกล้ฉัน are disabled for
  // the same window - the sequencing rule is the lead's).
  useEffect(() => {
    const map = mapRef.current;
    if (!map || !readyRef.current || switchingRef.current) return;
    applyPreset(map, preset, !reducedMotion(), appliedViewRef.current);
  }, [preset]);

  // The view switch (Phase 4): never the projection and the camera in
  // the same frame. setProjection reloads the tile managers, so the
  // sequence is projection → idle → camera, with the terrain and the
  // gauges re-seated after the transition (the layer is
  // projection-aware, but its ground contact reads the new terrain).
  // The tile-reload window is masked by a fade through the fog token -
  // a cut, not a fade, under reduced motion.
  useEffect(() => {
    const map = mapRef.current;
    if (!map || !readyRef.current) return;
    if (appliedViewRef.current === view || switchingRef.current) return;
    switchingRef.current = true;
    onViewSwitchingRef.current?.(true);
    const cut = reducedMotion();
    // The tile-reload mask: a DOM side effect of the switch lifecycle,
    // not React state - mounted here, removed in finish. A cut, not a
    // fade, under reduced motion (the overlay never mounts).
    const host = container.current?.parentElement ?? null;
    const fade = !cut && host ? document.createElement('div') : null;
    if (fade && host) {
      fade.className = 'lookdev__viewfade';
      fade.setAttribute('aria-hidden', 'true');
      host.appendChild(fade);
    }
    const finish = () => {
      appliedViewRef.current = view;
      switchingRef.current = false;
      fade?.remove();
      onViewSwitchingRef.current?.(false);
    };
    // The sky and the hillshade follow the mode (paint-level, cheap -
    // a setStyle would nuke the layers).
    map.setSky(skyFor(view));
    map.setPaintProperty('hillshade', 'hillshade-exaggeration', hillshadeExaggeration(view));
    if (view === '2d') {
      map.setTerrain(null);
      void map.setProjection({ type: 'mercator' });
      map.once('idle', () => {
        map.easeTo({ pitch: 0, bearing: 0, duration: cut ? 0 : 600, essential: true });
        map.once('moveend', finish);
      });
    } else {
      void map.setProjection({ type: 'globe' });
      map.once('idle', () => {
        map.setTerrain({ source: 'dem', exaggeration: exaggerationRef.current });
        gaugesRef.current?.setStations(stationsRef.current);
        const v = viewOf(presetRef.current);
        map.easeTo({
          pitch: v.pitch,
          bearing: v.bearing,
          duration: cut ? 0 : 600,
          essential: true,
        });
        map.once('moveend', finish);
      });
    }
  }, [view]);

  // Reduced motion is watched live, not read once (bug 7): if the
  // visitor turns it on while the page is open, the current view cuts
  // instead of waiting for the next flight to respect it.
  useEffect(() => {
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
    const onChange = () => {
      const map = mapRef.current;
      if (mq.matches && map && readyRef.current) {
        applyPreset(map, presetRef.current, false, appliedViewRef.current);
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
  // sample station (its geolocation twin lands in Phase 4). The button
  // is disabled for the view-switch window, so no mid-switch start.
  useEffect(() => {
    const map = mapRef.current;
    if (!map || !readyRef.current || flightRequest === 0) return;
    const flat = (v: { center: [number, number]; zoom: number; pitch: number; bearing: number }) =>
      appliedViewRef.current === '2d' ? { ...v, pitch: 0 } : v;
    if (reducedMotion()) {
      applyPreset(map, STATION, false, appliedViewRef.current);
      onFlightDoneRef.current();
      return;
    }
    applyPreset(map, NATIONAL, false, appliedViewRef.current);
    const t = window.setTimeout(() => {
      const m = mapRef.current;
      if (!m) return;
      m.flyTo({ ...flat(viewOf(STATION)), duration: 2000, curve: 1.5, essential: true });
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

function applyPreset(map: MLMap, preset: Preset, animate: boolean, view: ViewMode): void {
  if (preset.bounds) {
    map.fitBounds(preset.bounds, { padding: 72, pitch: 0, bearing: 0, animate, duration: 1600 });
    return;
  }
  // 2D is flat by definition: the authored pitches belong to the relief.
  const v = view === '2d' ? { ...viewOf(preset), pitch: 0 } : viewOf(preset);
  if (animate) {
    map.flyTo({ ...v, duration: 1600, curve: 1.4, essential: true });
  } else {
    map.jumpTo(v);
  }
}
