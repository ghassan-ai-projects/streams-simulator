# Pump station

Two worlds, their commissioned kind files and the first-slice campaign spec, for rule-promotion runs.

- `pump-station.domain.json`: the rotating-machinery pumps with a repairable maintenance loop. `schedule_maintenance` carries a `work_order`; an hour later the fault it names, if active on that pump, is cleared (so `sensor_drift` and `sensor_failure` are repaired too). `reduce_load` still lowers vibration without repairing a worn bearing.
- `pump-station-probe-b.domain.json`: the same pumps with a replacement vibration probe (another range, resolution and gain), so the `pump` kind is born with another digest.
- `kinds/pump.kind.json` and `kinds/pump-probe-b.kind.json`: the `pump` kind of each world, captured from a session birth (`streamsim kind capture`); each file's digest is the digest its world births the pump with. Both define the kind `pump`, so a host commissions one of them.
- `campaign/first-slice.campaign-spec.json`: 160 one-fault runs (learn, holdout, disturb). Set `host_socket` and `policies` to the host's own, then generate:

```bash
streamsim suite campaign --spec examples/pump-station/campaign/first-slice.campaign-spec.json --domains-dir examples/pump-station --out <dir>
```

Run one session of the campaign with `streamsim plantwire connect --plan <dir>/run-0001.session-plan.json --domains-dir examples/pump-station --out <run-dir> --slice-s 60`.

- The plan's `host_socket` must be an absolute path of at most 103 characters in a directory only you can enter; the committed `/tmp/streamsim-host.sock` is a placeholder. Without `policies`, every command is refused.
- A session never ends by itself: send SIGINT to `connect` when the world time reaches `duration_s`. Closing the host first leaves an incomplete run.
- `connect` writes `<run-dir>/label.json` when the plan holds exactly one disturbance. It names the fault: never give it to the consumer. Write the verdict to `<run-dir>/verdict.json` and run `streamsim score --run <run-dir>/run.json`. A verdict action's `command_id` is the DCMD `idempotency_key` and its `entity_id` is the instance (`site-a/pump-8`).
- `work_order` is the exact fault id: `bearing_wear`, `misalignment`, `sensor_drift`. A work order for another fault repairs nothing and does not score as the expected action.
- A 48 h run takes about 585 MB at `--slice-s 60`.
