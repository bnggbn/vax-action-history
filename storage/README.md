# VAX Storage SDK

Optional pure Go content-addressed storage, extracted from Pageforge Native on 2026-10-08. Standard library only; independently versioned from the VAX cryptographic SDK.

## Install

```sh
go get github.com/bnggbn/vax-action-history/storage@v0.1.0
```

Import `github.com/bnggbn/vax-action-history/storage`. The module requires Go 1.23. The VAX core module has no dependency on this storage module; consumers may use either or both. C and TypeScript storage ports are not included.

## Action sequence

Open a Store, PutText or PutBlob, Flush objects, then publish your application's revision/root reference. A read Session verifies stored bytes and returns Text, Blob or TextHash. See the executable [example](example_test.go).

| API | Responsibility |
| --- | --- |
| Open / OpenWithOptions | Bounded catalog loading and filesystem validation |
| PutText | UTF-8 content-defined chunks and ordered text tree |
| PutBlob | Opaque bytes; no JSON parsing, canonicalization or note schema |
| Flush | Publish staged catalog entries before publishing application metadata |
| Session | Limits for unique encoded bytes, object count and logical content |
| VerifyText / VerifyBlob | Verify hash, kind, byte length, order and text structure |
| Text / Blob | Restore one root; Blob returns a private byte copy |
| TextHash | Hash ordered verified leaves without constructing a full string |
| Files / VerifyFiles | Record and recheck an already verified dependency closure |

Store and Session are not safe for concurrent use. The caller must serialize mutations and provide exclusive writer ownership across processes. This package does not acquire a library/process lock. Options use bytes for InlineBytes and MiB for CatalogMiB. Open defaults to loose-object writes; OpenWithOptions enables catalog staging when InlineBytes is nonzero.

## Formats and compatibility

The existing Pageforge PFCO v1 and PFCA v1 formats are deliberately preserved. Renaming the blob API does not change kind 2, magic bytes, hash formulas, writer boundaries, root lengths or physical bytes. The original [golden vectors](testdata/objects-v1.json) are unchanged. Existing notes blobs remain opaque bytes; callers continue to interpret their own schemas.

- PFCO: ASCII PFCO, version 1, one kind byte, payload; object ID is SHA-256 of the entire encoded object. Kind 0 is UTF-8 text, 1 is an ordered branch, 2 is a blob.
- Branch: child count 2–16, then each child's 32-byte hash and 8-byte big-endian logical length. Reader verifies actual lengths and order; maximum depth 32.
- Text leaves: 16–64 KiB writer chunks, valid UTF-8 boundaries; empty text is an empty leaf. The fixed Gear policy uses the existing seed and high 13 bits. Readers do not recompute writer boundaries, so older policy roots remain readable.
- PFCA: ASCII PFCA, version 1, 4-byte big-endian count; sorted hash entries contain a 32-byte hash, 4-byte wire length, and full PFCO bytes. Catalog limits are checked independently of reachable-history limits.
- Loose object paths remain hash-prefix/hash.pfo; catalog filename remains catalog.pfca. Blob size, catalog size/count and each read session are bounded.

Blob bytes are not VAX-JCS bytes by default. Applications use the official VAX JCS/SAI APIs for their event commitments and decide whether to commit raw blob bytes, canonical data or restored text. A storage root is not interchangeable with a VAX SAI or a hash of the restored text.

## Errors

Use errors.As to inspect *storage.Error.Kind and errors.Is/errors.As to inspect the preserved underlying cause. Kinds: InvalidInput, InvalidOptions, LimitExceeded, Corrupt, Missing, IO, UnsupportedFormat and UnsafePath. Messages are English diagnostics and may contain filesystem causes; the application's API adapter should map kinds and redact private paths. The SDK has no HTTP statuses, authorization codes or application error dependency.

## Limits

PutText still scans the whole input and rebuilds branches; catalog Flush rewrites its bounded collection. Session retains unique reachable object payloads within its physical budget. Deduplication does not provide logarithmic edits, O(1) history checks or constant process memory.

Loose objects use same-directory temporary files, Sync and hard-link publication without overwriting existing objects; catalogs use temporary files, Sync and Rename. Hard-link support is required. These tests do not establish power-loss durability on every filesystem. No GC, append packs, range reads, inclusion-proof API, local tree edits, signing, synchronization, draft recovery or automatic stale-lock recovery is provided.

File hashes verify bytes against supplied references. Callers must obtain trusted roots/metadata and validate their VAX events; storage alone cannot detect replacement of the entire history. VerifyFiles is only for closures from a successful Session whose matching application metadata has already been checked.

## Tests and releases

Run go test ./... and go vet ./... in storage/. Tests cover Unicode/CRLF, both boundary policies, reuse, kind isolation, exact golden bytes/hashes, depth/byte/count limits, malformed catalogs, corruption, missing files and error causes. All fixtures use temporary storage.

CI runs on Windows and Linux with Go 1.23 and Go 1.26. Module tags use storage/vX.Y.Z. The storage workflow only tests and vets code, with read-only repository permissions; it does not publish GitHub releases or trigger the root C/TypeScript/npm release workflow. Core SAI/JCS bytes are unchanged.
