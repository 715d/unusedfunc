# Context Files

This directory is for temporary multi-session handoff notes whose decisions, blockers, or state are not readily recoverable from code, Git, or task tracking.

## File Naming Convention

When creating new context files during development:

Format: `YYYYMMDD-topic-description.md`

Examples:
- `20250831-rta-callgraph-analysis.md`
- `20251106-performance-investigation.md`

## Maintenance Policy

**Create context files** when:
- A multi-session handoff needs non-obvious decisions, blockers, or investigation state

**Delete context files** after verified completion:
- Promote durable findings to the appropriate long-lived document
- Delete the remaining handoff note; Git preserves historical context

Main documentation structure:
- `docs/architecture.md` - Design decisions and non-obvious constraints
- `docs/workflows.md` - Commands and repeatable procedures
- `docs/reference/` - Project-specific third-party contracts and operational details
