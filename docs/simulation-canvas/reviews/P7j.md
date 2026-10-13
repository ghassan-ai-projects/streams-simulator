# P7j review record — component module, compose, lint, catalog tools

- **Reviewed**: `55a4165..daad36e` on `component-module` (2026-10-13)
- **Reviewer**: one independent Sonnet subagent, with lab libraries, 19 compose runs, 300 random mutations and `-race` runs.
- **Outcome**: 1 High, 5 Medium, 7 Low, plus three red gates. All fixed in `fix(component,lint): P7j review findings` except as noted.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | High | The `streamsim/` namespace was reserved by directory position: a customer directory given first could pose as the built-in library, and the built-in vanished when a flag was given | Owner decision: the built-in library is embedded in the binary (`components/` package) and always loads first; flags and `STREAMSIM_COMPONENTS_DIRS` add customer directories; `streamsim/` is refused outside the embedded library |
| 2 | Med | Conservation detector `inputs`/`outputs` were not checked or qualified | Checked at load, required on removal, qualified in compose |
| 3 | Med | Link types keyed by (port type, kinds) cross-wired `follow` for two consuming ports of one type | Link types are `<to kind>__<to port>__<from kind>` (COMPONENTS §5, §7) |
| 4 | Med | Library text nested in `args_schema` reached the operator | Annotation keywords stripped recursively in copied schemas |
| 5 | Med | `pinned_at_bound` fired on scheduled flags | Only integrated states are judged |
| 6 | Med | `not_settled` cyclic exemption ignored `follow` and flows | Reachability through `follow` and flows |
| 7 | Low | Two parameters on one target silently last-wins | Load error |
| 8 | Low | A state clamped flat at a bound was not reported | Integrated states at a bound reported regardless of span |
| 9 | Low | Connection endpoints beyond the id-template count were accepted | Refused (count, rendered form) |
| 10 | Low | Span scaling used the intersected range | Converted point span, documented |
| 11 | Low | File paths exposed through the MCP library tools | Dropped from listings |
| 12 | Low | Stale release manifest | Refreshed if `make manifest` runs without keys, else at release |
| 13 | Low | Cosmetics: `utc_offset_s` without a period, `not_applicable` length, empty env entries | Validated / aligned / ignored; repeated error prefixes left |

Sound per the reviewer: preset and parameter order, keyed targets, unit conversion and range intersection, all seven instrument fault
expansions, qualification, determinism of compose and lint, digests, the catalog install lock (leaf lock, `-race` clean), the
operator role registering no component tools, kernel-only file reads, repository rules and coverage.
