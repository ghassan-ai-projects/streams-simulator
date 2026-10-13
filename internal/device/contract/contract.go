// Package contract embeds the vendored Agentic Stream device wire-protocol
// schemas (see SOURCE.md). The conformance fixtures beside them are test
// data for the device module.
package contract

import "embed"

//go:embed schemas/*.json
var Schemas embed.FS
