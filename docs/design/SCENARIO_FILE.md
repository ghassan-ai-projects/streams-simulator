# Scenario file

- **Status**: accepted design for phase P2 of the simulation canvas goal
  (`docs/simulation-canvas/PLAN.md`); contract
  [`scenario-v0.1`](../contracts/scenario-v0.1.schema.json).
- **Why**: a repeatable test should be a file a person, the canvas or a coding agent can write,
  review in git and run, instead of a long list of flags; and a run that someone drove by hand
  should be exportable as such a file.

## 1. Terms

| Term | Meaning | Not to be confused with |
|---|---|---|
| **Scenario file** | A JSON document (`scenario-v0.1`) naming a domain, adapter, seed and world setup, plus an ordered list of director commands | A domain's **profile** (`profiles[]`: entity count, duration, fault weights) and a **suite scenario** (one generated, labelled entry of a graded suite) |
| **Scenario command** | One director operation with its arguments, exactly as the run records it in its command log, without `seq` and `at_ns` | An MCP tool call (the scenario command is what the call records) |

## 2. Document

```json
{
  "schema_version": "0.1",
  "id": "pump-bearing-wear",
  "description": "Bearing wear on pump 1 with a duplicate burst.",
  "domain": "rotating-machinery",
  "adapter": "native-jsonl",
  "seed": 11,
  "world": { "start_time_ns": 1767225600000000000, "profile": "" },
  "commands": [
    { "op": "fault.inject",  "args": { "entity_id": "site-a/pump-1", "fault": "bearing_wear", "onset_ns": 1767225660000000000, "params": null } },
    { "op": "perturb.apply", "args": { "perturbation": "duplicate_burst", "params": null, "from_ns": 1767225720000000000, "until_ns": 1767226200000000000 } },
    { "op": "clock.advance", "args": { "to_ns": 1767229200000000000 } }
  ]
}
```

| Field | Meaning |
|---|---|
| `domain`, `adapter` | Ids resolved in `--domains-dir` and `--adapters-dir` |
| `seed` | World seed |
| `world.start_time_ns` | World start; required |
| `world.profile` | Domain profile, optional |
| `world.entity_ids` | Optional explicit entity ids; when omitted the domain's default population is used |
| `world.noiseless`, `world.force_failure_mode`, `world.clock_multiplier` | Optional world-identity settings, as in the run artifact's `world_config`; `force_failure_mode` is one of the effector failure modes |
| `description` | For readers; the run does not use it |
| `commands` | Executed in order. The run ends after the last command; a scenario without `clock.advance` emits nothing |

Times are absolute nanoseconds, exactly as the run records them, so a run reproduces without
any conversion. A time is written as a plain JSON integer: a fraction or an exponent (`1.7e18`)
is refused, because it is not an exact time. The sink and output directory are not part of the scenario: they are execution
choices (`--sink`, `--sink-target`, `--out`).

### Commands

The closed set, with the arguments the run records (`internal/run`):

| `op` | `args` |
|---|---|
| `clock.advance` | `to_ns` |
| `fault.inject` | `entity_id`, `fault`, `onset_ns`, `params` (object or null) |
| `fault.clear` | `fault_id`, `at_ns` |
| `perturb.apply` | `perturbation`, `params` (object or null), `from_ns`, `until_ns` |
| `perturb.clear` | `perturb_id` |
| `effector.invoke` | `effector`, `entity_id`, `command_id`, `args` (object), `at_ns` |
| `entity.add` | `entity_id`, `at_ns` |
| `entity.retire` | `entity_id`, `reason`, `at_ns` |

A clock advance carries only its target. Whether a live session waited for its consumer at an
advance (`await_consumer` in the command log) is pacing of that session, not an input to the
world: the consumer's own effector calls are recorded as commands at their times, so the world
is the same without the wait, and a scenario run has no consumer to wait for. A scenario run
records its advances with `await_consumer: false`, as replay does.

`env.inject` is not a scenario command: it is record-only today (`DECISIONS.md` D-14) and
changes nothing a replay can reproduce. The run's lifecycle records (`world.create`,
`run.begin`, `run.end`, `clock.run`) are not commands either.

## 3. Execution

`streamsim run --scenario <file>` validates the file against `scenario-v0.1`, resolves the
domain and adapter, creates the run from the scenario's world settings and executes the
commands **through the run module's command executor**, the same code that replays an artifact's
command log. There is one execution path:

- `run` with flags (`--domain`, `--seed`, `--fault`, `--perturb`, `--effector`, `--duration`, …)
  first builds a scenario in memory, then executes it.
- `run --scenario` refuses the flags that the scenario defines (`--domain`, `--adapter`,
  `--seed`, `--profile`, `--start-time`, `--duration`, `--fault`, `--perturb`, `--effector`)
  with `usage_invalid`.

**Run-wide rules.** Three rules hold for every run, however it is driven (flags, scenario file,
director, replay), so a scenario check and a live run never disagree:

- A world starts between 0 and `world.MaxStartNS` (half the 64-bit range, about the year
  2116), so no event time can overflow.
- A run's clock may not move more than 30 days of simulated time past its start
  (`run.MaxRunSpanNS`): a run keeps its records in memory, so the bound makes every run finite.
  `--duration` is bounded the same way.
