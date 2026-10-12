# P5b review record — overrides, links, system entities, board

- **Reviewed**: the P5b working-tree diff against `751abbe` (2026-10-12)
- **Reviewer**: one independent Sonnet subagent (correctness of overrides and structure), per the owner's direction to
  keep reviews light.
- **Outcome**: 1 High, 2 Medium, 5 Low. Resolved below, accepted, or carried.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| F1 | High | `isConfirmationChannel` scanned only the kinds' own effectors, so the `silent_no_effect` check (non-negotiable 8) ignored the confirmation channels of an effector added by an override | Fixed: the check reads the entity's effective kind; `TestAnAddedEffectorsConfirmationChannelReportsTheCounterfactualUnderSilentNoEffect` (an added effector whose only confirmation channel is an added channel, `silent_no_effect` forced, the channel reports the counterfactual 5 below the state) |
| F2 | Med | A name was refused only when it fit its own set's `id_template`, so it could fit another set's template and then collide with a rendered or churned id | Fixed: a name fitting any set's template is refused at `/instances/i/names/j`; spec section 8 and the contract description updated; a case in `TestInstanceSetsOfNamesAreCheckedAtTheirPointers` |
| F3 | Med | The audit took its channel list from `KindChannelNames`, which unions the channels of every override of the kind | Fixed: the channels come from the injected entity's effective kind; shipped domains unchanged (behaviour pin matches); `TestTheAuditChannelsComeFromTheInjectedEntitysEffectiveKind` |
| F4 | Low | The scenario pre-check validates effector arguments against the kind's bound while the dry run refuses a value outside an entity's narrowed bound | Accepted and documented (spec section 9, Limits; evidence Known limits) |
| F5 | Low | Peer pooling is by kind: the sigma of a `peer_residual` detector on an added channel pools peers that lack the channel | Carried to P6a as R5b-01 |
| F6 | Low | Suites sample faults of the base kind: a fault added by an override is only drawn for entities of its set | Accepted and documented (spec section 9, Limits; the suite test asserts every injected fault is carried by its entity) |
| F7 | Low | Narrowing validation gaps: duplicate narrow entries for one name are not refused; an argument `minimum` above the declared `maximum` is not refused | Carried to P6a as R5b-02 |
| F8 | Low | A profile's `entity_count` is ignored when the first instance set has names only | Carried to P6a as R5b-03 |
