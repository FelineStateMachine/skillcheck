# Privacy

`skilltrace` stores a sanitized event contract, derived episodes, random evidence tokens, aggregate workflows, and capability metadata. It does not store raw exchanges, prompts, responses, file contents, source paths in presentation records, or credentials.

Evidence tokens resolve locally only while the source fingerprint still matches. Tokens are not source locators and exported reports contain references rather than raw evidence. HTML reports are self-contained, make no network requests, and use export-only aliases.

Normal text/JSON output exposes stable public error codes and messages, never internal causes. Diagnostic logging is opt-in, bounded, local, and accepts only sanitized messages. Treat local catalogs and diagnostics as private files and apply normal user-directory permissions.
