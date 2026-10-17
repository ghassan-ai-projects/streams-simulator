# P5 graph CSS partition — validation pending

This source-only unit partitions graph node styling by responsibility while preserving the React Flow zoom-tier contract.

- Split `canvas/src/ui/graph/internal/nodes.module.css` into node, pin, frame, readout, and mark CSS modules. The resulting files are 224, 161, 105, 194, and 130 lines respectively; each remains below the default 300-line cap.
- Updated graph component class bindings and combined the five module-local tier classes on the graph root. `data-noted` preserves the prior fault-only `.noted` condition; node-root and hall markers scope the selectors that cross the stylesheet boundary.
- Removed 30 local `!important` declarations from pin/handle rules using higher-specificity graph-owned selectors. The two existing minimap-position declarations remain in `vendor.module.css`.
- Moved minimap node colors from global `families.css` rules into `GraphControls.module.css`; the minimap callback supplies the generated local classes.
- Lowered the nodes stylesheet cap from 695 to 300 lines and its important allowance from 30 to zero. The emitted important count is updated from 32 to 2 to reflect the 30 removed graph declarations and the two retained vendor declarations. The registered z-index and positioned-declaration values are unchanged.

Validation is pending by the owner’s production-first instruction. No Canvas tests, typecheck, architecture check, build, browser review, or emitted-CSS verification has been run for this unit. Prettier was invoked on the changed Canvas files and register.
