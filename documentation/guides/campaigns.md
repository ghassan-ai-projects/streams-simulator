# Campaigns

> Status: Implemented (generation and offline scoring); running sessions needs a plant wire host that is your own. Authority: `internal/suite`, `internal/plantwire`, `contracts/campaign-v0.1.schema.json` and the [CLI reference](../reference/cli.md#suite-campaign). Verified by: `suite campaign` run on the committed pump-station spec (160 runs generated). Last verified: 2026-11-03.

A campaign is an ordered list of runs, each with one or more injected faults and a phase label, plus one session plan per run. It is for testing a consumer (a stream processor behind a plant wire host) over many runs, with the world continuing from run to run, and for scoring each run from the files it left. The campaign is a pure function of its spec: the same spec gives the same files byte for byte.

## Spec shape

The committed example is `examples/pump-station/campaign/first-slice.campaign-spec.json`. The fields, as the CLI reference defines them:

| Field | Meaning |
| --- | --- |
| `seed` | Seeds the campaign and each run's seed. |
| `host_socket`, `policies` | Copied into every session plan. Set them to the host's own and generate again. Without `policies` every command is refused. |
| `duration_s` | Length of each run, in world seconds. |
| `fault_start_window_s` | A fault starts within this window. |
| `negative_duration_s` | How long the negative-class fault lasts. |
| `phases[]` | `name`, `domain`, `runs`, and either `negative_share` with `faults` (weights of the positive faults), or the multi-fault options below. |

Every domain a phase names must be in `--domains-dir`. The schema is [`campaign-v0.1`](../../contracts/campaign-v0.1.schema.json).

## Generate

```bash
streamsim suite campaign --spec examples/pump-station/campaign/first-slice.campaign-spec.json \
  --domains-dir examples/pump-station --out <dir>
```

`--out` must not already hold a campaign. It receives `campaign.json` (per run: phase, domain, seed, duration, session-plan file; no fault, expected action, deadline or label) and `run-NNNN.session-plan.json`. Each plan's `start_s` starts its world where the previous run ended, so a consumer that keeps state across runs sees time move forward. The injected fault appears only in the session plans, which go to `plantwire connect` unread.

### Several faults per run

- `faults_per_run` (default 1): positive faults per run. With more than 1, the phase's `runs` times `faults_per_run` slots are split by weight (largest remainders), shuffled with the campaign seed and dealt to the runs; `negative_share` is then refused.
- `negatives_per_run` (default 0): negative-class disturbances added to every run.
- Each disturbance of a run is on a different entity that declares the fault and starts inside `fault_start_window_s`. A phase is refused when `faults_per_run` plus `negatives_per_run` exceeds the entities available.
- `faults_per_run: 0` with no `faults` is a warm-up phase: no disturbance, or only the negatives.

## Run one session

```bash
streamsim plantwire connect --plan <dir>/run-0001.session-plan.json \
  --domains-dir examples/pump-station --out <run-dir> --slice-s 60
```

- `--out` must be a new folder for every session; a folder holding any run file is refused.
- The plan's `host_socket` must be an absolute path of at most 103 characters in a directory only you can enter. The committed `/tmp/streamsim-host.sock` is a placeholder.
- A campaign's session plans carry no `horizon_s`, so a session does not end by itself: interrupt `connect` (SIGINT) when world time reaches `duration_s`. Closing the host first leaves an incomplete run.
- Pace the host with a `--delay-ms` or a session writes a very large run folder; a 48 h run takes about 585 MB at `--slice-s 60`. Plan disk use for the whole campaign (160 runs at that size is on the order of 90 GB).

## Labels and scoring

`connect` writes `label.json` when the plan holds exactly one disturbance, and `labels.json` (an array in plan order) otherwise, empty for a warm-up. Labels name the fault: never give them to the consumer. Write the consumer's verdict to `<run-dir>/verdict.json`, then:

```bash
streamsim score --run <run-dir>/run.json --domains-dir examples/pump-station
```

`score` reads whichever label file the folder holds and replays the artifact, so an edited ledger cannot change a grade. With several labels it returns one scorecard per label, each judged on its own entity, plus `unexpected` (detections on entities with no label, actions and effector calls on entities neither labelled nor named by a label's `expected_actions`, all false). An effector call counts only if it names the fault (`work_order` is the exact fault id). Details: [artifacts](../reference/artifacts.md) and [truth and scoring](../architecture/truth-and-scoring.md).

## Comparing arms

Because a campaign is reproducible and its phases are labelled, the same spec can be run against two configurations of a consumer, for example with learning on and off, and the scorecards compared per phase (a learn phase, a holdout phase, a disturb phase on a changed world). The simulator knows nothing about the consumer; the comparison is yours. Compare runs with `streamsim report compare --a <run-dir> --b <run-dir>`.

## Limits

- The runner for the whole campaign is your harness; the simulator generates plans and scores single runs.
- Plans carry the fault: keep them away from the consumer.
- A policy digest that differs from the host's refuses every command; see [physical devices](physical-devices.md).

## Next reads

- [Benchmark workflow](benchmark-workflow.md)
- [Physical devices guide](physical-devices.md)
- [CLI reference: `suite campaign`](../reference/cli.md#suite-campaign), [`score`](../reference/cli.md#score)
- [The canvas](canvas.md)
