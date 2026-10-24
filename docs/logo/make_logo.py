"""Draws the gobees logo as SVG, one version for dark and one for light backgrounds.

The icon shows the main concepts of the model: a colony (hexagon nest), its foraging range (dashed circle),
flower patches made of layers (concentric circles in habitat colours) and a bumblebee on its way to a patch.
The word mark is set in Go Bold, with the "o" drawn as a hexagon (gobees-logo-*.svg).

A second design draws the letters themselves (gobees-letters-*.svg): g and b with flower patches as bowls,
o as the nest, e as blossoms, s as the dotted flight path of a bumblebee.

Usage:  python make_logo.py [--font path/to/Go-Bold.ttf]
Needs:  fontTools (pip install fonttools). Without --font, Go-Bold.ttf is taken from the Go module cache
        (module golang.org/x/image).
"""

import argparse
import glob
import math
import os
import subprocess

from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.ttLib import TTFont

HERE = os.path.dirname(os.path.abspath(__file__))

GO_BLUE = "#00ADD8"  # Go brand colour
HABITATS = ["#E07F96", "#7C50A4", "#1D9F78"]  # NetLogo colours of grassland, improved grass, cereals
FLOWER = "#F5B72A"

THEMES = {
    "dark": {"honey": "#F5B72A", "ink": "#F2F2F2", "range": "#F5B72A", "stripe": "#1D1D1D"},
    "light": {"honey": "#E09A0B", "ink": "#1D1D1D", "range": "#C98A0A", "stripe": "#1D1D1D"},
}


def find_font():
    gomodcache = subprocess.run(["go", "env", "GOMODCACHE"], capture_output=True, text=True, check=True).stdout.strip()
    hits = sorted(glob.glob(os.path.join(gomodcache, "golang.org", "x", "image@*", "font", "gofont", "ttfs", "Go-Bold.ttf")))
    if not hits:
        raise SystemExit("Go-Bold.ttf not found in the Go module cache, pass --font")
    return hits[-1]


def hexagon(cx, cy, r, rot=30):
    pts = []
    for i in range(6):
        a = math.radians(60 * i + rot)
        pts.append((cx + r * math.cos(a), cy + r * math.sin(a)))
    return pts


def poly(pts):
    return "M" + " L".join(f"{x:.2f},{y:.2f}" for x, y in pts) + " Z"


def word_mark(font, text, x, baseline, size, colours, hex_o):
    """Returns SVG paths for text. colours: one colour per character. hex_o: draw "o" as a hexagon ring."""
    upem = font["head"].unitsPerEm
    s = size / upem
    glyphs = font.getGlyphSet()
    cmap = font.getBestCmap()
    glyf = font["glyf"]
    ring = 0.15 * size  # stroke width of the hexagon "o", about the stem width of Go Bold
    out = []
    for ch, col in zip(text, colours):
        name = cmap[ord(ch)]
        adv = font["hmtx"][name][0] * s
        if ch == "o" and hex_o:
            g = glyf[name]
            cx = x + (g.xMin + g.xMax) / 2 * s
            cy = baseline - (g.yMin + g.yMax) / 2 * s
            r = (g.yMax - g.yMin) / 2 * s * 1.04
            d = poly(hexagon(cx, cy, r)) + " " + poly(hexagon(cx, cy, r - ring / math.cos(math.radians(30))))
            out.append(f'<path d="{d}" fill="{col}" fill-rule="evenodd"/>')
        else:
            pen = SVGPathPen(glyphs)
            glyphs[name].draw(pen)
            out.append(f'<path transform="translate({x:.2f},{baseline:.2f}) scale({s:.5f},{-s:.5f})" '
                       f'd="{pen.getCommands()}" fill="{col}"/>')
        x += adv
    return "\n  ".join(out), x


