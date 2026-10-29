// The camera presets, kept out of the 3D chunk so the page shell can
// name them without pulling three.js in.
import { lonLatToWorld } from './projection';

export interface Preset {
  key: string;
  label: string;
  lon: number;
  lat: number;
  /** Camera offset from the target, world km: x east, y up, z south. */
  offset: [number, number, number];
  /** Fixture station id to select with this preset. */
  station?: number;
}

// Named so callers never index the array (the lint config forbids
// non-null assertions, and this reads better anyway).
export const NATIONAL: Preset = {
  key: 'national',
  label: 'ทั้งประเทศ',
  lon: 100.9,
  lat: 15.4,
  offset: [-560, 1420, 1330],
};
export const PLAIN: Preset = {
  key: 'plain',
  label: 'ลุ่มเจ้าพระยา',
  lon: 100.35,
  lat: 15.15,
  offset: [-170, 430, 350],
};
export const STATION: Preset = {
  key: 'station',
  label: 'สองพี่น้อง (น้ำล้นตลิ่ง)',
  lon: 100.05017,
  lat: 14.21608,
  offset: [-52, 108, 92],
  station: 736, // THA010 in the fixture
};

export const PRESETS: Preset[] = [NATIONAL, PLAIN, STATION];

export function targetFor(preset: Preset): { x: number; y: number; z: number } {
  const { x, z } = lonLatToWorld(preset.lon, preset.lat);
  return { x, y: 0, z };
}

export function positionFor(preset: Preset): { x: number; y: number; z: number } {
  const t = targetFor(preset);
  return { x: t.x + preset.offset[0], y: t.y + preset.offset[1], z: t.z + preset.offset[2] };
}
