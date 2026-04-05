# Update *.studio.internal references

**Area:** lamina CLAUDE.md, docs
**Discovered:** 2026-04-03

The lamina root CLAUDE.md still references `*.studio.internal` as the
internal domain convention. This was retired in favour of per-host naming:
`*.hestia.internal` and `*.limen.internal`.

Grep the workspace for any remaining `studio.internal` references in
configs, docs, and scripts. Update them all.
