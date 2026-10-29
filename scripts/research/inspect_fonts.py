"""Inspect Thai + Latin font files for the typography shortlist.

Usage (from the project root, isolated mode, font paths as arguments):

    .cache/venv/bin/python -I scripts/research/inspect_fonts.py \
        [--var wght=400,wdth=100] [--json OUT.json] FONT [FONT ...]

Reports, per font file:
  - name table: family, version, designer, manufacturer, licence text and URL
  - static or variable, fvar axes and named instances
  - Thai, Basic Latin, Latin-1 and selected symbol coverage
  - GSUB/GPOS feature tags per script (thai, latn, DFLT)
  - whether tnum/lnum/pnum/onum/case/zero actually change digits when
    shaped with HarfBuzz, both in a Latin run and in a Thai run (digits
    inside a Thai sentence are itemised as Thai script by browsers)
  - vertical metrics, USE_TYPO_METRICS
  - ink extents after real shaping (GSUB ccmp + GPOS mark/mkmk) of Thai
    test strings and a stress corpus of stacked marks, and the CSS
    line-heights at which that ink leaves the line box
  - a loop heuristic: how many of 17 letters whose looped forms carry a
    closed head loop have a counter (a contour wound opposite to the
    outer contour). Open loops are missed, so confirm visually with
    render_specimen.py
  - Thai and Latin proportions and the advance width of a sample sentence

The script only reads the files passed in. It writes nothing unless
--json is given.
"""

import argparse
import json
import sys

import uharfbuzz as hb
from fontTools.pens.areaPen import AreaPen
from fontTools.pens.recordingPen import RecordingPen
from fontTools.ttLib import TTFont

THAI_ASSIGNED = list(range(0x0E01, 0x0E3B)) + list(range(0x0E3F, 0x0E5C))
BASIC_LATIN = list(range(0x20, 0x7F))
LATIN1 = list(range(0xA0, 0x100))
SYMBOLS = {
    0x00A0: "no-break space",
    0x00B0: "degree",
    0x00B1: "plus-minus",
    0x00B3: "superscript three",
    0x00B7: "middle dot",
    0x00D7: "multiplication",
    0x2009: "thin space",
    0x200B: "zero width space",
    0x202F: "narrow no-break space",
    0x2013: "en dash",
    0x2018: "left single quote",
    0x2019: "right single quote",
    0x201C: "left double quote",
    0x201D: "right double quote",
    0x2026: "ellipsis",
    0x2191: "up arrow",
    0x2193: "down arrow",
    0x2212: "minus sign",
    0x25CC: "dotted circle",
}

BRIEF_STRINGS = ["น้ำล้นตลิ่ง", "ผู้ใหญ่บ้าน", "ปิ่นเกล้า", "จังหวัดพระนครศรีอยุธยา"]
NATIONAL_SENTENCE = (
    "ตอนนี้ 22 สถานีน้ำล้นตลิ่ง ส่วนใหญ่อยู่ลุ่มเจ้าพระยาตอนล่างและภาคใต้ตอนบน"
)
# Real words: tall consonant + upper vowel + tone, and deep below-base forms.
REAL_WORDS = BRIEF_STRINGS + [
    NATIONAL_SENTENCE,
    "ปั๊มน้ำ", "ฝั่งแม่น้ำ", "ฟื้นฟู", "ปี่", "ป้อม", "ฝั่งธนบุรี", "ป่าสัก",
    "แม่น้ำปิง", "อ.ศรีสัชนาลัย", "จ.ปราจีนบุรี", "เขื่อนภูมิพล",
    "กฎหมาย", "ปฏิบัติ", "ปฏิทิน", "ฐาน", "ญาติ", "ฤดูฝน", "บุญ", "ครู",
    "ฎีกา", "อุทกวิทยา", "กุ้ง", "ปู่", "ฟู้ด", "ญี่ปุ่น", "ฤๅษี",
]
TALL = "ปฝฟฬ"
UPPER_V = ["", "ั", "ิ", "ี", "ึ", "ื", "็"]
TONES = ["", "่", "้", "๊", "๋", "์"]
STRESS = []
for c in TALL + "กนส":
    for v in UPPER_V:
        for t in TONES:
            if v or t:
                STRESS.append(c + v + t)
    for t in TONES[1:]:
        STRESS.append(c + t + "ำ")
