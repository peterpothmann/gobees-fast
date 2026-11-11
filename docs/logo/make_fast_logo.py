"""Draws the gobees-fast logo as SVG, for dark backgrounds (gobees-fast-logo-dark.svg).

The icon shows the main concepts of the model: a colony (nest), its foraging range (dashed circle),
flower patches made of layers (concentric circles in habitat colours) and a bumblebee flying to a patch.
The word mark is "gobees" in Go Bold, followed by "fast".

Bumblebees build no hexagonal comb: their nest is a clump of round wax pots for nectar and pollen and of
brood cells and cocoons. The nest is drawn this way. The bee has the round, furry body and the colour bands of
Bombus terrestris: yellow collar, black thorax, yellow band, black band, white tail.

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
NECTAR = "#B87A08"  # inside of an open wax pot
TAIL = "#F4F1E8"  # white tail of the bumblebee
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


def word_mark(font, text, x, baseline, size, colours):
    """Returns SVG paths for text and the end x. colours: one colour per character."""
    s = size / font["head"].unitsPerEm
    glyphs = font.getGlyphSet()
    cmap = font.getBestCmap()
    out = []
    for ch, col in zip(text, colours):
        name = cmap[ord(ch)]
        pen = SVGPathPen(glyphs)
        glyphs[name].draw(pen)
        out.append(f'<path transform="translate({x:.2f},{baseline:.2f}) scale({s:.5f},{-s:.5f})" '
                   f'd="{pen.getCommands()}" fill="{col}"/>')
        x += font["hmtx"][name][0] * s
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
    """A bumblebee centred at (cx, cy), heading in direction angle (degrees).

    Body seen from above, from the tail to the head: white tail, black band, yellow band, black thorax,
    yellow collar. A ring of dots around the body stands for the fur.
    """
    w = size
    rx, ry = 0.62 * w, 0.5 * w  # radii of the body
    bend = 0.1 * w  # the bands curve around the round body

    def band(x0, x1, colour):
        return (f'<path d="M{x0:.1f},{-ry:.1f} Q{x0 - bend:.1f},0 {x0:.1f},{ry:.1f} L{x1:.1f},{ry:.1f} '
                f'Q{x1 - bend:.1f},0 {x1:.1f},{-ry:.1f} Z" fill="{colour}"/>')

    bands = "\n      ".join([
        band(-rx - bend, -0.40 * w, TAIL),
        band(-0.40 * w, -0.20 * w, STRIPE),
        band(0.02 * w, 0.34 * w, STRIPE),
    ])
    return f'''<g transform="translate({cx},{cy}) rotate({angle})">
    <clipPath id="bee-body"><ellipse cx="0" cy="0" rx="{rx:.1f}" ry="{ry:.1f}"/></clipPath>
    <ellipse cx="{-0.05*w:.1f}" cy="{-0.62*w:.1f}" rx="{0.40*w:.1f}" ry="{0.20*w:.1f}" transform="rotate(-25 {-0.05*w:.1f} {-0.62*w:.1f})" fill="{INK}" fill-opacity="0.35" stroke="{INK}" stroke-width="{0.05*w:.1f}"/>
    <ellipse cx="{0.22*w:.1f}" cy="{-0.56*w:.1f}" rx="{0.32*w:.1f}" ry="{0.16*w:.1f}" transform="rotate(-50 {0.22*w:.1f} {-0.56*w:.1f})" fill="{INK}" fill-opacity="0.25" stroke="{INK}" stroke-width="{0.05*w:.1f}"/>
    <ellipse cx="0" cy="0" rx="{rx + 0.03*w:.1f}" ry="{ry + 0.03*w:.1f}" fill="none" stroke="{INK}" stroke-width="{0.11*w:.1f}" stroke-dasharray="0 {0.085*w:.1f}" stroke-linecap="round"/>
    <ellipse cx="0" cy="0" rx="{rx:.1f}" ry="{ry:.1f}" fill="{HONEY}"/>
    <g clip-path="url(#bee-body)">
      {bands}
    </g>
    <ellipse cx="0" cy="0" rx="{rx:.1f}" ry="{ry:.1f}" fill="none" stroke="{INK}" stroke-width="{0.06*w:.1f}"/>
    <circle cx="{0.62*w:.1f}" cy="0" r="{0.24*w:.1f}" fill="{STRIPE}" stroke="{INK}" stroke-width="{0.06*w:.1f}"/>
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


def wax_pot(cx, cy, r, open_pot):
    """A round wax cell of the nest. open_pot: a nectar pot, open at the top (dark centre), else a closed cocoon."""
    out = f'<circle cx="{cx:.1f}" cy="{cy:.1f}" r="{r}" fill="{HONEY}" stroke="{INK}" stroke-width="5"/>'
    if open_pot:
        out += f'\n  <circle cx="{cx:.1f}" cy="{cy:.1f}" r="{r * 0.55:.1f}" fill="{NECTAR}"/>'
    return out


def nest(cx, cy):
    """A bumblebee nest: a clump of round wax pots and cocoons of different sizes, no regular comb."""
    pots = [  # dx, dy, radius, open pot
        (-40, -6, 14, True),
        (-24, 22, 24, False),
        (22, 28, 21, True),
        (-4, -14, 22, False),
        (32, -8, 15, False),
    ]
    return "\n  ".join(wax_pot(cx + dx, cy + dy, r, o) for dx, dy, r, o in pots)


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
        nest(cx, cy + 10),
        # bumblebee flying to the large patch
        fast_bee(cx + 42, cy - 84, 52, 12),
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
