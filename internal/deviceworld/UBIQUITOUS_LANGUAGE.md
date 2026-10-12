# Ubiquitous language — deviceworld

The bridge that makes the simulated world the oracle behind the device
emulator's wire loop: a device command energizes an output because the
world's effector actually applied a physical effect.

| Term | Meaning | Code name | Wire / artifact name |
| --- | --- | --- | --- |
| Binding | A validated, data-defined mapping from one device target to a world effector invocation | `Binding` | `*.bindings.json` |
| Binding catalog | The strict JSON document that declares every binding | `LoadBindings` | `thermal.bindings.json` |
| Plant | The device-facing port backed by the world: applies commands, drives safe stops | `Plant` | – |
| Target | A device output a command addresses (for example a fan) | `Target` | `target` |
| Energized | What the world reports as physically applied — the observed truth, independent of the acknowledgement | `PlantEffect` energized flag | `energized` |
| Safe stop | The declared safe state a target is driven to on lease expiry or reboot | `SafeStop` | `safe_stop` |
| Composition | The pairing of a capability catalog and a world; validated before a listener opens | `ValidateBindings`, `ValidateComposition` | – |
| Open | Build the device a serve command describes: capabilities, fault schedule, then, when a binding catalog is given, the world, the entity, the bindings and the plant, in that order; with no catalog the device has no plant | `OpenDevice`, `OpenOptions` | `device serve --world` |
| Device clock | Microseconds of the injected clock since the plant was bound | `deviceClock` | `not_before_mono_us` |

## Retired words

| Word | Replaced by | Why |
| --- | --- | --- |
| `deviceworld.New` returning a bare `*Plant` | `New(...) (*Plant, error)` with `ErrNoWorld` | fail-closed constructor |
| `prepareDeviceWorld`, `bindDeviceWorld` (in `cli`) | `OpenDevice` | composing a device with a world is the bridge's work |
