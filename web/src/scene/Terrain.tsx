// The terrain: one displaced mesh whose look is entirely ours (ADR
// 0004). The vertex shader lifts the plane by the heightmap; the
// fragment shader shades it: a north-west key with a soft fill, a
// quiet hypsometric tint graded by elevation, paddy toning on the
// flat plain, neighbour desaturation through the provinces mask, and
// aerial perspective toward the fog. The sea is calm and never
// shaded as rock.
import { useMemo, useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import type { Heightfield } from './heightfield';
import { KM_PER_DEG_LAT, KM_PER_DEG_LON } from './projection';
import { colors, relief } from '../design/tokens';

const VERT = /* glsl */ `
  uniform sampler2D uHeight;
  uniform float uExaggeration; // world km per real km of height
  uniform vec2 uWorldSize;     // the plane's extent in km

  varying vec2 vUvRaster;      // u east, v south, as the PNG is laid out
  varying vec3 vWorldPos;

  float sampleT(vec2 uv) {
    return texture2D(uHeight, uv).r * 2047.0;
  }

  void main() {
    // The plane's own v runs north at v=1; the raster's rows run
    // south from row 0, so flip.
    vUvRaster = vec2(uv.x, 1.0 - uv.y);
    float code = sampleT(vUvRaster);
    float s = max(code - 1.0, 0.0) / 2046.0;
    float metres = s * s * 3300.0;

    vec3 pos = position;
    pos.y = metres / 1000.0 * uExaggeration;
    vec4 world = modelMatrix * vec4(pos, 1.0);
    vWorldPos = world.xyz;
    gl_Position = projectionMatrix * viewMatrix * world;
  }
`;

const FRAG = /* glsl */ `
  uniform sampler2D uHeight;
  uniform sampler2D uMask;
  uniform float uExaggeration;
  uniform vec2 uTexel;       // 1 / texture size
  uniform vec2 uKmPerTexel;  // world km per texel, after exaggeration bookkeeping
  uniform vec3 uAlbedoLow;
  uniform vec3 uAlbedoHigh;
  uniform vec3 uNeighbour;
  uniform vec3 uSea;
  uniform vec3 uFog;
  uniform vec3 uCameraPos;
  uniform float uFogNear;
  uniform float uFogFar;

  varying vec2 vUvRaster;
  varying vec3 vWorldPos;

  float sampleT(vec2 uv) {
    return texture2D(uHeight, uv).r * 2047.0;
  }
  float metresAt(vec2 uv) {
    float code = sampleT(uv);
    float s = max(code - 1.0, 0.0) / 2046.0;
    return s * s * 3300.0;
  }

  void main() {
    float code = sampleT(vUvRaster);
    float tha = smoothstep(0.35, 0.65, texture2D(uMask, vUvRaster).r);

    vec3 colour;
    if (code < 0.5) {
      // The sea: dark and calm, no rock shading, a whisper of the
      // depth gradient toward the shore so the coast reads.
      colour = uSea;
    } else {
      // Height and slope from central differences. The vertical scale
      // carries the exaggeration; the horizontal one does not, which is
      // the point of exaggerating.
      float hL = metresAt(vUvRaster - vec2(uTexel.x, 0.0));
      float hR = metresAt(vUvRaster + vec2(uTexel.x, 0.0));
      float hN = metresAt(vUvRaster - vec2(0.0, uTexel.y));
      float hS = metresAt(vUvRaster + vec2(0.0, uTexel.y));
      vec3 normal = normalize(vec3(
        (hL - hR) / 1000.0 * uExaggeration / (2.0 * uKmPerTexel.x),
        1.0,
        (hN - hS) / 1000.0 * uExaggeration / (2.0 * uKmPerTexel.y)
      ));

      float metres = metresAt(vUvRaster);

      // Quiet hypsometric tint, graded by elevation.
      float t = smoothstep(0.0, 2200.0, metres);
      vec3 albedo = mix(uAlbedoLow, uAlbedoHigh, t);
      // Paddy toning on the low, flat land.
      float flatness = 1.0 - clamp(length(normal.xz) * 6.0, 0.0, 1.0);
      albedo = mix(albedo, uAlbedoLow * 1.08, flatness * (1.0 - t) * 0.5 * tha);
      // Neighbours desaturate and cool; Thailand keeps its colour.
      float grey = dot(albedo, vec3(0.299, 0.587, 0.114));
      vec3 neighbour = mix(vec3(grey), uNeighbour, 0.5);
      albedo = mix(neighbour, albedo, tha);

      // One north-west key (cartographic convention) plus a soft fill.
      vec3 keyDir = normalize(vec3(-0.52, 0.72, -0.45));
      vec3 fillDir = normalize(vec3(0.45, 0.35, 0.55));
      float key = max(dot(normal, keyDir), 0.0);
      float fill = max(dot(normal, fillDir), 0.0);
      // A moonlit Lambert with a floor: nothing black, nothing at full
      // albedo. NoToneMapping (the Canvas is flat), so this scale is the
      // only thing standing between the tokens and the screen.
      vec3 lit = albedo * (0.22 + 0.55 * key + 0.16 * fill);

      // Aerial perspective: the far terrain sinks into the night air.
      float d = distance(uCameraPos, vWorldPos);
      float fog = smoothstep(uFogNear, uFogFar, d);
      colour = mix(lit, uFog, fog * 0.85);
    }

    gl_FragColor = vec4(colour, 1.0);
    // A raw ShaderMaterial gets no automatic output transform: without
    // these chunks the linear working-space values reach the screen
    // unencoded and every token colour renders far too dark.
    #include <tonemapping_fragment>
    #include <colorspace_fragment>
  }
`;

export function Terrain({
  heightfield,
  exaggeration,
  segments = 448,
}: {
  heightfield: Heightfield;
  exaggeration: number;
  segments?: number;
}) {
  const material = useRef<THREE.ShaderMaterial>(null);

  const world = useMemo(() => {
    const widthKm = 19 * KM_PER_DEG_LON;
    const heightKm = 21 * KM_PER_DEG_LAT;
    const cols = Math.round((segments * widthKm) / heightKm);
    const geometry = new THREE.PlaneGeometry(widthKm, heightKm, cols, segments);
    geometry.rotateX(-Math.PI / 2);
    return { widthKm, heightKm, cols, geometry };
  }, [segments]);

  const uniforms = useMemo(
    () => ({
      uHeight: { value: heightfield.texture },
      uMask: { value: heightfield.mask },
      uExaggeration: { value: 1 },
      uTexel: {
        value: new THREE.Vector2(1 / heightfield.png.width, 1 / heightfield.png.height),
      },
      uKmPerTexel: {
        value: new THREE.Vector2(
          (19 * KM_PER_DEG_LON) / heightfield.png.width,
          (21 * KM_PER_DEG_LAT) / heightfield.png.height,
        ),
      },
      uAlbedoLow: { value: new THREE.Color(relief.albedoLow) },
      uAlbedoHigh: { value: new THREE.Color(relief.albedoHigh) },
      uNeighbour: { value: new THREE.Color(relief.neighbour) },
      uSea: { value: new THREE.Color(colors.sea) },
      uFog: { value: new THREE.Color(relief.fog) },
      uCameraPos: { value: new THREE.Vector3() },
      uFogNear: { value: 900 },
      uFogFar: { value: 2600 },
    }),
    [heightfield],
  );

  useFrame(({ camera }) => {
    const u = material.current?.uniforms;
    const cam = u?.uCameraPos;
    const exagg = u?.uExaggeration;
    if (cam && exagg) {
      (cam.value as THREE.Vector3).copy(camera.position);
      exagg.value = exaggeration;
    }
  });

  return (
    <mesh geometry={world.geometry} frustumCulled={false}>
      <shaderMaterial
        ref={material}
        vertexShader={VERT}
        fragmentShader={FRAG}
        uniforms={uniforms}
      />
    </mesh>
  );
}
