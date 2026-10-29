"""Stitch Terrarium tiles, reproject to an equirectangular grid, mask the sea
and measure candidate heightmap encodings.

Research script, not the production pipeline.

Usage (run with python -I):
  python -I heightmap_measure.py \
      --tiles geodata/.cache/terrarium --z 8 \
      --bbox 92 3 111 24 --lat0 13 \
      --land geodata/.cache/naturalearth/ne_10m_land/ne_10m_land.shp \
      --heights 1024 1536 2048 \
      --out geodata/.cache/heightmap_candidates

Output grid: plate carree (regular lon/lat), with columns chosen so that a
pixel is square once x is scaled by cos(lat0). Row 0 is the north edge.

Writes candidate files to --out and prints a JSON report to stdout.
"""
import argparse
import io
import json
import lzma
import math
import os
import zlib

import brotli
import numpy as np
import shapefile
import zstandard
from PIL import Image, ImageDraw
from shapely.geometry import box, shape

try:
    import oxipng
except ImportError:  # optional
    oxipng = None

Image.MAX_IMAGE_PIXELS = None

# Lower Chao Phraya plain test box (W, S, E, N): Ayutthaya, Ang Thong,
# Pathum Thani and southern Sing Buri. Kept west of the Lop Buri and Saraburi
# hills (a first, wider box reached p95 = 75 m because it caught them).
PLAIN = (100.15, 14.00, 100.60, 15.00)
# Bangkok and Samut Prakan lowlands, used to show why "h <= 0" is not a sea mask.
BKK = (100.30, 13.55, 100.90, 14.00)


def lon2x(lon, z):
    return (lon + 180.0) / 360.0 * (1 << z)


def lat2y(lat, z):
    return (1.0 - np.arcsinh(np.tan(np.radians(lat))) / math.pi) / 2.0 * (1 << z)


def stitch(tiles, z, bbox):
    w, s, e, n = bbox
    x0, x1 = int(math.floor(lon2x(w, z))), int(math.floor(lon2x(e, z) - 1e-9))
    y0, y1 = int(math.floor(lat2y(n, z))), int(math.floor(lat2y(s, z) - 1e-9))
    nx, ny = x1 - x0 + 1, y1 - y0 + 1
    mos = np.zeros((ny * 256, nx * 256), dtype=np.float32)
    for x in range(x0, x1 + 1):
        for y in range(y0, y1 + 1):
            p = os.path.join(tiles, str(z), str(x), "%d.png" % y)
            a = np.asarray(Image.open(p).convert("RGB"), dtype=np.float32)
            h = a[..., 0] * 256.0 + a[..., 1] + a[..., 2] / 256.0 - 32768.0
            mos[(y - y0) * 256:(y - y0 + 1) * 256, (x - x0) * 256:(x - x0 + 1) * 256] = h
    return mos, x0, y0


def bilinear(img, fy, fx):
    """Sample img at fractional rows fy (1D) and columns fx (1D), separable."""
    fy = np.clip(fy, 0, img.shape[0] - 1.000001)
    fx = np.clip(fx, 0, img.shape[1] - 1.000001)
    iy, ix = np.floor(fy).astype(int), np.floor(fx).astype(int)
    wy, wx = (fy - iy)[:, None], (fx - ix)[None, :]
    rows0, rows1 = img[iy], img[iy + 1]
    top = rows0[:, ix] * (1 - wx) + rows0[:, ix + 1] * wx
    bot = rows1[:, ix] * (1 - wx) + rows1[:, ix + 1] * wx
    return (top * (1 - wy) + bot * wy).astype(np.float32)


def reproject(mos, x0, y0, z, bbox, rows, cols):
    """Mercator mosaic -> plate carree grid of rows x cols (pixel centres)."""
    w, s, e, n = bbox
    lats = n - (np.arange(rows) + 0.5) * (n - s) / rows
    lons = w + (np.arange(cols) + 0.5) * (e - w) / cols
    fy = (lat2y(lats, z) - y0) * 256.0 - 0.5
    fx = (lon2x(lons, z) - x0) * 256.0 - 0.5
    return bilinear(mos, fy, fx)


