# Privacy

`skilltrace` stores a sanitized event contract, derived episodes, random evidence tokens, aggregate workflows, and capability metadata. It does not store raw exchanges, prompts, responses, file contents, source paths in presentation records, or credentials.

Sessions are labelled with a project name and a git branch so work can be grouped and compared across a machine. The project name is the bare directory name of the session's working directory and never the path that contains it: a session run in `~/Developer/lofi` is recorded as `lofi`. Parent directories, home directory names, and absolute paths are discarded at ingest and are never written to the catalog or to exported reports. Scan errors likewise name a source by base name only.

Evidence tokens resolve locally only while the source fingerprint still matches. Tokens are not source locators and exported reports contain references rather than raw evidence. HTML reports are self-contained, make no network requests, and use export-only aliases.

Normal text/JSON output exposes stable public error codes and messages, never internal causes. Diagnostic logging is opt-in, bounded, local, and accepts only sanitized messages. Treat local catalogs and diagnostics as private files and apply normal user-directory permissions.
