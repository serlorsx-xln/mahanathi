// The gauges as a MapLibre custom layer (ADR 0007, brief 04 step 2).
// One instanced draw call for all 804 stations: each gauge is a pair of
// crossed slim blades seated on the terrain - a quiet slot with a
// bank-full cap, the water fill inside it, and above the cap the fill
// rises in the alert colour. The encoding is geometric before it is
// coloured, so the level survives colour-blind vision and monochrome
// printing alike.
//
// Positioning uses MapLibre's own `projectTileFor3D` prelude (mercator
// xy + elevation in metres), which is globe- and mercator-correct, and
// `queryTerrainElevation` for ground contact. Instance buffers are
// rebuilt only when their inputs change - stations or exaggeration -
// never per frame (brief 04, bug 5); the camera-dependent quantities
// (blade width on screen, fog) are per-frame uniforms.
import {
  MercatorCoordinate,
  type Map as MLMap,
  type CustomLayerInterface,
  type CustomRenderMethodInput,
} from 'maplibre-gl';
import { colors, relief, situations, stale as staleToken } from '../design/tokens';
import { addDiscLayer } from './discs';
import type { StationState } from './stations';

const CAP = 0.78; // where the bank-full cap sits on the pole
const CAP_WIDTH = 0.045;
const BASE_RING = 0.06; // the darker ground-contact ring at the base
const BASE_DISCARD = 0.004; // no sliver at the foot of the blade
const POLE_WIDTH_M = 2600; // the blade's real width at close range
const OVERFLOW_SPAN = 0.7; // fill is clamped at 170%; the last 70% maps to CAP..1
const EARTH_R_M = 6371008.8;

/** Pole height by level: 4 and 5 prominent, 1-3 small and quiet. */
function poleHeightM(level: number | null): number {
  if (level === 5) return 44000;
  if (level === 4) return 34000;
  if (level === null) return 12000;
  return 17000;
}

function levelColour(level: number | null): [number, number, number] {
  const s = situations.find((x) => x.level === level);
  const hex = s ? s.hex : staleToken.hex;
  return [
    parseInt(hex.slice(1, 3), 16) / 255,
    parseInt(hex.slice(3, 5), 16) / 255,
    parseInt(hex.slice(5, 7), 16) / 255,
  ];
}

function hexToRgb(hex: string): [number, number, number] {
  return [
    parseInt(hex.slice(1, 3), 16) / 255,
    parseInt(hex.slice(3, 5), 16) / 255,
    parseInt(hex.slice(5, 7), 16) / 255,
  ];
}

/** The overflow level (5), by name rather than array index. */
const [, , , , OVERFLOW] = situations;

// Two crossed quads per blade, so the slim profile reads from any
// azimuth (a single flat quad is edge-on half the time). Local blade
// space: x/z on the ground (one of them ±0.5 per quad), y up the pole.
const QUAD_VERTS = new Float32Array([
  // quad in the east-west plane
  -0.5, 0, 0, 0.5, 0, 0, 0.5, 1, 0, -0.5, 1, 0,
  // quad in the north-south plane
  0, 0, -0.5, 0, 0, 0.5, 0, 1, 0.5, 0, 1, -0.5,
]);
const QUAD_INDEX = new Uint16Array([0, 1, 2, 0, 2, 3, 4, 5, 6, 4, 6, 7]);

// No #version here: the render() call site puts the projection prelude
// (which owns #version and the projection uniforms/functions) first,
// per MapLibre's custom-layer shader contract.
const VERT = `
in vec3 aLocal;
in vec2 aMerc;
in float aGround;
in float aHeight;
in float aWidth;
in float aFill;
in float aOverflow;
in vec3 aColour;
in float aStale;
in float aSuspect;

uniform float uWidthScale;
uniform vec3 uCamMerc;
uniform float uMercPerM;
uniform float uElevScale;

out float vFill;
out float vOverflow;
out vec3 vColour;
out float vStale;
out float vSuspect;
out float vLocalY;
out float vDistM;

void main() {
  vFill = aFill;
  vOverflow = aOverflow;
  vColour = aColour;
  vStale = aStale;
  vSuspect = aSuspect;
  vLocalY = aLocal.y;

  vec2 pos = aMerc + aLocal.xz * aWidth * uWidthScale;
  float elevM = aGround + 2.0 + aLocal.y * aHeight;
  // Globe projection takes elevation in metres; mercator takes
  // conformal mercator z. One uniform switches the unit.
  gl_Position = projectTileFor3D(pos, elevM * uElevScale);
  // Fog distance in the mercator+altitude space the camera reports;
  // over Thailand's span the flat-space error is a fraction of the
  // fog's own softness, which is what the approximation is for.
  vec3 worldPos = vec3(pos, elevM * uMercPerM);
  vDistM = distance(worldPos, uCamMerc) / uMercPerM;
}
`;

