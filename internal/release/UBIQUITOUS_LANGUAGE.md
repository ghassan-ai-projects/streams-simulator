# Ubiquitous language — release

A *release* here is the signed statement of exactly what a benchmark claims to
be: which simulator build, which domain and adapter files, which reference
consumer and suite settings, written by whom and independently reviewed by
whom.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Release manifest | The document that pins one release | `Body`, `Document` | `release-manifest.json`, `release-manifest-v0.1` |
| Manifest body | The identity and digests, encoded canonically (RFC 8785); the bytes a signature covers | `Canonical` | `manifest` |
| Identity | Simulator build, author, independent reviewer and the instant of stamping | `Identity` | `sim`, `author`, `reviewer`, `created_at` |
| File digest | The canonical digest of one domain or adapter file's bytes, keyed by file name | `fileDigests` | `domains`, `adapters` |
| Seed | The 64 hex characters of an ed25519 private key | `Seed` | `--key` file |
| Signature | The ed25519 signature of the SHA-256 digest of the canonical body, hex encoded; empty when unsigned | `Sign` | `signature` |
| Publication | Digest, build, sign and write, in that order | `Publish` | `manifest` command |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `cli` manifest code | the `release` module | digesting, signing and writing a manifest are release rules, not command-line wiring |
