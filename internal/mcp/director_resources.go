package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (d *Director) readCatalogResource(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	entries := d.Catalog.List("")
	data, _ := json.Marshal(entries)
	return jsonResource("sim://catalog", data), nil
}

func (d *Director) readDomainResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	id, err := templateParam(req, "id")
	if err != nil {
		return nil, fmt.Errorf("mcp: %w", err)
	}
	compiled, err := d.Catalog.Describe(id)
	if err != nil {
		return nil, errTool(CodeDomainInvalid, "%v", err)
	}
	data, _ := json.Marshal(compiled.Spec)
	return jsonResource(req.Params.URI, data), nil
}

func jsonResource(uri string, data []byte) *mcp.ReadResourceResult {
	return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: uri, MIMEType: "application/json", Text: string(data)}}}
}