- A perturbation window may not end before it starts (`until_ns` 0 means no end): such a window
  could never act, so applying it is refused rather than silently doing nothing.
- An entity may not be added at a time before the world clock: the world would have to simulate a
  history that already passed.

**Checks before execution.** Nothing is delivered and no sink is opened until the scenario has
passed three checks.

1. **Contract.** The file must satisfy `scenario-v0.1`. A command's `op` selects the schema of its
   `args`, so a malformed command is reported at the argument concerned. Ids and names are 1 to
   256 characters without control characters. A value the schema admits but a scenario cannot
   hold (a seed beyond 64 bits) is reported at its field.
2. **Static check.** Everything that can be known before the run unfolds is checked, and every
   command's violations are reported, not only the first command's. A command with an inexact time is not checked
   further.

   | Checked | Rule | Pointer |
   |---|---|---|
   | `world.profile` | One of the domain's profiles | `/world/profile` |
   | World settings | The world can be created from them | `/world` |
   | Every time argument | A plain integer that fits in 64 bits | `/commands/<i>/args/<time>` |
   | `clock.advance` | `to_ns` not before the previous advance or the start, and within the run span | `/commands/<i>/args/to_ns` |
   | `fault.inject` | The domain declares the fault; `params` holds only parameters it accepts (`severity`, finite, non-negative) | `/commands/<i>/args/fault`, `/commands/<i>/args/params` |
   | `perturb.apply` | The perturbation is in the catalog; `params` holds only its declared parameters, in range; the window can act | `/commands/<i>/args/perturbation`, `/commands/<i>/args/params`, `/commands/<i>/args/until_ns` |
   | `effector.invoke` | The domain declares the effector; `args` satisfy its `args_schema` | `/commands/<i>/args/effector`, `/commands/<i>/args/args` |

3. **Dry run.** Which entities exist, and which fault and perturbation ids a run assigns, depend
   on the run as it unfolds: churning domains birth and retire entities on their own schedule.
   These facts are not guessed. The scenario is executed through the run's own command executor
   on the same world with emission turned off, so nothing is delivered and nothing is opened. The
   first command the run would refuse is reported at `/commands/<i>`, with the run's reason (an
   entity that does not exist or was retired, an id already in use, a `fault.clear` or
   `perturb.clear` naming an id the run never assigned, a backdated `entity.add`). The dry run
   costs one extra pass of the world: about 5% on a three-day k8s run, and about 0.6 s on a
   three-day pond run.

A violation in a scenario file is `scenario_invalid` (exit 1, `source` = the file). A scenario
built from flags reports the same violations as `usage_invalid` (exit 2); each issue names the
flag occurrence behind it (`--effector "start_aerator@site-a/pond-1@30": …`) and carries no
pointer, since the user wrote flags, not a document. The run module also refuses an unknown
profile whenever a run is created, whatever the surface, replay included: an artifact recorded
with a profile its domain does not declare no longer replays (pre-release; no backward
compatibility).

**Refusals during execution.** Scenario runs, dry runs and replay share one command executor with
an explicit refusal policy:

- An `effector.invoke` that the world refuses as it runs, through an interlock or a command id
  reused for a different request, is a recorded outcome, as a live consumer would receive it,
  and the scenario goes on. A scenario may reuse a command id deliberately, to test idempotency.
- Any other refusal stops the scenario at that command. The dry run finds these before anything
  is delivered. A failure of the execution itself, such as a sink that cannot be written, fails
  the run with `run_failed`.
- Replay keeps every effector refusal, because the log it replays already holds the refusals the
  original run got.

## 4. Export

`streamsim scenario from-run <run.json> --out <file> [--domains-dir d] [--adapters-dir d]`
writes the scenario that reproduces the run and prints `{"scenario_written": <file>, "commands": <n>, "omitted": <m>, "refused": <r>}`, where `omitted`
counts the command-log entries that are not scenario commands. It copies domain, adapter,
seed and world settings from the artifact, and copies every command-log entry whose `op` is a
scenario command, in order, without `seq` and `at_ns` and, for a clock advance, without
`await_consumer` (§2). The scenario owns its copies. Before writing, the export checks the
scenario against `scenario-v0.1` and then exactly as `run --scenario` would (§3), against the
domain and adapter found in the given directories. The artifact's domain and adapter digests must match the
ones found (otherwise `replay_mismatch`). Effector calls the world refused at admission (an
entity that does not exist or has gone, an undeclared effector, refused arguments) changed
nothing, so they are dropped and counted in `refused`. A run whose remaining commands cannot all
be expressed as a runnable scenario writes nothing and fails with `scenario_invalid`: its `source` is the
artifact, and its pointers locate the scenario that would have been written. `replay` and `verify` still reproduce
such a run from its artifact.

**Round trip.** For a run whose command log holds only scenario commands (every run made by the
CLI), `run` → `scenario from-run` → `run --scenario` yields an identical command log, trace
digest and world digest. For a director-driven run that exports, the trace and world digests are identical; the command log differs only by the lifecycle and record-only entries the
scenario omits and by `await_consumer`, which a scenario run records as `false`.

## 5. Determinism

A scenario is a pure function of its bytes and the referenced domain and adapter. The run
artifact of `run --scenario` records the same inputs as any run (domain and adapter digests,
seed, world config, command log), so `verify` works unchanged.
