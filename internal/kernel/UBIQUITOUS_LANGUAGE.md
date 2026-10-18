# Ubiquitous language — kernel

The *kernel* is the one place the simulator touches the operating system. Module edges call it; domain and app layers never
do. It holds mechanisms only: no record, file name or format belongs to it.

| Term | Meaning | Code name |
| --- | --- | --- |
| Regular read | A read that refuses anything but a regular file (a symbolic link, a named pipe and a directory are refused without being opened) and a file larger than the caller's limit | `ReadRegular`, `ErrNotRegular`, `ErrTooLarge` |
| Walked read | `OpenRegular` opens a regular file for streaming: a relative path is walked component by component and no component may be a symbolic link (`ErrSymlink`, `OpenRegularIn` with its base); an absolute path is checked at its last component only, because a system's own links above it cannot be told from a planted one | `OpenRegular`, `ErrSymlink` |
| Atomic write | A file replaced in one step: written to a temporary beside it, synced, renamed, the directory synced; owner-only | `WriteAtomic` |
| Private directory | A directory created owner-only (0700) | `PrivateDir` |
| Log | An owner-only append-only file behind a buffer: flushed at command boundaries, synced and closed at the end | `Log`, `OpenLog` |
| Owner-only socket | A unix socket published 0600 by rename, so it is never reachable before it is private; a stale one is replaced, a live one or a regular file is refused | `Listen`, `ListenPrivate` |
| Private dial | A connection made only to an owner-only socket in an owner-only directory | `DialPrivate` |
| Interrupt | A context that ends on interrupt or SIGTERM | `UntilInterrupted` |