for c in "ฎฏฐญฤฦปกรฬ":
    STRESS.append(c)
    for v in ["ุ", "ู", "ฺ"]:
        STRESS.append(c + v)
        for t in ["่", "้", "๊", "๋"]:
            STRESS.append(c + v + t)
STRESS += ["ฤๅ", "ฦๅ"]
LATIN_STRINGS = ["Hxdgjpqy", "ÁÉÎÕÜÅÇ", "0123456789", "Mahanathi 14:00"]

DIGITS = "0123456789"
NUM_FEATURES = ["tnum", "lnum", "pnum", "onum", "zero", "case"]
# "case" is tested on punctuation between capitals, not on digits.
CASE_TEST = "(H)-[H]:H\u00b7H"
# Letters whose looped forms have a closed head loop (a counter). Their
# loopless forms usually have no counter at all.
LOOP_TEST = "บปนพฟมวหยลสรคถภดต"


def name(font, nid):
    rec = font["name"].getName(nid, 3, 1, 0x409) or font["name"].getName(nid, 1, 0, 0)
    return str(rec) if rec else None


def script_features(font, tag):
    out = {}
    if tag not in font:
        return out
    table = font[tag].table
    if not table.ScriptList or not table.FeatureList:
        return out
    feats = table.FeatureList.FeatureRecord
    for srec in table.ScriptList.ScriptRecord:
        tags = set()
        langsys = [srec.Script.DefaultLangSys] if srec.Script.DefaultLangSys else []
        langsys += [l.LangSys for l in srec.Script.LangSysRecord]
        for ls in langsys:
            for idx in ls.FeatureIndex:
                tags.add(feats[idx].FeatureTag)
        out[srec.ScriptTag] = sorted(tags)
    return out


def hb_font(path, variations):
    blob = hb.Blob.from_file_path(path)
    face = hb.Face(blob)
    font = hb.Font(face)
    if variations:
        font.set_variations(variations)
    return font


def shape(font, text, script="Thai", lang="th", features=None):
    buf = hb.Buffer()
    buf.add_str(text)
    buf.direction = "ltr"
    buf.script = script
    buf.language = lang
    hb.shape(font, buf, features or {})
    return buf.glyph_infos, buf.glyph_positions


def ink(font, text, script="Thai", lang="th"):
    infos, poss = shape(font, text, script, lang)
    top, bottom, notdef = None, None, 0
    for info, pos in zip(infos, poss):
        if info.codepoint == 0:
            notdef += 1
        ext = font.get_glyph_extents(info.codepoint)
        if ext is None or ext.height == 0:
            continue
        t = pos.y_offset + ext.y_bearing
        b = t + ext.height
        top = t if top is None else max(top, t)
        bottom = b if bottom is None else min(bottom, b)
    return top, bottom, notdef


def corpus_extent(font, strings, script="Thai", lang="th"):
    best_top, best_bottom = (None, None), (None, None)
    notdef_strings = []
    for s in strings:
        t, b, nd = ink(font, s, script, lang)
        if nd:
            notdef_strings.append(s)
        if t is not None and (best_top[0] is None or t > best_top[0]):
            best_top = (t, s)
        if b is not None and (best_bottom[0] is None or b < best_bottom[0]):
            best_bottom = (b, s)
    return {
        "max_top": best_top[0],
        "max_top_string": best_top[1],
        "min_bottom": best_bottom[0],
        "min_bottom_string": best_bottom[1],
        "strings_with_notdef": notdef_strings,
    }


def line_height_limits(top, bottom, asc, desc, upm):
    """CSS line-heights (unitless) below which shaped ink leaves the line box.

    The line box is the content area (asc + |desc|) plus half-leading on each
    side. ink_box: smallest line-height where top and bottom ink both stay
    inside the line box. collide: smallest line-height where the deepest
    below-mark of one line cannot touch the tallest stack of the next line.
    """
    d = abs(desc)
    l_top = (2 * top - asc + d) / upm
    l_bot = (2 * abs(bottom) - d + asc) / upm
    return {
        "ink_box": round(max(l_top, l_bot), 3),
        "collide": round((top + abs(bottom)) / upm, 3),
    }


