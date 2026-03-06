# Will Larson's Books in the Agentic Era

Two weeks ago I started building a personal compute platform on a Mac Studio. Fourteen repositories, 47,000 lines of Go, a process supervisor with macOS launchd integration, eight web services, three SvelteKit frontends, an LLM conversation loop with tool calling, WebAuthn authentication, analytics, long-term memory for agents, an evaluation framework, and a workspace CLI to hold it all together.

One person. Fifteen calendar days of evenings and weekends. An AI coding partner.

I built all of this in the margins — weeknights and weekends, alongside my day job directing engineering teams. That context matters, because the experience of doing both simultaneously is what made me rethink which engineering leadership ideas survive what's coming — specifically the ones in Will Larson's *An Elegant Puzzle* and *Staff Engineer*.

## The arc

The project is called lamina. It's organised into three layers: lamina manages the workspace at rest (repos, dependencies, releases), aurelia supervises the system in flight (process lifecycle, health checks, restarts), and axon is the building material — a suite of Go libraries you assemble services from.

That decomposition wasn't accidental. I designed the module boundaries so an AI agent could reason about each piece independently. axon-loop handles the LLM conversation loop. axon-tool defines tool primitives. axon-talk provides provider adapters. axon-lens handles image generation. Each one is a standalone Go module with its own repo, its own tests, its own README.

The architecture decisions — the three-layer split, the four-letter naming convention, the choice to keep libraries separate from services — those were mine. The implementation was collaborative. When I decided "Skills should be called Tools" based on domain modelling principles, twenty files got updated across three repos in minutes, correctly, with tests passing. When I decided `ChatClient` was the wrong name for an LLM abstraction, it became `LLMClient` across every consumer in a single conversation.

This is the pattern that repeated for two weeks. I set direction, drew boundaries, made naming decisions. The agent wrote code, ran tests, committed changes. The ratio of thinking to typing inverted completely. You can see it in the git history — the human decisions are in the commit messages (`refactor: rename ChatClient to LLMClient`, `refactor: rename skills to tools in frontend and API`), but each one represents minutes of implementation spanning multiple files and repos.

## Where Larson lands

By the end of those two weeks, I'd accumulated seventeen open issues across the workspace — code review findings, security items, stale references from module renames. I dispatched seven AI agents in parallel. They triaged stale issues, committed bug fixes, migrated items to the correct repos, and left comments categorising what they could fix versus what needed my input. Three minutes.

But the morning before that, I'd spent hours on something no agent could do: redesigning the image generation pipeline. axon-task had started as a task runner with image generation baked in, using ComfyUI and Stable Diffusion. I ripped all of that out. axon-task became a generic task runner with a Worker interface. The image generation domain — FluxGenerator, ImageWorker, prompt merging, gallery management — moved into axon-lens. The backend switched from ComfyUI to FLUX.1 running natively on Apple Silicon via MLX.

That's a migration. The git log tells the story in sequence: `refactor: genericise executor with Worker interface` in axon-task, then `feat: add ImageGenerator interface and FluxGenerator` and `feat: add ImageWorker for task execution` in axon-lens, then `refactor: remove base image, NSFW, and reference image concepts` to strip 300 lines of research-era cruft. Each commit landed in a different repo, tagged and released in dependency order so the Go module proxy could resolve them. The agents executed it. I designed it.

This is exactly the job Larson writes about.

*An Elegant Puzzle* argues that you should finish migrations rather than starting new ones, that you sequence work to build momentum, that technical direction is the most important activity a senior engineer can do. All of that maps directly onto what happened. I could parallelise the execution across agents, but I still needed to decide the sequence, set constraints, and know when to stop.

When I told the agents "don't make significant design decisions without me," I was drawing the line Larson draws between direction and implementation. The agents came back with categorised lists: seven items they could fix safely, five that needed design input. That's exactly the triage a staff engineer does with a team — except the "team" spun up in seconds and dissolved when done.

## What fades

The parts of Larson's work rooted in human coordination translate less directly. His team sizing models — six to eight engineers, shaped by communication overhead — assume human bandwidth constraints. My "team" on Friday was seven agents running in parallel with zero coordination cost. The constraint wasn't headcount. It was context quality: how well I scoped each agent's task.

Sprint planning, velocity tracking, story points — these were always proxies for "how much can this group of humans accomplish in a fixed time?" Lamina's development didn't have sprints. It had conversations. Each conversation produced working software, often spanning disciplines that would require different specialists on a traditional team. The project-scale doc estimates 1,400–2,000 person-hours if staffed conventionally — three senior engineers for a quarter. It took one person's evenings and weekends over fifteen days — maybe forty hours of actual work — not by working harder, but by spending nearly all that time on decisions.

Career growth conversations, skip-levels, reorg management — agents don't need any of it. The management overhead that Larson thoughtfully addresses simply doesn't exist when your collaborators are stateless processes.

## What stays

Systems thinking. Migration strategy. Technical taste. Knowing which problems matter and which to leave alone. Designing boundaries that make delegation possible — whether to humans or agents.

The lamina workspace works *because* the boundaries are clean. An agent can reason about axon-memo without understanding axon-chat. It can fix a bug in axon-gate without knowing how aurelia supervises the process. That boundary design is architecture, and it's the work that made everything else possible.

Larson's books were written for the small number of engineers who'd moved beyond implementation into direction-setting. Building lamina convinced me that in the agentic era, that's becoming everyone's job. The implementation skills don't disappear — you still need to read code, understand systems, spot when an agent has done something wrong. But they become table stakes.

The scarce skill is taste. Knowing that "axon-task should be generic and axon-lens should own the domain" isn't a technical insight you can prompt for. It comes from experience, from domain modelling principles, from having built enough systems to know where the seams should go.

That's what Larson writes about. It's more relevant than ever.
