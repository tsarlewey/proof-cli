package business

import (
	"net/url"

	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// DocumentsEndpoint is the base endpoint for documents API
	DocumentsEndpoint = "/v1/documents"
)

// SigningDesignationGroup represents a group of signing designations
type SigningDesignationGroup struct {
	Name        string `json:"name,omitempty"`         // Name of the group, maxLength: 5000, regex: /^[a-zA-Z0-9_-()]{1,64}$/
	MinRequired int    `json:"min_required,omitempty"` // Minimum number of designations that must be fulfilled
	MaxRequired int    `json:"max_required,omitempty"` // Maximum number of designations that can be fulfilled
}

// SigningDesignation represents a signing designation for a document
type SigningDesignation struct {
	SignerIdentifier             string                   `json:"signer_identifier,omitempty"`              // External ID for signers, witness{n} for witnesses, notary for notaries, maxLength: 5000
	PageNumber                   int                      `json:"page_number,omitempty"`                    // Page number starting at 0
	X                            int                      `json:"x,omitempty"`                              // X coordinate from bottom left
	Y                            int                      `json:"y,omitempty"`                              // Y coordinate from bottom left
	Height                       int                      `json:"height,omitempty"`                         // Height
	Width                        int                      `json:"width,omitempty"`                          // Width
	Hint                         string                   `json:"hint,omitempty"`                           // Hint text, maxLength: 5000
	Type                         string                   `json:"type,omitempty"`                           // Type of field designation (see API docs for enum values)
	SigningDesignationGroup      *SigningDesignationGroup `json:"signing_designation_group,omitempty"`      // Group configuration
	Optional                     bool                     `json:"optional,omitempty"`                       // Whether designation is optional
	Instruction                  string                   `json:"instruction,omitempty"`                    // Instruction for signer, maxLength: 5000
	PrimaryDesignationIdentifier string                   `json:"primary_designation_identifier,omitempty"` // Unique ID for primary designation, maxLength: 5000
	ConditionalOnPrimary         string                   `json:"conditional_on_primary,omitempty"`         // ID for conditional designation, maxLength: 5000
}

// AddDocumentParams represents the request body for adding a document to a transaction
type AddDocumentParams struct {
	Filename                     string                    `json:"filename,omitempty"`                       // Plain language name for the document, maxLength: 5000
	Resource                     string                    `json:"resource,omitempty"`                       // Document file resource (URL, Base64, or template permalink), maxLength: 42000000
	Document                     string                    `json:"document,omitempty"`                       // Document content
	Requirement                  string                    `json:"requirement,omitempty"`                    // Completion requirement (notarization, esign, identity_confirmation, readonly, non_essential), maxLength: 5000
	NotarizationRequired         bool                      `json:"notarization_required,omitempty"`          // Whether notarization is required
	WitnessRequired              bool                      `json:"witness_required,omitempty"`               // Whether additional witness must be present
	BundlePosition               int                       `json:"bundle_position,omitempty"`                // Position in document bundle
	EsignRequired                bool                      `json:"esign_required,omitempty"`                 // Whether e-signature is required
	IdentityConfirmationRequired bool                      `json:"identity_confirmation_required,omitempty"` // Whether identity confirmation is required
	SigningRequiresMeeting       bool                      `json:"signing_requires_meeting,omitempty"`       // Whether signing requires a meeting
	Vaulted                      bool                      `json:"vaulted,omitempty"`                        // Whether to store authoritative copy in eVault
	AuthorizationHeader          string                    `json:"authorization_header,omitempty"`           // Header for fetching doc URLs (format: "header_name:header_value"), maxLength: 5000
	CustomerCanAnnotate          bool                      `json:"customer_can_annotate,omitempty"`          // Whether signer can add annotations
	PDFBookmarked                bool                      `json:"pdf_bookmarked,omitempty"`                 // Whether document is bookmarked PDF (splits by bookmarks)
	TrackingID                   string                    `json:"tracking_id,omitempty"`                    // External tracking identifier, maxLength: 5000
	TextTagSyntax                string                    `json:"text_tag_syntax,omitempty"`                // Syntax used by text tags, maxLength: 5000
	SigningDesignations          []SigningDesignation      `json:"signing_designations,omitempty"`           // Array of signing designations
	SigningDesignationGroups     []SigningDesignationGroup `json:"signing_designation_groups,omitempty"`     // Array of signing designation groups
}

