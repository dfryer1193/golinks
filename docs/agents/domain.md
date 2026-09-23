# Domain Docs

How the engineering skills should consume this repo's domain documentation when exploring the codebase.

## Before exploring, read these

- **CONTEXT.md** at the repo root, or
- **CONTEXT-MAP.md** at the repo root if it exists: it points at one CONTEXT.md per context. Read each one relevant to the topic.
- **docs/adr/**: read ADRs that touch the area you're about to work in. In multi-context repos, also check src/<context>/docs/adr/ for context-scoped decisions.

If any of these files don't exist, proceed silently. Don't flag their absence; don't suggest creating them upfront. The /domain-modeling skill (reached via /grill-with-docs and /improve-codebase-architecture) creates them lazily when terms or decisions actually get resolved.

## File structure

Single-context repo (most repos):

/
├── CONTEXT.md
├── docs/adr/
│   ├── 0001-event-sourced-orders.md
│   └── 0002-postgres-for-write-model.md
└── src/

## Use the glossary's vocabulary

When your output names a domain concept (in an issue title, a refactor proposal, a hypothesis, a test name), use the term as defined in CONTEXT.md. Don't drift to synonyms the glossary explicitly avoids.

If the concept you need isn't in the glossary yet, that's a signal: either you're inventing language the project doesn't use (reconsider) or there's a real gap (note it for /domain-modeling).

## Flag ADR conflicts

If your output contradicts an existing ADR, surface it explicitly rather than silently overriding:

> _Contradicts ADR-0007 (event-sourced orders), but worth reopening because…_

## Update CONTEXT.md inline

When a term is resolved, update CONTEXT.md right there. Don't batch these up: capture them as they happen. Use the format in CONTEXT-FORMAT.md.

## CONTEXT.md should be a glossary

CONTEXT.md should be totally devoid of implementation details. Do not treat CONTEXT.md as a spec, a scratch pad, or a repository for implementation decisions. It is a glossary and nothing else.

## Offer ADRs sparingly

Only offer to create an ADR when all three are true:

1. Hard to reverse: the cost of changing your mind later is meaningful
2. Surprising without context: a future reader will wonder "why did they do it this way?"
3. The result of a real trade-off: there were genuine alternatives and you picked one for specific reasons

If any of the three is missing, skip the ADR. Use the format in ADR-FORMAT.md.