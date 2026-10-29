#!/usr/bin/env bash
# Estimate self-hosted WOFF2 sizes for Thai + Basic Latin + digits/punctuation.
#
# Usage (from the project root):
#   scripts/research/subset_sizes.sh LABEL FONT [AXES]
#     LABEL  short name used for output files
#     FONT   path to a TTF/OTF
#     AXES   optional, for variable fonts: e.g. "wght=400 wdth=100".
#            When given, an instance is cut first with
#            fontTools.varLib.instancer (pin every axis for a static font, or
#            only some axes for a partial instance), and the full variable
#            font is also subset, so both sizes are reported.
#
# Runs pyftsubset (python -I -m fontTools.subset, the same entry point) from
# the project venv. Output goes to .cache/fonts/_subset/. Two numbers per
# file: all OpenType layout features kept with hinting, and the same without
# hinting. Codepoints a font lacks are skipped by pyftsubset.
set -euo pipefail

PY=".cache/venv/bin/python"
OUT=".cache/fonts/_subset"
mkdir -p "$OUT"

# Thai block (assigned), Basic Latin, and the symbols the product is likely to
# set: nbsp, degree, plus-minus, superscript three, middle dot, multiplication,
# thin/zero-width/narrow nbsp spaces, en dash, curly quotes, ellipsis, minus,
# dotted circle (shapers insert it for orphaned Thai marks).
UNICODES="U+0020-007E,U+00A0,U+00B0,U+00B1,U+00B3,U+00B7,U+00D7,U+0E01-0E3A,U+0E3F-0E5B,U+2009,U+200B,U+202F,U+2013,U+2018,U+2019,U+201C,U+201D,U+2026,U+2212,U+25CC"

label="$1"
font="$2"
axes="${3:-}"

subset() {
  local src="$1" dst="$2"
  "$PY" -I -m fontTools.subset "$src" --unicodes="$UNICODES" \
    --layout-features='*' --flavor=woff2 --output-file="$dst.hinted.woff2"
  "$PY" -I -m fontTools.subset "$src" --unicodes="$UNICODES" \
    --layout-features='*' --no-hinting --desubroutinize --flavor=woff2 \
    --output-file="$dst.unhinted.woff2"
  local h u
  h=$(wc -c < "$dst.hinted.woff2" | tr -d ' ')
  u=$(wc -c < "$dst.unhinted.woff2" | tr -d ' ')
  printf '%-48s hinted %7d B (%5.1f KB)  unhinted %7d B (%5.1f KB)\n' \
    "$(basename "$dst")" "$h" "$(echo "$h/1024" | bc -l)" "$u" "$(echo "$u/1024" | bc -l)"
}

if [ -n "$axes" ]; then
  inst="$OUT/$label.instance.ttf"
  # shellcheck disable=SC2086
  "$PY" -I -m fontTools.varLib.instancer "$font" $axes --output "$inst" >/dev/null
  subset "$inst" "$OUT/$label.instance"
  subset "$font" "$OUT/$label.variable"
else
  subset "$font" "$OUT/$label"
fi
