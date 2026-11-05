"""Draws the gobees-fast logo as SVG, for dark backgrounds (gobees-fast-logo-dark.svg).

The icon shows the main concepts of the model: a colony (hexagon nest), its foraging range (dashed circle),
flower patches made of layers (concentric circles in habitat colours) and a bumblebee flying to a patch.
The word mark is "gobees" in Go Bold, with the "o" drawn as a hexagon, followed by "fast".
Three cues show speed:
- "fast" is set in Go Bold slanted forward, in Go blue,
- three speed lines in front of "fast", as in the Go brand logo,
- the bumblebee trails speed streaks.

Usage:  python make_fast_logo.py [--font path/to/Go-Bold.ttf]
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
HONEY = "#F5B72A"
INK = "#F2F2F2"  # outlines, light on the dark background
RANGE = "#F5B72A"
STRIPE = "#1D1D1D"

SLANT = 14  # forward slant of "fast" in degrees


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


def word_mark(font, text, x, baseline, size, colours):
    """Returns SVG paths for text and the end x. colours: one colour per character. "o" is a hexagon ring."""
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
        if ch == "o":
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


def slanted(font, text, x, baseline, size, colour):
    """Returns SVG paths for text in Go Bold, slanted forward by SLANT degrees, and the end x."""
    s = size / font["head"].unitsPerEm
    glyphs = font.getGlyphSet()
    cmap = font.getBestCmap()
    out = []
    for ch in text:
        name = cmap[ord(ch)]
        pen = SVGPathPen(glyphs)
        glyphs[name].draw(pen)
        out.append(f'<path transform="translate({x:.2f},{baseline:.2f}) skewX({-SLANT}) scale({s:.5f},{-s:.5f})" '
                   f'd="{pen.getCommands()}" fill="{colour}"/>')
        x += font["hmtx"][name][0] * s
    return "\n  ".join(out), x


def speed_lines(right, ys, lengths, width, colour):
    """Horizontal speed lines with round caps, right-aligned at x = right (shifted by the slant)."""
    out = []
    for y, length in zip(ys, lengths):
        r = right + (ys[-1] - y) * math.tan(math.radians(SLANT))  # upper lines end further right, along the slant
        out.append(f'<path d="M{r - length:.1f},{y} L{r:.1f},{y}" stroke="{colour}" stroke-width="{width}" '
                   f'stroke-linecap="round"/>')
    return "\n  ".join(out)


def bee(cx, cy, size, angle):
    """A bumblebee centred at (cx, cy), heading in direction angle (degrees)."""
    w = size
    return f'''<g transform="translate({cx},{cy}) rotate({angle})">
    <ellipse cx="{-0.15*w:.1f}" cy="{-0.55*w:.1f}" rx="{0.42*w:.1f}" ry="{0.22*w:.1f}" transform="rotate(-25 {-0.15*w:.1f} {-0.55*w:.1f})" fill="{INK}" fill-opacity="0.35" stroke="{INK}" stroke-width="{0.05*w:.1f}"/>
    <ellipse cx="{0.15*w:.1f}" cy="{-0.5*w:.1f}" rx="{0.34*w:.1f}" ry="{0.18*w:.1f}" transform="rotate(-50 {0.15*w:.1f} {-0.5*w:.1f})" fill="{INK}" fill-opacity="0.25" stroke="{INK}" stroke-width="{0.05*w:.1f}"/>
    <ellipse cx="0" cy="0" rx="{0.62*w:.1f}" ry="{0.44*w:.1f}" fill="{HONEY}" stroke="{INK}" stroke-width="{0.06*w:.1f}"/>
    <path d="M{-0.12*w:.1f},{-0.43*w:.1f} Q{-0.2*w:.1f},0 {-0.12*w:.1f},{0.43*w:.1f} L{0.08*w:.1f},{0.43*w:.1f} Q{0.0*w:.1f},0 {0.08*w:.1f},{-0.43*w:.1f} Z" fill="{STRIPE}"/>
    <path d="M{-0.48*w:.1f},{-0.28*w:.1f} Q{-0.56*w:.1f},0 {-0.48*w:.1f},{0.28*w:.1f} L{-0.36*w:.1f},{0.37*w:.1f} Q{-0.45*w:.1f},0 {-0.36*w:.1f},{-0.37*w:.1f} Z" fill="{STRIPE}"/>
    <circle cx="{0.62*w:.1f}" cy="0" r="{0.26*w:.1f}" fill="{STRIPE}" stroke="{INK}" stroke-width="{0.06*w:.1f}"/>
    <path d="M{0.78*w:.1f},{-0.12*w:.1f} q{0.18*w:.1f},{-0.2*w:.1f} {0.3*w:.1f},{-0.22*w:.1f} M{0.78*w:.1f},{0.12*w:.1f} q{0.18*w:.1f},{0.2*w:.1f} {0.3*w:.1f},{0.22*w:.1f}" stroke="{INK}" stroke-width="{0.05*w:.1f}" fill="none" stroke-linecap="round"/>
  </g>'''


def fast_bee(cx, cy, size, angle):
    """The bumblebee with three speed streaks behind it."""
    w = size
    streaks = []
    for dy, x0, x1 in ((-0.28, -0.95, -1.9), (0.0, -0.85, -2.4), (0.28, -0.95, -1.6)):
        streaks.append(f'<path d="M{x0 * w:.1f},{dy * w:.1f} L{x1 * w:.1f},{dy * w:.1f}" stroke="{GO_BLUE}" '
                       f'stroke-width="{0.11 * w:.1f}" stroke-linecap="round"/>')
    trail = f'<g transform="translate({cx},{cy}) rotate({angle})">\n    ' + "\n    ".join(streaks) + "\n  </g>"
    return trail + "\n  " + bee(cx, cy, size, angle)


def flower_patch(cx, cy, r, colour):
    """A flower patch: habitat circle with flower layers inside."""
    return (f'<circle cx="{cx}" cy="{cy}" r="{r}" fill="{colour}"/>\n  '
            f'<circle cx="{cx}" cy="{cy}" r="{r*0.62:.1f}" fill="{FLOWER}" stroke="{INK}" stroke-opacity="0.0"/>\n  '
            f'<circle cx="{cx}" cy="{cy}" r="{r*0.26:.1f}" fill="{colour}"/>')


def icon():
    cx, cy, R = 150, 150, 124
    parts = [
        # foraging range
        f'<circle cx="{cx}" cy="{cy}" r="{R}" fill="none" stroke="{RANGE}" stroke-width="7" '
        f'stroke-dasharray="20 14" stroke-linecap="round" opacity="0.8"/>',
        # flower patches on the edge of the range
        flower_patch(cx + R * math.cos(math.radians(-28)), cy + R * math.sin(math.radians(-28)), 30, HABITATS[0]),
        flower_patch(cx + R * math.cos(math.radians(205)), cy + R * math.sin(math.radians(205)), 22, HABITATS[2]),
        flower_patch(cx + R * math.cos(math.radians(110)), cy + R * math.sin(math.radians(110)), 17, HABITATS[1]),
        # nest: a small comb of three cells
        f'<path d="{poly(hexagon(cx - 26, cy + 26, 29))}" fill="{HONEY}" stroke="{INK}" stroke-width="5" stroke-linejoin="round"/>',
        f'<path d="{poly(hexagon(cx + 24.2, cy + 26, 29))}" fill="{HONEY}" stroke="{INK}" stroke-width="5" stroke-linejoin="round"/>',
        f'<path d="{poly(hexagon(cx - 0.9, cy - 17.5, 29))}" fill="{HONEY}" stroke="{INK}" stroke-width="5" stroke-linejoin="round"/>',
        # bumblebee flying to the large patch
        fast_bee(cx + 40, cy - 84, 56, 12),
    ]
    return "\n  ".join(parts)


def logo(font):
    size, baseline = 200, 205
    text, end = word_mark(font, "gobees", 330, baseline, size, [GO_BLUE, GO_BLUE] + [HONEY] * 4)
    # speed lines in the x-height band, then "fast"
    xh = 0.53 * size  # x-height of Go Bold
    lines_left = end + 46
    lines = speed_lines(lines_left + 78, [baseline - xh * 0.82, baseline - xh * 0.5, baseline - xh * 0.18],
                        [62, 86, 48], 15, GO_BLUE)
    fast, end = slanted(font, "fast", lines_left + 112, baseline, size, GO_BLUE)
    width = math.ceil(end + size * math.tan(math.radians(SLANT)) * 0.75 + 20)
    return f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} 300" width="{width}" height="300">
  <title>gobees-fast</title>
  {icon()}
  {text}
  {lines}
  {fast}
</svg>
'''


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--font", help="path to Go-Bold.ttf")
    args = ap.parse_args()
    font = TTFont(args.font or find_font())
    path = os.path.join(HERE, "gobees-fast-logo-dark.svg")
    with open(path, "w", encoding="utf-8", newline="\n") as f:
        f.write(logo(font))
    print("wrote", path)


if __name__ == "__main__":
    main()
