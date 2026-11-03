# Fidelity of embedding outputs (little-endian float32, one 384-d vector per line of speed.jsonl).
# Usage: fid.py <ref.npy> <results_dir> <arm>=<file.bin> ...
# Prints per-arm cosine vs the PyTorch reference (mean, p05, min), then pairwise cosine between gte-go arms.
# Writes <results_dir>/fidelity.json. A missing or mis-sized output is reported as FAILED, never skipped.
import sys, json, os
import numpy as np

ref, rdir, pairs = sys.argv[1], sys.argv[2], sys.argv[3:]
R = np.load(ref).astype(np.float64); R /= np.linalg.norm(R, axis=1, keepdims=True)
res, vecs = {}, {}
for p in pairs:
    arm, path = p.split('=', 1)
    try:
        G = np.fromfile(path, dtype='<f4').astype(np.float64)
        if G.size != R.size:
            raise ValueError(f'size {G.size} != {R.size}')
        G = G.reshape(R.shape); G /= np.maximum(np.linalg.norm(G, axis=1, keepdims=True), 1e-12)
        c = (R * G).sum(1)
        res[arm] = {'n': int(len(c)), 'mean': float(c.mean()), 'p05': float(np.percentile(c, 5)), 'min': float(c.min())}
        vecs[arm] = G
        print(f"fidelity arm={arm} n={len(c)} mean={c.mean():.6f} p05={np.percentile(c, 5):.6f} min={c.min():.6f}")
    except Exception as e:  # noqa: BLE001
        res[arm] = {'error': str(e)}
        print(f"fidelity arm={arm} FAILED: {e}")
go = [a for a in vecs if a.startswith('gtego')]
cross = {}
for i, a in enumerate(go):
    for b in go[i + 1:]:
        c = (vecs[a] * vecs[b]).sum(1)
        cross[f'{a}~{b}'] = {'mean': float(c.mean()), 'min': float(c.min())}
        print(f"cross {a} vs {b} mean={c.mean():.7f} min={c.min():.7f}")
json.dump({'vs_ref': res, 'gtego_cross': cross}, open(os.path.join(rdir, 'fidelity.json'), 'w'), indent=1)
