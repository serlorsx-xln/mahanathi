"""Print a compact table from inspect_fonts.py JSON reports.

Usage:
    .cache/venv/bin/python -I scripts/research/summarize_reports.py REPORT.json [...]
"""

import json
import sys


def yn(v):
    return "Y" if v else "n"


def main(paths):
    for p in paths:
        with open(p, encoding="utf-8") as fh:
            rows = json.load(fh)
        for i in rows:
            m, n, c = i["metrics"], i["numeric"], i["coverage"]
            lat, tha = n["Latn"], n["Thai"]
            gs = set(i["gsub_by_script"].get("thai", [])) | set(i["gsub_by_script"].get("latn", [])) | set(i["gsub_by_script"].get("DFLT", []))
            gp_thai = i["gpos_by_script"].get("thai", [])
            real, stress = i["ink_real_words"], i["ink_stress"]
            lp = i["loop_heuristic"]
            print(f"== {i['family']} {i['subfamily']} | {i['version']} | {i['designer']} / {i['manufacturer']} | glyphs {i['num_glyphs']} | var {i['axes']}")
            print(f"   lic: {(i['license'] or '')[:75]!r} | thai {c['thai']} latin {c['basic_latin']} latin1 {c['latin1']} | missing sym: {[s.split(' ',1)[1] for s in c['symbols_missing']]}")
            print(f"   num feats in GSUB: {sorted(x for x in gs if x in ('tnum','lnum','pnum','onum','case','zero','ss01','ss02','ss03','ss04','ss05','salt','calt'))} | thai GPOS {gp_thai} | thai GSUB ccmp {yn('ccmp' in i['gsub_by_script'].get('thai', []))}")
            print(f"   digits default tabular {yn(lat['default_tabular'])} adv {sorted(set(lat['default_advances']))} | tnum effective latn {yn(lat['tnum']['effective'])} thai {yn(tha['tnum']['effective'])} | onum {yn(lat['onum']['effective'])} pnum {yn(lat['pnum']['effective'])} case {yn(lat['case']['effective'])} | tnum+lnum tabular in thai run {yn(tha['tnum+lnum_tabular'])}")
            print(f"   metrics upm {m['unitsPerEm']} hhea {m['hhea_ascender']}/{m['hhea_descender']}/{m['hhea_lineGap']} typo {m['typo_ascender']}/{m['typo_descender']}/{m['typo_lineGap']} win {m['win_ascent']}/{m['win_descent']} USE_TYPO {yn(m['USE_TYPO_METRICS'])} normal-lh {m['hhea_normal_line_height']}")
            print(f"   real ink top {real.get('max_top_em')} ({real.get('max_top_string')[:14]!r}) bottom {real.get('min_bottom_em')} ({real.get('min_bottom_string')[:14]!r}) over-asc {real.get('top_over_hhea_ascender')} below-desc {real.get('bottom_below_hhea_descender')} | lh {real.get('line_height_limits')}")
            print(f"   stress ink top {stress.get('max_top_em')} ({stress.get('max_top_string')!r}) bottom {stress.get('min_bottom_em')} ({stress.get('min_bottom_string')!r}) | lh {stress.get('line_height_limits')}")
            print(f"   loop: {lp['with_counter']}/{lp['total']} letters with a counter | props {i['proportions']}")


if __name__ == "__main__":
    main(sys.argv[1:])