const FRAG = `#version 300 es
precision highp float;

uniform vec3 uSteel;
uniform vec3 uTrack;
uniform vec3 uAlert;
uniform vec3 uFog;
uniform float uFogStart;
uniform float uFogEnd;

in float vFill;
in float vOverflow;
in vec3 vColour;
in float vStale;
in float vSuspect;
in float vLocalY;
in float vDistM;

out vec4 fragColor;

void main() {
  // Water fill from the base; the cap at CAP marks bank-full. Above
  // the cap the blade exists only where the water does, and the last
  // 70% of the clamped fill maps exactly onto CAP..1 (brief 04, bug 9:
  // nothing was allowed to clip past the pole top).
  float waterTop = ${CAP} * min(vFill, 1.0)
    + vOverflow * clamp(vFill - 1.0, 0.0, ${OVERFLOW_SPAN}) * (1.0 - ${CAP}) / ${OVERFLOW_SPAN};
  // atCap is not stale-gated: the stale branch below runs first and
  // keeps only the cap stroke, so the gate here would have discarded
  // every fragment of a stale pole (the hollow cap never drew - the
  // lead's review caught the dead branch).
  bool atCap = abs(vLocalY - ${CAP}) < ${CAP_WIDTH};
  bool inWater = vLocalY <= waterTop && vLocalY > ${BASE_DISCARD};

  float alpha = 1.0;
  vec3 colour;
  if (vStale > 0.5 || (vSuspect > 0.5 && !inWater)) {
    // Hollow: the cap stroke only, never a fill - a stale or suspect
    // reading must not read as a level (bugs 8 and the stale rule).
    if (!atCap) discard;
    colour = uSteel * (vStale > 0.5 ? 0.75 : 0.9);
  } else if (atCap) {
    colour = uSteel;
  } else if (inWater) {
    colour = vLocalY > ${CAP} ? uAlert : vColour;
    if (vSuspect > 0.5) alpha = 0.45;
  } else if (vLocalY < ${CAP} && vLocalY > ${BASE_RING}) {
    colour = uTrack; // the empty slot above the water, below the cap
    if (vSuspect > 0.5) alpha = 0.45;
  } else if (vLocalY <= ${BASE_RING} && vLocalY > ${BASE_DISCARD}) {
    colour = uTrack * 0.55; // the ground-contact base ring, darker
  } else {
    discard;
  }

  float fog = smoothstep(uFogStart, uFogEnd, vDistM);
  colour = mix(colour, uFog, fog);
  // Premultiplied output: the map sets blendFunc(ONE, ONE_MINUS_SRC_ALPHA).
  fragColor = vec4(colour * alpha, alpha);
}
`;

interface InstanceRow {
  merc: [number, number];
  ground: number;
  height: number;
  width: number;
  fill: number;
  overflow: number;
  colour: [number, number, number];
  stale: number;
  suspect: number;
  lon: number;
  lat: number;
}

export class GaugesLayer implements CustomLayerInterface {
  readonly id = 'gauges';
  readonly type = 'custom' as const;
  readonly renderingMode = '3d' as const;

  private map: MLMap | null = null;
  private gl: WebGL2RenderingContext | null = null;
  private program: WebGLProgram | null = null;
  private vao: WebGLVertexArrayObject | null = null;
  private buffers: WebGLBuffer[] = [];
  private instBuffer: WebGLBuffer | null = null;
  private count = 0;
  private rows: InstanceRow[] = [];
  private idleBound: () => void;

  constructor() {
    // Ground contact follows the terrain tiles as they load at finer
    // zooms; idle is the only moment that is both settled and often
    // enough. It never runs per frame.
    this.idleBound = () => {
      this.syncGrounds();
    };
  }

