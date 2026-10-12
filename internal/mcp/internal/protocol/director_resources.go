package protocol

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (d directorTools) readCatalogResource(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	data, _ := json.Marshal(d.CatalogList("")["domains"])
	return jsonResource("sim://catalog", data), nil
}

func (d directorTools) readDomainResource(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	id, err := templateParam(req, "id")
	if err != nil {
		return nil, fmt.Errorf("mcp: %w", err)
	}
	spec, err := d.DomainSpec(id)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(spec)
	return jsonResource(req.Params.URI, data), nil
}

func jsonResource(uri string, data []byte) *mcp.ReadResourceResult {
	return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: uri, MIMEType: "application/json", Text: string(data)}}}
}
