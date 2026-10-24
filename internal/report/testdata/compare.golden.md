# Run comparison: r-3h6wx9adsnlft and r-3h6wx9adsnlft

Run A is the reference. This document lists what differs; it does not rank the runs.

## Inputs

| Field | A | B |
|---|---|---|
| Run ID | `r-3h6wx9adsnlft` | `r-3h6wx9adsnlft` |
| Mode | evaluate | evaluate |
| Seed | 7 | 7 |
| Simulator version | 0.3.0 | 0.3.0 |
| Domain | aquaculture-pond 0.1.0 `sha256:dfb37e19c2c3137194d23652a5531a07605fb909bce59e2f63724191545adbe1` | aquaculture-pond 0.1.0 `sha256:dfb37e19c2c3137194d23652a5531a07605fb909bce59e2f63724191545adbe1` |
| Adapter | native-jsonl 0.1.0 `sha256:6c01da6f343fa8e671962f14df015450fdf2327690147413310c8b0242241804` | native-jsonl 0.1.0 `sha256:6c01da6f343fa8e671962f14df015450fdf2327690147413310c8b0242241804` |
| Time mode | stepped | stepped |
| Start | 2026-01-01T00:00:00Z | 2026-01-01T00:00:00Z |
| End | 2026-01-01T00:02:00Z (+120s) | 2026-01-01T00:02:00Z (+120s) |
| Simulated length | 120s | 120s |
| Complete | yes | yes |
| Verification | files\_match | files\_differ |

## Comparability limits

- `files_unverified`: Run B: its files are not shown to match the digests the run recorded (files\_differ).
- `kpi_truth_absent`: Run A: it records no whole-run KPI results, so it contributes no result rows.
- `kpi_truth_absent`: Run B: it records no whole-run KPI results, so it contributes no result rows.
- `single_run_per_side`: Each side is one run. A difference between two single runs does not establish that a change caused it, and a seed is one draw of the world's randomness.

## Plan differences

### Faults

Identical: 0. Different: 1. Only in A: 0. Only in B: 0.

#### Different

| Run | Entity | Fault | Onset | Cleared | Differs in |
|---|---|---|---|---|---|
| A | site-a/pond-1 | aerator\_failure | 2026-01-01T00:00:30Z (+30s) |  | cleared\_by, cleared\_ns |
| B | site-a/pond-1 | aerator\_failure | 2026-01-01T00:00:30Z (+30s) | 2026-01-01T00:01:30Z (+90s) |  |

### Perturbations

Identical: 2. Different: 0. Only in A: 0. Only in B: 0.

### Environment

Identical: 0. Different: 0. Only in A: 0. Only in B: 0.

### Effector calls

Identical: 1. Different: 0. Only in A: 1. Only in B: 0.

#### Only in A

| At | Entity | Effector | Args | Outcome |
|---|---|---|---|---|
| 2026-01-01T00:01:40Z (+100s) | site-a/pond-2 | start\_aerator | `{"level":0.5,"pond_id":"site-a/pond-2"}` | mode=silent\_no\_effect accepted=yes effect\_applied=yes ack\_latency\_ms=491 |

## Results

Scope: whole run of each run. Simulated length covered by KPI truth: A none, B none.

### KPIs

None.

### Alarms

None.

### Alarm system

None.

## Windowed results

not_available: The run report carries whole-run KPI truth only (mean, minimum, maximum and samples over the whole run), so no event-relative window can be computed from it and none is shown.
