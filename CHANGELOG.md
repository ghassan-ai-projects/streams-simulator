# Changelog

This project has not established a formal release stream yet. Until it does, changes are recorded in pull requests and the repository history.

## Unreleased

- **CLI output contract** (`docs/design/CLI_SURFACE.md`, `cli-output-v0.1`): every command except
  `help` and `mcp` prints exactly one JSON document on stdout, for failures too. Failures carry a
  stable code, the file at fault, every contract issue with a JSON pointer when known, and for
  `verify`/`adapter verify`/`score` the result that diverged.
- **Behaviour change:** argument errors that are not flag-parse errors (missing or unknown verb,
  missing required argument, malformed `--fault`/`--effector` value) now exit 2 with
  `usage_invalid` instead of 1. A failed `verify` prints the error document with the replay
  report as `result` instead of the bare report.
- Domain, adapter, run-artifact, verdict, label and ledger failures are typed
  (`model.ValidationError`) instead of text; `domain.ErrUnknownDomain` names an unknown id.
- **Behaviour change:** argument values are checked before running: `--sink` must be a known sink,
  `http-push` needs `--sink-target`, `--duration`, `suite --n`, `refconsumer --window` and
  `--threshold` must be positive, `suite` needs `--domain`, and `--profile` must name one of the
  domain's profiles (a `run` with an unknown profile used to succeed and ignore it). All exit 2.
- **Behaviour change:** `score` checks the verdict and label against their contracts, decodes the
  ledger and state history strictly, and loads the run artifact first; `-h` works after a verb
  (`domain validate -h`, `device -h`); flags may precede a verb (`adapter --adapters-dir d verify f`).
- `run.ErrDigestMismatch`, `adapter.Lookup`/`adapter.ErrUnknownAdapter` and
  `model.ValidateGroundTruth` are new typed entry points.
- Gates: `godox` and `nolintlint` in lint; a consumer-knowledge gate with a stale-exception check; a
  pinned `device serve` transcript; failure documents in the behaviour pin (stdout and stderr
  separately); `make behaviour-pin-check`; `make phase-gate`.

- Introduced a curated public documentation set under `documentation/`.
- Documented the distinction between the public set and the engineering archive under `docs/`.
- Added explicit limitations and evidence guidance for the current working-tree inventory.

When a release process is established, each entry should name contract changes, behavior changes, migration notes, and evidence for the exact release commit.