def digit_test(font, script, lang, feats):
    """Return (glyph ids, advances) for 0-9 shaped with the given features."""
    infos, poss = shape(font, DIGITS, script, lang, feats)
    return [i.codepoint for i in infos], [p.x_advance for p in poss]


def numeric_features(font):
    res = {}
    for script, lang in (("Latn", "en"), ("Thai", "th")):
        base_g, base_a = digit_test(font, script, lang, {})
        row = {"default_advances": base_a,
               "default_tabular": len(set(base_a)) == 1}
        for f in NUM_FEATURES:
            if f == "case":
                # glyph ids and vertical offsets of punctuation between caps
                def sig(feats):
                    infos, poss = shape(font, CASE_TEST, script, lang, feats)
                    return [(i.codepoint, p.y_offset) for i, p in zip(infos, poss)]
                row[f] = {"effective": sig({}) != sig({"case": True}),
                          "tabular_after": None}
                continue
            g, a = digit_test(font, script, lang, {f: True})
            row[f] = {
                "effective": g != base_g or a != base_a,
                "tabular_after": len(set(a)) == 1,
            }
        # tnum + lnum together, as the brief's font-variant-numeric sets.
        g, a = digit_test(font, script, lang, {"tnum": True, "lnum": True})
        row["tnum+lnum_tabular"] = len(set(a)) == 1
        res[script] = row
    # Thai digits
    infos, poss = shape(font, "๐๑๒๓๔๕๖๗๘๙", "Thai", "th")
    res["thai_digit_advances"] = [p.x_advance for p in poss]
    infos, poss = shape(font, "๐๑๒๓๔๕๖๗๘๙", "Thai", "th", {"tnum": True})
    res["thai_digits_tnum_tabular"] = len(set(p.x_advance for p in poss)) == 1
    return res


def contour_areas(glyphset, gname):
    """Signed area of each contour of a glyph."""
    pen = RecordingPen()
    glyphset[gname].draw(pen)
    areas, cur = [], None
    for op, args in pen.value:
        if op == "moveTo":
            cur = AreaPen()
            cur.moveTo(*args)
        elif op in ("closePath", "endPath"):
            if cur is not None:
                cur.closePath()
                areas.append(cur.value)
            cur = None
        else:
            getattr(cur, op)(*args)
    return areas


def hole_count(glyphset, gname):
    """Contours wound opposite to the largest contour, i.e. counters.

    Overlapping contours in variable fonts share the outer winding, so
    they are not counted. A looped Thai letter's head loop is a counter.
    """
    areas = contour_areas(glyphset, gname)
    if not areas:
        return 0
    outer = max(areas, key=abs)
    return sum(1 for a in areas if a * outer < 0 and abs(a) > 1)


def loop_heuristic(tt, location):
    cmap = tt.getBestCmap()
    gs = tt.getGlyphSet(location=location) if location else tt.getGlyphSet()
    counts = {}
    for ch in LOOP_TEST:
        g = cmap.get(ord(ch))
        if g:
            counts[ch] = hole_count(gs, g)
    looped = sum(1 for v in counts.values() if v >= 1)
    return {"letters": "".join(counts), "with_counter": looped,
            "total": len(counts),
            "per_letter": "".join(str(min(v, 9)) for v in counts.values())}


def proportions(font, upm):
    def top_of(ch, script="Thai", lang="th"):
        t, b, _ = ink(font, ch, script, lang)
        return t
    x_h = top_of("x", "Latn", "en")
    cap = top_of("H", "Latn", "en")
    asc_l = top_of("d", "Latn", "en")
    thai_body = top_of("น")
    thai_tall = top_of("ป")
    infos, poss = shape(font, NATIONAL_SENTENCE)
    width = sum(p.x_advance for p in poss)
    infos, poss = shape(font, "The river gauge at Bang Sai reads 3.42 m", "Latn", "en")
    lwidth = sum(p.x_advance for p in poss)
    r = lambda v: None if v is None else round(v / upm, 3)
    return {
        "latin_x_height": r(x_h), "latin_cap": r(cap), "latin_ascender": r(asc_l),
        "thai_body_na": r(thai_body), "thai_tall_pa": r(thai_tall),
        "thai_body_over_x_height": None if not (x_h and thai_body) else round(thai_body / x_h, 3),
        "national_sentence_advance_em": r(width),
        "latin_sample_advance_em": r(lwidth),
    }


