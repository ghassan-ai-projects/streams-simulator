# Validation — 2026-10-14

Source commit: `12c1f2e`. No Go production files, contracts, domain data or gates changed.

## Passed

| Check | Evidence |
|---|---|
| `make quick` | Build, vet, short tests and architecture rules passed with required local socket access. |
| `make ci-check` | Formatting, documentation, tidy, build, vet, lint, gate ratchet, coverage, tests, deadcode, vulnerability scan and fuzz soak passed. |
| `make canvas-check` | Audit, formatting, types, lint, dependencies, duplication, unused exports, coverage, build and Chrome end-to-end checks passed. |
| Canvas tests | 52 files, 232 tests passed. |
| Chrome end-to-end | 29 tests passed against the real binary and files. |
| Final desktop matrix | 36 captures; no axe findings, console errors or document-width overflow. |
| Enola enforcement | `cycles,layers` at confidence 0.8: no structural regression. |
| Source review | Delegated diffs reviewed; `git diff main -- test/architecture scripts AGENTS.md` is empty. |
| Whitespace | `git diff --check` passed. |

Build, tests and vulnerability scanning used Go 1.27.2. Homebrew golangci-lint 2.14.0 used Go 1.27.1 for analysis through a temporary local launcher because its importer cannot decode the newer toolchain's export format. The older duplicate lint binary was excluded from PATH. The launcher changed no tracked command, configuration or check. Pre-commit is not installed.

## Existing behavior-pin blocker

The final `make phase-gate` passed CI, canvas and architecture, then failed `behaviour-pin-check` against the committed pin. Twenty entries differ: eighteen run artifacts, the manifest file and unknown-command stderr.

To distinguish an introduced regression from an existing mismatch, the unchanged starting commit `9500b76` was exported to an isolated temporary checkout and its original behavior-pin script ran under the same toolchain. Its 201 generated hashes are byte-identical to the final source result. Both differ from the committed pin by the same twenty entries.

The identical 201-entry comparison outputs have SHA-256 `f98f78bd424bed37bdf0d0305f067f3e35ec959fbc87dc823a6e76252d9ec180`. The before-versus-after diff is empty; the inventory remains temporary test output rather than a second committed pin.

The comparison proves this UI round did not change those CLI outputs. The committed pin was not refreshed or weakened. The complete phase gate remains blocked by that pre-existing mismatch; this is not an all-gates-green or release claim.

## Architecture comparison

The baseline was pinned before edits and was not re-pinned. Both receipts use Enola 0.4.25 and extractor v276:

- Before: `sha256:e3517a8e642e569b4e61ecf235a4199abf72d60fe2d50e4199841fe290163dee`.
- After: `sha256:377637234db7279629732c43ec492d9f8b1b559575e502456cd95dc51888a618`.

The diff reports zero introduced findings. Changed symbols and coupling stay in the canvas presentation/model work. Scope flags were reviewed: `canvas/.tsbuild/app/src/ui/plant` is ignored generated build output, `canvas/src/ui` is the affected parent UI module, and this task's `docs` folder is the requested review evidence. No unrelated tracked source change was found. Go architecture tests and canvas dependency checks remain authoritative for detailed ownership.

The screenshots cover the implemented desktop flows. These checks do not establish full WCAG conformance, every domain's conceptual completeness, deferred P10/P11 capabilities or the full campus performance budgets.
