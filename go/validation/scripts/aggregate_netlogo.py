"""Aggregate a NetLogo BehaviorSpace table to quantiles per step and metric, and embed them in a Go file.

Usage (from the repository root):
    python go/validation/scripts/aggregate_netlogo.py step3_netlogo.csv step3

Reads the --table output of run_netlogo_experiment.ps1 and writes go/validation/reference/netlogo_<name>.csv with
the quantiles 5, 25, 50, 75 and 95 % of every metric at every step, and netlogo_<name>_runs.csv with the summaries
of every run (validation.RunSummaries): the value at step 0 (period 0, kind "value"), and per year of 365 steps
(period 1, 2, ..., steps 1 to 365, 366 to 730, ...) the maximum and the mean (kinds "max" and "mean"). go/validation/reference.go embeds the file, the
comparison test go/sys/validation_<name>_test.go reads it.

The metric names are given by position, in the order of the <metric> elements of the experiment file.
"""
import csv
import statistics
import sys

METRICS = {
    'step3': ['foodsources', 'masterpatches', 'nestsites', 'nestsite_area', 'colonies_founded',
              'queens_searching', 'queens_hibernating', 'mean_queen_weight', 'mean_emerging_date'],
    'step4a': ['day', 'nectar_l', 'pollen_kg'],
    'step6': ['colonies', 'dead_colonies', 'eggs', 'larvae', 'pupae', 'workers', 'males', 'young_queens',
              'energy_kJ', 'pollen_g', 'trips', 'switch_point', 'queen_production', 'nectar_left_l',
              'pollen_left_kg'],
    'step7': ['colonies', 'dead_colonies', 'colonies_founded', 'eggs', 'larvae', 'pupae', 'workers', 'males',
              'young_queens', 'hibernating_queens', 'energy_kJ', 'pollen_g', 'trips', 'nectar_left_l',
              'pollen_left_kg'],
    'step7_badgers': ['colonies', 'dead_colonies', 'colonies_founded', 'eggs', 'larvae', 'pupae', 'workers',
                      'males', 'young_queens', 'hibernating_queens', 'energy_kJ', 'pollen_g', 'trips',
                      'nectar_left_l', 'pollen_left_kg', 'badgers'],
    'step7_crop': ['colonies', 'dead_colonies', 'colonies_founded', 'eggs', 'larvae', 'pupae', 'workers', 'males',
                   'young_queens', 'hibernating_queens', 'energy_kJ', 'pollen_g', 'trips', 'nectar_left_l',
                   'pollen_left_kg', 'foodsources'],
}
# step 8: the metrics of step 7 plus a few per experiment
STEP7 = METRICS['step7']
METRICS['step8_ibm'] = STEP7 + ['ibm_colonies', 'bee_agents']
METRICS['step8_genetics'] = STEP7 + ['diploid_males']
METRICS['step8_species'] = STEP7 + ['colonies_B_terrestris', 'colonies_B_pascuorum', 'colonies_B_lapidarius',
                                    'colonies_B_hortorum', 'colonies_B_pratorum', 'colonies_B_hypnorum', 'psithyrus']
METRICS['step8_weather'] = STEP7 + ['foraging_period_s']
METRICS['step8_settings'] = STEP7 + ['foodsources', 'queens_removed']


def main():
    if len(sys.argv) != 3 or sys.argv[2] not in METRICS:
        sys.exit(__doc__)
    table, name = sys.argv[1], sys.argv[2]
    metrics = METRICS[name]

    with open(table, encoding='utf-8', newline='') as f:
        rows = list(csv.reader(f))
    # 6 lines of BehaviorSpace header, then the column names, then the data
    header_index = next(i for i, r in enumerate(rows) if r and r[0] == '[run number]')
    header, data = rows[header_index], rows[header_index + 1:]
    # variables set by the experiment (enumeratedValueSet) come as extra columns between [run number] and [step]
    step_col = header.index('[step]')
    if len(header) != step_col + 1 + len(metrics):
        sys.exit(f'expected {len(metrics)} metrics, the table has {len(header) - step_col - 1} columns')

    values = {}  # (step, metric) -> list of values over all runs
    per_run = {}  # (run, metric) -> {step: value}
    runs = set()
    for r in data:
        if not r:
            continue
        runs.add(r[0])
        step = int(r[step_col])
        for m, v in zip(metrics, r[step_col + 1:]):
            values.setdefault((step, m), []).append(float(v))
            per_run.setdefault((int(r[0]), m), {})[step] = float(v)

    lines = ['step,metric,n,q05,q25,q50,q75,q95']
    for (step, m), v in sorted(values.items(), key=lambda x: (x[0][0], metrics.index(x[0][1]))):
        if len(v) > 1:
            q = statistics.quantiles(v, n=20, method='inclusive')  # q[0] = 5 %, q[4] = 25 %, ... q[18] = 95 %
            qs = [q[0], q[4], statistics.median(v), q[14], q[18]]
        else:
            qs = [v[0]] * 5
        lines.append(f'{step},{m},{len(v)},' + ','.join(repr(x) for x in qs))

    out = f'go/validation/reference/netlogo_{name}.csv'
    with open(out, 'w', encoding='utf-8', newline='\n') as f:
        f.write('\n'.join(lines) + '\n')
    print(f'{len(runs)} runs, {len(values)} step/metric pairs -> {out}')

    # summaries per run, in the order of the run numbers
    lines = ['run,metric,period,kind,value']
    for (run, m) in sorted(per_run, key=lambda x: (x[0], metrics.index(x[1]))):
        steps = per_run[(run, m)]
        last = max(steps)
        lines.append(f'{run},{m},0,value,{steps[0]!r}')
        for year in range(1, (last - 1) // 365 + 2):
            v = [steps[s] for s in range((year - 1) * 365 + 1, min(year * 365, last) + 1)]
            lines.append(f'{run},{m},{year},max,{max(v)!r}')
            lines.append(f'{run},{m},{year},mean,{sum(v) / len(v)!r}')
    out = f'go/validation/reference/netlogo_{name}_runs.csv'
    with open(out, 'w', encoding='utf-8', newline='\n') as f:
        f.write('\n'.join(lines) + '\n')
    print(f'{len(lines) - 1} run summaries -> {out}')


if __name__ == '__main__':
    main()