def bee(cx, cy, size, angle, t):
    """A bumblebee centred at (cx, cy), heading in direction angle (degrees)."""
    w = size
    return f'''<g transform="translate({cx},{cy}) rotate({angle})">
    <ellipse cx="{-0.15*w:.1f}" cy="{-0.55*w:.1f}" rx="{0.42*w:.1f}" ry="{0.22*w:.1f}" transform="rotate(-25 {-0.15*w:.1f} {-0.55*w:.1f})" fill="{t['ink']}" fill-opacity="0.35" stroke="{t['ink']}" stroke-width="{0.05*w:.1f}"/>
    <ellipse cx="{0.15*w:.1f}" cy="{-0.5*w:.1f}" rx="{0.34*w:.1f}" ry="{0.18*w:.1f}" transform="rotate(-50 {0.15*w:.1f} {-0.5*w:.1f})" fill="{t['ink']}" fill-opacity="0.25" stroke="{t['ink']}" stroke-width="{0.05*w:.1f}"/>
    <ellipse cx="0" cy="0" rx="{0.62*w:.1f}" ry="{0.44*w:.1f}" fill="{t['honey']}" stroke="{t['ink']}" stroke-width="{0.06*w:.1f}"/>
    <path d="M{-0.12*w:.1f},{-0.43*w:.1f} Q{-0.2*w:.1f},0 {-0.12*w:.1f},{0.43*w:.1f} L{0.08*w:.1f},{0.43*w:.1f} Q{0.0*w:.1f},0 {0.08*w:.1f},{-0.43*w:.1f} Z" fill="{t['stripe']}"/>
    <path d="M{-0.48*w:.1f},{-0.28*w:.1f} Q{-0.56*w:.1f},0 {-0.48*w:.1f},{0.28*w:.1f} L{-0.36*w:.1f},{0.37*w:.1f} Q{-0.45*w:.1f},0 {-0.36*w:.1f},{-0.37*w:.1f} Z" fill="{t['stripe']}"/>
    <circle cx="{0.62*w:.1f}" cy="0" r="{0.26*w:.1f}" fill="{t['stripe']}" stroke="{t['ink']}" stroke-width="{0.06*w:.1f}"/>
    <path d="M{0.78*w:.1f},{-0.12*w:.1f} q{0.18*w:.1f},{-0.2*w:.1f} {0.3*w:.1f},{-0.22*w:.1f} M{0.78*w:.1f},{0.12*w:.1f} q{0.18*w:.1f},{0.2*w:.1f} {0.3*w:.1f},{0.22*w:.1f}" stroke="{t['ink']}" stroke-width="{0.05*w:.1f}" fill="none" stroke-linecap="round"/>
  </g>'''


def flower_patch(cx, cy, r, colour, t):
    """A flower patch: habitat circle with flower layers inside."""
    return (f'<circle cx="{cx}" cy="{cy}" r="{r}" fill="{colour}"/>\n  '
            f'<circle cx="{cx}" cy="{cy}" r="{r*0.62:.1f}" fill="{FLOWER}" stroke="{t["ink"]}" stroke-opacity="0.0"/>\n  '
            f'<circle cx="{cx}" cy="{cy}" r="{r*0.26:.1f}" fill="{colour}"/>')


def icon(t):
    cx, cy, R = 150, 150, 124
    parts = [
        # foraging range
        f'<circle cx="{cx}" cy="{cy}" r="{R}" fill="none" stroke="{t["range"]}" stroke-width="7" '
        f'stroke-dasharray="20 14" stroke-linecap="round" opacity="0.8"/>',
        # flower patches on the edge of the range
        flower_patch(cx + R * math.cos(math.radians(-28)), cy + R * math.sin(math.radians(-28)), 30, HABITATS[0], t),
        flower_patch(cx + R * math.cos(math.radians(205)), cy + R * math.sin(math.radians(205)), 22, HABITATS[2], t),
        flower_patch(cx + R * math.cos(math.radians(110)), cy + R * math.sin(math.radians(110)), 17, HABITATS[1], t),
        # nest: a small comb of three cells
        f'<path d="{poly(hexagon(cx - 26, cy + 26, 29))}" fill="{t["honey"]}" stroke="{t["ink"]}" stroke-width="5" stroke-linejoin="round"/>',
        f'<path d="{poly(hexagon(cx + 24.2, cy + 26, 29))}" fill="{t["honey"]}" stroke="{t["ink"]}" stroke-width="5" stroke-linejoin="round"/>',
        f'<path d="{poly(hexagon(cx - 0.9, cy - 17.5, 29))}" fill="{t["honey"]}" stroke="{t["ink"]}" stroke-width="5" stroke-linejoin="round"/>',
        # flight path from the nest towards the large patch
        f'<path d="M{cx - 2},{cy - 52} Q{cx - 6},{cy - 92} {cx + 12},{cy - 92}" fill="none" stroke="{t["ink"]}" '
        f'stroke-width="4" stroke-dasharray="2 10" stroke-linecap="round" opacity="0.8"/>',
        bee(cx + 42, cy - 80, 56, 20, t),
    ]
    return "\n  ".join(parts)


def logo(font, theme):
    t = THEMES[theme]
    text, end = word_mark(font, "gobees", 330, 205, 200, [GO_BLUE, GO_BLUE] + [t["honey"]] * 4, hex_o=True)
    width = math.ceil(end + 20)
    return f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} 300" width="{width}" height="300">
  <title>gobees</title>
  {icon(t)}
  {text}
