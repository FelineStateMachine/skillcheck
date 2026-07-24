# skilltrace

`skilltrace` discovers local AI skills and turns supported Codex, Claude, and eligible Hugging Face traces into sanitized usage episodes, workflows, comparisons, and offline reports.

## Machine-wide sync

`skilltrace sync` walks your Claude and Codex trace directories, scans everything new or changed since the last run, and skips the rest:

```sh
skilltrace sync                              # ~/.claude/projects and ~/.codex/sessions
skilltrace sync --claude-root PATH --codex-root PATH
```

Sessions are labelled by project (the working directory's folder name only — `~/Developer/lofi` is recorded as `lofi`) and git branch. Tool use is captured with real tool names and success/failure, so the workflow view shows which tools a skill leans on, how often, and where it errors.

## Install and run

Download the native archive for your platform or build with Go 1.24:

```sh
go build ./cmd/skilltrace
./skilltrace
```

The no-argument command opens the TUI. With no `--skill-root`, discovery searches the
project and global skill directories for every supported harness — `.codex/skills` and
`.claude/skills` under the working directory, and under your home directory.

```sh
skilltrace                                   # discovery and episodes
skilltrace --skill-root ~/.claude/skills     # pin discovery to one root
skilltrace --scan-input trace.jsonl          # enable the scan key
skilltrace --policy policy.yaml              # enable the policy view
skilltrace --skill nzip --left a.yaml --right b.yaml   # enable the comparison view
```

TUI routes: discovery (`enter` opens episodes), episodes (`w` opens the workflow map),
plus the policy and comparison views when their inputs are configured. Press `?` for keys.

Headless commands never prompt and support `--format json`:

```sh
skilltrace discover --skill-root .codex/skills --format json
skilltrace source scan --input trace.jsonl --harness codex --format json
skilltrace analyze --skill nzip --scope current --format json
skilltrace compare --skill nzip --left left.yaml --right right.yaml --format json
skilltrace policy validate --file policy.yaml --format json
skilltrace export html --output report.html --format json
```

JSON stdout uses a versioned envelope. Long-running progress is NDJSON on stderr. Exit codes are `0` success, `1` failure, `2` usage, `3` unavailable/not found, `4` conflict, `5` writer busy, and `130` cancelled. The TUI follows the same contract: it exits `2` when stdout is not a terminal and `130` on `ctrl+c`.

See `docs/configuration.md`, `docs/privacy.md`, and `docs/trace-support.md` for operational details.

## Verification

Run `make verify` for formatting, vet, Go tests, report tests/build, and generated-file drift checks. Release archives are produced by `VERSION=1.0.0 make package`.
