// The /lookdev scene: terrain, gauges, draped lines, the camera rig
// with its three presets and one authored flight. Everything visible
// comes from the tokens; every number comes from the fixture.
import { useEffect, useMemo, useRef } from 'react';
import { Canvas, useFrame, useThree } from '@react-three/fiber';
import { CameraControls, Html } from '@react-three/drei';
import type CameraControlsImpl from 'camera-controls';
import * as THREE from 'three';
import { BorderRivers } from './BorderRivers';
import { Gauges, type StationState } from './Gauges';
import { Terrain } from './Terrain';
import type { Heightfield } from './heightfield';
import { NATIONAL, STATION, positionFor, targetFor, type Preset } from './presets';
import { lonLatToWorld } from './projection';
import { colors } from '../design/tokens';

function easeOutExpo(t: number): number {
  return t >= 1 ? 1 : Math.pow(2, -10 * t) * -1 + 1;
}

function v(x: number, y: number, z: number): THREE.Vector3 {
  return new THREE.Vector3(x, y, z);
}

// The brief: "target bounded to Thailand". Panning past the country
// leaves the visitor staring at fog; the target (and the camera with
// it, so the view direction is preserved) is pulled back inside.
const TARGET_BOUNDS = (() => {
  const nw = lonLatToWorld(97.3, 20.6);
  const se = lonLatToWorld(105.8, 5.5);
  return { minX: nw.x, maxX: se.x, minZ: se.z, maxZ: nw.z };
})();

function clampTarget(c: CameraControlsImpl, camera: THREE.Camera): void {
  const t = c.getTarget(new THREE.Vector3());
  const x = Math.min(Math.max(t.x, TARGET_BOUNDS.minX), TARGET_BOUNDS.maxX);
  const z = Math.min(Math.max(t.z, TARGET_BOUNDS.minZ), TARGET_BOUNDS.maxZ);
  if (x === t.x && z === t.z) return;
  void c.setLookAt(
    camera.position.x + (x - t.x),
    camera.position.y,
    camera.position.z + (z - t.z),
    x,
    t.y,
    z,
  );
}

interface Flight {
  from: { pos: THREE.Vector3; target: THREE.Vector3 };
  to: { pos: THREE.Vector3; target: THREE.Vector3 };
  start: number;
  duration: number;
}

