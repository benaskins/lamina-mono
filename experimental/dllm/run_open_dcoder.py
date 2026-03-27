#!/usr/bin/env python3
"""Quick test of Open-dCoder-0.5B on Apple Silicon."""

import sys
import time
import importlib.util
import torch
from transformers import AutoTokenizer, AutoConfig

MODEL_PATH = "/Users/benaskins/.local/share/models/Open-dCoder-0.5B"

# Load the custom modeling module directly from the model directory
# This avoids trust_remote_code issues with HuggingFace caching
spec = importlib.util.spec_from_file_location(
    "modeling_qwen2", f"{MODEL_PATH}/modeling_qwen2.py"
)
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)
Qwen2ForCausalLM = mod.Qwen2ForCausalLM
MDMGenerationConfig = mod.MDMGenerationConfig

if torch.backends.mps.is_available():
    device = "mps"
    print("Using Metal (MPS) backend")
else:
    device = "cpu"
    print("WARNING: MPS not available, falling back to CPU")

print(f"Loading model from {MODEL_PATH}...")
t0 = time.time()

tokenizer = AutoTokenizer.from_pretrained(MODEL_PATH)
config = AutoConfig.from_pretrained(MODEL_PATH)
config._attn_implementation = "sdpa"
model = Qwen2ForCausalLM.from_pretrained(
    MODEL_PATH,
    config=config,
    torch_dtype=torch.float16,
).to(device).eval()

print(f"Model loaded in {time.time() - t0:.1f}s")

prompt = sys.argv[1] if len(sys.argv) > 1 else "def is_palindrome(s: str) -> bool:"

print(f"\nPrompt: {prompt}\n")
print("=" * 60)

t0 = time.time()
input_ids = tokenizer(prompt, return_tensors="pt").input_ids.to(device)

# Build config manually to avoid GenerationConfig.update() tuple issue
gen_cfg = MDMGenerationConfig()
gen_cfg.max_new_tokens = 256
gen_cfg.steps = 128
gen_cfg.temperature = 0.4
gen_cfg.top_p = 0.95
gen_cfg.mask_token_id = 151643

with torch.no_grad():
    outputs = model.diffusion_generate(inputs=input_ids, generation_config=gen_cfg)

elapsed = time.time() - t0
generated = tokenizer.decode(outputs.sequences[0], skip_special_tokens=True)
tokens_generated = outputs.sequences.shape[1] - input_ids.shape[1]

print(generated)
print("=" * 60)
print(f"\n{tokens_generated} tokens in {elapsed:.1f}s ({tokens_generated/elapsed:.1f} tok/s)")
