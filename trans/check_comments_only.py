"""Check that two NetLogo files differ only in comments and blank lines.

Run after every annotation step to make sure the model code itself is unchanged.

Usage (from the repository root):
    python tools/check_comments_only.py ORIGINAL.nlogo ANNOTATED.nlogo
"""
import re
import sys


def code_lines(path):
    """Return (line number, code without comment) for every line that contains code."""
    out = []
    for i, line in enumerate(open(path, encoding='utf-8', errors='replace').read().split('\n'), 1):
        code = re.sub(r'"(?:\\.|[^"\\])*"', lambda m: m.group(0).replace(';', '\0'), line)
        code = code.split(';')[0].replace('\0', ';').rstrip()
        if code.strip():
            out.append((i, code))
    return out


def main():
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    a, b = code_lines(sys.argv[1]), code_lines(sys.argv[2])
    for (ia, ca), (ib, cb) in zip(a, b):
        if ca != cb:
            print(f'DIFF original line {ia} vs annotated line {ib}:\n  {ca}\n  {cb}')
            sys.exit(1)
    if len(a) != len(b):
        print(f'DIFF number of code lines: {len(a)} vs {len(b)}')
        sys.exit(1)
    print(f'OK: {len(a)} code lines identical, only comments changed')


if __name__ == '__main__':
    main()
