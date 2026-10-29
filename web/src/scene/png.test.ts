/// <reference types="node" />
import { deflateSync } from 'node:zlib';
import { describe, expect, it } from 'vitest';
import { crc32, decodeGray16Png } from './png';

function chunk(type: string, data: Uint8Array): Uint8Array {
  const out = new Uint8Array(12 + data.length);
  new DataView(out.buffer).setUint32(0, data.length);
  for (let i = 0; i < 4; i++) out[4 + i] = type.charCodeAt(i);
  out.set(data, 8);
  new DataView(out.buffer).setUint32(8 + data.length, crc32(out.subarray(4, 8 + data.length)));
  return out;
}

function png(
  width: number,
  height: number,
  rows: number[][] /* filter byte, then big-endian pixels */,
): Uint8Array {
  const ihdr = new Uint8Array(13);
  const dv = new DataView(ihdr.buffer);
  dv.setUint32(0, width);
  dv.setUint32(4, height);
  ihdr[8] = 16; // bit depth
  ihdr[9] = 0; // grayscale
  const body = new Uint8Array(rows.flat());
  const parts = [
    new Uint8Array([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    chunk('IHDR', ihdr),
    chunk('IDAT', new Uint8Array(deflateSync(body))),
    chunk('IEND', new Uint8Array(0)),
  ];
  const total = parts.reduce((n, p) => n + p.length, 0);
  const out = new Uint8Array(total);
  let at = 0;
  for (const p of parts) {
    out.set(p, at);
    at += p.length;
  }
  return out;
}

describe('decodeGray16Png', () => {
  it('decodes a PIL-encoded 16-bit grayscale image', async () => {
    // Written by Pillow (mode I;16, optimize=True) with row filters
    // [none, up, average, none, none]; the expected samples come from
    // the numpy array that produced it. This is the encoder of the
    // shipped heightmap, so it is the oracle that matters.
    const b64 =
      'iVBORw0KGgoAAAANSUhEUgAAAAcAAAAFEAAAAAD8YXVHAAAAVklEQVR42gFLALT/AGbcsz1cAw6VavRH1LyuAq5Qs5kdFcSsbAADvQ6nA02iFibtKs11+U8SOe+MADT9AQNX6ev1nf0l24EJAL8Uu5MUmaBEy9E5TRVw3DMhTkyZBR0AAAAASUVORK5CYII=';
    const bytes = Uint8Array.from(atob(b64), (c) => c.charCodeAt(0));
    const { width, height, data } = await decodeGray16Png(bytes);
    expect([width, height]).toEqual([7, 5]);
    // The expected samples are Pillow's own reading of the file
    // (list(img.getdata())), so the oracle is the encoder itself.
    expect(Array.from(data)).toEqual([
      26332, 45885, 23555, 3733, 27380, 18388, 48302, 5164, 26326, 31000, 53825, 55028, 19089,
      51797, 22456, 29933, 25516, 26603, 38718, 33440, 38150, 13565, 259, 22505, 60405, 40445, 9691,
      33033, 48916, 48019, 5273, 41028, 52177, 14669, 5488,
    ]);
  });

  it('decodes the Sub and Paeth filters', async () => {
    // 2x3 pixels; filtering is byte-wise with bpp = 2.
    const img = png(2, 3, [
      [0, 0x30, 0x00, 0x40, 0x00], // row 0: 0x3000, 0x4000
      [1, 0x10, 0x00, 0x10, 0x00], // Sub: pixel1's high byte += pixel0's
      [4, 0x00, 0x01, 0x00, 0x02], // Paeth against row 1
    ]);
    const { data } = await decodeGray16Png(img);
    // Row 0: none.
    expect(data[0]).toBe(0x3000);
    expect(data[1]).toBe(0x4000);
    // Row 1 (Sub): pixel0 has no left, pixel1's high byte adds pixel0's.
    expect(data[2]).toBe(0x1000);
    expect(data[3]).toBe(0x2000);
    // Row 2 (Paeth): byte0 has up 0x10 and picks it; byte1 has all-zero
    // predictors; byte2 has up 0x20 and picks it; byte3 has left 0x01.
    expect(data[4]).toBe(0x1001);
    expect(data[5]).toBe(0x2003);
  });

  it('rejects 8-bit depth with the depth error, not a CRC error', async () => {
    const bad = png(1, 1, [[0, 0x12, 0x34]]);
    bad[8 + 8 + 8] = 8; // IHDR data starts at 16; the depth byte is offset 8 within it
    // Mend the IHDR CRC (it covers the type and data bytes) so the
    // decoder reaches the depth check.
    new DataView(bad.buffer).setUint32(8 + 8 + 13, crc32(bad.subarray(12, 8 + 8 + 13)));
    await expect(decodeGray16Png(bad)).rejects.toThrow('16-bit grayscale');
  });

  it('rejects corrupted data instead of returning garbage', async () => {
    const img = png(2, 1, [[0, 0x01, 0x02, 0x03, 0x04]]);
    img.set([(img[41] ?? 0) ^ 0xff], 41); // a byte inside the compressed stream
    await expect(decodeGray16Png(img)).rejects.toThrow();
  });
});

describe('crc32', () => {
  it('matches the standard test vector', () => {
    expect(crc32(new Uint8Array([49, 50, 51, 52, 53, 54, 55, 56, 57]))).toBe(0xcbf43926);
  });
});