  onAdd(map: MLMap, gl: WebGL2RenderingContext): void {
    this.map = map;
    this.gl = gl;

    // Geometry and instance buffers only. The program is compiled in
    // render(), where the projection prelude (which defines
    // projectTileFor3D) is available - it is not, at onAdd time.
    const local = gl.createBuffer();
    const index = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, local);
    gl.bufferData(gl.ARRAY_BUFFER, QUAD_VERTS, gl.STATIC_DRAW);
    this.buffers.push(local);

    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, index);
    gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, QUAD_INDEX, gl.STATIC_DRAW);
    this.buffers.push(index);

    this.instBuffer = gl.createBuffer();
    // Stations may have been set before the map finished loading; the
    // rows are already there, only the upload was impossible.
    if (this.rows.length) this.upload();
    map.on('idle', this.idleBound);
  }

  /** Emergency exit: replace this layer with the disc encoding. The
   *  poles are the signature, but a dead frame loop is worse than flat
   *  circles - and the discs were built as the guaranteed fallback. */
  private degradeToDiscs(e: unknown): void {
    console.error('gauge shader failed, falling back to discs', e);
    const map = this.map;
    if (!map) return;
    const stations = this.lastStations;
    // Swapping layers from inside render() would mutate the layer list
    // mid-iteration; the swap happens after the frame.
    window.setTimeout(() => {
      if (!this.map) return;
      this.map.removeLayer(this.id);
      addDiscLayer(this.map, stations);
    }, 0);
  }

  onRemove(map: MLMap): void {
    map.off('idle', this.idleBound);
    const gl = this.gl;
    if (!gl) return;
    if (this.vao) gl.deleteVertexArray(this.vao);
    for (const b of this.buffers) gl.deleteBuffer(b);
    if (this.instBuffer) gl.deleteBuffer(this.instBuffer);
    if (this.program) gl.deleteProgram(this.program);
    this.vao = null;
    this.buffers = [];
    this.instBuffer = null;
    this.program = null;
    this.map = null;
    this.gl = null;
  }

  /** The raw stations, kept so a shader failure can degrade to the
   *  disc layer without the caller's help. */
  private lastStations: StationState[] = [];

  /** Rebuild the instances. Called when stations or exaggeration change
   *  - the inputs the encoding depends on - and never per frame. */
  setStations(stations: StationState[]): void {
    this.lastStations = stations;
    const rows: InstanceRow[] = stations.map((s) => {
      const mc = MercatorCoordinate.fromLngLat({ lng: s.lon, lat: s.lat });
      const pct = s.pct ?? 0;
      return {
        merc: [mc.x, mc.y],
        ground: this.map?.queryTerrainElevation([s.lon, s.lat]) ?? 0,
        height: poleHeightM(s.level),
        width: POLE_WIDTH_M * mc.meterInMercatorCoordinateUnits(),
        fill: Math.max(0, Math.min(pct, 100 + OVERFLOW_SPAN * 100)) / 100,
        overflow: pct > 100 ? 1 : 0,
        colour: levelColour(s.stale ? null : s.level),
        stale: s.stale ? 1 : 0,
        suspect: s.suspect?.length ? 1 : 0,
        lon: s.lon,
        lat: s.lat,
      };
    });
    this.rows = rows;
    this.upload();
    this.map?.triggerRepaint();
  }

  /** Re-query the ground elevations (the terrain refines as tiles load;
   *  the values also change with exaggeration, which the caller covers
   *  by calling setStations again). */
  private syncGrounds(): void {
    const map = this.map;
    if (!map || this.rows.length === 0) return;
    let changed = false;
    for (const r of this.rows) {
      const g = map.queryTerrainElevation([r.lon, r.lat]);
      if (g !== null && Math.abs(g - r.ground) > 0.5) {
        r.ground = g;
        changed = true;
      }
    }
    if (changed) {
      this.upload();
      map.triggerRepaint();
    }
  }

  private upload(): void {
    const gl = this.gl;
    if (!gl || !this.instBuffer) return;
    const data = new Float32Array(this.rows.length * 12);
    this.rows.forEach((r, i) => {
      const o = i * 12;
      data[o] = r.merc[0];
      data[o + 1] = r.merc[1];
      data[o + 2] = r.ground;
      data[o + 3] = r.height;
      data[o + 4] = r.width;
      data[o + 5] = r.fill;
      data[o + 6] = r.overflow;
      data[o + 7] = r.colour[0];
      data[o + 8] = r.colour[1];
      data[o + 9] = r.colour[2];
      data[o + 10] = r.stale;
      data[o + 11] = r.suspect;
    });
    gl.bindBuffer(gl.ARRAY_BUFFER, this.instBuffer);
    gl.bufferData(gl.ARRAY_BUFFER, data, gl.DYNAMIC_DRAW);
    this.count = this.rows.length;
  }

  render(gl: WebGL2RenderingContext, options: CustomRenderMethodInput): void {
    if (this.count === 0) return;
    const map = this.map;
    if (!map) return;

    // The prelude changes with the projection; the program carries the
    // prelude, so it is compiled per projection variant (and once). A
    // throw here would abort MapLibre's whole frame loop, so a compile
    // failure degrades to the disc layer instead (the guaranteed
    // fallback the brief asked for, reached by its emergency exit).
    const variant = options.shaderData.variantName;
    if (variant !== this.compiledVariant) {
      const vert = `#version 300 es\n${options.shaderData.vertexShaderPrelude}\n${options.shaderData.define}\n${VERT}`;
      try {
        this.recompile(gl, vert, variant);
      } catch (e) {
        this.degradeToDiscs(e);
        return;
      }
    }
    const prog = this.program;
    if (!prog || !this.vao) return;
    gl.useProgram(prog);

    // Per-frame uniforms only - no buffer is touched while the camera
    // moves (bug 5). Screen-constant blade width, clamped both ways:
    // never thinner than ~3.5 px (below that the blades dissolve into
    // the anti-aliasing at the country view, measured), never wider
    // than ~36 px (above that the landing-zoom blade becomes a
    // monolithic slab - the gate critique measured 24% of the canvas
    // as alert red before this cap).
    const zoom = map.getZoom();
    const pitch = (map.getPitch() * Math.PI) / 180;
    const bearing = (map.getBearing() * Math.PI) / 180;
    const center = map.getCenter();
    const centerMc = MercatorCoordinate.fromLngLat(center, 0);
    const metersPerPx = (40075016.7 * Math.cos((center.lat * Math.PI) / 180)) / (512 * Math.pow(2, zoom));
    const widthM = Math.min(
      Math.max(POLE_WIDTH_M, 3.5 * metersPerPx),
      36 * metersPerPx,
    );

    // The camera position, derived from the public camera state (there
    // is no free-camera getter on Map in v6): straight down at pitch 0,
    // swung back and up as it pitches, bearing the facing direction.
    const mercPerM = centerMc.meterInMercatorCoordinateUnits();
    const dMerc =
      0.5 * gl.drawingBufferHeight / Math.tan(options.fov / 2) / (512 * Math.pow(2, zoom));
    const camMerc: [number, number, number] = [
      centerMc.x - dMerc * Math.sin(pitch) * Math.sin(bearing),
      centerMc.y + dMerc * Math.sin(pitch) * Math.cos(bearing),
      dMerc * Math.cos(pitch),
    ];
    const altM = Math.max(camMerc[2] / mercPerM, 1);
    // One continuous fog (brief 04): the same token the sky and horizon
    // use, reached near the horizon distance for this altitude.
    const fogEnd = Math.sqrt(2 * EARTH_R_M * altM);
    // The fog keeps one continuous token colour, but starts at 55% of
    // the horizon distance: at the country view most of Thailand sits
    // inside that radius, so the gauge field keeps its colour instead
    // of dissolving into the fog the critique measured.
    const fogStart = fogEnd * 0.55;

    const u = (name: string): WebGLUniformLocation | null =>
      gl.getUniformLocation(prog, name);

    // MapLibre's projection prelude declares its uniforms but does not
    // set them in our program - that is the custom layer's job, every
    // frame (the matrices move with the camera). Without these the
    // shader computes gl_Position from zeroed matrices and nothing
    // renders anywhere.
    const pd = options.defaultProjectionData;
    gl.uniformMatrix4fv(u('u_projection_matrix'), false, new Float32Array(pd.mainMatrix));
    gl.uniformMatrix4fv(
      u('u_projection_fallback_matrix'),
      false,
      new Float32Array(pd.fallbackMatrix),
    );
    gl.uniform4fv(u('u_projection_tile_mercator_coords'), pd.tileMercatorCoords);
    gl.uniform4fv(u('u_projection_clipping_plane'), pd.clippingPlane);
    gl.uniform1f(u('u_projection_transition'), pd.projectionTransition);
    gl.uniform1f(u('u_projection_clip_antimeridian'), pd.clipAntimeridian ? 1 : 0);
    gl.uniform1f(u('uWidthScale'), widthM / POLE_WIDTH_M);
    gl.uniform3f(u('uCamMerc'), camMerc[0], camMerc[1], camMerc[2]);
    gl.uniform1f(u('uMercPerM'), mercPerM);
    gl.uniform1f(
      u('uElevScale'),
      options.shaderData.variantName === 'globe' ? 1 : mercPerM,
    );
    gl.uniform3fv(u('uSteel'), hexToRgb(colors.gaugeSteel));
    gl.uniform3fv(u('uTrack'), hexToRgb(colors.gaugeTrack));
    gl.uniform3fv(u('uAlert'), hexToRgb(OVERFLOW.hex));
    gl.uniform3fv(u('uFog'), hexToRgb(relief.fog));
    gl.uniform1f(u('uFogStart'), fogStart);
    gl.uniform1f(u('uFogEnd'), fogEnd);

    gl.bindVertexArray(this.vao);
    gl.drawElementsInstanced(gl.TRIANGLES, QUAD_INDEX.length, gl.UNSIGNED_SHORT, 0, this.count);
    gl.bindVertexArray(null);
  }

  private compiledVariant = '';

  /** Compile the program with the current projection prelude, and bind
   *  a fresh VAO against it. Rare: once per projection variant. */
  private recompile(gl: WebGL2RenderingContext, vert: string, variant: string): void {
    const compile = (kind: number, src: string): WebGLShader => {
      const sh = gl.createShader(kind);
      if (!sh) throw new Error(`gauge shader (${variant}): allocation failed`);
      gl.shaderSource(sh, src);
      gl.compileShader(sh);
      if (!gl.getShaderParameter(sh, gl.COMPILE_STATUS)) {
        throw new Error(`gauge shader (${variant}): ${gl.getShaderInfoLog(sh)}`);
      }
      return sh;
    };
    const prog = gl.createProgram();
    gl.attachShader(prog, compile(gl.VERTEX_SHADER, vert));
    gl.attachShader(prog, compile(gl.FRAGMENT_SHADER, FRAG));
    gl.linkProgram(prog);
    if (!gl.getProgramParameter(prog, gl.LINK_STATUS)) {
      throw new Error(`gauge program: ${gl.getProgramInfoLog(prog)}`);
    }
    if (this.program) gl.deleteProgram(this.program);
    if (this.vao) gl.deleteVertexArray(this.vao);
    this.program = prog;
    this.compiledVariant = variant;

    this.vao = gl.createVertexArray();
    gl.bindVertexArray(this.vao);
    this.bindAttributes(gl, prog);
    gl.bindVertexArray(null);
  }

  private bindAttributes(gl: WebGL2RenderingContext, prog: WebGLProgram): void {
    const [local, index] = this.buffers;
    if (!local || !index || !this.instBuffer) return;
    gl.bindBuffer(gl.ARRAY_BUFFER, local);
    const aLocal = gl.getAttribLocation(prog, 'aLocal');
    gl.enableVertexAttribArray(aLocal);
    gl.vertexAttribPointer(aLocal, 3, gl.FLOAT, false, 0, 0);
    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, index);

    gl.bindBuffer(gl.ARRAY_BUFFER, this.instBuffer);
    const stride = 12 * 4;
    const layout: [number, number, number][] = [
      [gl.getAttribLocation(prog, 'aMerc'), 2, 0],
      [gl.getAttribLocation(prog, 'aGround'), 1, 2],
      [gl.getAttribLocation(prog, 'aHeight'), 1, 3],
      [gl.getAttribLocation(prog, 'aWidth'), 1, 4],
      [gl.getAttribLocation(prog, 'aFill'), 1, 5],
      [gl.getAttribLocation(prog, 'aOverflow'), 1, 6],
      [gl.getAttribLocation(prog, 'aColour'), 3, 7],
      [gl.getAttribLocation(prog, 'aStale'), 1, 10],
      [gl.getAttribLocation(prog, 'aSuspect'), 1, 11],
    ];
    for (const [loc, size, off] of layout) {
      if (loc < 0) throw new Error(`gauge attribute not found at offset ${off}`);
      gl.enableVertexAttribArray(loc);
      gl.vertexAttribPointer(loc, size, gl.FLOAT, false, stride, off * 4);
      gl.vertexAttribDivisor(loc, 1);
    }
    gl.bindVertexArray(null);
  }
}
