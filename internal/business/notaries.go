package business

import (
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// NotariesEndpoint is the base endpoint for notaries API
	NotariesEndpoint = "/v1/notaries"
)

// CreateNotaryParams represents the request body for creating a notary
type CreateNotaryParams struct {
	Email       string `json:"email"`                 // Required: Unique email address for the new notary account, maxLength: 5000
	USStateAbbr string `json:"us_state_abbr"`         // Required: Two-letter state abbreviation where notary is commissioned, maxLength: 2
	FirstName   string `json:"first_name"`            // Required: First name of the notary (must match commission), maxLength: 5000
	MiddleName  string `json:"middle_name,omitempty"` // Optional: Middle name of the notary (must match commission), maxLength: 5000
	LastName    string `json:"last_name"`             // Required: Last name of the notary (must match commission), maxLength: 5000
}

// UpdateNotaryParams represents the request body for updating a notary
type UpdateNotaryParams struct {
	Email       string `json:"email,omitempty"`         // Email address for the notary account, maxLength: 5000
	USStateAbbr string `json:"us_state_abbr,omitempty"` // Two-letter state abbreviation where notary is commissioned, maxLength: 2
	FirstName   string `json:"first_name,omitempty"`    // First name of the notary (must match commission), maxLength: 5000
	MiddleName  string `json:"middle_name,omitempty"`   // Middle name of the notary (must match commission), maxLength: 5000
	LastName    string `json:"last_name,omitempty"`     // Last name of the notary (must match commission), maxLength: 5000
}

// ListNotariesParams represents query parameters for listing notaries
type ListNotariesParams struct {
	OrganizationID string `json:"organization_id,omitempty"` // ID of organization to pull notaries from (required for parent org API keys accessing child orgs)
	USStateAbbr    string `json:"us_state_abbr,omitempty"`   // Two-letter state abbreviation where notary is commissioned
}

// GetAllNotaries retrieves all notaries with optional pagination
func GetAllNotaries(client *utils.ProofClient, params *ListNotariesParams) ([]byte, error) {
	queryParams := utils.BuildQueryParams(params)
	path := NotariesEndpoint
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}
	return client.Get(path)
}

// GetNotary retrieves a specific notary by ID
func GetNotary(client *utils.ProofClient, notaryID string) ([]byte, error) {
	path := NotariesEndpoint + "/" + notaryID
	return client.Get(path)
}

// UpdateNotary updates an existing notary
func UpdateNotary(client *utils.ProofClient, notaryID string, params *UpdateNotaryParams) ([]byte, error) {
	path := NotariesEndpoint + "/" + notaryID
	return client.Put(path, params)
}

// DeleteNotary deletes a notary by ID
func DeleteNotary(client *utils.ProofClient, notaryID string) ([]byte, error) {
	path := NotariesEndpoint + "/" + notaryID
	return client.Delete(path)
}

// CreateNotary creates a new notary
func CreateNotary(client *utils.ProofClient, params *CreateNotaryParams) ([]byte, error) {
	return client.Post(NotariesEndpoint, params)
}
