package real_estate

import (
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// NotarizationRecordsEndpoint is the base endpoint for notarization records API
	NotarizationRecordsEndpoint = "/mortgage/v2/notarization_records"
)

// NotarizationRecord represents a notarization record
type NotarizationRecord struct {
	ID                  string `json:"id,omitempty"`
	TransactionID       string `json:"transaction_id,omitempty"`
	NotaryID            string `json:"notary_id,omitempty"`
	SignerID            string `json:"signer_id,omitempty"`
	DocumentID          string `json:"document_id,omitempty"`
	Status              string `json:"status,omitempty"`
	NotarizationType    string `json:"notarization_type,omitempty"`
	NotarizationDate    string `json:"notarization_date,omitempty"` // ISO 8601 timestamp
	CompletedAt         string `json:"completed_at,omitempty"`      // ISO 8601 timestamp
	CreatedAt           string `json:"created_at,omitempty"`        // ISO 8601 timestamp
	UpdatedAt           string `json:"updated_at,omitempty"`        // ISO 8601 timestamp
	SignatureCount      int    `json:"signature_count,omitempty"`
	AffirmationCount    int    `json:"affirmation_count,omitempty"`
	AcknowledgmentCount int    `json:"acknowledgment_count,omitempty"`
	JuratCount          int    `json:"jurat_count,omitempty"`
	CertificateURL      string `json:"certificate_url,omitempty"`
	AuditTrailURL       string `json:"audit_trail_url,omitempty"`
}

// ListNotarizationRecordsParams represents query parameters for listing notarization records
type ListNotarizationRecordsParams struct {
	Limit         int    `json:"limit,omitempty"`
	Offset        int    `json:"offset,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	NotaryID      string `json:"notary_id,omitempty"`
	Status        string `json:"status,omitempty"`
}

// ListNotarizationRecordsResponse represents the response from listing notarization records
type ListNotarizationRecordsResponse struct {
	Count    int                  `json:"count"`
	Next     string               `json:"next,omitempty"`
	Previous string               `json:"previous,omitempty"`
	Results  []NotarizationRecord `json:"results"`
}

// GetNotarizationRecord retrieves a specific notarization record by ID
func GetNotarizationRecord(client *utils.ProofClient, recordID string) ([]byte, error) {
	path := NotarizationRecordsEndpoint + "/" + recordID
	return client.Get(path)
}

// ListNotarizationRecords retrieves a list of notarization records
func ListNotarizationRecords(client *utils.ProofClient, params *ListNotarizationRecordsParams) ([]byte, error) {
	path := NotarizationRecordsEndpoint

	// Build query parameters
	queryParams := utils.BuildQueryParams(params)
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	return client.Get(path)
}
