# P2c review record — CLI to wiring only

- **Reviewed**: the P2c working-tree diff against `87f49e3` (2026-10-12)
- **Reviewer**: one independent subagent (module boundaries and behaviour preservation), per the
  owner's direction to keep reviews light.
- **Outcome**: behaviour byte-identical, module boundaries sound. 0 High, 2 Medium, 5 Low, 2 carried.
  Resolved below or carried in [FINDINGS.md](../../../governance/FINDINGS.md).

| # | Sev | Finding | Disposition |
|---|---|---|---|
| M1a | Med | The CLI gate matched `pkgDir == internal/cli/internal/app`, so a nested package (`app/logic`) passed | Fixed: any package under `internal/cli` other than the facade, `app`, `files` and `process` fails; synthetic cases for a nested package and for a package the CLI never had, and an injected `app/logic/logic.go` in the real tree fails the gate |
| M1b | Med | Logic appended to a listed file passed; one 1900-line total was the only cap | Fixed: each `cliAppFiles` entry declares the facades it may import (only `score.go` imports `grading`, only `suite.go` imports `suite`, and so on) and its own line cap (current size plus headroom); `encoding/json` only in `cli.go`; an added import and an added 40 lines in the real `score.go` each fail the gate |
| M1c | Med | The CLI's own edges were not checked as wiring | Fixed: `files` may import `fmt` and `os` only (40 lines), `process` `io`, `os` and `time` only (35 lines); a module import, a socket import and growth past the cap each fail; an injected `net/http` in the real `process.go` fails the gate |
| M2a | Med | The suite workflow (catalog load and describe, profile admission, file name, marshal, write, `SampleNS`) was still in the CLI | Fixed: `suite.Publish` (`suite/internal/app`, `suite/internal/files`); `DefaultSampleNS` and `FileName` are named in the suite domain; an unknown profile is a typed `*suite.UnknownProfileError` the CLI restates as the usage failure with the same message. `suite` becomes a core with an edge |
| M2b | Med | `run`, `replay` and `scenario from-run` still loaded inputs and sequenced several run calls in the CLI | Fixed: `run.RunScenario`, `run.ReplayFile` and `run.ExportScenarioFile` each load their own inputs, so each command makes one call; a scenario violation is a typed `*run.ScenarioCheckError` so the CLI still restates it for a scenario file or for flags; error text is unchanged |
| L1 | Low | `--start-time` range check used `world.MaxStartNS` in the CLI | Fixed without changing behaviour: `run.MaxStartNS` re-exports the world's bound (as `MaxRunSpanNS` is), so the CLI no longer imports `world`; the check and its usage text are unchanged |
| L2 | Low | `cmdVerify` decided from `report["matches"] == true` on a map | Fixed: `cmdVerify` and `cmdReplay` share a typed `*run.ReplayResult` |
| L3 | Low | The rule for when `adapter verify` fails lived in the CLI | Fixed: `adapter.VerifyFailure(res)` returns the failure message and whether a declared check diverged; `adapterVerdict` only shapes the document |
| L4 | Low | `run.ExportScenario` returned two bare ints | Fixed: it returns `ScenarioExport{Scenario, Omitted, Refused, Raw}`; only the file entry is on the facade |
| L5 | Low | Garbled doc comment on `transport.BindOperator` | Fixed |
| C1 | Low | The `files` Read/Write wrappers are duplicated across modules | Carried (R2c-01) |
| C2 | Low | Modules emit the `streamsim: ` prefix themselves | Carried (R2c-02); R2-03 already covers the doubled prefix |
