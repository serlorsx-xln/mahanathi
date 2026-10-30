// The MapLibre style, built from the tokens (ADR 0007). The single
// source of truth rule holds here too: no colour in this file is
// written by hand - every one comes from web/src/design/tokens.ts,
// which the generated tokens.css mirrors for the DOM chrome.
//
// The basemap layers come from protomaps-themes-base (BSD-3, the
// style for our self-built ODbL PMTiles). Its DARK flavor is close to
// our world but not ours, so the two loudest surfaces - background and
// water - are retinted to the tokens. Thai labels are deferred until
// we self-host a Thai glyph set (fonts.protomaps.com does not resolve;
// the about page and SOURCES.md credit OSM/Protomaps regardless).
import type { LayerSpecification, StyleSpecification } from 'maplibre-gl';
import { layers as protomapsLayers, namedTheme } from 'protomaps-themes-base';
import { colors, relief, situations } from '../design/tokens';

/** Mix two hex colours; the ratios (not the colours) are the only
 *  hand-set values in this file - every base colour is a token. */
function mix(a: string, b: string, t: number): string {
  const pa = [1, 3, 5].map((i) => parseInt(a.slice(i, i + 2), 16));
  const pb = [1, 3, 5].map((i) => parseInt(b.slice(i, i + 2), 16));
  const out = pa.map((v, i) => Math.round(v + ((pb[i] ?? v) - v) * t));
  return `#${out.map((v) => v.toString(16).padStart(2, '0')).join('')}`;
}

/** Retint a protomaps layer's paint into the token world. The dark
 *  theme's own greys (#1f1f1f earth, #3d3d3d roads, #111 buildings)
 *  took over the close zooms before this mapping existed - the gate
 *  critique measured 77.7% of the landing frame as Protomaps grey.
 *  Every replacement colour is derived from the tokens above. */
function retint(layer: LayerSpecification): LayerSpecification {
  if (layer.type === 'background') {
    return { ...layer, paint: { 'background-color': colors.ground } };
  }
  const id = layer.id;
  if (layer.type === 'fill' && id.startsWith('water')) {
    return { ...layer, paint: { ...layer.paint, 'fill-color': colors.sea } };
  }
  if (layer.type === 'fill' && id.startsWith('earth')) {
    return { ...layer, paint: { ...layer.paint, 'fill-color': colors.ground } };
  }
  if (layer.type === 'fill' && id.startsWith('landcover')) {
    // A whisper of the normal-level situation green over the ground -
    // the token, not a hand-written hex (this file's own rule).
    const normal = situations.find((s) => s.level === 3);
    const green = normal ? normal.hex : colors.ground;
    return {
      ...layer,
      paint: { ...layer.paint, 'fill-color': mix(colors.ground, green, 0.06) },
    };
  }
  if (layer.type === 'fill' && id.startsWith('buildings')) {
    return {
      ...layer,
      paint: { ...layer.paint, 'fill-color': mix(colors.ground, '#ffffff', 0.04) },
    };
  }
  if (layer.type === 'line' && id.startsWith('roads')) {
    // Quiet tracks: the ground lifted just enough to read as structure.
    return {
      ...layer,
      paint: { ...layer.paint, 'line-color': mix(colors.ground, '#ffffff', 0.12) },
    };
  }
  return layer;
}

export interface StyleInput {
  /** The PMTiles basemap is present (HEAD /geodata/thailand.pmtiles
   *  answered 200); without it we ship terrain, borders and rivers
   *  only, which is still the committed world. */
  basemap: boolean;
  /** The view mode (Phase 4): '2d' flattens to mercator with the full
   *  night fog and a softened hillshade; '3d' (the default) keeps the
   *  globe, the terrain and the atmosphere. */
  view?: ViewMode;
}

/** The view modes (Phase 4's Google-Maps pattern). */
export type ViewMode = '2d' | '3d';

