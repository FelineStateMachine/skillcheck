# Configuration

The catalog and settings live in the operating system's user configuration directory. `SKILLTRACE_CATALOG` overrides the catalog path, `SKILLTRACE_SKILL_ROOT` overrides the TUI skill root, and `SKILLTRACE_SCAN_INPUT` supplies an optional launch-time scan source.

Project policy is read from the repository and takes precedence by stable definition ID over global policy. Local policy is machine-specific and must not be committed. Policy documents use the strict versioned YAML contract: unknown or duplicate keys are rejected. Preview before apply; apply is revision-bound and reports conflicts rather than overwriting concurrent edits.

Custom source roots are harness-specific. Source forget, skill clear, and catalog reset remove derived catalog state only; they never modify traces, installed skills, imports, or exported reports.
