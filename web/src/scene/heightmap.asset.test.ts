/// <reference types="node" />
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';
import { decodeGray16Png } from './png';
import { lonLatToUv, sampleMetres } from './projection';

// The real asset, decoded the way the browser decodes it, checked
// against geography anyone can verify on a map. If the pipeline, the
// PNG decoder or the projection drifts, one of these heights is wrong.
const ASSET = resolve(import.meta.dirname, '../../../geodata/assets/heightmap.png');

async function loadAsset() {
  const bytes = new Uint8Array(readFileSync(ASSET));
  return decodeGray16Png(bytes);
}

describe('the shipped heightmap', () => {
  it('has the documented shape', async () => {
    const png = await loadAsset();
    expect(png.width).toBe(1815);
    expect(png.height).toBe(2048);
    // Sea is reserved code 0; land tops out at 1 + 2046.
    let max = 0;
    let sea = 0;
    for (const v of png.data) {
      if (v > max) max = v;
      if (v === 0) sea++;
    }
    expect(max).toBeLessThanOrEqual(1 + 2046);
    // Thailand tops out at 2,565 m (code ~1804); Fansipan just inside
    // the bbox is 3,143 m, blurred and coarse-quantised as a neighbour.
    expect(max).toBeGreaterThan(1700);
    expect(sea).toBeGreaterThan(0);
  });

  it('knows the ground under five known places', async () => {
    const png = await loadAsset();
    const at = (lon: number, lat: number) => {
      const { u, v } = lonLatToUv(lon, lat);
      return sampleMetres(png, u, v);
    };
    // The lower Chao Phraya plain: single-digit to low-double metres.
    expect(at(100.56, 14.35)).toBeLessThan(30); // Ayutthaya, in the plain
    // Doi Inthanon, the country's highest point (2,565 m).
    expect(at(98.4883, 18.5883)).toBeGreaterThan(2200);
    // The Khorat plateau: a few hundred metres.
    const korat = at(102.8, 14.9);
    expect(korat).toBeGreaterThan(120);
    expect(korat).toBeLessThan(350);
    // The Gulf of Thailand is sea: exactly zero.
    expect(at(101.5, 9.0)).toBe(0);
    expect(at(100.9, 7.5)).toBe(0);
    // The Andaman Sea, west of the peninsula.
    expect(at(96.5, 8.5)).toBe(0);
    // The Mekong at Nong Khai: the river valley is low but not sea.
    expect(at(102.75, 17.9)).toBeGreaterThan(150);
  });
});
