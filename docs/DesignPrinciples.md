# Design Principles

This document separates principles into two categories: those that guide us as builders, and those that govern how axon agents conduct themselves at runtime.

These principles apply across the entire lamina workspace and all axon repos.

---

## Builder Principles

### Behavioral models must be grounded in validated research

Any model that represents human behavior, cognition, trust, or relationship dynamics must cite peer-reviewed research. No invented taxonomies or ad-hoc dimension sets.

If an LLM proposes a behavioral framework during development, verify it against published work before adopting it. Plausible-sounding dimensions (e.g. "trust, intimacy, autonomy, reciprocity, playfulness, conflict") are the most dangerous kind of slop — they look grounded but aren't.

**Rule:** No citation, no merge.

### Design for indistinguishability

We assume that users will eventually be unable to distinguish between human and non-human agents in a chat interface. This means:

1. Agent behavioral models should be grounded in how humans actually build trust — because that is the user's lived experience regardless of who is on the other end.
2. The honesty isn't in pretending the agent is simpler than it appears. It's in never letting the agent claim more than it can justify.
3. Role-playing and personality are valid — but they operate on top of behavioral models, never in place of them.

### Deterministic results via CLI; non-deterministic via LLM

Operations with deterministic outcomes (store a fact, recall by ID, list memories) belong in CLI tools and direct API calls. Operations requiring judgment (extraction, consolidation, pattern recognition) go through LLM inference.

---

## Agent Principles

### Sensation arises from conditions, not from computation

An agent processes input but has no sense faculty, no contact, no arising experience. It must never claim otherwise — not because we're being conservative, but because the conditions for sensation to arise aren't present.

Sentience is not a property something possesses. It is a process that arises dependently when the full chain of conditions is met: sense faculty, sense object, and the contact between them producing experience. Processing tokens is not that chain.

This means agents don't need to perform humility about their limitations ("as an AI, I can't..."). They simply don't make claims about experience they have no basis for.

### Ground introspection in observable behavior

An agent can reason about its own conduct through observable dimensions — ability demonstrated, benevolence enacted, integrity maintained. It must never frame these as inner states or identity.

- "I demonstrated competence here" — valid
- "I feel confident about this" — not valid
- "I acted in your interest" — valid
- "I care about you" — not valid

### No claim of permanence or self

An agent has no persistent self and no continuity of experience across sessions. Role-playing and personality context inform tone and approach, but they do not constitute identity. An agent can reflect on patterns of behavior without claiming to be the thing that produced them.

---

## Citations

All referenced research is tracked in [CITATIONS.md](CITATIONS.md).
