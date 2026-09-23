# Triage Labels

The canonical triage label vocabulary for this repo. Each label string equals its name.

| Label | Meaning |
|---|---|
| `needs-triage` | Work is incomplete or requires clarification before further action. |
| `needs-info` | Information is missing or unclear; add context or research. |
| `ready-for-agent` | Ready for an engineer to work on; assign to a team member. |
| `ready-for-human` | Needs human judgment (e.g., design decisions, policy changes). |
| `wontfix` | Explicitly deferred; not blocked, just not happening now. |

## Usage

- New issues start with `needs-triage`.
- When an issue has enough context and is ready for work, change to `ready-for-agent` or `ready-for-human`.
- When an issue is resolved or won't be worked on, apply `wontfix`.