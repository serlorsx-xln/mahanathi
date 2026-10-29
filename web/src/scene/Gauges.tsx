// The gauges: one InstancedMesh, one draw call (ADR 0004). Each gauge
// is a slim vertical bar seated on the terrain: a quiet slot with a
// bank-full cap, the water fill inside it, and - when the river is
// over its banks - a fill that rises above the cap in the alert
// colour. The encoding is geometric before it is coloured, so the
// level survives colour-blind vision and monochrome printing alike.
import { useMemo, useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import type { Heightfield } from './heightfield';
import { lonLatToUv, lonLatToWorld, sampleMetres } from './projection';
import { colors, situations, stale as staleToken } from '../design/tokens';

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
  sentence: string | null;
  measuredAt: string | null;
}

const CAP = 0.78; // where the bank-full cap sits on the pole
const CAP_WIDTH = 0.045;
const POLE_WIDTH_KM = 2.6;

function levelColour(level: number | null): THREE.Color {
  const s = situations.find((x) => x.level === level);
  return new THREE.Color(s ? s.hex : staleToken.hex);
}

const VERT = /* glsl */ `
  attribute float aFill;      // fraction of the ground-to-bank span that is water
  attribute float aOverflow;  // 1 when the water is above bank-full
  attribute vec3 aColour;
  attribute float aStale;

  varying float vFill;
  varying float vOverflow;
  varying vec3 vColour;
  varying float vStale;
  varying float vLocalY;      // 0 at the base, 1 at the pole top

  void main() {
    vFill = aFill;
    vOverflow = aOverflow;
    vColour = aColour;
    vStale = aStale;
    vLocalY = position.y;
    gl_Position = projectionMatrix * modelViewMatrix * instanceMatrix * vec4(position, 1.0);
  }
`;

const FRAG = /* glsl */ `
  uniform vec3 uSteel;
  uniform vec3 uTrack;
  uniform vec3 uAlert;

  varying float vFill;
  varying float vOverflow;
  varying vec3 vColour;
  varying float vStale;
  varying float vLocalY;

  void main() {
    // Water fill from the base; the cap at CAP marks bank-full. Above
    // the cap the pole exists only where the water does.
    float waterTop = ${CAP} * min(vFill, 1.0) + vOverflow * max(0.0, vFill - 1.0) * (1.0 - ${CAP}) / 0.6;
    bool atCap = abs(vLocalY - ${CAP}) < ${CAP_WIDTH} && vStale < 0.5;
    bool inWater = vLocalY <= waterTop && vLocalY > 0.004;

    vec3 colour;
    if (vStale > 0.5) {
      // Hollow: the cap only, a quiet stroke.
      if (!atCap) discard;
      colour = uSteel * 0.75;
    } else if (atCap) {
      colour = uSteel;
    } else if (inWater) {
      colour = vLocalY > ${CAP} ? uAlert : vColour;
    } else if (vLocalY < ${CAP} && vLocalY > 0.004) {
      // The empty slot above the water, below the cap.
      colour = uTrack;
    } else {
      discard;
    }

    gl_FragColor = vec4(colour, 1.0);
    // As in Terrain: a raw ShaderMaterial must encode its own output,
    // or every colour reaches the screen as unencoded linear values.
    #include <tonemapping_fragment>
    #include <colorspace_fragment>
  }
`;

export function Gauges({
  stations,
  heightfield,
  exaggeration,
}: {
  stations: StationState[];
  heightfield: Heightfield;
  exaggeration: number;
}) {
  const mesh = useRef<THREE.InstancedMesh>(null);
  const material = useRef<THREE.ShaderMaterial>(null);

  const geometry = useMemo(() => {
    // A unit box with its base at y = 0; the shader reads local y as
    // the fraction up the pole.
    const g = new THREE.BoxGeometry(1, 1, 1);
    g.translate(0, 0.5, 0);
    return g;
  }, []);

  const { matrices, fills, overflows, colours, stales } = useMemo(() => {
    const n = stations.length;
    const matrices = new Float32Array(n * 16);
    const fills = new Float32Array(n);
    const overflows = new Float32Array(n);
    const colours = new Float32Array(n * 3);
    const stales = new Float32Array(n);
    const m = new THREE.Matrix4();
    const q = new THREE.Quaternion();
    const scale = new THREE.Vector3();
    const pos = new THREE.Vector3();

    stations.forEach((s, i) => {
      // Pole height by level: 4 and 5 prominent, 1-3 small and quiet.
      const heightKm = s.level === 5 ? 44 : s.level === 4 ? 34 : s.level === null ? 12 : 17;
      const { u, v } = lonLatToUv(s.lon, s.lat);
      const ground = sampleMetres(heightfield.png, u, v);
      const { x, z } = lonLatToWorld(s.lon, s.lat);
      pos.set(x, (ground / 1000) * exaggeration - 0.5, z);
      scale.set(POLE_WIDTH_KM, heightKm, POLE_WIDTH_KM);
      m.compose(pos, q, scale);
      m.toArray(matrices, i * 16);

      const pct = s.pct ?? 0;
      fills[i] = Math.max(0, Math.min(pct, 170)) / 100;
      overflows[i] = pct > 100 ? 1 : 0;
      const c = levelColour(s.stale ? null : s.level);
      colours[i * 3] = c.r;
      colours[i * 3 + 1] = c.g;
      colours[i * 3 + 2] = c.b;
      stales[i] = s.stale ? 1 : 0;
    });
    return { matrices, fills, overflows, colours, stales };
  }, [stations, heightfield, exaggeration]);

  const uniforms = useMemo(
    () => ({
      uSteel: { value: new THREE.Color(colors.gaugeSteel) },
      uTrack: { value: new THREE.Color(colors.gaugeTrack) },
      uAlert: { value: new THREE.Color(situations[4].hex) },
    }),
    [],
  );

  useFrame(() => {
    const inst = mesh.current;
    if (!inst) return;
    const m = new THREE.Matrix4();
    for (let i = 0; i < stations.length; i++) {
      m.fromArray(matrices, i * 16);
      inst.setMatrixAt(i, m);
    }
    inst.instanceMatrix.needsUpdate = true;
  });

  return (
    <instancedMesh
      key={stations.length}
      ref={mesh}
      args={[geometry, undefined, stations.length]}
      frustumCulled={false}
    >
      <shaderMaterial
        ref={material}
        vertexShader={VERT}
        fragmentShader={FRAG}
        uniforms={uniforms}
      />
      <instancedBufferAttribute attach="geometry-attributes-aFill" args={[fills, 1]} />
      <instancedBufferAttribute attach="geometry-attributes-aOverflow" args={[overflows, 1]} />
      <instancedBufferAttribute attach="geometry-attributes-aColour" args={[colours, 3]} />
      <instancedBufferAttribute attach="geometry-attributes-aStale" args={[stales, 1]} />
    </instancedMesh>
  );
}