// The night air, twice - both skies are the fog token and nothing else
// (the no-hand-hex rule holds; the blend ratios are the only hand-set
// numbers, like the retint's). Mercator carries the fog matrix (the
// full night fog the engine supports there); the globe does not - its
// fog matrix is a documented no-op (ADR 0007 addendum) - so the globe
// sky pays for it in atmosphere instead. The console warning becomes
// a designed trade: each mode gets the sky its projection can render.
function skyFor(view: ViewMode): NonNullable<StyleSpecification['sky']> {
  const blend =
    view === '2d'
      ? (['interpolate', ['linear'], ['zoom'], 0, 0.7, 6, 0.35, 12, 0] as never)
      : (['interpolate', ['linear'], ['zoom'], 0, 0.9, 6, 0.55, 12, 0] as never);
  return {
    'sky-color': relief.fog,
    'horizon-color': relief.fog,
    'fog-color': relief.fog,
    'fog-ground-blend': 0.55,
    'horizon-fog-blend': 0.35,
    'atmosphere-blend': blend,
  };
}

/** The hillshade softens when the terrain is flat: in 2D the shading
 *  is structure, not relief drama. */
export function hillshadeExaggeration(view: ViewMode): number {
  return view === '2d' ? 0.25 : 0.45;
}

export function buildStyle({ basemap, view = '3d' }: StyleInput): StyleSpecification {
  const base: LayerSpecification[] = basemap
    ? protomapsLayers('pmt', namedTheme('dark'), {}).map(retint)
    : [];

  const style: StyleSpecification = {
    version: 8,
    sources: {
      ...(basemap
        ? {
            pmt: {
              type: 'vector',
              url: 'pmtiles:///geodata/thailand.pmtiles',
              attribution: '© ผู้ร่วมแก้ไข OpenStreetMap · Protomaps',
            },
          }
        : {}),
      dem: {
        type: 'raster-dem',
        tiles: ['/geodata/dem/{z}/{x}/{y}.png'],
        encoding: 'terrarium',
        tileSize: 256,
        minzoom: 4,
        maxzoom: 8,
        attribution: '',
      },
      borders: { type: 'geojson', data: { type: 'FeatureCollection', features: [] } },
      rivers: { type: 'geojson', data: { type: 'FeatureCollection', features: [] } },
    },
    layers: [
      {
        id: 'night',
        type: 'background',
        paint: { 'background-color': colors.ground },
      },
      ...base,
      {
        // The relief: hillshade over everything the basemap draws, so
        // roads and landcover are shaded by the same light. The sea is
        // flat and calm by construction - elevation 0 shades to nothing.
        // Shadow is the sea's dark; the highlight is the silt albedo,
        // so lit ridges read as moonlit sediment, the world's material.
        id: 'hillshade',
        type: 'hillshade',
        source: 'dem',
        paint: {
          'hillshade-exaggeration': hillshadeExaggeration(view),
          'hillshade-shadow-color': colors.sea,
          'hillshade-highlight-color': relief.albedoLow,
          'hillshade-accent-color': colors.sea,
        },
      },
      {
        id: 'rivers',
        type: 'line',
        source: 'rivers',
        paint: {
          'line-color': colors.accentWater,
          'line-width': ['interpolate', ['linear'], ['zoom'], 4, 0.5, 8, 1.2, 12, 2.5],
          'line-opacity': 0.7,
        },
        layout: { 'line-cap': 'round', 'line-join': 'round' },
      },
      {
        id: 'borders',
        type: 'line',
        source: 'borders',
        paint: {
          'line-color': colors.borderStrong,
          'line-width': ['interpolate', ['linear'], ['zoom'], 4, 0.7, 8, 1.4, 12, 2],
        },
      },
    ],
    sky: skyFor(view),
  };
  return style;
}

/** The sky for the other mode, applied at runtime on a view switch
 *  (setSky, not setStyle - a style swap would nuke the layers). */
export { skyFor };
