"""Write the parameter files of the step 8 experiments: the example farm parameters with a few values changed.

Usage (from the repository root):
    python go/validation/scripts/parameter_files.py

Reads netlogo/_SYSTEM_Example_Farm_Parameters.csv and writes go/validation/experiments/<experiment>_Parameters.csv
for every experiment in VARIANTS. Only the listed values change, everything else (format, quoting, line endings)
stays as in the original. NetLogo reads the file through MyParametersFile (set in the experiment), the gobees
comparison test parses the same file (go/validation/reference.go embeds it).
"""
import csv
import sys

SOURCE = 'netlogo/_SYSTEM_Example_Farm_Parameters.csv'

# Values as they appear in the CSV file: strings in NetLogo quotes inside CSV quotes ("""text"""), booleans as
# true/false, numbers as they are.
VARIANTS = {
    'step8_ibm': {
        'COLONIES_IBM': '10',
    },
    'step8_genetics': {
        'SexLocus?': 'true',
        'UnlimitedMales?': 'false',
    },
    'step8_species': {
        'BeeSpeciesInitialQueensListAsString': '"""B_terrestris 150 B_pascuorum 150 B_lapidarius 150 B_hortorum 150 '
                                               'B_pratorum 150 B_hypnorum 150 Psithyrus 100"""',
    },
    'step8_weather': {
        'Weather': '"""foragingHoursExample"""',
        'ForagingMortalityModel': '"""low"""', 
    },
    'step8_settings': {
        'WinterMortality?': 'false',
        'MergeHedges?': 'false', 
        'MinSizeFoodSources?': 'false',
        'RemoveEmptyFoodSources?': 'false',
        'KeepDeadColonies?': 'false',
        'MaxHibernatingQueens': '150', 
    },
}


def main():
    raw = open(SOURCE, 'rb').read().decode('utf-8')
    newline = '\r\n' if '\r\n' in raw else '\n'
    lines = raw.split(newline)
    header = lines[0].split(',')
    cells = lines[1].split(',')
    # the values contain no commas, so splitting at commas keeps the original quoting
    if len(cells) != len(next(csv.reader([lines[1]]))) or len(cells) != len(header):
        sys.exit('unexpected format of ' + SOURCE)
    names = [h.strip().strip('"').strip() for h in header]

    for experiment, changes in VARIANTS.items():
        new = list(cells)
        for name, value in changes.items():
            new[names.index(name)] = value
        out = f'go/validation/experiments/{experiment}_Parameters.csv'
        with open(out, 'wb') as f:
            f.write(newline.join([lines[0], ','.join(new)] + lines[2:]).encode('utf-8'))
        print(f'{out}: ' + ', '.join(f'{k} = {v}' for k, v in changes.items()))


if __name__ == '__main__':
    main()
