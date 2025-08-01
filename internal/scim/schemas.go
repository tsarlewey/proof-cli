package scim

import (
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

// AuthenticationScheme represents an authentication scheme
type AuthenticationScheme struct {
	Type             string `json:"type"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	DocumentationURI string `json:"documentationUri"`
	Primary          bool   `json:"primary"`
}

// ServiceProviderConfigFeature represents a feature configuration
type ServiceProviderConfigFeature struct {
	Supported bool `json:"supported"`
}

// ServiceProviderConfigMeta represents metadata for service provider config
type ServiceProviderConfigMeta struct {
	Location     string `json:"location"`
	ResourceType string `json:"resourceType"`
}

// ServiceProviderConfig represents the service provider configuration
type ServiceProviderConfig struct {
	Schemas               []string                     `json:"schemas"`
	Name                  string                       `json:"name"`
	DocumentationURI      string                       `json:"documentationUri"`
	AuthenticationSchemes []AuthenticationScheme       `json:"authenticationSchemes"`
	Patch                 ServiceProviderConfigFeature `json:"patch"`
	Bulk                  ServiceProviderConfigFeature `json:"bulk"`
	Filter                ServiceProviderConfigFeature `json:"filter"`
	ChangePassword        ServiceProviderConfigFeature `json:"changePassword"`
	Sort                  ServiceProviderConfigFeature `json:"sort"`
	Etag                  ServiceProviderConfigFeature `json:"etag"`
	Meta                  ServiceProviderConfigMeta    `json:"meta"`
}

// ResourceTypeMeta represents metadata for a resource type
type ResourceTypeMeta struct {
	Location     string `json:"location"`
	ResourceType string `json:"resourceType"`
}

// ResourceType represents a SCIM resource type
type ResourceType struct {
	Description string           `json:"description"`
	Endpoint    string           `json:"endpoint"`
	ID          string           `json:"id"`
	Meta        ResourceTypeMeta `json:"meta"`
	Name        string           `json:"name"`
	Schema      string           `json:"schema"`
	Schemas     []string         `json:"schemas"`
}

// ResourceTypesResponse represents the response from listing resource types
type ResourceTypesResponse struct {
	ItemsPerPage int            `json:"itemsPerPage"`
	Resources    []ResourceType `json:"Resources"`
	Schemas      []string       `json:"schemas"`
	StartIndex   int            `json:"startIndex"`
	TotalResults int            `json:"totalResults"`
}

// GetUserSchema retrieves the user schema for an organization
func GetUserSchema(client *utils.ProofClient, organizationID string) ([]byte, error) {
	path := SCIMEndpoint + "/" + organizationID + "/Schemas/Users"
	return client.Get(path, scimOptions)
}

// GetServiceProviderConfig retrieves the service provider configuration
func GetServiceProviderConfig(client *utils.ProofClient, organizationID string) ([]byte, error) {
	path := SCIMEndpoint + "/" + organizationID + "/ServiceProviderConfig"
	return client.Get(path, scimOptions)
}

// GetResourceTypes retrieves the supported resource types
func GetResourceTypes(client *utils.ProofClient, organizationID string) ([]byte, error) {
	path := SCIMEndpoint + "/" + organizationID + "/ResourceTypes"
	return client.Get(path, scimOptions)
}
