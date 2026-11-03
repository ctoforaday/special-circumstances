# Summarises runs.tsv + fidelity.json into a Markdown table and summary.json. Usage: summarize.py <results_dir>
import sys, os, re, json, statistics

R = sys.argv[1]
rows = [l.rstrip('\n').split('\t') for l in open(os.path.join(R, 'runs.tsv'))][1:]
fid = json.load(open(os.path.join(R, 'fidelity.json')))['vs_ref']
host = open(os.path.join(R, 'host.txt')).read()
m = re.search(r'Model name:\s*(.+)', host)
cpu = m.group(1).strip() if m else 'unknown'
flags = re.search(r'== cpuinfo flags.*\n(.*)\n', host)
arms, out = [], {}
for ph, rnd, arm, th, st, load, res in rows:
    if arm not in arms:
        arms.append(arm)
    p = re.search(r'per_s=([0-9.]+)', res)
    v = float(p.group(1)) if (p and st == '0') else None
    out.setdefault(arm, {}).setdefault(ph, []).append({'round': int(rnd), 'exit': int(st), 'per_s': v, 'load': load, 'raw': res})
print(f"cpu: {cpu}\n\nflags: {flags.group(1).strip() if flags else ''}\n")
print('| arm | 1-thread median | rounds | 4-thread | fidelity mean | p05 | min | exits |')
print('|---|---|---|---|---|---|---|---|')
summ = {'cpu': cpu, 'arms': {}}
for a in arms:
    t1 = [x['per_s'] for x in out[a].get('t1', [])]
    good = [x for x in t1 if x is not None]
    med = statistics.median(good) if len(good) == 3 else None
    t4 = [x['per_s'] for x in out[a].get('t4', [])]
    f = fid.get(a, {})
    exits = [x['exit'] for ph in out[a].values() for x in ph]
    fmt = lambda v: 'FAIL' if v is None else f'{v:.2f}'
    print(f"| {a} | {fmt(med)} | {' / '.join(fmt(x) for x in t1)} | {fmt(t4[0] if t4 else None)} | "
          f"{f.get('mean', float('nan')):.5f} | {f.get('p05', float('nan')):.5f} | {f.get('min', float('nan')):.5f} | {exits} |")
    summ['arms'][a] = {'t1': t1, 't1_median': med, 't4': t4, 'fidelity': f, 'exits': exits}
json.dump(summ, open(os.path.join(R, 'summary.json'), 'w'), indent=1)
