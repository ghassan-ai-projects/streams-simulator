# Plant wire contract generators

`schemas.py` writes `contracts/plant-wire/schemas/`; `fixtures.py` writes `contracts/plant-wire/fixtures/` (except its
README) and prints the contract digest for `contracts/plant-wire/DIGEST`. After a change, copy the schemas into
`internal/plantwire/contract/schemas/` and run `go test ./test/contract/ ./internal/plantwire/...`.

```bash
python3 scripts/plant-wire/schemas.py contracts/plant-wire/schemas
python3 scripts/plant-wire/fixtures.py contracts/plant-wire/fixtures contracts/plant-wire/schemas | cut -d' ' -f1 > contracts/plant-wire/DIGEST
```
