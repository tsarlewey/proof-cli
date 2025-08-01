package real_estate

import (
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

// VerifyAddressRequest represents a request to verify an address
type VerifyAddressRequest struct {
	StreetAddress *Address `json:"street_address"` // Address to verify
}

// VerifyAddressResponse represents a response from address verification
type VerifyAddressResponse struct {
	IsValid               bool                   `json:"is_valid"`
	RecordingJurisdiction *RecordingJurisdiction `json:"recording_jurisdiction,omitempty"`
	TitleAgencies         []TitleAgency          `json:"title_agencies,omitempty"`
	SuggestedAddress      *Address               `json:"suggested_address,omitempty"`
}

// RecordingJurisdiction represents a recording jurisdiction
type RecordingJurisdiction struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`         // Jurisdiction name
	State       string   `json:"state,omitempty"`        // State abbreviation
	County      string   `json:"county,omitempty"`       // County name
	Address     *Address `json:"address,omitempty"`      // Jurisdiction address
	PhoneNumber string   `json:"phone_number,omitempty"` // Contact phone number
	Website     string   `json:"website,omitempty"`      // Website URL
}

// TitleAgency represents a title agency
type TitleAgency struct {
	ID            string   `json:"id,omitempty"`
	Name          string   `json:"name,omitempty"`           // Agency name
	Address       *Address `json:"address,omitempty"`        // Agency address
	PhoneNumber   string   `json:"phone_number,omitempty"`   // Contact phone number
	Email         string   `json:"email,omitempty"`          // Contact email
	Website       string   `json:"website,omitempty"`        // Website URL
	LicenseNumber string   `json:"license_number,omitempty"` // License number
}

// TitleUnderwriter represents a title underwriter
type TitleUnderwriter struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`         // Underwriter name
	Address     *Address `json:"address,omitempty"`      // Underwriter address
	PhoneNumber string   `json:"phone_number,omitempty"` // Contact phone number
	Email       string   `json:"email,omitempty"`        // Contact email
	Website     string   `json:"website,omitempty"`      // Website URL
}

// ListRecordingJurisdictionsParams represents query parameters for listing recording jurisdictions
type ListRecordingJurisdictionsParams struct {
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
	State  string `json:"state,omitempty"`  // Filter by state
	County string `json:"county,omitempty"` // Filter by county
}

// ListRecordingJurisdictionsResponse represents the response from listing recording jurisdictions
type ListRecordingJurisdictionsResponse struct {
	Count    int                     `json:"count"`
	Next     string                  `json:"next,omitempty"`
	Previous string                  `json:"previous,omitempty"`
	Results  []RecordingJurisdiction `json:"results"`
}

// ListTitleAgenciesParams represents query parameters for listing title agencies
type ListTitleAgenciesParams struct {
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
	State  string `json:"state,omitempty"` // Filter by state
}

// ListTitleAgenciesResponse represents the response from listing title agencies
type ListTitleAgenciesResponse struct {
	Count    int           `json:"count"`
	Next     string        `json:"next,omitempty"`
	Previous string        `json:"previous,omitempty"`
	Results  []TitleAgency `json:"results"`
}

// ListTitleUnderwritersResponse represents the response from listing title underwriters
type ListTitleUnderwritersResponse struct {
	Count    int                `json:"count"`
	Next     string             `json:"next,omitempty"`
	Previous string             `json:"previous,omitempty"`
	Results  []TitleUnderwriter `json:"results"`
}

// VerifyAddress verifies a street address and returns recording jurisdiction info
func VerifyAddress(client *utils.ProofClient, request *VerifyAddressRequest) ([]byte, error) {
	path := "/mortgage/v2/transactions/verify_address"
	return client.Post(path, request)
}

// ListTitleUnderwriters retrieves a list of title underwriters
func ListTitleUnderwriters(client *utils.ProofClient) ([]byte, error) {
	path := "/mortgage/v2/transactions/title_underwriters"
	return client.Get(path)
}

// ListRecordingJurisdictions retrieves a list of recording jurisdictions
func ListRecordingJurisdictions(client *utils.ProofClient, params *ListRecordingJurisdictionsParams) ([]byte, error) {
	path := "/mortgage/v2/recording_locations"

	// Build query parameters
	queryParams := utils.BuildQueryParams(params)
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	return client.Get(path)
}

// ListTitleAgencies retrieves a list of title agencies
func ListTitleAgencies(client *utils.ProofClient, params *ListTitleAgenciesParams) ([]byte, error) {
	path := "/mortgage/v2/title_agencies"

	// Build query parameters
	queryParams := utils.BuildQueryParams(params)
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	return client.Get(path)
}
