#!/usr/bin/env python3
"""Quick test of Apple's DiffuCoder-7B-Instruct on Apple Silicon."""

import sys
import time
import torch
from transformers import AutoTokenizer, AutoConfig, AutoModel

MODEL_PATH = "/Users/benaskins/.local/share/models/DiffuCoder-7B-Instruct"

# Use MPS (Metal) on Apple Silicon
if torch.backends.mps.is_available():
    device = "mps"
    print("Using Metal (MPS) backend")
else:
    device = "cpu"
    print("WARNING: MPS not available, falling back to CPU")

print(f"Loading model from {MODEL_PATH}...")
t0 = time.time()

tokenizer = AutoTokenizer.from_pretrained(MODEL_PATH, trust_remote_code=True)
model = AutoModel.from_pretrained(
    MODEL_PATH,
    trust_remote_code=True,
    torch_dtype=torch.float16,
).to(device)

print(f"Model loaded in {time.time() - t0:.1f}s")

# Default prompt or take from CLI
prompt = sys.argv[1] if len(sys.argv) > 1 else "Write a Go function that checks if a string is a palindrome."

print(f"\nPrompt: {prompt}\n")
print("=" * 60)

t0 = time.time()
inputs = tokenizer(prompt, return_tensors="pt").to(device)

# DiffuCoder uses diffusion generation — the generate method is on DreamGenerationMixin
with torch.no_grad():
    outputs = model.diffusion_generate(
        inputs["input_ids"],
        attention_mask=inputs["attention_mask"],
        max_new_tokens=512,
        output_history=False,
        return_dict_in_generate=True,
        steps=128,
        temperature=0.4,
        top_p=0.95,
        top_k=None,
        alg="origin",
    )

elapsed = time.time() - t0
generated_ids = outputs.sequences[0]
# Strip the input tokens
generated = tokenizer.decode(generated_ids[inputs["input_ids"].shape[1]:], skip_special_tokens=True)
tokens_generated = len(generated_ids) - inputs["input_ids"].shape[1]

print(generated)
print("=" * 60)
print(f"\n{tokens_generated} tokens in {elapsed:.1f}s ({tokens_generated/elapsed:.1f} tok/s)")