</svg>
'''


# ---- Second design: the letters themselves show the model ------------------------------------------------
#
# g, b: the bowl is a flower patch (ring in a habitat colour with a yellow flower layer)
# o:    a hexagon, the nest
# e, e: blossoms, petals on a ring with the gap and the bar of an "e"
# s:    a dotted flight path, with the bumblebee taking off at its end

BASE, XH, STROKE = 210, 120, 28  # baseline, x-height, stroke width
MID = BASE - XH / 2
BOWL = XH / 2 - STROKE / 2  # centre-line radius of a round bowl


def ring(cx, colour, t):
    return (f'<circle cx="{cx}" cy="{MID}" r="{BOWL}" fill="none" stroke="{colour}" stroke-width="{STROKE}"/>\n  '
            f'<circle cx="{cx}" cy="{MID}" r="{BOWL}" fill="none" stroke="{FLOWER}" stroke-width="6"/>')


def letter_g(x, colour, t):
    cx = x + XH / 2
    sx = x + XH - STROKE / 2
    hook = f'M{sx},{MID - BOWL} L{sx},{BASE + 18} A{BOWL},{BOWL * 0.9} 0 0 1 {x + 14},{BASE + 30}'
    return (f'<path d="{hook}" fill="none" stroke="{colour}" stroke-width="{STROKE}" stroke-linecap="round"/>\n  '
            + ring(cx, colour, t)), x + XH


def letter_o(x, t):
    r = XH / 2 - STROKE * 0.42
    cx = x + r * math.cos(math.radians(30)) + STROKE / 2
    cell = poly(hexagon(cx, MID, r))
    return (f'<path d="{cell}" fill="none" stroke="{t["honey"]}" stroke-width="{STROKE * 0.85:.1f}" '
            f'stroke-linejoin="round"/>'), cx + r * math.cos(math.radians(30)) + STROKE / 2


def letter_b(x, colour, t):
    sx = x + STROKE / 2
    cx = x + XH / 2
    return (f'<path d="M{sx},{BASE - XH - 62} L{sx},{BASE - 4}" stroke="{colour}" stroke-width="{STROKE}" stroke-linecap="round"/>\n  '
            + ring(cx, colour, t)), x + XH


def letter_e(x, colour, t):
    cx = x + XH / 2
    parts = []
    # petals along the ring, leaving the gap of the "e" at the lower right
    for k in range(10):
        a = k * 36
        if a == 36:  # the gap of the "e" at the lower right
            continue
        px, py = cx + BOWL * math.cos(math.radians(a)), MID + BOWL * math.sin(math.radians(a))
        parts.append(f'<ellipse cx="{px:.1f}" cy="{py:.1f}" rx="{STROKE * 0.62:.1f}" ry="{STROKE * 0.5:.1f}" '
                     f'transform="rotate({a} {px:.1f} {py:.1f})" fill="{colour}"/>')
    parts.append(f'<path d="M{cx - BOWL},{MID} L{cx + BOWL},{MID}" stroke="{FLOWER}" '
                 f'stroke-width="{STROKE * 0.6:.1f}" stroke-linecap="round"/>')
    return "\n  ".join(parts), x + XH


def letter_s(x, t):
    w = XH * 0.78
    d = (f'M{x + 6},{BASE - 14} C{x + 30},{BASE + 4} {x + w + 4},{BASE - 2} {x + w - 4},{BASE - 36} '
         f'C{x + w - 12},{MID + 2} {x + 8},{MID + 8} {x + 8},{MID - 30} '
         f'C{x + 8},{BASE - XH + 2} {x + w - 14},{BASE - XH - 4} {x + w + 2},{BASE - XH + 10}')
    path = (f'<path d="{d}" fill="none" stroke="{GO_BLUE}" stroke-width="{STROKE * 0.75:.1f}" '
            f'stroke-dasharray="0 {STROKE * 0.95:.1f}" stroke-linecap="round"/>')
    return path + "\n  " + bee(x + w + 26, BASE - XH - 12, 58, -35, t), x + w + 70


def letters_logo(theme):
    t = THEMES[theme]
    gap = 22
    parts = []
    x = 20
    for draw in (lambda x: letter_g(x, GO_BLUE, t),
                 lambda x: letter_o(x, t),
                 lambda x: letter_b(x, HABITATS[0], t),
                 lambda x: letter_e(x, HABITATS[1], t),
                 lambda x: letter_e(x, HABITATS[2], t),
                 lambda x: letter_s(x, t)):
        svg, x = draw(x)
        parts.append(svg)
        x += gap
    width = math.ceil(x)
    return f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} 300" width="{width}" height="300">
  <title>gobees</title>
  {chr(10).join("  " + p for p in parts).lstrip()}
</svg>
'''


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--font", help="path to Go-Bold.ttf")
    args = ap.parse_args()
    font = TTFont(args.font or find_font())
    for theme in THEMES:
        for name, svg in ((f"gobees-logo-{theme}.svg", logo(font, theme)),
                          (f"gobees-letters-{theme}.svg", letters_logo(theme))):
            path = os.path.join(HERE, name)
            with open(path, "w", encoding="utf-8", newline="\n") as f:
                f.write(svg)
            print("wrote", path)


if __name__ == "__main__":
    main()
