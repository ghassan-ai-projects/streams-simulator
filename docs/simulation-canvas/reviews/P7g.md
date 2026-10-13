# P7g review record — component library contracts

- **Reviewed**: the P7g working tree against `e2b73b9` (2026-10-13)
- **Reviewer**: one independent Sonnet subagent.
- **Outcome**: 2 High, 10 Medium, 6 Low. Fixed in the contracts, or registered.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | High | An instrument fault would silence the channel's noise: the engine reads an absent `noise_scale_at_full` as 0, the schema documents 1 | Compose writes `noise_scale_at_full: 1` (COMPONENTS §3); `TestComposedInstrumentFaultsKeepTheChannelNoise`; the engine default is R7g-01 (P7i) |
| 2 | High | Instrument fault expansion was not derivable; `noise` could not be produced | §3 expansion table (bias state, ids, detector and SNR, offset and drift with `duration_s`); `noise` removed until P7i |
| 3 | Med | Link type names could collide; links lose their ports | `link_type` is `<type>__<from kind>__<to kind>`; ports on links come with P7h (§5) |
| 4 | Med | A consuming port and "one entry is one kind" | §5: a linked instance reads the aggregate in place of its own dynamics, an unlinked one keeps them (P7h) |
| 5 | Med | Unit conversion undefined | §3: gain and offset from the vocabulary, ranges converted, magnitudes in the instrument's unit; tags and families must match |
| 6 | Med | Instrument channel members untyped | Validated by the domain validator after compose; an instrument may name no state (§3); typed `$ref`s when the compiler supports cross-schema references |
| 7 | Med | Component kinds could not hold `flows` or `system` | Allowed |
| 8 | Med | Array positions in targets, name lengths, enum parameters, default preset, load checks | Keyed paths (§2), length limits, `cases` for enums, `default_preset` required with presets, a load-check list |
| 9 | Med | Provenance and metadata gaps | Defaults recorded, instrument digests, `site` with areas and channel `tag` in the domain (domain-spec §20); `area` must be declared; counts ≥ 1; `names` with `id_template` allowed |
| 10 | Med | Chiller physics | `running` bounded; heat rejection = load + power; undercharge lowers power; supply near setpoint; steady-state initials per preset; stop removed until products (R7g-02, P7i) |
| 11 | Med | Absolute magnitudes ignore size | Instrument `scaling: span`; presets scale fault deltas |
| 12 | Med | Library descriptions could leak | §9: never copied into a domain; composed effectors carry no description |
| 13 | Low | Extensibility | Vocabulary merges entry by entry; open class; library layout; (id, version) key; customer `streamsim/` refused |
| 14 | Low | Realism details | RTD drift 0.05 °C/h over a day; span-relative meter noise; AHU fault reduces the load it passes on |
| 15 | Low | Stale design text | component-library.md §3 now points to the contracts |
| 16 | Low | Weak tests | Vocabulary consistency, point and instrument matching, provenance digests, noise kept under a fault |
