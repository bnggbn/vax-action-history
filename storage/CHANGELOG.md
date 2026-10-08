# Changelog

## v0.1.0 — 2026-10-08

- Extract Pageforge content-addressed storage and tests into an optional Go module.
- Preserve PFCO/PFCA wire bytes, kind numbers, chunk policy and existing golden hashes.
- Generalize notes APIs to PutBlob/VerifyBlob/Blob without changing payloads.
- Introduce storage-specific error kinds with preserved causes and no application dependencies.
- Document writer ownership, budgets, publication limits and unimplemented operations.
- Add read-only Windows/Linux test CI with independent storage module tags.
