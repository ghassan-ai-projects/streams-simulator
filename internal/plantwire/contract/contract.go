// Package contract embeds the plant wire record schemas, a byte-identical copy of contracts/plant-wire/schemas.
package contract

import "embed"

//go:embed schemas/*.json
var Schemas embed.FS
