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
import { colors, relief } from '../design/tokens';

/** Retint a protomaps layer's paint to the token world. */
function retint(layer: LayerSpecification): LayerSpecification {
  if (layer.type === 'background') {
    return { ...layer, paint: { 'background-color': colors.ground } };
  }
  if (layer.type === 'fill' && layer.id.startsWith('water')) {
    return { ...layer, paint: { ...layer.paint, 'fill-color': colors.sea } };
  }
  return layer;
}

export interface StyleInput {
  /** The PMTiles basemap is present (HEAD /geodata/thailand.pmtiles
   *  answered 200); without it we ship terrain, borders and rivers
   *  only, which is still the committed world. */
  basemap: boolean;
}

export function buildStyle({ basemap }: StyleInput): StyleSpecification {
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
          'hillshade-exaggeration': 0.45,
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
    sky: {
      // The night air: our fog token is the sky's horizon and the fog
      // the far terrain sinks into (brief 04: one continuous fog).
      'sky-color': relief.fog,
      'horizon-color': relief.fog,
      'fog-color': relief.fog,
      'fog-ground-blend': 0.55,
      'horizon-fog-blend': 0.35,
      'atmosphere-blend': ['interpolate', ['linear'], ['zoom'], 0, 0.8, 6, 0.4, 12, 0],
    },
  };
  return style;
}