def inspect(path, variations):
    tt = TTFont(path, lazy=True)
    upm = tt["head"].unitsPerEm
    cmap = tt.getBestCmap()
    os2 = tt["OS/2"]
    hhea = tt["hhea"]
    is_var = "fvar" in tt
    loc = None
    if is_var:
        loc = {a.axisTag: a.defaultValue for a in tt["fvar"].axes}
        loc.update(variations or {})
    hbf = hb_font(path, loc if is_var else None)

    def missing(cps):
        return [f"U+{c:04X}" for c in cps if c not in cmap]

    info = {
        "file": path,
        "family": name(tt, 16) or name(tt, 1),
        "subfamily": name(tt, 17) or name(tt, 2),
        "version": name(tt, 5),
        "head_fontRevision": round(tt["head"].fontRevision, 3),
        "designer": name(tt, 9),
        "manufacturer": name(tt, 8),
        "vendor_id": os2.achVendID,
        "copyright": name(tt, 0),
        "license": name(tt, 13),
        "license_url": name(tt, 14),
        "outline": "CFF" if "CFF " in tt or "CFF2" in tt else "glyf",
        "num_glyphs": tt["maxp"].numGlyphs,
        "variable": is_var,
        "axes": [
            {"tag": a.axisTag, "min": a.minValue, "default": a.defaultValue, "max": a.maxValue}
            for a in tt["fvar"].axes
        ] if is_var else None,
        "named_instances": [
            name(tt, i.subfamilyNameID) for i in tt["fvar"].instances
        ] if is_var else None,
        "measured_at": loc,
        "usWeightClass": os2.usWeightClass,
        "coverage": {
            "thai": f"{sum(1 for c in THAI_ASSIGNED if c in cmap)}/{len(THAI_ASSIGNED)}",
            "thai_missing": missing(THAI_ASSIGNED),
            "basic_latin": f"{sum(1 for c in BASIC_LATIN if c in cmap)}/{len(BASIC_LATIN)}",
            "basic_latin_missing": missing(BASIC_LATIN),
            "latin1": f"{sum(1 for c in LATIN1 if c in cmap)}/{len(LATIN1)}",
            "symbols_missing": [f"U+{c:04X} {n}" for c, n in SYMBOLS.items() if c not in cmap],
            "total_codepoints": len(cmap),
        },
        "gsub_by_script": script_features(tt, "GSUB"),
        "gpos_by_script": script_features(tt, "GPOS"),
        "numeric": numeric_features(hbf),
        "metrics": {
            "unitsPerEm": upm,
            "hhea_ascender": hhea.ascent,
            "hhea_descender": hhea.descent,
            "hhea_lineGap": hhea.lineGap,
            "typo_ascender": os2.sTypoAscender,
            "typo_descender": os2.sTypoDescender,
            "typo_lineGap": os2.sTypoLineGap,
            "win_ascent": os2.usWinAscent,
            "win_descent": os2.usWinDescent,
            "USE_TYPO_METRICS": bool(os2.fsSelection & (1 << 7)),
            "sxHeight": getattr(os2, "sxHeight", None),
            "sCapHeight": getattr(os2, "sCapHeight", None),
            "hhea_normal_line_height": round(
                (hhea.ascent - hhea.descent + hhea.lineGap) / upm, 3),
        },
    }
    real = corpus_extent(hbf, REAL_WORDS)
    stress = corpus_extent(hbf, STRESS)
    latin = corpus_extent(hbf, LATIN_STRINGS, "Latn", "en")
    asc, desc = hhea.ascent, hhea.descent
    for block in (real, stress, latin):
        if block["max_top"] is not None and block["min_bottom"] is not None:
            block["max_top_em"] = round(block["max_top"] / upm, 3)
            block["min_bottom_em"] = round(block["min_bottom"] / upm, 3)
            block["top_over_hhea_ascender"] = block["max_top"] - asc
            block["bottom_below_hhea_descender"] = desc - block["min_bottom"]
            block["line_height_limits"] = line_height_limits(
                block["max_top"], block["min_bottom"], asc, desc, upm)
    info["ink_real_words"] = real
    info["ink_stress"] = stress
    info["ink_latin"] = latin
    brief = {}
    for s in BRIEF_STRINGS:
        t, b, nd = ink(hbf, s)
        brief[s] = {"top_em": round(t / upm, 3), "bottom_em": round(b / upm, 3),
                    "notdef": nd}
    info["brief_strings"] = brief
    info["loop_heuristic"] = loop_heuristic(tt, loc if is_var else None)
    info["proportions"] = proportions(hbf, upm)
    return info


