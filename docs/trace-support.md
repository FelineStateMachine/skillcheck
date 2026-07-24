# Trace Support

| Source | Support | Important limits |
| --- | --- | --- |
| Codex native JSONL | Supported | Unknown record versions become exclusions; model/actor fields depend on source evidence. |
| Claude native history | Supported | Tool correlation and subagent attribution degrade conservatively when records are incomplete. |
| Hugging Face normalized rows | Conditional | Requires a supported harness and retained raw trace; normalized-only rows are rejected. |

Health is reported as full, partial, unsupported, malformed, or unavailable. Partial sources remain inspectable with capability and exclusion disclosures. Unsupported formats are never guessed into the canonical event contract. Append scanning requires matching file identity, parser revision, and digest checkpoints; replacement, truncation, or digest mismatch triggers a full scan.

Benchmark output reports aggregate and harness/model/tier slices. Private acceptance manifests stay local and must contain sanitized evidence-token expectations rather than transcript content.
