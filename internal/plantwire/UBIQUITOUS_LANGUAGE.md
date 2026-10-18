# plantwire — ubiquitous language

| Term | Meaning |
|---|---|
| Node | The simulated world speaking the plant wire to a host |
| Host | The plant supervisor the node dials |
| Session | One connection: its own `session_id`, both sequences from 0 |
| Birth | `NBIRTH` (kinds once) and `DBIRTH` (one device, its board, boot and current values) |
| Slice | One step of world time; the node sends its data, then `NSYNC`, and waits for `HSYNC` |
| Barrier | The wait for `HSYNC`, capped in wall time; three missed barriers are host loss |
| Lane | Transmit priority (`control` before `data`); `seq` is assigned at transmission |
| Rebirth | `NBIRTH` with the next `bd_seq` and every `DBIRTH`, in the same session |
| Epoch floor | The highest `authority_epoch` the host announced |
| Kind file | One `NBIRTH.kinds` entry without its `kind_digest`, in `*.kind.json`; the interface a device implements (`KindFiles`, `KindSource`) |
| Regular kind file | A kind file is read only as a regular file of at most 1 MiB (`MaxKindFileBytes`), one JSON document, with no member named twice (`ErrDuplicateMember`) and a canonical form (`ErrNotCanonical`: no number beyond the range of a double) |
| Implements | A domain kind names the kind file it implements; `CheckImplements` requires the kind's metrics and commands, safe-stop flags included, as `KindTemplates` gives them, to equal the kind file after canonicalisation, and names the first differing member. It reads kind files only for a domain that implements one, and then only the files of the kinds it implements (`ImplementedKinds`, `SelectKindSources`): a file is known by the kind it declares, or by its file name when it cannot be read or parsed, so a malformed file of another kind is never read |
| Implements checked | What `CheckImplements` returns for a domain that passed: one `{kind, kind_digest}` per implementing kind, an empty list when no kind implements (`model.ImplementsChecked`) |
| Scaffold | `Scaffold` reads kind files, hands them to a scaffold function the caller injects (the plant wire module may not import the component module), and checks that the domain it built implements them: the one entry for a twin |
| Twin rules | A function the caller injects (the plant wire module may not import the component module) that lists everything a twin needs a kind file to state; `kind validate --twin` applies it to a file or to one document on standard input and reports every issue at once (`TwinRules`, `ErrTwinRequirements`) |
| Kind stream | One kind document read from standard input under the limit and strictness of a kind file, named `<stdin>` (`ValidateKindStream`) |
| Capture | The draft kind files of one `NBIRTH` for a person to review: per announced kind the entry without `kind_digest`, the digest announced, the issues the draft still has as a kind file and, against kind directories, a comparison with the file of the same name (`new`, `same`, `differs` at the first differing pointer); `{environment, bd_seq, kinds}` (`CaptureFile`, `CaptureStream`, `ParseCapture`). It writes nothing |
| Birth is a claim | A birth is captured only when it validates against the `NBIRTH` schema, speaks this contract (`ProtocolInvalid`, `ContractUnsupported`) and every announced `kind_digest` is recomputed and holds (`KindMisdigested`); no kind is announced twice; the sealed director stream is never taken (`ErrBirthRefused`, `ErrSealedStream`, `ErrNoBirth`) |
| Capture source | One `NBIRTH` record, or the first one of a JSON Lines stream: bare, or the `record` of a journal `wire` record with `direction` `node` (`FindBirth`); a file up to 32 MiB (`MaxCaptureFileBytes`), standard input up to 1 MiB |
| Stand-in | A scripted plant wire node for tests (`internal/plantwire/internal/standin`): a hardware board behind a gateway that births kind files, plays readings, link loss and rebirth in script time and answers commands from a script. Test support: imported by tests only, no I/O, no clock, never part of the binary |
| Kind files for | `WithKindFiles` reads kind files and hands them to a caller-injected function, so a surface can use them without the plant wire module knowing the use (`Scaffold` is built on it) |
| Safe stop | A command the domain flags `safe_stop: true` on its effector; the node announces it as `safe_stop` in `NBIRTH.kinds` and runs it without freshness, epoch or policy checks. There is no node option that names safe stops: the announced flag and the executed behaviour are the domain's, and `implements` holds them equal to the kind file |
