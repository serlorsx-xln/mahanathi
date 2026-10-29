// The draped lines: the Thai border as a fine, precise line, and the
// 13 rivers with width by river class. Both are lifted a fixed
// distance above the sampled surface; the terrain's vertical
// exaggeration is applied to the samples, so the lines follow the
// relief at every exaggeration.
import { useMemo } from 'react';
import { Line } from '@react-three/drei';
import * as THREE from 'three';
import { feature } from 'topojson-client';
import type { Heightfield } from './heightfield';
import { lonLatToUv, lonLatToWorld, sampleMetres } from './projection';
import { colors } from '../design/tokens';

type LonLat = [number, number];

/** Lift above the surface in world km, scaled with the exaggeration so
 *  steep terrain never swallows the lines. */
const LIFT_KM = 3;

function drape(heightfield: Heightfield, points: LonLat[], exaggeration: number): THREE.Vector3[] {
  const lift = LIFT_KM * (0.6 + exaggeration * 0.04);
  return points.map(([lon, lat]) => {
    const { u, v } = lonLatToUv(lon, lat);
    const ground = sampleMetres(heightfield.png, u, v);
    const { x, z } = lonLatToWorld(lon, lat);
    return new THREE.Vector3(x, (ground / 1000) * exaggeration + lift, z);
  });
}

// The 13 shipped rivers, with the Mekong and the Chao Phraya widest
// and the rest by their place in the network (docs/research/
// geodata-sources.md, decision 7).
const RIVER_WIDTH: Record<number, number> = {
  215354: 3.0, // Mekong
  227210: 2.4, // Chao Phraya
  227319: 1.9, // Ping
  392516: 1.9, // Chi
  392517: 1.9, // Mun
  228063: 1.6, // Tha Chin
  229599: 1.6, // Pa Sak
  14030694: 1.6, // Yom
  19264378: 1.5, // Bang Pakong
  227317: 1.4, // Nan
  2676664: 1.4, // Tapi
  11700349: 1.4, // Wang
  12911358: 1.4, // Mae Klong
};

export function BorderRivers({
  heightfield,
  bordersTopo,
  riversGeo,
  exaggeration,
}: {
  heightfield: Heightfield;
  bordersTopo: unknown;
  riversGeo: GeoJSON.FeatureCollection;
  exaggeration: number;
}) {
  const borderLines = useMemo(() => {
    const topo = bordersTopo as {
      objects: Record<string, { type: string; geometries: Record<string, unknown>[] }>;
    };
    const entry = Object.entries(topo.objects)[0];
    if (entry === undefined) throw new Error('the topology has no objects');
    const [name, object] = entry;
    // Only the lines that touch Thailand; drawn like every other land
    // border (ADR 0004: one line style, no disputed-area emphasis).
    const thai = object.geometries.filter((g) => {
      const p = (g.properties ?? {}) as { ADM0_A3_L?: string; ADM0_A3_R?: string };
      return p.ADM0_A3_L === 'THA' || p.ADM0_A3_R === 'THA';
    });
    const geo = feature(
      { ...topo, objects: { [name]: { type: 'GeometryCollection', geometries: thai } } } as never,
      { type: 'GeometryCollection', geometries: thai } as never,
    ) as unknown as GeoJSON.FeatureCollection;
    const lines: LonLat[][] = [];
    for (const f of geo.features) {
      const g = f.geometry;
      if (g.type === 'MultiLineString') lines.push(...(g.coordinates as LonLat[][]));
      else if (g.type === 'LineString') lines.push(g.coordinates as LonLat[]);
    }
    return lines.map((line) => drape(heightfield, line, exaggeration));
  }, [heightfield, bordersTopo, exaggeration]);

  const riverLines = useMemo(() => {
    return riversGeo.features.map((f) => {
      const osm = f.properties?.osm as number;
      const geoms =
        f.geometry.type === 'MultiLineString'
          ? (f.geometry.coordinates as LonLat[][])
          : f.geometry.type === 'LineString'
            ? [f.geometry.coordinates as LonLat[]]
            : [];
      return {
        width: RIVER_WIDTH[osm] ?? 1.3,
        points: geoms.map((line) => drape(heightfield, line, exaggeration)),
      };
    });
  }, [heightfield, riversGeo, exaggeration]);

  return (
    <group>
      {borderLines.map((points, i) => (
        <Line
          key={`border-${i}`}
          points={points}
          color={colors.borderStrong}
          lineWidth={1.1}
          transparent
          opacity={0.9}
        />
      ))}
      {riverLines.flatMap(({ points, width }, fi) =>
        points.map((p, i) => (
          <Line
            key={`river-${fi}-${i}`}
            points={p}
            color={colors.accentWater}
            lineWidth={width}
            transparent
            opacity={0.55}
          />
        )),
      )}
    </group>
  );
}
