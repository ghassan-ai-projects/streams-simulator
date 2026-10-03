# Go Style Context

## Core Rules

- Use `context.Context` as the first parameter for I/O, blocking, or cancellable work.
- Return `error` last.
- Wrap errors with `%w` and enough local context to debug failures.
- Use `log/slog` instead of `fmt.Println` for logging.
- Document every exported symbol.
- Prefer standard library packages like `cmp`, `maps`, and `slices` over new helper dependencies.
- Keep packages small, lowercase, and singular.
- Prefer direct code over indirection when both are maintainable.
- Add new abstractions only after a real second use, repeated logic, or a clear testing need appears.

## Layering Rules

- Keep flags and application wiring in `cli`, and MCP protocol handling in `mcp`.
- Keep simulation, delivery transforms, projections and scoring in their business packages: `world`, `perturb`, `adapter`, `truth` and `score`.
- Keep command logs, delivery ledgers, artifacts, replay and quiescence in `run`.
- Keep shared records in `model`, data loading in `domain`, and schema validation in `jsonschema`.
- Keep the device transport in `device`; only `deviceworld` bridges it to the world.
- Follow the direct dependency graph enforced by `test/architecture`. Follow the file-size and function-review rules in root `AGENTS.md`.
- Define interfaces in the consumer package when possible.

## Naming Rules

- Use `ID`, `URL`, `HTTP`, `JSON`, `API`, `YAML`, `MCP`, `SQL` consistently.
- Prefer `ErrXxx` for sentinel errors and `XxxError` for error types.
- Do not mix acronym styles such as `Http` and `HTTP`.

## Avoid

- persistence abstractions beyond the current file-based JSON/JSONL artifact model without an accepted design change
- `init()` outside configuration/bootstrap cases
- global mutable state
- "future-proof" interfaces or wrapper layers with no current consumer need
- helpers that save only one or two lines while hiding behavior
- placeholder TODO logic shipped as if complete
- unrelated refactors in the same change
