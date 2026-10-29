// Camera presets for the MapLibre rig (ADR 0007). The national view is
// a bounds fit, computed against the real viewport, not a fixed zoom;
// the close-ups are hand-set. Same three named views as the hand-built
// relief so the DOM contracts (and the e2e) carry over.

export interface Preset {
  key: string;
  label: string;
  /** Fixed view for the close-ups. */
  view?: {
    center: [number, number];
    zoom: number;
    pitch: number;
    bearing: number;
  };
  /** Bounds fit for the national view. */
  bounds?: [[number, number], [number, number]];
  /** Fixture station id to select with this preset. */
  station?: number;
}

export const NATIONAL: Preset = {
  key: 'national',
  label: 'ทั้งประเทศ',
  bounds: [
    [97.3, 5.6],
    [105.8, 20.5],
  ],
};

export const PLAIN: Preset = {
  key: 'plain',
  label: 'ลุ่มเจ้าพระยา',
  view: { center: [100.35, 15.15], zoom: 8.4, pitch: 55, bearing: -18 },
};

export const STATION: Preset = {
  key: 'station',
  label: 'สองพี่น้อง (น้ำล้นตลิ่ง)',
  view: { center: [100.05017, 14.21608], zoom: 12.2, pitch: 60, bearing: -24 },
  station: 736, // THA010 in the fixture
};

export const PRESETS: Preset[] = [NATIONAL, PLAIN, STATION];
