// The station record the fixture carries and every map layer consumes.
// One definition, so the disc layer (tomorrow), the gauge layer and
// the station sheet agree on what a station is.
export interface StationState {
  id: number;
  code: string;
  name: string;
  lat: number;
  lon: number;
  province: string | null;
  river: string | null;
  wl: number | null;
  bankFull: number | null;
  ground: number | null;
  pct: number | null;
  level: number | null;
  stale: boolean;
  /** The data-quality flags the fixture records (e.g. ["jump"]); a
   *  suspect reading must never render as fully trusted (brief 04,
   *  bug 8). */
  suspect?: string[];
  sentence: string | null;
  measuredAt: string | null;
}
