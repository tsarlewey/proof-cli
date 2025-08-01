package business

import (
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// OrganizationEndpoint is the base endpoint for organization API
	OrganizationEndpoint = "/v1/organization"
	// PartnersEndpoint is the endpoint for partners API
	PartnersEndpoint = "/v1/partners"
)

// UpdatePartnerOrganizationParams represents the request body for updating a partner organization
type UpdatePartnerOrganizationParams struct {
	Branding *Branding `json:"branding,omitempty"` // Co-branding related attributes for the partner organization
}

// Branding represents the branding configuration for a partner organization
type Branding struct {
	LogoImage string `json:"logo_image,omitempty"` // Supported formats: valid image URL or base64 encoded image data, maxLength: 15000000
}

// CreatePartnerOrganizationParams represents the request body for creating a partner organization
type CreatePartnerOrganizationParams struct {
	Name     string    `json:"name"`               // Required: Name of the partner organization, maxLength: 5000
	Email    string    `json:"email"`              // Required: A unique email associated with this organization, maxLength: 5000
	Branding *Branding `json:"branding,omitempty"` // Optional: Co-branding related attributes for the partner organization
}

// GetOrganizationInformation retrieves organization information
func GetOrganizationInformation(client *utils.ProofClient) ([]byte, error) {
	return client.Get(OrganizationEndpoint)
}

// CreatePartnerOrganization creates a new partner organization
func CreatePartnerOrganization(client *utils.ProofClient, params *CreatePartnerOrganizationParams) ([]byte, error) {
	return client.Post(PartnersEndpoint, params)
}

// UpdatePartnerOrganization updates an existing partner organization
func UpdatePartnerOrganization(client *utils.ProofClient, id string, params *UpdatePartnerOrganizationParams) ([]byte, error) {
	path := PartnersEndpoint + "/" + id
	return client.Patch(path, params)
}
