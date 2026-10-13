// Package components embeds the built-in component library (vocabulary,
// equipment, instruments and example compositions) so the binary carries it.
package components

import (
	"embed"
	"io/fs"
)

//go:embed vocabulary.json equipment/*.component.json instruments/*.instrument.json compositions/*.composition.json
var library embed.FS

func FS() fs.FS { return library }
