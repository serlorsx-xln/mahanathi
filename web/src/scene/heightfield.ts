// Loading the heightmap and building what the scene needs from it:
// the CPU code copy (draping, seating gauges), the R16F-ish half-float
// texture for the vertex shader, and the Thailand mask rasterised from
// the provinces (neighbour desaturation and the crisp country edge).
import * as THREE from 'three';
import { feature } from 'topojson-client';
import { decodeGray16Png, type Gray16Png } from './png';

export interface Heightfield {
  png: Gray16Png;
  /** Half-float bits of code / 2047, the shader's height source. */
  textureData: Uint16Array;
  texture: THREE.DataTexture;
  /** The Thailand mask, 1 inside the country, same uv layout. */
  mask: THREE.DataTexture;
}

export async function loadHeightfield(
  heightmapUrl: string,
  provincesTopo: unknown,
): Promise<Heightfield> {
  const bytes = new Uint8Array(await (await fetch(heightmapUrl)).arrayBuffer());
  const png = await decodeGray16Png(bytes);

  const textureData = new Uint16Array(png.data.length);
  for (let i = 0; i < png.data.length; i++) {
    textureData[i] = THREE.DataUtils.toHalfFloat((png.data[i] ?? 0) / 2047);
  }
  const texture = new THREE.DataTexture(
    textureData,
    png.width,
    png.height,
    THREE.RedFormat,
    THREE.HalfFloatType,
  );
  texture.minFilter = THREE.LinearFilter;
  texture.magFilter = THREE.LinearFilter;
  texture.wrapS = THREE.ClampToEdgeWrapping;
  texture.wrapT = THREE.ClampToEdgeWrapping;
  texture.needsUpdate = true;

  return { png, textureData, texture, mask: rasteriseMask(provincesTopo, 512) };
}

/** Rasterise the province polygons into a single-channel mask texture. */
function rasteriseMask(topo: unknown, size: number): THREE.DataTexture {
  const objects = (topo as { objects: Record<string, unknown> }).objects;
  const first = Object.values(objects)[0];
  if (first === undefined) throw new Error('the topology has no objects');
  const geo = feature(
    topo as never,
    (objects.provinces ?? first) as never,
  ) as unknown as GeoJSON.FeatureCollection;

  const w = size;
  const h = Math.round((size * 2048) / 1815);
  const canvas = document.createElement('canvas');
  canvas.width = w;
  canvas.height = h;
  const ctx = canvas.getContext('2d');
  if (!ctx) throw new Error('no 2d context for the mask');
  ctx.fillStyle = '#000';
  ctx.fillRect(0, 0, w, h);
  ctx.fillStyle = '#fff';
  for (const f of geo.features) {
    const polys =
      f.geometry.type === 'Polygon'
        ? [f.geometry.coordinates]
        : f.geometry.type === 'MultiPolygon'
          ? f.geometry.coordinates
          : [];
    for (const poly of polys) {
      for (const ring of poly) {
        ctx.beginPath();
        ring.forEach((point, i) => {
          const [lon, lat] = point as [number, number];
          const x = ((lon - 92) / 19) * w;
          const y = ((24 - lat) / 21) * h;
          if (i === 0) ctx.moveTo(x, y);
          else ctx.lineTo(x, y);
        });
        ctx.closePath();
        ctx.fill();
      }
    }
  }
  const img = ctx.getImageData(0, 0, w, h);
  const data = new Uint8Array(w * h);
  for (let i = 0; i < w * h; i++) data[i] = img.data[i * 4] ?? 0;
  const tex = new THREE.DataTexture(data, w, h, THREE.RedFormat, THREE.UnsignedByteType);
  tex.minFilter = THREE.LinearFilter;
  tex.magFilter = THREE.LinearFilter;
  tex.wrapS = THREE.ClampToEdgeWrapping;
  tex.wrapT = THREE.ClampToEdgeWrapping;
  tex.needsUpdate = true;
  return tex;
}
