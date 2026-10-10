# Ubiquitous language — world

The world is what physically happened: hidden state evolving under declared
dynamics, faults and effectors. What an observer later receives is the
`perturb` module's business, not this one's.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| World | One seeded, single-goroutine simulated world built from a compiled domain | `World` | `world_id` |
| Entity | One simulated producer (a pond, a pump) with its own hidden state | `Entity` | `entity_id`, `entity_type` |
| Hidden state | A true physical quantity the consumer never sees directly | `StateValue` | state names in the domain spec |
| Reading | The noise-free value of a channel (the observed form is a native event) | `Reading` | – |
| Native event | One emission of a channel, before any perturbation | `model.SimEvent` | `sim-event-v0.1` |
| Fault | A declared physical fault injected on an entity at an onset time | `InjectFault`, `FaultInfo` | `fault_id` |
| Effector | A declared, risk-classed actuation a consumer may invoke | `InvokeEffector` | `effector` |
| Command id | The idempotency key of one effector invocation | `commandID` | `command_id` |
| Effector call | The director-side record of one invocation, the scoring authority | `EffectorCall` | `effector_calls` |
| Failure mode | How an invocation went: ok, slow, ack_lost, reject, partial, confirmed_no_effect, silent_no_effect | `Mode*` constants | `mode` |
| Interlock | An independent safety predicate that refuses an invocation terminally | `ErrInterlockRefused` | `interlock_refused` |
| Kick | A scheduled state delta an applied effect produces | `PendingKicks` | – |
| Substream | A named seeded random stream; names fix the draw order | `substream` | `<world>/<entity>/<channel>/<purpose>` |
| Churn | Autonomous birth and retirement of entities on a schedule | `birthAutonomous`, `Retire` | – |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `PendingEvents`, `HiddenStateSnapshot`, `ErrEffectorRefused` | removed | no caller anywhere |
| `World.Seed`, `ClockNS`, `Noiseless`, `EmitDisabled` (exported fields) | unexported | never read outside the module |
| `DynamicsFor` | `dynamicsFor` | internal lookup |
| `recordCall` positional parameters | `invocation`, `callOutcome` | one request value, one outcome value |
