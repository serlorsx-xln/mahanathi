"""Promote self-hosted WOFF2 subsets into web/public/fonts/.

Cuts the specimen set from the cached source fonts with the same
pyftsubset options as the research (scripts/research/subset_sizes.sh),
renames the IBM Plex Sans Thai subsets so no Reserved Font Name
survives (OFL condition 3; subsets are modified versions), and copies
each family's OFL.txt beside them (OFL condition 2, "as files").

Usage (from the project root):
  .cache/venv/bin/python -I scripts/research/promote_fonts.py [FAMILY ...]

With no FAMILY argument it promotes the full specimen set; named
families (e.g. `mahanathi-sans noto-serif-thai`) promote only those.
The final pairing after ADR 0005 is promoted by naming its families.
"""
import os
import re
import shutil
import subprocess
import sys

PY = sys.executable
SRC = ".cache/fonts"
OUT = "web/public/fonts"

# Thai block (assigned), Basic Latin, and the symbols the product sets:
# the same set as subset_sizes.sh.
UNICODES = (
    "U+0020-007E,U+00A0,U+00B0,U+00B1,U+00B3,U+00B7,U+00D7,"
    "U+0E01-0E3A,U+0E3F-0E5B,U+2009,U+200B,U+202F,U+2013,U+2018,"
    "U+2019,U+201C,U+201D,U+2026,U+2212,U+25CC"
)

# key: css family, source file (relative to .cache/fonts), axes to
# instance, hinted or not, and whether the OFL rename applies.
# UI faces keep their hinting (small Thai text on Windows); the large
# sentence serifs drop it (the bytes buy nothing at display sizes).
SPECIMEN = [
    # family              source                                            axes              hinted  rename
    ("Mahanathi Sans", "ibmplexsansthai/IBMPlexSansThai-Regular.ttf", "", True, True),
    ("Mahanathi Sans", "ibmplexsansthai/IBMPlexSansThai-SemiBold.ttf", "", True, True),
    ("Anuphan", "anuphan/Anuphan[wght].ttf", "wght=400", True, False),
    ("Anuphan", "anuphan/Anuphan[wght].ttf", "wght=600", True, False),
    ("Noto Sans Thai", "notosansthai/NotoSansThai[wdth,wght].ttf", "wght=400 wdth=100", True, False),
    ("Noto Sans Thai", "notosansthai/NotoSansThai[wdth,wght].ttf", "wght=600 wdth=100", True, False),
    ("Noto Serif Thai", "notoserifthai/NotoSerifThai[wdth,wght].ttf", "wght=400 wdth=100", False, False),
    ("Pridi", "pridi/Pridi-Regular.ttf", "", False, False),
    ("Trirong", "trirong/Trirong-Regular.ttf", "", False, False),
]

# OFL condition 3: no Reserved Font Name in a modified version. The
# Plex name table carries "Plex" in every identifying record.
RENAMES = [
    ("IBM Plex Sans Thai", "Mahanathi Sans"),
    ("IBMPlexSansThai", "MahanathiSans"),
    ("Plex", "Mahanathi"),
]

OFL_FOR = {
    "Mahanathi Sans": "ibmplexsansthai",
    "Anuphan": "anuphan",
    "Noto Sans Thai": "notosansthai",
    "Noto Serif Thai": "notoserifthai",
    "Pridi": "pridi",
    "Trirong": "trirong",
}


def slug(family):
    return re.sub(r"[^a-z0-9]+", "-", family.lower()).strip("-")


def weight_of(axes, source, default="400"):
    for part in axes.split():
        if part.startswith("wght="):
            return part.split("=")[1]
    # Static sources carry the weight in their file name.
    for weight, suffix in ((600, "SemiBold"), (500, "Medium"), (700, "Bold")):
        if source.rsplit("/", 1)[-1].rsplit("-", 1)[-1].split(".")[0] == suffix:
            return str(weight)
    return default


def rename_rfn(path, family):
    from fontTools.ttLib import TTFont

    font = TTFont(path)
    table = font["name"]
    for record in table.names:
        if record.nameID not in (1, 3, 4, 6, 16, 17):
            continue
        try:
            text = record.toUnicode()
        except UnicodeDecodeError:
            continue
        new = text
        for old, repl in RENAMES:
            new = new.replace(old, repl)
        if new != text:
            record.string = new.encode(record.getEncoding())
    font.save(path)


def promote(family, source, axes, hinted, rename):
    weight = weight_of(axes, source)
    stem = "%s-%s" % (slug(family), weight)
    src = os.path.join(SRC, source)
    tmp = os.path.join(OUT, "." + stem + ".src.ttf")

    if axes:
        subprocess.run(
            [PY, "-I", "-m", "fontTools.varLib.instancer", src, *axes.split(), "--output", tmp],
            check=True, capture_output=True)
        src = tmp

    out = os.path.join(OUT, stem + ".woff2")
    cmd = [PY, "-I", "-m", "fontTools.subset", src,
           "--unicodes=" + UNICODES, "--layout-features=*", "--flavor=woff2",
           "--output-file=" + out]
    if not hinted:
        cmd += ["--no-hinting", "--desubroutinize"]
    subprocess.run(cmd, check=True, capture_output=True)
    if os.path.exists(tmp):
        os.remove(tmp)

    if rename:
        rename_rfn(out, family)
    return out


def main():
    wanted = set(sys.argv[1:])
    os.makedirs(OUT, exist_ok=True)
    done = set()
    for family, source, axes, hinted, rename in SPECIMEN:
        if wanted and slug(family) not in wanted and family not in wanted:
            continue
        out = promote(family, source, axes, hinted, rename)
        print("promoted %-32s %6d B  (axes: %s)" % (os.path.basename(out), os.path.getsize(out), axes or "static"))
        done.add(family)
    for family in sorted(done):
        ofl = os.path.join(SRC, OFL_FOR[family], "OFL.txt")
        shutil.copyfile(ofl, os.path.join(OUT, slug(family) + "-OFL.txt"))
        print("licence  %s-OFL.txt" % slug(family))


if __name__ == "__main__":
    main()
