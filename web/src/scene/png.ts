// Decode 16-bit grayscale PNGs without a canvas, because canvases
// reduce 16-bit images to 8 bits and the heightmap's sqrt-11 codes
// live in those upper bits (ADR 0004). Supports exactly what our
// pipeline emits: color type 0 (grayscale), bit depth 16, no
// interlace. Inflation runs on DecompressionStream, so this works in
// a worker or on the main thread with no dependencies.

export interface Gray16Png {
  width: number;
  height: number;
  /** Big-endian sample values, row-major, length width * height. */
  data: Uint16Array;
}

const PNG_SIGNATURE = [0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a];

const CRC_TABLE = (() => {
  const table = new Uint32Array(256);
  for (let n = 0; n < 256; n++) {
    let c = n;
    for (let k = 0; k < 8; k++) {
      c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    }
    table[n] = c >>> 0;
  }
  return table;
})();

/** Standard CRC-32, as PNG chunks use it. */
export function crc32(bytes: Uint8Array): number {
  let c = 0xffffffff;
  for (const byte of bytes) {
    c = (CRC_TABLE[(c ^ byte) & 0xff] ?? 0) ^ (c >>> 8);
  }
  return (c ^ 0xffffffff) >>> 0;
}

export async function decodeGray16Png(bytes: Uint8Array): Promise<Gray16Png> {
  for (let i = 0; i < PNG_SIGNATURE.length; i++) {
    if (bytes[i] !== (PNG_SIGNATURE[i] ?? -1)) {
      throw new Error('not a PNG');
    }
  }

  let offset = PNG_SIGNATURE.length;
  let width = 0;
  let height = 0;
  let bitDepth = 0;
  let colorType = -1;
  let interlace = 0;
  const idat: Uint8Array[] = [];
  let idatLength = 0;
  let sawIend = false;

  while (offset + 8 <= bytes.length) {
    const view = new DataView(bytes.buffer, bytes.byteOffset + offset, bytes.byteLength - offset);
    const length = view.getUint32(0);
    const type = String.fromCharCode(
      bytes[offset + 4] ?? 0,
      bytes[offset + 5] ?? 0,
      bytes[offset + 6] ?? 0,
      bytes[offset + 7] ?? 0,
    );
    const dataStart = offset + 8;
    if (dataStart + length + 4 > bytes.length) {
      throw new Error('truncated chunk ' + type);
    }
    const chunk = bytes.subarray(dataStart, dataStart + length);
    const crcView = new DataView(bytes.buffer, bytes.byteOffset + dataStart + length, 4);
    if (crcView.getUint32(0) !== crc32(bytes.subarray(offset + 4, dataStart + length))) {
      throw new Error('CRC mismatch in chunk ' + type);
    }

    if (type === 'IHDR') {
      const ihdr = new DataView(chunk.buffer, chunk.byteOffset, chunk.byteLength);
      width = ihdr.getUint32(0);
      height = ihdr.getUint32(4);
      bitDepth = chunk[8] ?? 0;
      colorType = chunk[9] ?? -1;
      interlace = chunk[12] ?? 1;
    } else if (type === 'IDAT') {
      idat.push(chunk);
      idatLength += chunk.length;
    } else if (type === 'IEND') {
      sawIend = true;
      break;
    }
    offset = dataStart + length + 4; // skip the CRC
  }

  if (!sawIend || width === 0) {
    throw new Error('incomplete PNG');
  }
  if (bitDepth !== 16 || colorType !== 0) {
    throw new Error(`expected 16-bit grayscale, got depth ${bitDepth} type ${colorType}`);
  }
  if (interlace !== 0) {
    throw new Error('interlaced PNGs are not supported');
  }

  const inflated = await inflate(concat(idat, idatLength));

  // One filter byte per row, then width samples of 2 bytes each.
  const bpp = 2; // filter unit: bytes per pixel
  const stride = width * bpp;
  if (inflated.length !== height * (stride + 1)) {
    throw new Error(
      `inflated size ${inflated.length} does not fit ${height} rows of ${stride + 1}`,
    );
  }

  const rows = unfilter(inflated, height, stride, bpp);
  const data = new Uint16Array(width * height);
  for (let y = 0; y < height; y++) {
    const row = rows[y];
    if (row === undefined) throw new Error(`missing row ${String(y)}`);
    for (let x = 0; x < width; x++) {
      data[y * width + x] = ((row[x * 2] ?? 0) << 8) | (row[x * 2 + 1] ?? 0);
    }
  }
  return { width, height, data };
}

async function inflate(data: Uint8Array): Promise<Uint8Array> {
  const stream = new DecompressionStream('deflate');
  const writer = stream.writable.getWriter();
  void writer.write(data as BufferSource);
  void writer.close();
  const chunks: Uint8Array[] = [];
  let total = 0;
  const reader = stream.readable.getReader();
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    chunks.push(value);
    total += value.length;
  }
  return concat(chunks, total);
}

function concat(parts: Uint8Array[], total: number): Uint8Array {
  const out = new Uint8Array(total);
  let at = 0;
  for (const p of parts) {
    out.set(p, at);
    at += p.length;
  }
  return out;
}

/** Undo the per-row PNG filters in place, returning the pixel rows. */
function unfilter(inflated: Uint8Array, height: number, stride: number, bpp: number): Uint8Array[] {
  const rows: Uint8Array[] = [];
  const raw = new Uint8Array(height * stride);
  for (let y = 0; y < height; y++) {
    const filter = inflated[y * (stride + 1)] ?? -1;
    const src = inflated.subarray(y * (stride + 1) + 1, (y + 1) * (stride + 1));
    const cur = raw.subarray(y * stride, (y + 1) * stride);
    const prev = y > 0 ? raw.subarray((y - 1) * stride, y * stride) : null;
    cur.set(src);
    switch (filter) {
      case 0: // None
        break;
      case 1: // Sub
        for (let i = bpp; i < stride; i++) cur[i] = ((cur[i] ?? 0) + (cur[i - bpp] ?? 0)) & 0xff;
        break;
      case 2: // Up
        if (prev) for (let i = 0; i < stride; i++) cur[i] = ((cur[i] ?? 0) + (prev[i] ?? 0)) & 0xff;
        break;
      case 3: // Average
        for (let i = 0; i < stride; i++) {
          const left = i >= bpp ? (cur[i - bpp] ?? 0) : 0;
          const up = prev ? (prev[i] ?? 0) : 0;
          cur[i] = ((cur[i] ?? 0) + ((left + up) >> 1)) & 0xff;
        }
        break;
      case 4: // Paeth
        for (let i = 0; i < stride; i++) {
          const left = i >= bpp ? (cur[i - bpp] ?? 0) : 0;
          const up = prev ? (prev[i] ?? 0) : 0;
          const upLeft = prev && i >= bpp ? (prev[i - bpp] ?? 0) : 0;
          cur[i] = ((cur[i] ?? 0) + paeth(left, up, upLeft)) & 0xff;
        }
        break;
      default:
        throw new Error(`unknown filter ${String(filter)}`);
    }
    rows.push(cur);
  }
  return rows;
}

function paeth(left: number, up: number, upLeft: number): number {
  const p = left + up - upLeft;
  const pl = Math.abs(p - left);
  const pu = Math.abs(p - up);
  const pd = Math.abs(p - upLeft);
  if (pl <= pu && pl <= pd) return left;
  if (pu <= pd) return up;
  return upLeft;
}
