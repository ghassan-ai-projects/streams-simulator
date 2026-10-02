package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func boolArg(args map[string]any, key string) bool {
	if args == nil {
		return false
	}
	b, _ := args[key].(bool)
	return b
}

func mapArg(args map[string]any, key string) map[string]any {
	if args == nil {
		return nil
	}
	if m, ok := args[key].(map[string]any); ok {
		return m
	}
	return nil
}

// templateParam extracts a value from a resolved resource URI template. The
// SDK resolves templates before invoking the handler, so the parameter is
// parsed from the concrete URI.
func templateParam(req *mcp.ReadResourceRequest, key string) (string, error) {
	uri := ""
	if req.Params != nil {
		uri = req.Params.URI
	}
	switch key {
	case "id":
		const prefix = "sim://domains/"
		if len(uri) > len(prefix) && uri[:len(prefix)] == prefix {
			id := uri[len(prefix):]
			if end := indexByte(id, '/'); end >= 0 {
				id = id[:end]
			}
			if id != "" {
				return id, nil
			}
		}
		return "", errTool(CodeDomainInvalid, "cannot parse domain id from %q", uri)
	}
	return "", errTool(CodeDomainInvalid, "unknown template parameter %q", key)
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
