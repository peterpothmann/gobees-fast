"""Draws the gobees logo as pixel art (SVG of square pixels), for dark and light backgrounds.

Every glyph and sprite is an ASCII grid below: one character per pixel, "." is transparent.
Edit a grid and run the script again:  python make_pixel_logo.py

Concepts in the picture: the "o" is the hive (a honeycomb cell), a bumblebee flies from it on a dotted
path to the flowers, and the ground is made of flower patches in habitat colours.
"""

import os

HERE = os.path.dirname(os.path.abspath(__file__))
PX = 12  # size of one pixel in the SVG

# Colours per character. "ink" and "ink2" depend on the theme.
PALETTE = {
    "B": "#00ADD8",  # Go blue (g)
    "H": "#F5B72A",  # honey (b, e, e, s)
    "h": "#C98A0A",  # dark honey (hive)
    "y": "#F5B72A",  # bee body
    "k": "#1D1D1D",  # bee stripes and head
    "e": "#FFFFFF",  # eye
    "w": "#BFE6F5",  # wings
    "Y": "#FFD84A",  # flower centre
    "p": "#E07F96",  # petals: grassland pink
    "v": "#9C6BC9",  # petals: violet
    "c": "#5FA8E8",  # petals: blue
    "G": "#3E9B4F",  # stem, grass
    "g": "#2E7A3D",  # dark grass
    "1": "#1D9F78",  # ground: cereals (NetLogo turquoise)
    "2": "#E07F96",  # ground: grassland
    "3": "#7C50A4",  # ground: improved grass
    "4": "#345DA9",  # ground: hedgerow
}
THEMES = {
    # K: antennae, o: flight dots, O: outline of the bee (None: not drawn)
    "dark": {"K": "#F2F2F2", "o": "#F2F2F2", "O": "#6E7681"},
    "light": {"K": "#1D1D1D", "o": "#1D1D1D", "O": None},
}

# Letters: 13 rows. Rows 0-2 ascender, rows 3-9 x-height (row 9 = baseline), rows 10-12 descender.
GLYPHS = {
    "g": """
.....
.....
.....
.BBBB
B...B
B...B
B...B
B...B
.BBBB
....B
....B
B...B
.BBB.""",
    # The "o" is a hive (skep): a dome of straw bands, the round entrance is the counter of the "o".
    "o": """
....hhh....
..hhHHHhh..
.hHHHHHHHh.
.hhhhhhhhh.
hHHH...HHHh
hhh.....hhh
hHH.....HHh
hhhh...hhhh
hHHHHHHHHHh
hhhhhhhhhhh
...........
...........
...........""",
    "b": """
H....
H....
H....
H.HH.
HH..H
H...H
H...H
H...H
HH..H
H.HH.
.....
.....
.....""",
    "e": """
.....
.....
.....
.HHH.
H...H
H...H
HHHHH
H....
H...H
.HHH.
.....
.....
.....""",
    "s": """
.....
.....
.....
.HHHH
H....
H....
.HHH.
....H
....H
HHHH.
.....
.....
.....""",
}

BEE = """
.....ww..ww...
....wwww.www.K
.....wwwwww.K.
..yykyykyykk..
.yyykyykyykkk.
kyyykyykyykek.
.yyykyykyykkk.
..yykyykyykk.."""

FLOWERS = {
    "p": """
.ppp.
ppYpp
.ppp.
..G..
G.G..
.GG..
..G..""",
    "v": """
.vvv.
vvYvv
.vvv.
..G..
..G.G
..GG.
..G..""",
    "c": """
.ccc.
ccYcc
.ccc.
..G..
G.G..
.GG..
..G..""",
}


def grid(s):
    return [row for row in s.strip("\n").split("\n")]


def stamp(canvas, art, x, y):
    for r, row in enumerate(grid(art)):
        for c, ch in enumerate(row):
            if ch != ".":
                canvas[(x + c, y + r)] = ch


def build():
    canvas = {}
    top = 7  # rows above the letters for the flight path and the bee
    x = 2
    starts = {}
    for i, ch in enumerate("gobees"):
        starts[i] = x
        stamp(canvas, GLYPHS[ch], x, top)
        x += len(grid(GLYPHS[ch])[0]) + 2
    word_end = x
    base = top + 9  # baseline row

    # Flowers to the right of the word, standing on the ground.
    fx = word_end + 1
    for k, name in enumerate(["p", "v", "c"]):
        stamp(canvas, FLOWERS[name], fx + k * 6, base + 3 - 6 + (k % 2))
    width = fx + 3 * 6 + 1

    # Ground: flower patches in habitat colours, with grass on top.
    ground = base + 4
    patches = "1111111222222333333444444111111222222333333444444111111"
    for c in range(width):
        canvas[(c, ground)] = "G" if c % 3 else "g"
        canvas[(c, ground + 1)] = patches[c % len(patches)]

    # Flight path out of the entrance of the hive, over "bees", to the bee that lands on the flowers.
    hole_x, hole_y = starts[1] + 5, top + 5
    bee_x, bee_y = fx - 1, top - 6
    path = []
    steps = 120
    for i in range(steps + 1):
        t = i / steps
        # quadratic curve: entrance -> high above "b e e s" -> tail of the bee
        cx, cy = (hole_x + bee_x) / 2, -6
        px = (1 - t) ** 2 * hole_x + 2 * (1 - t) * t * cx + t ** 2 * (bee_x - 1)
        py = (1 - t) ** 2 * hole_y + 2 * (1 - t) * t * cy + t ** 2 * (bee_y + 4)
        path.append((round(px), round(py)))
    seen = []
    for p in path:
        if p not in seen:
            seen.append(p)
    for i, p in enumerate(seen[1:-1]):
        if i % 3 == 0 and p not in canvas:
            canvas[p] = "o"
    stamp(canvas, BEE, bee_x, bee_y)
    # Outline around the bee ("O"), visible only on dark backgrounds, where the black head would vanish.
    bee_px = [(bee_x + c, bee_y + r) for r, row in enumerate(grid(BEE)) for c, ch in enumerate(row) if ch != "."]
    for (px, py) in bee_px:
        for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1)):
            q = (px + dx, py + dy)
            if q not in canvas:
                canvas[q] = "O"

    height = ground + 2
    return canvas, width, height


def svg(canvas, width, height, theme):
    colours = dict(PALETTE, **THEMES[theme])
    rects = []
    for y in range(height):
        x = 0
        while x < width:
            ch = canvas.get((x, y))
            if ch is None or colours[ch] is None:
                x += 1
                continue
            run = 1
            while canvas.get((x + run, y)) == ch:
                run += 1
            rects.append(f'<rect x="{x * PX}" y="{y * PX}" width="{run * PX}" height="{PX}" fill="{colours[ch]}"/>')
            x += run
    w, h = width * PX, height * PX
    body = "\n  ".join(rects)
    return f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {w} {h}" width="{w}" height="{h}" shape-rendering="crispEdges">
  <title>gobees</title>
  {body}
</svg>
'''


def main():
    canvas, width, height = build()
    for theme in THEMES:
        path = os.path.join(HERE, f"gobees-pixel-{theme}.svg")
        with open(path, "w", encoding="utf-8", newline="\n") as f:
            f.write(svg(canvas, width, height, theme))
        print("wrote", path, f"({width} x {height} pixels)")


if __name__ == "__main__":
    main()
