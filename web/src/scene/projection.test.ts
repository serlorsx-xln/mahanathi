import { describe, expect, it } from 'vitest';
import {
  BBOX,
  HM,
  TOP,
  codeToMetres,
  lonLatToUv,
  lonLatToWorld,
  sampleMetres,
  worldToLonLat,
} from './projection';

describe('codeToMetres', () => {
  it('maps the reserved sea code and the land range endpoints', () => {
    expect(codeToMetres(0)).toBe(0);
    expect(codeToMetres(1)).toBeCloseTo(0, 6);
    // 1 + 2046 is the top code: full scale.
    expect(codeToMetres(1 + TOP)).toBeCloseTo(HM, 6);
  });

  it('is the inverse of the pipeline encoding', () => {
    // The pipeline writes 1 + round(sqrt(h / HM) * 2046).
    for (const h of [0.5, 12, 137, 1567.3, HM - 0.7]) {
      const code = 1 + Math.round(Math.sqrt(h / HM) * TOP);
      const back = codeToMetres(code);
      // Round-trip error is at most half a code step at that height.
      const oneStepUp = codeToMetres(code + 1);
      expect(back).toBeGreaterThanOrEqual(h - (oneStepUp - back));
      expect(back).toBeLessThanOrEqual(h + (oneStepUp - back));
    }
  });
});

describe('uv and world projections', () => {
  it('corners round-trip exactly', () => {
    const corners = [
      [BBOX.west, BBOX.north],
      [BBOX.east, BBOX.north],
      [BBOX.west, BBOX.south],
      [BBOX.east, BBOX.south],
    ];
    for (const [lon, lat] of corners as [number, number][]) {
      const { x, z } = lonLatToWorld(lon, lat);
      const back = worldToLonLat(x, z);
      expect(back.lon).toBeCloseTo(lon, 9);
      expect(back.lat).toBeCloseTo(lat, 9);
    }
  });

  it('u grows east and v grows south, as the PNG is laid out', () => {
    const nw = lonLatToUv(100, 20);
    const se = lonLatToUv(105, 10);
    expect(nw.u).toBeLessThan(se.u);
    expect(nw.v).toBeLessThan(se.v);
  });

  it('world z grows south, x grows east, and the origin is the bbox centre', () => {
    const centre = lonLatToWorld(101.5, 13.5);
    expect(centre.x).toBeCloseTo(0, 9);
    expect(centre.z).toBeCloseTo(0, 9);
    const south = lonLatToWorld(101.5, 12);
    expect(south.z).toBeGreaterThan(0);
    const east = lonLatToWorld(103, 13.5);
    expect(east.x).toBeGreaterThan(0);
  });

  it('the world extent matches the raster aspect', () => {
    const ne = lonLatToWorld(111, 24);
    const sw = lonLatToWorld(92, 3);
    const widthKm = ne.x - sw.x;
    const heightKm = sw.z - ne.z;
    // 19 deg of longitude against 21 of latitude, at the 13 N scales.
    expect(widthKm).toBeCloseTo(19 * 108.4847, 3);
    expect(heightKm).toBeCloseTo(21 * 110.6305, 3);
    // The raster is 1815 x 2048 px on those same extents, so its
    // aspect ratio and the world's agree to the pipeline's 0.1%.
    expect(widthKm / heightKm).toBeCloseTo(1815 / 2048, 2);
  });
});

describe('sampleMetres', () => {
  const codes = {
    width: 2,
    height: 2,
    data: Uint16Array.from([0, 1 + 1023, 1 + 511, 1 + 2045]),
  };

  it('samples the four corners exactly', () => {
    expect(sampleMetres(codes, 0, 0)).toBeCloseTo(0, 5);
    expect(sampleMetres(codes, 1, 0)).toBeCloseTo(codeToMetres(1 + 1023), 5);
    expect(sampleMetres(codes, 0, 1)).toBeCloseTo(codeToMetres(1 + 511), 5);
    expect(sampleMetres(codes, 1, 1)).toBeCloseTo(codeToMetres(1 + 2045), 5);
  });

  it('interpolates at the centre', () => {
    const tl = codeToMetres(1 + 1023);
    const bl = codeToMetres(1 + 511);
    const br = codeToMetres(1 + 2045);
    const expected = (0 + tl + bl + br) / 4;
    expect(sampleMetres(codes, 0.5, 0.5)).toBeCloseTo(expected, 5);
  });

  it('clamps outside the raster instead of wrapping', () => {
    expect(sampleMetres(codes, -0.5, 0)).toBeCloseTo(sampleMetres(codes, 0, 0), 5);
    expect(sampleMetres(codes, 1.5, 1)).toBeCloseTo(sampleMetres(codes, 1, 1), 5);
  });
});