def box_blur(a, r):
    """Separable box blur of radius r pixels (edges clamped)."""
    out = a.astype(np.float64)
    for axis in (0, 1):
        pad = [(0, 0), (0, 0)]
        pad[axis] = (r + 1, r)
        p = np.pad(out, pad, mode="edge")
        c = np.cumsum(p, axis=axis)
        n = out.shape[axis]
        hi = np.take(c, np.arange(2 * r + 1, 2 * r + 1 + n), axis=axis)
        lo = np.take(c, np.arange(0, n), axis=axis)
        out = (hi - lo) / (2 * r + 1)
    return out.astype(np.float32)


def raster_land(shp, bbox, rows, cols, ss=4, only_a3=None):
    """Fraction of each output pixel covered by Natural Earth land polygons
    (or, with only_a3, by one admin-0 unit of ne_10m_admin_0_countries)."""
    w, s, e, n = bbox
    clip = box(w - 0.5, s - 0.5, e + 0.5, n + 0.5)
    W, H = cols * ss, rows * ss
    im = Image.new("L", (W, H), 0)
    dr = ImageDraw.Draw(im)

    def px(coords):
        return [((x - w) / (e - w) * W, (n - y) / (n - s) * H) for x, y in coords]

    r = shapefile.Reader(shp)
    fields = [f[0] for f in r.fields[1:]]
    for sr in r.iterShapeRecords():
        sh = sr.shape
        if only_a3 is not None and sr.record[fields.index("ADM0_A3")] != only_a3:
            continue
        if sh.bbox[2] < w or sh.bbox[0] > e or sh.bbox[3] < s or sh.bbox[1] > n:
            continue
        g = shape(sh.__geo_interface__).intersection(clip)
        if g.is_empty:
            continue
        polys = [g] if g.geom_type == "Polygon" else [p for p in getattr(g, "geoms", []) if p.geom_type == "Polygon"]
        for p in polys:
            dr.polygon(px(p.exterior.coords), fill=255)
            for hole in p.interiors:
                dr.polygon(px(hole.coords), fill=0)
    small = im.resize((cols, rows), Image.BOX)
    return np.asarray(small, dtype=np.float32) / 255.0


PILLOW_SIZES = {}


def png_bytes(img, optimize=True):
    b = io.BytesIO()
    img.save(b, format="PNG", optimize=optimize, compress_level=9)
    raw = b.getvalue()
    PILLOW_SIZES["last"] = len(raw)
    best = raw
    if oxipng is not None:
        try:
            o = oxipng.optimize_from_memory(raw, level=6)
            if len(o) < len(best):
                best = o
        except Exception:
            pass
    return raw, best


def webp_bytes(img):
    b = io.BytesIO()
    img.save(b, format="WEBP", lossless=True, quality=100, method=6, exact=True)
    return b.getvalue()


def med_residual(u):
    """JPEG-LS median edge predictor (a = left, b = up, c = up-left, zero
    outside the image). Returns int16 residuals; lossless, decodable in one
    sequential pass."""
    x = u.astype(np.int32)
    a = np.zeros_like(x); a[:, 1:] = x[:, :-1]
    b = np.zeros_like(x); b[1:, :] = x[:-1, :]
    c = np.zeros_like(x); c[1:, 1:] = x[:-1, :-1]
    mn, mx = np.minimum(a, b), np.maximum(a, b)
    pred = np.where(c >= mx, mn, np.where(c <= mn, mx, a + b - c))
    return (x - pred).astype(np.int16)


def bin_sizes(res16):
    """Compressed sizes of an int16 residual array: interleaved little-endian
    bytes and planar (all high bytes, then all low bytes)."""
    le = res16.tobytes()
    v = res16.astype(np.int32) & 0xFFFF
    planar = (v >> 8).astype(np.uint8).tobytes() + (v & 255).astype(np.uint8).tobytes()
    return {"br11": len(brotli.compress(le, quality=11, lgwin=24)),
            "br11_planar": len(brotli.compress(planar, quality=11, lgwin=24)),
            "gzip9_planar": len(zlib.compress(planar, 9)),
            "zstd19_planar": len(zstandard.ZstdCompressor(level=19).compress(planar))}


