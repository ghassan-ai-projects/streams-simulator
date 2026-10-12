package mcp

import "github.com/ghassan-ai-projects/streams-simulator/internal/mcp/internal/protocol"

// ServeOptions say which director to serve: where its domains and adapters
// are loaded from, where its runs publish evidence, and the address that
// binds the operator role over streamable HTTP ("" binds none).
type ServeOptions = protocol.ServeOptions
