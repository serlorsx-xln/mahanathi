// The disc fallback (brief 04): the same stations as flat circles, so
// the lead can choose between poles and discs on visual merit. Behind
// ?gauges=discs. Same encoding rules as the poles where a circle can
// carry them: colour by situation, stale is a hollow stroke (never a
// fill), suspect is a translucent fill.
import type { GeoJSONSource, Map as MLMap } from 'maplibre-gl';
import { situations, stale as staleToken } from '../design/tokens';
import type { StationState } from './stations';

export const DISC_SOURCE = 'gauges-discs';
export const DISC_LAYER = 'gauges-discs';

export function stationsToDiscGeoJson(stations: StationState[]): GeoJSON.FeatureCollection {
  return {
    type: 'FeatureCollection',
    features: stations.map((s) => ({
      type: 'Feature' as const,
      geometry: { type: 'Point' as const, coordinates: [s.lon, s.lat] },
      properties: {
        id: s.id,
        level: s.level,
        stale: s.stale,
        suspect: !!(s.suspect && s.suspect.length > 0),
      },
    })),
  };
}

export function addDiscLayer(map: MLMap, stations: StationState[]): void {
  // The level->colour match: every situation level, stale as default.
  const levelMatch: unknown[] = ['match', ['get', 'level']];
  for (const s of situations) levelMatch.push(s.level, s.hex);
  levelMatch.push(staleToken.hex);

  map.addSource(DISC_SOURCE, {
    type: 'geojson',
    data: stationsToDiscGeoJson(stations),
  });
  map.addLayer({
    id: DISC_LAYER,
    type: 'circle',
    source: DISC_SOURCE,
    paint: {
      'circle-color': levelMatch as never,
      // Stale: hollow - the stroke only, never a fill. Suspect: the
      // fill at less than full trust. Otherwise a solid fill.
      'circle-opacity': [
        'case',
        ['get', 'stale'],
        0,
        ['get', 'suspect'],
        0.45,
        1,
      ],
      'circle-stroke-color': levelMatch as never,
      'circle-stroke-width': ['case', ['get', 'stale'], 1.5, 0],
      'circle-stroke-opacity': 0.9,
      'circle-radius': ['interpolate', ['linear'], ['zoom'], 4, 2.2, 8, 3.5, 12, 6],
    },
  });
}

export function setDiscData(map: MLMap, stations: StationState[]): void {
  const source = map.getSource<GeoJSONSource>(DISC_SOURCE);
  if (source) {
    void source.setData(stationsToDiscGeoJson(stations));
  }
}
