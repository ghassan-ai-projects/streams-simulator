# P7b review record — plant wire node sessions

- **Reviewed**: the P7b working-tree changes against `af8ec2d` (2026-10-12)
- **Reviewer**: one independent Sonnet subagent, with probes running sessions over all eight shipped domains against a
  schema-validating host.
- **Outcome**: 4 High, 8 Medium, 6 Low. All High and Medium fixed; Low fixed or recorded.

| # | Sev | Finding | Disposition |
|---|---|---|---|
| 1 | High | Five shipped domains use `/` in entity ids, which the wire id pattern refuses | Fixed: a deterministic wire id per entity (`/` → `:`, other unsafe characters → `_`, distinct on collision) with a reverse map for host requests; `TestEveryShippedDomainSpeaksValidWire` runs a session over every shipped domain against a validating host |
| 2 | High | Channel descriptions (author notes) carried fault ids and hidden mechanisms to the wire | Fixed: no domain description is sent; argument-schema descriptions are stripped at every depth (SPEC §5) |
| 3 | High | The truth canary used one clean domain and a formatted onset the wire never carries | Fixed: the canary runs over every shipped domain with an off-grid onset in nanoseconds, every fault id and the hidden failure modes; `TestAFaultChangesNoBirth` proves births are identical with and without a fault |
| 4 | High | A schema-invalid host record lost its `seq` and the next record ended the session as a gap | Fixed: an invalid record still consumes its `seq`; an invalid `DCMD` gets a `malformed` receipt (SPEC §7); `TestAMalformedCommandIsAnsweredAndTheSessionGoesOn` |
| 5 | Med | Barrier acknowledgement rule ambiguous | Fixed: SPEC §8 states `HSYNC.through_seq` ≥ the `NSYNC`'s own `seq` |
| 6 | Med | `Send` had no write deadline or context | Fixed: a write deadline of the barrier cap per send; a cancelled context closes the socket |
| 7 | Med | The reader goroutine could block forever after `Close` | Fixed: the reader stops when the link closes |
| 8 | Med | A goodbye followed by a hang-up was an error; other hang-ups were not host loss | Fixed: records are processed before the error; a hang-up is `ErrHostLost`; test covers both |
| 9 | Med | Options unvalidated (zero slice loops forever) | Fixed: options are validated (slice, end, version, environment, session id) |
| 10 | Med | A query or command for an unknown device invented state | Fixed: a query is ignored; a command is refused `unknown_device` |
| 11 | Med | Handshake edge cases (first host `seq`, host version range, lower-epoch goodbye) | Fixed: the first host record must be `seq` 0; the version must be inside both ranges; a lower-epoch `STATE` is ignored before its `online` flag |
| 12 | Med | Wire data depended on the run's adapter | Fixed: `run.SetDeliveryObserver`, a post-perturbation, pre-adapter hook |
| 13 | Low | Gap and duplicate handling simplified for an ordered stream | Documented in SPEC §7 |
| 14 | Low | Socket check followed symlinks and ignored type and owner | Fixed: `Lstat`, socket and directory types, owner uid |
| 15 | Low | Non-finite values and value-less channels | Fixed: non-finite samples and channels without a data type are not sent (SPEC §12) |
| 16 | Low | Implicit boards missing from `NBIRTH`; `bd_seq` and `prior_last_seq` not settable | Fixed: unboarded devices are listed as their own boards; both are options |
| 17 | Low | `app` imports the `uds` edge for `Serve` | Kept: the facade may only delegate, so the composition lives in `app` (the architecture gate allows it) |
| 18 | Low | A birth value could be `null` for any quality | Fixed: `null` only with `warming_up` (schema), `DSTATE` values never `null` |
