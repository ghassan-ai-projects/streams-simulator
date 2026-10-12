# Run report: r-3h6wx9adsnlft

mode: evaluate | complete: yes | ended: yes | verification: files_match

## Run

| Field | Value |
|---|---|
| Run ID | `r-3h6wx9adsnlft` |
| Mode | evaluate |
| Complete | yes |
| Ended | yes |
| Unblinded | no |
| Simulator version | 0.3.0 |
| Domain | aquaculture-pond 0.1.0 `sha256:dfb37e19c2c3137194d23652a5531a07605fb909bce59e2f63724191545adbe1` |
| Adapter | native-jsonl 0.1.0 `sha256:6c01da6f343fa8e671962f14df015450fdf2327690147413310c8b0242241804` |
| Seed | 7 |
| Sink | inproc |
| Time mode | stepped |
| Start | 2026-01-01T00:00:00Z |
| End | 2026-01-01T00:02:00Z (+120s) |
| Entities | 8 |

## Digests

| Field | Value |
|---|---|
| World | `sha256:5c7456ac4fe6b8b6d703efd5b558dbac5cb9b6bc5d59fcc6b3998d23882b61d9` |
| Trace | `sha256:8dc7384ce72c35382466ad0cf5c6092b2287cea91a0d62c6fde4464392ed0014` |
| Observer journal | `sha256:a162e96357eb16362aa4977e7760223a473cdd5d12cad7932e90a0ca50de51d0` |
| Director journal | `sha256:c4db42470bc98dcb32abe365f82a5e12c9b238626d75c002dfe0ec1aae96e4e0` |

## Verification

Status: files_match

| Check | Status | Recorded | Computed |
|---|---|---|---|
| trace | match | `sha256:8dc7384ce72c35382466ad0cf5c6092b2287cea91a0d62c6fde4464392ed0014` | `sha256:8dc7384ce72c35382466ad0cf5c6092b2287cea91a0d62c6fde4464392ed0014` |
| journal | match | `sha256:a162e96357eb16362aa4977e7760223a473cdd5d12cad7932e90a0ca50de51d0` | `sha256:a162e96357eb16362aa4977e7760223a473cdd5d12cad7932e90a0ca50de51d0` |
| sealed_journal | match | `sha256:c4db42470bc98dcb32abe365f82a5e12c9b238626d75c002dfe0ec1aae96e4e0` | `sha256:c4db42470bc98dcb32abe365f82a5e12c9b238626d75c002dfe0ec1aae96e4e0` |

## Deliveries

| Field | Value |
|---|---|
| Emitted | 52 |
| Observed by the consumer | 21 |
| Ledger rows | 53 |
| Delivered | 21 |
| Dropped | 32 |

### Delivered by reason

| Reason | Rows |
|---|---|
| duplicated | 1 |
| ok | 20 |

### Dropped by reason

| Reason | Rows |
|---|---|
| dropped_by_perturbation | 32 |

## Faults

| Seq | At | Action | Entity | Fault | Onset | Params |
|---|---|---|---|---|---|---|
| 0 | 2026-01-01T00:00:00Z (+0s) | inject | site-a/pond-1 | aerator_failure | 2026-01-01T00:00:30Z (+30s) |  |

## Perturbations

| Seq | At | Action | Perturbation | From | Until | Params |
|---|---|---|---|---|---|---|
| 1 | 2026-01-01T00:00:00Z (+0s) | apply | drop | 2026-01-01T00:00:50Z (+50s) | 2026-01-01T00:01:02Z (+62s) | `{"rate":1.0}` |
| 2 | 2026-01-01T00:00:00Z (+0s) | apply | duplicate_burst | 2026-01-01T00:01:02Z (+62s) | 2026-01-01T00:02:00Z (+120s) | `{"rate":0.5}` |

## Environment

None.

## Effector calls

| Seq | At | Command | Effector | Entity | Args | Outcome |
|---|---|---|---|---|---|---|
| 3 | 2026-01-01T00:00:40Z (+40s) | act-1 | start_aerator | site-a/pond-1 | `{"level":1.0,"pond_id":"site-a/pond-1"}` | mode=silent_no_effect accepted=yes effect_applied=yes ack_latency_ms=644 |
| 4 | 2026-01-01T00:01:40Z (+100s) | act-2 | start_aerator | site-a/pond-2 | `{"level":0.5,"pond_id":"site-a/pond-2"}` | mode=silent_no_effect accepted=yes effect_applied=yes ack_latency_ms=491 |

## Consumer

| Field | Value |
|---|---|
| Name | fixture-consumer |
| Version | 1.0.0 |
| Arm | baseline |
| Config digest | `sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef` |