def plain_slice(bbox, rows, cols, region=PLAIN):
    w, s, e, n = bbox
    pw, ps, pe, pn = region
    r0 = int((n - pn) / (n - s) * rows)
    r1 = int(math.ceil((n - ps) / (n - s) * rows))
    c0 = int((pw - w) / (e - w) * cols)
    c1 = int(math.ceil((pe - w) / (e - w) * cols))
    return slice(r0, r1), slice(c0, c1)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--tiles", required=True)
    ap.add_argument("--z", type=int, required=True)
    ap.add_argument("--bbox", type=float, nargs=4, required=True)
    ap.add_argument("--lat0", type=float, default=13.0)
    ap.add_argument("--land", required=True)
    ap.add_argument("--heights", type=int, nargs="+", default=[1024, 1536, 2048])
    ap.add_argument("--out", required=True)
    ap.add_argument("--sqrt-bits", type=int, nargs="*", default=[10, 11, 12])
    ap.add_argument("--skip-basic", action="store_true",
                    help="only run the sqrt-N-bit variants (faster re-runs)")
    ap.add_argument("--tha", default=None,
                    help="ne_10m_admin_0_countries.shp; enables the Thailand-priority variants")
    ap.add_argument("--band-km", type=float, default=25.0)
    ap.add_argument("--nb-blur", type=int, default=2,
                    help="box-blur radius (px, applied 3 times) for neighbour terrain; 0 = no blur")
    ap.add_argument("--prio-bits", type=int, nargs="*", default=[12, 11, 10])
    a = ap.parse_args()
    os.makedirs(a.out, exist_ok=True)
    bbox = tuple(a.bbox)
    w, s, e, n = bbox
    rep = {"bbox": bbox, "z": a.z, "lat0": a.lat0, "plain_box": PLAIN}

    mos, x0, y0 = stitch(a.tiles, a.z, bbox)
    rep["mosaic_px"] = list(mos.shape)
    # Native-resolution size of the bbox inside the mosaic.
    nat_rows = int(round((lat2y(s, a.z) - lat2y(n, a.z)) * 256))
    nat_cols = int(round((lon2x(e, a.z) - lon2x(w, a.z)) * 256))
    rep["native_bbox_px"] = [nat_rows, nat_cols]

    km_per_deg_lat = 110.574 + (111.694 - 110.574) * (a.lat0 / 90.0)  # rough, ~110.7 at 13 deg
    km_per_deg_lon = 111.320 * math.cos(math.radians(a.lat0))
    width_km = (e - w) * km_per_deg_lon
    height_km = (n - s) * km_per_deg_lat
    rep["extent_km"] = [round(width_km, 1), round(height_km, 1)]

    # Reproject once at native-ish resolution, then box-filter down per target.
    ref_rows = nat_rows
    ref_cols = int(round(ref_rows * width_km / height_km))
    eq = reproject(mos, x0, y0, a.z, bbox, ref_rows, ref_cols)
    rep["equirect_native_px"] = [ref_rows, ref_cols]
    rep["native_m_per_px"] = round(height_km * 1000 / ref_rows, 1)

    results = []
    for H in a.heights:
        Wd = int(round(H * width_km / height_km))
        res = {"rows": H, "cols": Wd, "m_per_px": round(height_km * 1000 / H, 1),
               "megapixels": round(H * Wd / 1e6, 3)}
        hf = np.asarray(Image.fromarray(eq).resize((Wd, H), Image.BOX), dtype=np.float32)
        land = raster_land(a.land, bbox, H, Wd)
        ne_land = land >= 0.5
        res["ne_land_fraction"] = round(float(ne_land.mean()), 4)
        res["ne_sea_px_h_gt_1"] = int(np.sum(~ne_land & (hf > 1)))
        res["ne_land_px_h_le_0"] = int(np.sum(ne_land & (hf <= 0)))
        br, bc = plain_slice(bbox, H, Wd, BKK)
        bl = ne_land[br, bc]
        res["bkk_box_land_px"] = int(bl.sum())
        res["bkk_box_land_px_h_le_0"] = int(np.sum(bl & (hf[br, bc] <= 0)))
        res["bkk_box_land_px_h_le_1"] = int(np.sum(bl & (hf[br, bc] <= 1)))
        # Final mask: NE land, plus DEM islands NE does not carry.
        is_land = ne_land | (hf > 1.0)
        res["land_fraction"] = round(float(is_land.mean()), 4)
        res["land_px_h_le_0"] = int(np.sum(is_land & (hf <= 0)))
        res["land_px_h_lt_minus2"] = int(np.sum(is_land & (hf < -2)))
        res["sea_px_h_gt_5"] = int(np.sum(~is_land & (hf > 5)))
        res["dem_le0_fraction"] = round(float((hf <= 0).mean()), 4)
        res["land_h_min_max"] = [round(float(hf[is_land].min()), 1), round(float(hf[is_land].max()), 1)]
        res["sea_h_min"] = round(float(hf[~is_land].min()), 1)
        pr, pc = plain_slice(bbox, H, Wd)
        plain = hf[pr, pc]
        pl = is_land[pr, pc]
        pv = plain[pl]
        res["plain_px"] = int(pl.sum())
        res["plain_h_pct"] = {p: round(float(np.percentile(pv, p)), 2) for p in (1, 5, 50, 95, 99)}
        dx = np.diff(plain, axis=1)[pl[:, 1:] & pl[:, :-1]]
        res["plain_neighbour_diff_std_m"] = round(float(np.std(dx)), 3)
        res["plain_neighbour_absdiff_median_m"] = round(float(np.median(np.abs(dx))), 3)

        # Sea-flattened reference: sea = 0, land clamped to >= 0.
        ref = np.where(is_land, np.maximum(hf, 0.0), 0.0).astype(np.float32)
        hmax = float(ref.max())
        res["encode_hmax_m"] = round(hmax, 1)
        HM = math.ceil(hmax / 100.0) * 100.0  # range ceiling for 8-bit curves
        res["encode_range_m"] = HM

        enc = {}

        def record(name, data, decoded, note=""):
            err = (decoded - ref)[pr, pc][pl]
            d = {"bytes": len(data), "plain_rmse_m": round(float(np.sqrt(np.mean(err ** 2))), 3),
                 "plain_maxerr_m": round(float(np.max(np.abs(err))), 3),
                 "plain_distinct_levels": int(len(np.unique(decoded[pr, pc][pl]))),
                 "land_rmse_m": round(float(np.sqrt(np.mean(((decoded - ref)[is_land]) ** 2))), 3)}
            if note:
                d["note"] = note
            if name.endswith(".png"):
                d["pillow_bytes"] = PILLOW_SIZES.get("last")
            enc[name] = d
            return d

        def save(name, data):
            with open(os.path.join(a.out, "h%d_%s" % (H, name)), "wb") as f:
                f.write(data)

        if not a.skip_basic:
            # 1. Terrarium RGB, raw (bathymetry kept, fractional channel kept).
            v = hf.astype(np.float64) + 32768.0
            r_, g_ = np.floor(v / 256), np.floor(v % 256)
            b_ = np.floor((v - np.floor(v)) * 256)
            rgb = np.stack([r_, g_, b_], -1).astype(np.uint8)
            raw, best = png_bytes(Image.fromarray(rgb))
            dec = (rgb[..., 0] * 256.0 + rgb[..., 1] + rgb[..., 2] / 256.0 - 32768.0)
            dec = np.where(is_land, np.maximum(dec, 0), 0)
            record("terrarium_raw.png", best, dec, "bathymetry and 1/256 m fraction kept")
            save("terrarium_raw.png", best)

            # 2. Terrarium RGB, sea flat, 1 m integer (B = 0).
            q = np.round(ref)
            v = q + 32768.0
            rgb = np.stack([np.floor(v / 256), np.floor(v % 256), np.zeros_like(v)], -1).astype(np.uint8)
            raw, best = png_bytes(Image.fromarray(rgb))
            record("terrarium_flat_1m.png", best, q)
            save("terrarium_flat_1m.png", best)
            enc["terrarium_flat_1m.webp"] = dict(enc["terrarium_flat_1m.png"])
            enc["terrarium_flat_1m.webp"]["bytes"] = len(webp_bytes(Image.fromarray(rgb)))

            # 3. 16-bit grayscale PNG, linear, sea reserved as 0, several steps.
            for step in (1.0, 0.5, 0.25, 0.1):
                u = np.where(is_land, 1 + np.round(ref / step), 0).astype(np.uint16)
                decoded = np.where(u > 0, (u.astype(np.float32) - 1) * step, 0)
                im16 = Image.fromarray(u)  # mode I;16
                raw, best = png_bytes(im16)
                nm = "gray16_step%gm.png" % step
                record(nm, best, decoded, "sea=0 reserved, land=1+round(h/step)")
                save(nm, best)
                # Same values packed hi/lo into R,G for WebP lossless.
                hi, lo = (u >> 8).astype(np.uint8), (u & 255).astype(np.uint8)
                rg = np.stack([hi, lo, np.zeros_like(hi)], -1)
                wb = webp_bytes(Image.fromarray(rg))
                enc["rg16_step%gm.webp" % step] = dict(enc[nm])
                enc["rg16_step%gm.webp" % step]["bytes"] = len(wb)
                save("rg16_step%gm.webp" % step, wb)
                # Raw little-endian uint16, row-delta, then general compressors.
                d = u.astype(np.int32)
                dd = np.diff(d, axis=1, prepend=0).astype(np.int16)
                rawb, deltab = u.tobytes(), dd.tobytes()
                enc["u16raw_step%gm.bin" % step] = dict(enc[nm])
                enc["u16raw_step%gm.bin" % step].update({
                    "bytes": len(brotli.compress(rawb, quality=11, lgwin=24)),
                    "note": "raw uint16 + brotli q11 (served with Content-Encoding: br)",
                    "zstd19": len(zstandard.ZstdCompressor(level=19).compress(rawb)),
                    "gzip9": len(zlib.compress(rawb, 9))})
                enc["u16delta_step%gm.bin" % step] = dict(enc[nm])
                enc["u16delta_step%gm.bin" % step].update({
                    "bytes": len(brotli.compress(deltab, quality=11, lgwin=24)),
                    "note": "row-delta int16 + brotli q11",
                    "zstd19": len(zstandard.ZstdCompressor(level=19).compress(deltab)),
                    "gzip9": len(zlib.compress(deltab, 9)),
                    "xz": len(lzma.compress(deltab, preset=9))})

            # 4. Half float (proxy for KTX2 R16F + zstd supercompression).
            f16 = np.where(is_land, ref, -1.0).astype(np.float16)
            dec = np.where(is_land, f16.astype(np.float32), 0)
            b16 = f16.tobytes()
            record("f16.bin", zstandard.ZstdCompressor(level=19).compress(b16), dec,
                   "float16, sea=-1, zstd 19 (KTX2 R16_SFLOAT + zstd proxy)")
            enc["f16.bin"]["brotli11"] = len(brotli.compress(b16, quality=11, lgwin=24))

            # 5. 8-bit curves, sea reserved as 0, land 1..255.
            curves = {
                "linear": (lambda h: h / HM, lambda t: t * HM),
                "sqrt": (lambda h: np.sqrt(h / HM), lambda t: (t ** 2) * HM),
                "log": (lambda h: np.log1p(h) / math.log1p(HM), lambda t: np.expm1(t * math.log1p(HM))),
            }
            for cname, (fwd, inv) in curves.items():
                t = fwd(np.maximum(ref, 0))
                u8 = np.where(is_land, 1 + np.round(t * 254), 0).astype(np.uint8)
                decoded = np.where(u8 > 0, inv((u8.astype(np.float32) - 1) / 254.0), 0).astype(np.float32)
                raw, best = png_bytes(Image.fromarray(u8))
                nm = "gray8_%s.png" % cname
                dd = record(nm, best, decoded, "sea=0 reserved, land 1..255 over 0..%d m" % HM)
                # analytic step sizes (metres per code) at given heights
                steps = {}
                for hh in (0, 5, 15, 30, 200, 1000, 2500):
                    t0 = float(fwd(np.float64(hh)))
                    steps[hh] = round(float(inv(np.float64(min(t0 + 1 / 254.0, 1.0)))) - hh, 3)
                dd["step_m_at"] = steps
                save(nm, best)
                wb = webp_bytes(Image.fromarray(u8).convert("RGB"))
                enc[nm.replace(".png", ".webp")] = dict(dd)
                enc[nm.replace(".png", ".webp")]["bytes"] = len(wb)
                save(nm.replace(".png", ".webp"), wb)

        # 6. sqrt curve at N bits in a 16-bit container (gray16 PNG, RG WebP,
        #    row-delta + brotli). Fine steps in the plain, coarse in mountains.
        for bits in a.sqrt_bits:
            top = (1 << bits) - 2
            t = np.sqrt(np.maximum(ref, 0) / HM)
            u = np.where(is_land, 1 + np.round(t * top), 0).astype(np.uint16)
            decoded = np.where(u > 0, (((u.astype(np.float64) - 1) / top) ** 2) * HM, 0).astype(np.float32)
            raw, best = png_bytes(Image.fromarray(u))
            nm = "sqrt%d_gray16.png" % bits
            dd = record(nm, best, decoded, "sqrt curve, %d bits, sea=0 reserved" % bits)
            steps = {}
            for hh in (0, 5, 15, 30, 200, 1000, 2500):
                tt = math.sqrt(hh / HM)
                steps[hh] = round(((tt + 1.0 / top) ** 2) * HM - hh, 3)
            dd["step_m_at"] = steps
            save(nm, best)
            hi, lo = (u >> 8).astype(np.uint8), (u & 255).astype(np.uint8)
            rg = np.stack([hi, lo, np.zeros_like(hi)], -1)
            wb = webp_bytes(Image.fromarray(rg))
            enc["sqrt%d_rg.webp" % bits] = dict(dd)
            enc["sqrt%d_rg.webp" % bits]["bytes"] = len(wb)
            save("sqrt%d_rg.webp" % bits, wb)
            dlt = np.diff(u.astype(np.int32), axis=1, prepend=0).astype(np.int16).tobytes()
            enc["sqrt%d_delta.bin" % bits] = dict(dd)
            enc["sqrt%d_delta.bin" % bits]["bytes"] = len(brotli.compress(dlt, quality=11, lgwin=24))
            enc["sqrt%d_delta.bin" % bits]["note"] = "row-delta int16 + brotli q11"
            sz = bin_sizes(med_residual(u))
            enc["sqrt%d_med.bin" % bits] = dict(dd)
            enc["sqrt%d_med.bin" % bits].update({"bytes": min(sz["br11"], sz["br11_planar"]),
                                                 "note": "JPEG-LS MED residual int16 + brotli q11", **sz})

        # 7. Thailand priority: full precision inside Thailand plus a fade band
        #    of about --band-km, neighbours blurred and coarsely quantised.
        if a.tha:
            km_px = height_km / H
            tmask = raster_land(a.tha, bbox, H, Wd, ss=2, only_a3="THA")
            band_px = max(1, int(round(a.band_km / km_px)))
            wgt = np.clip(box_blur(tmask, band_px) * 2.0, 0.0, 1.0)
            wgt = np.maximum(wgt, (tmask >= 0.5).astype(np.float32))
            r_px = a.nb_blur
            nb = box_blur(box_blur(box_blur(ref, r_px), r_px), r_px) if r_px > 0 else ref.copy()
            nb = np.where(is_land, np.maximum(nb, 0), 0)
            res["priority"] = {"band_px": band_px, "nb_blur_px": a.nb_blur, "tha_px_share": round(float((tmask >= 0.5).mean()), 4),
                               "full_weight_share": round(float((wgt >= 0.999).mean()), 4)}
            for bits, coarse in [(b, {12: 8, 11: 4, 10: 2}[b]) for b in a.prio_bits]:
                top = (1 << bits) - 2
                fine = np.round(np.sqrt(np.maximum(ref, 0) / HM) * top)
                crs = np.round(np.round(np.sqrt(nb / HM) * top) / coarse) * coarse
                v = np.where(wgt >= 0.999, fine, np.round(wgt * fine + (1 - wgt) * crs))
                u = np.where(is_land, 1 + v, 0).astype(np.uint16)
                decoded = np.where(u > 0, (((u.astype(np.float64) - 1) / top) ** 2) * HM, 0).astype(np.float32)
                raw, best = png_bytes(Image.fromarray(u))
                nm = "prio_sqrt%d_gray16.png" % bits
                dd = record(nm, best, decoded, "Thailand-priority, sqrt %d bits, neighbours blurred and /%d" % (bits, coarse))
                nbm = is_land & (wgt < 0.001)
                dd["neighbour_land_rmse_m"] = round(float(np.sqrt(np.mean((decoded - ref)[nbm] ** 2))), 2)
                thm = tmask >= 0.5
                dd["thai_land_rmse_m"] = round(float(np.sqrt(np.mean((decoded - ref)[thm & is_land] ** 2))), 3)
                save(nm, best)
                hi, lo = (u >> 8).astype(np.uint8), (u & 255).astype(np.uint8)
                wb = webp_bytes(Image.fromarray(np.stack([hi, lo, np.zeros_like(hi)], -1)))
                enc["prio_sqrt%d_rg.webp" % bits] = dict(dd)
                enc["prio_sqrt%d_rg.webp" % bits]["bytes"] = len(wb)
                save("prio_sqrt%d_rg.webp" % bits, wb)
                dlt = np.diff(u.astype(np.int32), axis=1, prepend=0).astype(np.int16)
                enc["prio_sqrt%d_delta.bin" % bits] = dict(dd)
                enc["prio_sqrt%d_delta.bin" % bits].update({"note": "row-delta int16", **bin_sizes(dlt)})
                enc["prio_sqrt%d_delta.bin" % bits]["bytes"] = enc["prio_sqrt%d_delta.bin" % bits]["br11"]
                sz = bin_sizes(med_residual(u))
                enc["prio_sqrt%d_med.bin" % bits] = dict(dd)
                enc["prio_sqrt%d_med.bin" % bits].update({"note": "JPEG-LS MED residual int16", **sz})
                enc["prio_sqrt%d_med.bin" % bits]["bytes"] = min(sz["br11"], sz["br11_planar"])
                if bits == 11:
                    with open(os.path.join(a.out, "h%d_prio_sqrt11_med_planar.bin" % H), "wb") as f:
                        v = med_residual(u).astype(np.int32) & 0xFFFF
                        f.write((v >> 8).astype(np.uint8).tobytes() + (v & 255).astype(np.uint8).tobytes())

        res["encodings"] = enc
        results.append(res)

        # Preview hillshade (NW light) for a visual sanity check of the stitch.
        if H == a.heights[-1]:
            gy, gx = np.gradient(ref * 10.0, height_km * 1000 / H)
            az, alt = math.radians(315), math.radians(45)
            slope = np.arctan(np.hypot(gx, gy))
            aspect = np.arctan2(-gx, gy)
            hs = np.sin(alt) * np.cos(slope) + np.cos(alt) * np.sin(slope) * np.cos(az - aspect)
            img = np.clip(hs * 255, 0, 255).astype(np.uint8)
            img[~is_land] = 30
            Image.fromarray(img).save(os.path.join(a.out, "preview_hillshade_%d.png" % H))

    rep["results"] = results
    print(json.dumps(rep, indent=1))


if __name__ == "__main__":
    main()