/** The camera rig: clamped orbit controls, preset moves, one flight. */
function Rig({
  preset,
  flightRequest,
  onFlightDone,
}: {
  preset: Preset;
  flightRequest: number;
  onFlightDone: () => void;
}) {
  const controls = useRef<CameraControlsImpl | null>(null);
  const flight = useRef<Flight | null>(null);
  const { camera, invalidate } = useThree();
  // The brief: reduced motion cuts instead of flying.
  const reduced = useMemo(
    () =>
      typeof window === 'undefined'
        ? false
        : window.matchMedia('(prefers-reduced-motion: reduce)').matches,
    [],
  );

  useEffect(() => {
    const c = controls.current;
    if (!c) return;
    // Snap to the first preset; later preset changes fly.
    void c.setLookAt(
      positionFor(preset).x,
      positionFor(preset).y,
      positionFor(preset).z,
      targetFor(preset).x,
      targetFor(preset).y,
      targetFor(preset).z,
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Preset buttons fly unless a flight was requested for them.
  useEffect(() => {
    const c = controls.current;
    if (!c) return;
    const pos = v(positionFor(preset).x, positionFor(preset).y, positionFor(preset).z);
    const target = v(targetFor(preset).x, targetFor(preset).y, targetFor(preset).z);
    if (reduced) {
      flight.current = null;
      void c.setLookAt(pos.x, pos.y, pos.z, target.x, target.y, target.z);
      invalidate();
      return;
    }
    flight.current = {
      from: { pos: camera.position.clone(), target: c.getTarget(new THREE.Vector3()) },
      to: { pos, target },
      start: performance.now(),
      duration: 1600,
    };
    invalidate();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [preset]);

  useEffect(() => {
    if (flightRequest === 0) return;
    const c = controls.current;
    if (!c) return;
    void c.setLookAt(
      positionFor(NATIONAL).x,
      positionFor(NATIONAL).y,
      positionFor(NATIONAL).z,
      targetFor(NATIONAL).x,
      targetFor(NATIONAL).y,
      targetFor(NATIONAL).z,
    );
    if (reduced) {
      // Cut straight to the station: snap, no flight, no glide.
      void c.setLookAt(
        positionFor(STATION).x,
        positionFor(STATION).y,
        positionFor(STATION).z,
        targetFor(STATION).x,
        targetFor(STATION).y,
        targetFor(STATION).z,
      );
      onFlightDone();
      invalidate();
      return;
    }
    // Let the snap land, then fly to the station.
    flight.current = {
      from: {
        pos: v(positionFor(NATIONAL).x, positionFor(NATIONAL).y, positionFor(NATIONAL).z),
        target: v(targetFor(NATIONAL).x, targetFor(NATIONAL).y, targetFor(NATIONAL).z),
      },
      to: {
        pos: v(positionFor(STATION).x, positionFor(STATION).y, positionFor(STATION).z),
        target: v(targetFor(STATION).x, targetFor(STATION).y, targetFor(STATION).z),
      },
      start: performance.now() + 120,
      duration: 1600,
    };
    invalidate();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [flightRequest]);

  useFrame(() => {
    const f = flight.current;
    const c = controls.current;
    if (!c) return;
    if (!f) {
      clampTarget(c, camera);
      return;
    }
    const t = (performance.now() - f.start) / f.duration;
    if (t < 0) {
      invalidate();
      return;
    }
    if (t >= 1) {
      void c.setLookAt(
        f.to.pos.x,
        f.to.pos.y,
        f.to.pos.z,
        f.to.target.x,
        f.to.target.y,
        f.to.target.z,
      );
      flight.current = null;
      onFlightDone();
      invalidate();
      return;
    }
    const e = easeOutExpo(t);
    const pos = f.from.pos.clone().lerp(f.to.pos, e);
    const target = f.from.target.clone().lerp(f.to.target, e);
    void c.setLookAt(pos.x, pos.y, pos.z, target.x, target.y, target.z);
    invalidate();
  });

  return (
    <CameraControls
      ref={controls}
      makeDefault
      minDistance={70}
      maxDistance={2800}
      minPolarAngle={0.12}
      maxPolarAngle={1.42}
      onStart={() => {
        // Any user input interrupts the flight.
        flight.current = null;
      }}
    />
  );
}

/** A DOM label projected from the selected station's gauge. Thai text
 *  never goes into WebGL (brief, section 4); Html keeps it in the DOM. */
function StationLabel({ station }: { station: StationState | null }) {
  const world = useMemo(() => {
    if (!station) return null;
    const { x, z } = lonLatToWorld(station.lon, station.lat);
    return new THREE.Vector3(x, 34, z);
  }, [station]);
  if (!station || !world) return null;
  return (
    <Html position={world} center zIndexRange={[20, 0]}>
      <div className="lookdev__station-label">{station.name}</div>
    </Html>
  );
}

export function LookdevScene({
  heightfield,
  bordersTopo,
  riversGeo,
  stations,
  selected,
  preset,
  flightRequest,
  onFlightDone,
  exaggeration,
  onReady,
}: {
  heightfield: Heightfield;
  bordersTopo: unknown;
  riversGeo: GeoJSON.FeatureCollection;
  stations: StationState[];
  selected: StationState | null;
  preset: Preset;
  flightRequest: number;
  onFlightDone: () => void;
  exaggeration: number;
  onReady?: () => void;
}) {
  return (
    <>
      <Canvas
        frameloop="demand"
        dpr={[1, 2]}
        flat
        gl={{ antialias: true, alpha: false }}
        camera={{ near: 1, far: 12000, fov: 42 }}
        onCreated={({ gl, scene }) => {
          gl.setClearColor(new THREE.Color(colors.ground));
          scene.background = new THREE.Color(colors.ground);
          onReady?.();
        }}
      >
        <Rig preset={preset} flightRequest={flightRequest} onFlightDone={onFlightDone} />
        <Terrain heightfield={heightfield} exaggeration={exaggeration} />
        <BorderRivers
          heightfield={heightfield}
          bordersTopo={bordersTopo}
          riversGeo={riversGeo}
          exaggeration={exaggeration}
        />
        <Gauges stations={stations} heightfield={heightfield} exaggeration={exaggeration} />
        <StationLabel station={selected} />
      </Canvas>
    </>
  );
}
