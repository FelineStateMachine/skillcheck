# skilltrace

`skilltrace` discovers local AI skills and turns supported Codex, Claude, and eligible Hugging Face traces into sanitized usage episodes, workflows, comparisons, and offline reports.

## Install and run

Download the native archive for your platform or build with Go 1.24:

```sh
go build ./cmd/skilltrace
./skilltrace
```

The no-argument command opens the TUI. Headless commands never prompt and support `--format json`:

```sh
skilltrace discover --skill-root .codex/skills --format json
skilltrace source scan --input trace.jsonl --harness codex --format json
skilltrace analyze --skill nzip --scope current --format json
skilltrace compare --skill nzip --left left.yaml --right right.yaml --format json
skilltrace policy validate --file policy.yaml --format json
skilltrace export html --output report.html --format json
```

JSON stdout uses a versioned envelope. Long-running progress is NDJSON on stderr. Exit codes are `0` success, `1` failure, `2` usage, `3` unavailable/not found, `4` conflict, `5` writer busy, and `130` cancelled.

See `docs/configuration.md`, `docs/privacy.md`, and `docs/trace-support.md` for operational details.

## Verification

Run `make verify` for formatting, vet, Go tests, report tests/build, and generated-file drift checks. Release archives are produced by `VERSION=1.0.0 make package`.
