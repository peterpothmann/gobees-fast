"""Draws the gobees-fast logo as SVG, one version for dark and one for light backgrounds.

The logo is the gobees logo (make_logo.py) with the word "fast" appended. Three cues show speed:
- "fast" is set in Go Bold slanted forward, in Go blue,
- three speed lines in front of "fast", as in the Go brand logo,
- the bumblebee of the icon trails speed streaks instead of the dotted flight path.

Usage:  python make_fast_logo.py [--font path/to/Go-Bold.ttf]
Needs:  fontTools (pip install fonttools) and make_logo.py in the same folder.
"""

import argparse
import math
import os

from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.ttLib import TTFont

from make_logo import (GO_BLUE, HABITATS, HERE, THEMES, bee, find_font, flower_patch, hexagon, poly,
                       word_mark)

SLANT = 14  # forward slant of "fast" in degrees


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


def fast_bee(cx, cy, size, angle, t):
    """The bumblebee of make_logo.py with three speed streaks behind it."""
    w = size
    streaks = []
    for dy, x0, x1 in ((-0.28, -0.95, -1.9), (0.0, -0.85, -2.4), (0.28, -0.95, -1.6)):
        streaks.append(f'<path d="M{x0 * w:.1f},{dy * w:.1f} L{x1 * w:.1f},{dy * w:.1f}" stroke="{GO_BLUE}" '
                       f'stroke-width="{0.11 * w:.1f}" stroke-linecap="round"/>')
    trail = f'<g transform="translate({cx},{cy}) rotate({angle})">\n    ' + "\n    ".join(streaks) + "\n  </g>"
    return trail + "\n  " + bee(cx, cy, size, angle, t)


def icon(t):
    """The icon of make_logo.py, with the dotted flight path replaced by speed streaks."""
    cx, cy, R = 150, 150, 124
    parts = [
        f'<circle cx="{cx}" cy="{cy}" r="{R}" fill="none" stroke="{t["range"]}" stroke-width="7" '
        f'stroke-dasharray="20 14" stroke-linecap="round" opacity="0.8"/>',
        flower_patch(cx + R * math.cos(math.radians(-28)), cy + R * math.sin(math.radians(-28)), 30, HABITATS[0], t),
        flower_patch(cx + R * math.cos(math.radians(205)), cy + R * math.sin(math.radians(205)), 22, HABITATS[2], t),
        flower_patch(cx + R * math.cos(math.radians(110)), cy + R * math.sin(math.radians(110)), 17, HABITATS[1], t),
        f'<path d="{poly(hexagon(cx - 26, cy + 26, 29))}" fill="{t["honey"]}" stroke="{t["ink"]}" stroke-width="5" stroke-linejoin="round"/>',
        f'<path d="{poly(hexagon(cx + 24.2, cy + 26, 29))}" fill="{t["honey"]}" stroke="{t["ink"]}" stroke-width="5" stroke-linejoin="round"/>',
        f'<path d="{poly(hexagon(cx - 0.9, cy - 17.5, 29))}" fill="{t["honey"]}" stroke="{t["ink"]}" stroke-width="5" stroke-linejoin="round"/>',
        fast_bee(cx + 40, cy - 84, 56, 12, t),
    ]
    return "\n  ".join(parts)


def logo(font, theme):
    t = THEMES[theme]
    size, baseline = 200, 205
    text, end = word_mark(font, "gobees", 330, baseline, size, [GO_BLUE, GO_BLUE] + [t["honey"]] * 4, hex_o=True)
    # speed lines in the x-height band, then "fast"
    xh = 0.53 * size  # x-height of Go Bold
    lines_left = end + 46
    lines = speed_lines(lines_left + 78, [baseline - xh * 0.82, baseline - xh * 0.5, baseline - xh * 0.18],
                        [62, 86, 48], 15, GO_BLUE)
    fast, end = slanted(font, "fast", lines_left + 112, baseline, size, GO_BLUE)
    width = math.ceil(end + size * math.tan(math.radians(SLANT)) * 0.75 + 20)
    return f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {width} 300" width="{width}" height="300">
  <title>gobees-fast</title>
  {icon(t)}
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
    for theme in THEMES:
        path = os.path.join(HERE, f"gobees-fast-logo-{theme}.svg")
        with open(path, "w", encoding="utf-8", newline="\n") as f:
            f.write(logo(font, theme))
        print("wrote", path)


if __name__ == "__main__":
    main()
