package business

import (
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// TemplatesEndpoint is the base endpoint for templates API
	TemplatesEndpoint = "/v1/templates"
)

// ListTemplatesParams represents query parameters for listing templates
type ListTemplatesParams struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// GetAllTemplates lists templates with optional pagination
func GetAllTemplates(client *utils.ProofClient, params *ListTemplatesParams) ([]byte, error) {
	// Build query parameters
	queryParams := utils.BuildQueryParams(params)

	path := TemplatesEndpoint
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	return client.Get(path)
}
