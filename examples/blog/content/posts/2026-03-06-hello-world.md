---
title: Hello, world
date: 2026-03-06
tags: [meta, go]
summary: First post — rolling a blog with axon.
---

This blog is a single Go binary built on [axon](https://github.com/benaskins/axon).

## How it works

Markdown files live in `content/posts/`. Each file has YAML front matter at the top:

```yaml
title: Hello, world
date: 2026-03-06
tags: [meta, go]
```

At startup the server reads every `.md` file, parses the front matter, and renders the body to HTML with [goldmark](https://github.com/yuin/goldmark). Posts are sorted newest-first and served from memory.

## Why not Hugo

Nothing wrong with Hugo. But a Go binary that serves its own content means:

- One deployment artifact
- Full control over routing, headers, caching
- Easy to extend — add an API, wire in auth, bolt on search
- Same stack as everything else in the cluster

The whole thing is about 200 lines of application code. The axon toolkit handles graceful shutdown, request logging, and Prometheus metrics.
