"""List global variables that a NetLogo procedure sets AND that appear in a parameter file.

Used for finding 6 in trans/netlogo-findings.md: ParametersProc runs after ReadAllParametersProc,
so any overlap would mean that a hard-coded value overwrites a value from the parameter file.

Usage (from the repository root):
    python trans/check_param_overlap.py
    python trans/check_param_overlap.py --proc ParametersProc --params netlogo/_SYSTEM_Parameters.csv
"""
import argparse
import csv
import re
import sys

NLOGO = 'netlogo/__BEE-STEWARD_RC_ColourVersionMB.nlogo'


def strip_comment(line):
    """Remove a NetLogo comment, ignoring semicolons inside string literals."""
    code = re.sub(r'"(?:\\.|[^"\\])*"', lambda m: m.group(0).replace(';', '\0'), line)
    return code.split(';')[0].replace('\0', ';')


def procedure_body(path, name):
    """Return the code lines of procedure `name` (between 'to name' and 'end')."""
    lines = open(path, encoding='utf-8', errors='replace').read().split('@#$#@#$#@')[0].split('\n')
    start = None
    for i, line in enumerate(lines):
        if start is None and re.match(r'^\s*to(-report)?\s+' + re.escape(name) + r'\b', line, re.I):
            start = i + 1
        elif start is not None and re.match(r'^\s*end\b', line, re.I):
            return [strip_comment(x) for x in lines[start:i]]
    sys.exit(f'procedure {name} not found in {path}')


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n')[0])
    ap.add_argument('--nlogo', default=NLOGO)
    ap.add_argument('--proc', default='ParametersProc')
    ap.add_argument('--params', default='netlogo/_SYSTEM_Parameters.csv')
    args = ap.parse_args()

    # NetLogo identifiers are case-insensitive, so compare in lower case.
    body = ' '.join(procedure_body(args.nlogo, args.proc))
    set_vars = {v.lower(): v for v in re.findall(r'\bset\s+(\S+)', body, re.I)}
    with open(args.params, encoding='utf-8-sig', newline='') as f:
        header = next(csv.reader(f))
    file_vars = {h.strip().lower(): h.strip() for h in header}

    overlap = sorted(set(set_vars) & set(file_vars))
    print(f'{args.proc} sets {len(set_vars)} variables, {args.params} has {len(file_vars)} parameters.')
    if overlap:
        print('Overlap (value from the file is overwritten):')
        for v in overlap:
            print(f'  {set_vars[v]}')
        sys.exit(1)
    print('No overlap.')


if __name__ == '__main__':
    main()
