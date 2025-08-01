package business

import (
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// NotarizationRecordsEndpoint is the base endpoint for notarization records API
	NotarizationRecordsEndpoint = "/v1/notarization_records"
)

// GetNotarizationRecordsParams represents query parameters for retrieving notarization records
type GetNotarizationRecordsParams struct {
	Limit              int    `json:"limit,omitempty"`                // Max number of notarization records to return. Defaults to 20
	Offset             int    `json:"offset,omitempty"`               // Number to offset the notarization records if limit is reached. Defaults to 0
	DocumentURLVersion string `json:"document_url_version,omitempty"` // Control documents and signer photo identifications download URLs. v1 for AWS S3 pre-signed URLs. v2 for Proof secure URLs. Default: v1
}

// GetNotarizationRecordParams represents query parameters for retrieving a specific notarization record
type GetNotarizationRecordParams struct {
	DocumentURLVersion string `json:"document_url_version,omitempty"` // Control documents and signer photo identifications download URLs. v1 for AWS S3 pre-signed URLs. v2 for Proof secure URLs. Default: v1
}

// GetNotarizationRecords retrieves meeting records with optional pagination and document URL version
func GetNotarizationRecords(client *utils.ProofClient, params *GetNotarizationRecordsParams) ([]byte, error) {

	queryParams := utils.BuildQueryParams(params)
	queryParams.Set("document_url_version", DocumentURLVersion)
	path := NotarizationRecordsEndpoint
	path += "?" + queryParams.Encode()

	return client.Get(path)
}

// GetNotarizationRecord retrieves a specific meeting record by ID
func GetNotarizationRecord(client *utils.ProofClient, recordID string, params *GetNotarizationRecordParams) ([]byte, error) {

	queryParams := utils.BuildQueryParams(params)
	queryParams.Set("document_url_version", DocumentURLVersion)
	path := NotarizationRecordsEndpoint + "/" + recordID
	path += "?" + queryParams.Encode()

	return client.Get(path)
}
