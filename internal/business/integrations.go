package business

import (
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// IntegrationsEndpoint is the base endpoint for integrations API
	IntegrationsEndpoint = "/v1/integrations"
)

// IntegrationConfiguration represents the configuration settings for an integration
type IntegrationConfiguration struct {
	AccountID   string `json:"account_id,omitempty"`  // Integration Account, maxLength: 5000
	Environment string `json:"environment,omitempty"` // Integration Environment, maxLength: 5000
}

// CreateIntegrationParams represents the request body for creating an integration
type CreateIntegrationParams struct {
	Name           string                    `json:"name,omitempty"`            // Name of Integration (ADOBE, DOCUTECH)
	OrganizationID string                    `json:"organization_id,omitempty"` // Organization to create integration for, maxLength: 5000
	Configuration  *IntegrationConfiguration `json:"configuration,omitempty"`   // Integration configuration settings
}

// CreateIntegration creates a new integration
func CreateIntegration(client *utils.ProofClient, params *CreateIntegrationParams) ([]byte, error) {
	return client.Post(IntegrationsEndpoint, params)
}
