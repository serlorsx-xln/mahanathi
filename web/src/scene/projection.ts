// The one projection the scene, the CPU sampler and the geodata
// pipeline share (ADR 0004). Coordinates: world X and Z are
// kilometres on the 13 N tangent plane, Y is height in kilometres
// after exaggeration. The raster is addressed by u (column fraction,
// west to east) and v (row fraction, north to south), exactly as the
// heightmap PNG is laid out.
//
// The constants are the ADR's decision of record. The geodata
// pipeline's raster aspect used a linear 110.736 km per degree of
// latitude, 0.10% above the meridian value here; the difference moves
// nothing but world scale and is recorded in the research doc.

export const BBOX = { west: 92, east: 111, south: 3, north: 24 } as const;
export const KM_PER_DEG_LON = 108.4847; // 111.320 * cos(13 deg)
export const KM_PER_DEG_LAT = 110.6305; // meridian arc at 13 deg
export const LON0 = (BBOX.west + BBOX.east) / 2; // 101.5
export const LAT0 = (BBOX.south + BBOX.north) / 2; // 13.5

/** Heightmap code range: sea is 0, land is 1 + round(sqrt(h/HM) * 2046). */
export const HM = 3300; // metres, the encoding ceiling (geodata/SOURCES.md)
export const TOP = 2046; // (1 << 11) - 2

/** Decode a heightmap code to metres above the sea flatten. */
export function codeToMetres(code: number): number {
  if (code <= 0) return 0;
  const t = (code - 1) / TOP;
  return t * t * HM;
}

/** Longitude/latitude to texture fractions (u columns, v rows). */
export function lonLatToUv(lon: number, lat: number): { u: number; v: number } {
  return {
    u: (lon - BBOX.west) / (BBOX.east - BBOX.west),
    v: (BBOX.north - lat) / (BBOX.north - BBOX.south),
  };
}

/** Longitude/latitude to world kilometres on the ground plane. */
export function lonLatToWorld(lon: number, lat: number): { x: number; z: number } {
  return {
    x: (lon - LON0) * KM_PER_DEG_LON,
    z: (LAT0 - lat) * KM_PER_DEG_LAT,
  };
}

/** World kilometres back to longitude/latitude (camera bounds, tests). */
export function worldToLonLat(x: number, z: number): { lon: number; lat: number } {
  return {
    lon: x / KM_PER_DEG_LON + LON0,
    lat: LAT0 - z / KM_PER_DEG_LAT,
  };
}

/**
 * Bilinear height sample in metres from the decoded heightmap codes.
 * u grows to the east, v to the south, matching the PNG layout.
 */
export function sampleMetres(
  codes: { width: number; height: number; data: Uint16Array },
  u: number,
  v: number,
): number {
  const x = Math.min(Math.max(u, 0), 1) * (codes.width - 1);
  const y = Math.min(Math.max(v, 0), 1) * (codes.height - 1);
  const x0 = Math.floor(x);
  const y0 = Math.floor(y);
  const x1 = Math.min(x0 + 1, codes.width - 1);
  const y1 = Math.min(y0 + 1, codes.height - 1);
  const fx = x - x0;
  const fy = y - y0;
  const m00 = codeToMetres(codes.data[y0 * codes.width + x0] ?? 0);
  const m10 = codeToMetres(codes.data[y0 * codes.width + x1] ?? 0);
  const m01 = codeToMetres(codes.data[y1 * codes.width + x0] ?? 0);
  const m11 = codeToMetres(codes.data[y1 * codes.width + x1] ?? 0);
  return m00 * (1 - fx) * (1 - fy) + m10 * fx * (1 - fy) + m01 * (1 - fx) * fy + m11 * fx * fy;
}
