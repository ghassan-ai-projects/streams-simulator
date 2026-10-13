# P7i evidence — world forms for plants

P7i lands in steps, one commit each. Domain-spec v0.2 §22 onwards holds the rules.

## Step 1 — prerequisites (R7g-01, R7e-04, R6a-01)

| Check | Result |
|---|---|
| An absent `noise_scale_at_full` keeps the channel noise; an explicit one scales it | `TestABiasWithoutANoiseScaleKeepsTheChannelNoise` |
| A state bound clamps only the side it names | `TestAStateBoundClampsOnlyTheSideItNames` |
| An assignment approaches its target from the value it replaces (natural, after an additive kick, after an earlier assignment, with a later kick, before it starts) | `TestAnAssignmentApproachesItsTargetFromTheValueItReplaces` |
| Lags, RC networks, regimes and dead times step within their stability bound; stores within their exchange bound, scale bounds included | `TestAnIntegratedStateMustStepWithinItsStabilityBound`, `TestEachIntegratedFormHasItsStabilityLimit` |
| A negative flow scale stops the flow | `TestANegativeScaleStateStopsAFlowInsteadOfReversingIt` |
| Id templates rendering a common id are refused (exact, both orders) | `TestTwoIDTemplatesOverlapWhenOneIDFitsBoth`, `TestAnInstanceSetMayNotRenderAnotherSetsIDs` |
| A lease argument cannot be negative | `TestALeaseArgumentMustNotAllowANegativeLease` |
| Shipped worlds unchanged | Behaviour pin: run, trace and replay hashes identical; only `catalog describe` output changed, because explicit zero bounds now survive the model (re-recorded with the v0.1 runtime equivalence digests) |

## Step 2 — operator truth boundary (R7e-09)

| Check | Result |
|---|---|
| The operator's effector view carries no description and no schema annotation at any depth; validation keywords stay; the director's schema is untouched | `TestTheOperatorSeesNoEffectorDescriptionOrSchemaAnnotation` |
| Shipped runs unchanged | Behaviour pin identical |

## Step 3 — algebraic forms `curve`, `product`, `sum` (domain-spec §23)

| Check | Result |
|---|---|
| Sum, clamped product, 1-D and 2-D curves and a curve of a sum equal their formulas of the inputs at the read time, between steps included | `TestAlgebraicFormsAreExactFunctionsOfTheirInputsAtTheReadTime` (1e-12) |
| A curve holds its end values and hits its points | `TestACurveClampsAtItsEndsAndHitsItsPoints` |
| A fault overlays an algebraic state | `TestAFaultOnAnAlgebraicStateOverlaysItsValue` |
| Table shape, axis order, arity and algebraic loops are validated; no `dt_ms` is needed | `TestCurveTablesAndProductsAreValidatedByShape` (9 cases) |

## Step 4 — `pid` controller (domain-spec §24)

| Check | Result |
|---|---|
| Closed loops with a first-order plant agree with an independent recurrence: PID with derivative, anti-windup with a setpoint drop, manual and back, direct action | `TestAClosedLoopAgreesWithAnIndependentRecurrence` (1e-9 over 1200 s) |
| The output leaves saturation within a step of the error reversing | `TestAControllerLeavesSaturationAsSoonAsItsErrorReverses` |
| Mutations caught | Without anti-windup both tests fail; without bumpless manual transfer the manual case fails |
| Reads, gains, manual pairing and output limits are validated | `TestAPIDControllerIsValidatedByItsReadsGainsAndLimits` (8 cases) |
