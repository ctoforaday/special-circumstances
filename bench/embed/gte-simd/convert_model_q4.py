#!/usr/bin/env python3
"""
Convert GTE-small model from safetensors to Q4-quantized .gtemodel format.

Usage:
    python convert_model_q4.py models/gte-small gte-small-q4.gtemodel

Produces a ~10MB file (6.4× smaller than FP32) with magic "GTE4".
Weight matrices are quantized to Q4_0 (4-bit symmetric, block size 32).
Biases and LayerNorm params remain FP32 for accuracy.
"""

import sys
import struct
import json
import numpy as np
from pathlib import Path

try:
    from safetensors import safe_open
except ImportError:
    print("Please install safetensors: pip install safetensors")
    sys.exit(1)

BLOCK_SIZE = 32


def quantize_q4_0(tensor):
    """Quantize a tensor to Q4_0 blocks.

    Each block: float32 scale + 16 packed bytes (32 nibbles).
    Nibble = round(value/scale) + 8, clamped to [0, 15].
    """
    flattened = tensor.flatten().astype(np.float32)
    n = flattened.size
    padding = (BLOCK_SIZE - (n % BLOCK_SIZE)) % BLOCK_SIZE
    if padding > 0:
        flattened = np.concatenate([flattened, np.zeros(padding, dtype="float32")])

    reshaped = flattened.reshape(-1, BLOCK_SIZE)

    # Scale = max(abs(block)) / 7
    scales = np.max(np.abs(reshaped), axis=1) / 7.0
    scales = scales.astype(np.float32)

    # Quantize: round(value / scale), clamp to [-8, 7], shift to [0, 15]
    inv_scales = np.where(scales != 0, 1.0 / scales, 0).reshape(-1, 1)
    quant = np.round(reshaped * inv_scales).clip(-8, 7).astype(np.int8)
    quant = (quant + 8).astype(np.uint8)  # Shift to unsigned [0, 15]

    # Pack: low elements in low nibble, high elements in high nibble
    packed = np.zeros((quant.shape[0], BLOCK_SIZE // 2), dtype=np.uint8)
    for i in range(BLOCK_SIZE // 2):
        packed[:, i] = quant[:, i] | (quant[:, i + BLOCK_SIZE // 2] << 4)

    return scales, packed


def main():
    if len(sys.argv) < 3:
        print(f"Usage: {sys.argv[0]} <model_dir> <output.gtemodel>")
        sys.exit(1)

    model_dir = Path(sys.argv[1])
    output_path = sys.argv[2]

    with open(model_dir / "config.json") as f:
        config = json.load(f)

    vocab_size = config["vocab_size"]
    hidden_size = config["hidden_size"]
    num_layers = config["num_hidden_layers"]
    num_heads = config["num_attention_heads"]
    intermediate_size = config["intermediate_size"]
    max_seq_length = config["max_position_embeddings"]

    print(f"Q4 Model conversion:")
    print(f"  vocab_size={vocab_size} hidden={hidden_size} layers={num_layers}")
    print(f"  heads={num_heads} intermediate={intermediate_size} max_seq={max_seq_length}")
    print(f"  Block size: {BLOCK_SIZE}, compression: ~6.4×")

    vocab = []
    with open(model_dir / "vocab.txt", "r", encoding="utf-8") as f:
        for line in f:
            vocab.append(line.rstrip("\n"))

    safetensors_path = model_dir / "model.safetensors"
    tensors = safe_open(safetensors_path, framework="numpy")

    total_fp32 = 0
    total_q4 = 0

    with open(output_path, "wb") as f:
        # Header
        f.write(b"GTE4")
        f.write(struct.pack("<I", vocab_size))
        f.write(struct.pack("<I", hidden_size))
        f.write(struct.pack("<I", num_layers))
        f.write(struct.pack("<I", num_heads))
        f.write(struct.pack("<I", intermediate_size))
        f.write(struct.pack("<I", max_seq_length))

        # Vocabulary
        for word in vocab:
            word_bytes = word.encode("utf-8")
            f.write(struct.pack("<H", len(word_bytes)))
            f.write(word_bytes)

        def write_fp32(name):
            nonlocal total_fp32
            tensor = tensors.get_tensor(name).astype("float32")
            data = tensor.tobytes()
            f.write(data)
            total_fp32 += len(data)
            return tensor.shape

        def write_q4(name):
            nonlocal total_q4
            tensor = tensors.get_tensor(name).astype("float32")
            scales, packed = quantize_q4_0(tensor)
            # Write interleaved: [scale(4B), packed(16B)] per block
            for s, p in zip(scales, packed):
                f.write(s.tobytes())
                f.write(p.tobytes())
            total_q4 += len(scales) * 20
            return tensor.shape

        # Embeddings (Q4)
        print("\nEmbeddings (Q4)...")
        write_q4("embeddings.word_embeddings.weight")
        write_q4("embeddings.position_embeddings.weight")
        write_q4("embeddings.token_type_embeddings.weight")
        write_fp32("embeddings.LayerNorm.weight")
        write_fp32("embeddings.LayerNorm.bias")

        # Transformer layers
        print("Transformer layers...")
        for l in range(num_layers):
            prefix = f"encoder.layer.{l}"
            write_q4(f"{prefix}.attention.self.query.weight")
            write_fp32(f"{prefix}.attention.self.query.bias")
            write_q4(f"{prefix}.attention.self.key.weight")
            write_fp32(f"{prefix}.attention.self.key.bias")
            write_q4(f"{prefix}.attention.self.value.weight")
            write_fp32(f"{prefix}.attention.self.value.bias")
            write_q4(f"{prefix}.attention.output.dense.weight")
            write_fp32(f"{prefix}.attention.output.dense.bias")
            write_fp32(f"{prefix}.attention.output.LayerNorm.weight")
            write_fp32(f"{prefix}.attention.output.LayerNorm.bias")
            write_q4(f"{prefix}.intermediate.dense.weight")
            write_fp32(f"{prefix}.intermediate.dense.bias")
            write_q4(f"{prefix}.output.dense.weight")
            write_fp32(f"{prefix}.output.dense.bias")
            write_fp32(f"{prefix}.output.LayerNorm.weight")
            write_fp32(f"{prefix}.output.LayerNorm.bias")
            print(f"  Layer {l} done")

        # Pooler
        write_q4("pooler.dense.weight")
        write_fp32("pooler.dense.bias")

    size_mb = Path(output_path).stat().st_size / 1024 / 1024
    print(f"\nSaved: {output_path} ({size_mb:.2f} MB)")
    print(f"  Q4 data: {total_q4/1024/1024:.2f} MB")
    print(f"  FP32 data: {total_fp32/1024/1024:.2f} MB")
    print(f"  Compression vs FP32: {63.4/size_mb:.1f}×")


if __name__ == "__main__":
    main()
