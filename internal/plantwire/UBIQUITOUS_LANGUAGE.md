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
| Implements | A domain kind names the kind file it implements; `CheckImplements` requires the kind's metrics and commands, as `KindTemplates` gives them, to equal the kind file after canonicalisation, and names the first differing member |