// AddDocument adds a document to a transaction
func AddDocument(client *utils.ProofClient, transactionID string, params *AddDocumentParams) ([]byte, error) {
	path := TransactionsEndpoint + "/" + transactionID + "/documents"

	// Add default document URL version
	queryParams := url.Values{}
	queryParams.Set("document_url_version", DocumentURLVersion)
	path += "?" + queryParams.Encode()

	return client.Post(path, params)
}

// UpdateDocumentParams represents the request body for updating a document
type UpdateDocumentParams struct {
	Name                         string `json:"name,omitempty"`                           // Plain language name for the document, maxLength: 5000
	DocumentURLVersion           bool   `json:"document_url_version,omitempty"`           // Optional param to test v2 document urls
	Requirement                  string `json:"requirement,omitempty"`                    // Completion requirement (notarization, esign, identity_confirmation, readonly, non_essential), maxLength: 5000
	NotarizationRequired         bool   `json:"notarization_required,omitempty"`          // Whether notarization is required
	WitnessRequired              bool   `json:"witness_required,omitempty"`               // Whether additional witness must be present
	IdentityConfirmationRequired bool   `json:"identity_confirmation_required,omitempty"` // Whether identity confirmation is required
	CustomerCanAnnotate          bool   `json:"customer_can_annotate,omitempty"`          // Whether signer can add annotations
	PDFBookmarked                bool   `json:"pdf_bookmarked,omitempty"`                 // Whether document is bookmarked PDF (splits by bookmarks)
	TrackingID                   string `json:"tracking_id,omitempty"`                    // External tracking identifier, maxLength: 5000
}

// UpdateDocument updates an existing document
func UpdateDocument(client *utils.ProofClient, documentID string, params *UpdateDocumentParams) ([]byte, error) {
	path := DocumentsEndpoint + "/" + documentID

	// Add default document URL version
	queryParams := url.Values{}
	queryParams.Set("document_url_version", DocumentURLVersion)
	path += "?" + queryParams.Encode()

	return client.Put(path, params)
}

// PatchDocument performs a partial update on an existing document
func PatchDocument(client *utils.ProofClient, documentID string, params *UpdateDocumentParams) ([]byte, error) {
	path := DocumentsEndpoint + "/" + documentID

	// Add default document URL version
	queryParams := url.Values{}
	queryParams.Set("document_url_version", DocumentURLVersion)
	path += "?" + queryParams.Encode()

	return client.Patch(path, params)
}

// DeleteDocument deletes a document by ID
func DeleteDocument(client *utils.ProofClient, documentID string) ([]byte, error) {
	path := DocumentsEndpoint + "/" + documentID
	return client.Delete(path)
}

// GetDocumentParams represents query parameters for getting a document
type GetDocumentParams struct {
	Encoding string `json:"encoding,omitempty"` // Can be "base64" or "uri". "uri" returns hosted URL (only after transaction completion)
}

// GetDocument retrieves a specific document from a transaction
func GetDocument(client *utils.ProofClient, transactionID string, documentID string, params *GetDocumentParams) ([]byte, error) {
	path := TransactionsEndpoint + "/" + transactionID + "/documents/" + documentID

	// Build query parameters
	queryParams := url.Values{}
	queryParams.Set("document_url_version", DocumentURLVersion)

	// Add optional encoding parameter
	if params != nil && params.Encoding != "" {
		queryParams.Set("encoding", params.Encoding)
	}

	path += "?" + queryParams.Encode()

	return client.Get(path)
}
