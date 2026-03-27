#!/usr/bin/env python3
"""Quick test of LLaDA-8B-Instruct on Apple Silicon.

Uses the diffusion generation approach from the LLaDA paper:
iterative masked denoising rather than autoregressive token generation.
"""

import sys
import time
import numpy as np
import torch
import torch.nn.functional as F
from transformers import AutoTokenizer, AutoModel

MODEL_PATH = "/Users/benaskins/.local/share/models/LLaDA-8B-Instruct"
MASK_ID = 126336  # LLaDA's [MASK] token id


def add_gumbel_noise(logits, temperature):
    """Add Gumbel noise for sampling."""
    if temperature == 0:
        return logits
    noise = torch.zeros_like(logits).uniform_(1e-30, 1.0)
    noise = -torch.log(-torch.log(noise))
    return logits / temperature + noise


def get_num_transfer_tokens(mask_index, steps):
    """Calculate how many tokens to unmask at each step."""
    mask_num = mask_index.sum(dim=-1, keepdim=True)
    base = mask_num // steps
    remainder = mask_num % steps
    num_transfer_tokens = torch.zeros(mask_num.shape[0], steps, dtype=torch.long, device=mask_index.device)
    for i in range(steps):
        num_transfer_tokens[:, i] = base.squeeze(-1)
        if i < remainder.max():
            num_transfer_tokens[:, i] += (i < remainder).long().squeeze(-1)
    return num_transfer_tokens


@torch.no_grad()
def generate(model, prompt, attention_mask=None, steps=128, gen_length=256,
             block_length=256, temperature=0.4, remasking='low_confidence'):
    """Diffusion-based generation for LLaDA."""
    x = torch.full((prompt.shape[0], prompt.shape[1] + gen_length), MASK_ID,
                   dtype=torch.long, device=prompt.device)
    x[:, :prompt.shape[1]] = prompt.clone()

    if attention_mask is not None:
        attention_mask = torch.cat([
            attention_mask,
            torch.ones((prompt.shape[0], gen_length), dtype=attention_mask.dtype, device=prompt.device)
        ], dim=-1)

    prompt_index = (x != MASK_ID)
    num_blocks = gen_length // block_length
    steps_per_block = steps // num_blocks

    for num_block in range(num_blocks):
        start = prompt.shape[1] + num_block * block_length
        end = prompt.shape[1] + (num_block + 1) * block_length
        block_mask_index = (x[:, start:end] == MASK_ID)
        num_transfer_tokens = get_num_transfer_tokens(block_mask_index, steps_per_block)

        for i in range(steps_per_block):
            mask_index = (x == MASK_ID)
            logits = model(x, attention_mask=attention_mask).logits

            logits_with_noise = add_gumbel_noise(logits, temperature=temperature)
            x0 = torch.argmax(logits_with_noise, dim=-1)

            if remasking == 'low_confidence':
                p = F.softmax(logits, dim=-1)
                x0_p = torch.squeeze(
                    torch.gather(p, dim=-1, index=torch.unsqueeze(x0, -1)), -1)
            elif remasking == 'random':
                x0_p = torch.rand((x0.shape[0], x0.shape[1]), device=x0.device)

            x0_p[:, prompt.shape[1] + (num_block + 1) * block_length:] = -np.inf

            x0 = torch.where(mask_index, x0, x)
            confidence = torch.where(mask_index, x0_p, -np.inf)

            transfer_index = torch.zeros_like(x0, dtype=torch.bool, device=x0.device)
            for j in range(confidence.shape[0]):
                _, select_index = torch.topk(confidence[j], k=num_transfer_tokens[j, i])
                transfer_index[j, select_index] = True
            x[transfer_index] = x0[transfer_index]

    return x


if __name__ == "__main__":
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

    prompt_text = sys.argv[1] if len(sys.argv) > 1 else "Write a Go function that checks if a string is a palindrome."

    # Apply chat template
    messages = [{"role": "user", "content": prompt_text}]
    chat_input = tokenizer.apply_chat_template(
        messages, add_generation_prompt=True, return_tensors="pt"
    ).to(device)

    print(f"\nPrompt: {prompt_text}\n")
    print("=" * 60)

    t0 = time.time()
    outputs = generate(
        model,
        chat_input,
        steps=128,
        gen_length=256,
        block_length=256,
        temperature=0.4,
    )
    elapsed = time.time() - t0

    generated = tokenizer.decode(
        outputs[0][chat_input.shape[1]:], skip_special_tokens=True
    )
    tokens_generated = outputs.shape[1] - chat_input.shape[1]

    print(generated)
    print("=" * 60)
    print(f"\n{tokens_generated} tokens in {elapsed:.1f}s ({tokens_generated/elapsed:.1f} tok/s)")