def summary(i):
    m = i["metrics"]
    n = i["numeric"]
    lines = [
        f"== {i['family']} {i['subfamily']}  ({i['file']})",
        f"   version: {i['version']} | designer: {i['designer']} | manufacturer: {i['manufacturer']} | vendor: {i['vendor_id']}",
        f"   licence: {(i['license'] or '')[:110]!r} | url: {i['license_url']}",
        f"   outline: {i['outline']} | glyphs: {i['num_glyphs']} | variable: {i['variable']} axes: {i['axes']} measured_at: {i['measured_at']}",
        f"   coverage: thai {i['coverage']['thai']} missing {i['coverage']['thai_missing']} | basic latin {i['coverage']['basic_latin']} | latin-1 {i['coverage']['latin1']} | symbols missing {i['coverage']['symbols_missing']}",
        f"   GSUB: {i['gsub_by_script']}",
        f"   GPOS: {i['gpos_by_script']}",
    ]
    for s in ("Latn", "Thai"):
        r = n[s]
        eff = ", ".join(f"{f}:{'Y' if r[f]['effective'] else 'n'}" for f in NUM_FEATURES)
        lines.append(f"   digits in {s} run: default tabular={r['default_tabular']} adv={r['default_advances']} | {eff} | tnum+lnum tabular={r['tnum+lnum_tabular']}")
    lines.append(f"   thai digits adv={n['thai_digit_advances']} tnum tabular={n['thai_digits_tnum_tabular']}")
    lines.append(
        f"   metrics: upm {m['unitsPerEm']} hhea {m['hhea_ascender']}/{m['hhea_descender']}/{m['hhea_lineGap']} "
        f"typo {m['typo_ascender']}/{m['typo_descender']}/{m['typo_lineGap']} win {m['win_ascent']}/{m['win_descent']} "
        f"USE_TYPO {m['USE_TYPO_METRICS']} xH {m['sxHeight']} capH {m['sCapHeight']} normal-lh {m['hhea_normal_line_height']}")
    for k in ("ink_real_words", "ink_stress", "ink_latin"):
        b = i[k]
        lines.append(
            f"   {k}: top {b.get('max_top')} ({b.get('max_top_em')} em, {b.get('max_top_string')!r}, {b.get('top_over_hhea_ascender')} over hhea asc) "
            f"bottom {b.get('min_bottom')} ({b.get('min_bottom_em')} em, {b.get('min_bottom_string')!r}, {b.get('bottom_below_hhea_descender')} below hhea desc) "
            f"lh-limits {b.get('line_height_limits')} notdef {b['strings_with_notdef'][:5]}")
    lines.append(f"   brief strings: {i['brief_strings']}")
    lines.append(f"   loop heuristic: {i['loop_heuristic']}")
    lines.append(f"   proportions: {i['proportions']}")
    return "\n".join(lines)


def main(argv):
    ap = argparse.ArgumentParser()
    ap.add_argument("fonts", nargs="+")
    ap.add_argument("--var", default="", help="e.g. wght=400,wdth=100")
    ap.add_argument("--json", default=None)
    a = ap.parse_args(argv)
    variations = {}
    if a.var:
        for kv in a.var.split(","):
            k, v = kv.split("=")
            variations[k] = float(v)
    results = []
    for p in a.fonts:
        info = inspect(p, variations)
        results.append(info)
        print(summary(info))
        print()
    if a.json:
        with open(a.json, "w", encoding="utf-8") as fh:
            json.dump(results, fh, ensure_ascii=False, indent=1)


if __name__ == "__main__":
    main(sys.argv[1:])
