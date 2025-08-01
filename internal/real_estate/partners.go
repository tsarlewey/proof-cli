package real_estate

import (
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// PartnersEndpoint is the base endpoint for partners API
	PartnersEndpoint = "/mortgage/v2/partners"
)

// Partner represents a partner organization
type Partner struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`         // Partner name, maxLength: 5000
	Type        string   `json:"type,omitempty"`         // Partner type (e.g., "title_agency", "lender")
	Email       string   `json:"email,omitempty"`        // Contact email, maxLength: 5000
	PhoneNumber string   `json:"phone_number,omitempty"` // Contact phone number, maxLength: 15
	Address     *Address `json:"address,omitempty"`      // Partner address
	Status      string   `json:"status,omitempty"`       // Partner status (e.g., "active", "inactive")
	CreatedAt   string   `json:"created_at,omitempty"`   // ISO 8601 creation timestamp
	UpdatedAt   string   `json:"updated_at,omitempty"`   // ISO 8601 last update timestamp
}

// CreatePartnerRequest represents a request to create a partner
type CreatePartnerRequest struct {
	Name        string   `json:"name"`                   // Partner name, maxLength: 5000
	Type        string   `json:"type"`                   // Partner type
	Email       string   `json:"email,omitempty"`        // Contact email, maxLength: 5000
	PhoneNumber string   `json:"phone_number,omitempty"` // Contact phone number, maxLength: 15
	Address     *Address `json:"address,omitempty"`      // Partner address
}

// UpdatePartnerRequest represents a request to update a partner
type UpdatePartnerRequest struct {
	Name        string   `json:"name,omitempty"`         // Partner name, maxLength: 5000
	Type        string   `json:"type,omitempty"`         // Partner type
	Email       string   `json:"email,omitempty"`        // Contact email, maxLength: 5000
	PhoneNumber string   `json:"phone_number,omitempty"` // Contact phone number, maxLength: 15
	Address     *Address `json:"address,omitempty"`      // Partner address
	Status      string   `json:"status,omitempty"`       // Partner status
}

// ListPartnersParams represents query parameters for listing partners
type ListPartnersParams struct {
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Type   string `json:"type,omitempty"`   // Filter by partner type
	Status string `json:"status,omitempty"` // Filter by status
}

// ListPartnersResponse represents the response from listing partners
type ListPartnersResponse struct {
	Count    int       `json:"count"`
	Next     string    `json:"next,omitempty"`
	Previous string    `json:"previous,omitempty"`
	Results  []Partner `json:"results"`
}

// ListPartners retrieves a list of partners
func ListPartners(client *utils.ProofClient, params *ListPartnersParams) ([]byte, error) {
	path := PartnersEndpoint

	// Build query parameters
	queryParams := utils.BuildQueryParams(params)
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	return client.Get(path)
}

// GetPartner retrieves a specific partner by ID
func GetPartner(client *utils.ProofClient, partnerID string) ([]byte, error) {
	path := PartnersEndpoint + "/" + partnerID
	return client.Get(path)
}

// CreatePartner creates a new partner
func CreatePartner(client *utils.ProofClient, request *CreatePartnerRequest) ([]byte, error) {
	path := PartnersEndpoint
	return client.Post(path, request)
}

// UpdatePartner updates an existing partner
func UpdatePartner(client *utils.ProofClient, partnerID string, request *UpdatePartnerRequest) ([]byte, error) {
	path := PartnersEndpoint + "/" + partnerID
	return client.Put(path, request)
}

// DeletePartner deletes a partner
func DeletePartner(client *utils.ProofClient, partnerID string) ([]byte, error) {
	path := PartnersEndpoint + "/" + partnerID
	return client.Delete(path)
}
