# Ubiquitous language — perturb

The delivery perturbation layer sits between the world and the adapter. The
world produces what physically happened; this module produces what the
observer got.

| Term | Meaning | Code name | Wire / ledger name |
| --- | --- | --- | --- |
| Perturbation | One named corruption of delivery from the fixed catalog | `Names`, the 19 name constants | `applied_perturbations[]` in the run artifact |
| Applied perturbation | A perturbation switched on with parameters and a time window | `applied` (domain layer), id `<name>-<n>` | `perturb_id` |
| Layer | The aggregate that holds the applied perturbations and their buffers for one world | `Layer` | – |
| Delivered record | One event after perturbation plus its delivery metadata | `Delivered` | ledger record |
| Delivery reason | Why a record looks as it does: ok, dropped, duplicated, delayed, mangled, rewritten, reordered | `model.Delivery*` | `delivery_reason` |
| Window | The `[from, until)` world-time interval in which a perturbation acts | `FromNS`, `UntilNS` | `from_ns`, `until_ns` |
| Flush | Release of records a windowed perturbation held back | `Layer.Flush` | – |
| Parameter contract | The closed set of options each perturbation accepts | `paramRules` | `params` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `Active` (exported type) | `applied` | never returned to a caller; the facade exposes ids only |
| switch ladder (`applyMultiplicity`, `applyDelivery`, …) | `transforms` table | one name, one function |
