# Export thenlper/gte-small to ONNX (float32), quantize dynamically to int8, and embed with ONNX Runtime.
# Usage: ort.py <hf_model_dir> <work_dir> export
#        ort.py <hf_model_dir> <work_dir> run <ort_f32|ort_int8> <threads> <input.jsonl> <output.bin>
# `run` prints one line with per_s (timing excludes session creation and a 3-text warm-up) and writes
# little-endian float32 L2-normalised mean-pooled vectors, one per input line.
import os, sys, json, time
import numpy as np

hf, work, mode = sys.argv[1], sys.argv[2], sys.argv[3]
f32 = os.path.join(work, 'gte-small.onnx'); i8 = os.path.join(work, 'gte-small.int8.onnx')
P = {'ort_f32': f32, 'ort_int8': i8}
from transformers import AutoTokenizer
tk = AutoTokenizer.from_pretrained(hf)

if mode == 'export':
    import torch
    from transformers import AutoModel

    class Wrap(torch.nn.Module):
        # The traced exporter rejects kwargs-with-empty-args on BertModel (use_cache clash); this wrapper takes
        # exactly the two graph inputs positionally and calls the model by keyword.
        def __init__(s, m): super().__init__(); s.m = m
        def forward(s, input_ids, attention_mask): return s.m(input_ids=input_ids, attention_mask=attention_mask).last_hidden_state

    m = AutoModel.from_pretrained(hf, dtype=torch.float32).eval()
    ex = tk(['hello world, a longer sentence'], return_tensors='pt')
    torch.onnx.export(Wrap(m), (ex['input_ids'], ex['attention_mask']), f32,
                      input_names=['input_ids', 'attention_mask'], output_names=['last_hidden_state'],
                      dynamic_axes={'input_ids': {0: 'b', 1: 's'}, 'attention_mask': {0: 'b', 1: 's'}, 'last_hidden_state': {0: 'b', 1: 's'}},
                      opset_version=17, dynamo=False)
    from onnxruntime.quantization import quantize_dynamic, QuantType
    quantize_dynamic(f32, i8, weight_type=QuantType.QInt8)
    print('exported', os.path.getsize(f32), os.path.getsize(i8)); sys.exit(0)

import onnxruntime as ort
name, th, inp, outp = sys.argv[4], int(sys.argv[5]), sys.argv[6], sys.argv[7]
texts = [json.loads(l)['text'] for l in open(inp, encoding='utf-8') if l.strip()]
so = ort.SessionOptions(); so.intra_op_num_threads = th; so.inter_op_num_threads = 1
so.execution_mode = ort.ExecutionMode.ORT_SEQUENTIAL; so.graph_optimization_level = ort.GraphOptimizationLevel.ORT_ENABLE_ALL
s = ort.InferenceSession(P[name], so, providers=['CPUExecutionProvider'])


def embed(ts):
    out = []
    for t in ts:
        e = tk(t, truncation=True, max_length=512, return_tensors='np')
        h = s.run(None, {'input_ids': e['input_ids'].astype(np.int64), 'attention_mask': e['attention_mask'].astype(np.int64)})[0]
        a = e['attention_mask'][..., None]; v = (h * a).sum(1) / a.sum(1); out.append(v[0] / np.linalg.norm(v[0]))
    return np.stack(out)


embed(texts[:3])
t = time.perf_counter(); V = embed(texts); d = time.perf_counter() - t
V.astype('<f4').tofile(outp)
print(f"{name} ort={ort.__version__} threads={th} n={len(texts)} secs={d:.2f} per_s={len(texts)/d:.2f} ms_each={d*1000/len(texts):.1f}", flush=True)
