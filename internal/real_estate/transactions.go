package real_estate

import (
	"time"

	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// TransactionsEndpoint is the base endpoint for real estate transactions API
	TransactionsEndpoint = "/mortgage/v2/transactions"
)

// ListTransactionsParams represents query parameters for listing transactions
type ListTransactionsParams struct {
	Limit                int        `json:"limit,omitempty"`
	Offset               int        `json:"offset,omitempty"`
	OrganizationID       string     `json:"organization_id,omitempty"`
	LoanNumber           string     `json:"loan_number,omitempty"`
	CreatedDateStart     *time.Time `json:"created_date_start,omitempty"`
	CreatedDateEnd       *time.Time `json:"created_date_end,omitempty"`
	LastUpdatedDateStart *time.Time `json:"last_updated_date_start,omitempty"`
	LastUpdatedDateEnd   *time.Time `json:"last_updated_date_end,omitempty"`
	TransactionStatus    string     `json:"transaction_status,omitempty"`
	DocumentURLVersion   string     `json:"document_url_version,omitempty"`
}

// Address represents a street address
type Address struct {
	Line1      string `json:"line1,omitempty"`       // House number and Street name, maxLength: 5000
	Line2      string `json:"line2,omitempty"`       // Unit number, maxLength: 5000
	City       string `json:"city,omitempty"`        // Town name, maxLength: 5000
	State      string `json:"state,omitempty"`       // Two character state abbreviation (e.g. "CO"), maxLength: 2
	PostalCode string `json:"postal_code,omitempty"` // Five or ten digit postal code, maxLength: 10
}

// Signer represents a transaction signer
type Signer struct {
	Email          string                 `json:"email,omitempty"`           // Email is required unless you provide a recipient_group object, maxLength: 5000
	FirstName      string                 `json:"first_name,omitempty"`      // First name, maxLength: 5000
	MiddleName     string                 `json:"middle_name,omitempty"`     // Middle name, maxLength: 5000
	LastName       string                 `json:"last_name,omitempty"`       // Last name, maxLength: 5000
	PhoneNumber    string                 `json:"phone_number,omitempty"`    // Signer phone number with country code, maxLength: 15
	ExternalID     string                 `json:"external_id,omitempty"`     // External ID, maxLength: 5000
	RecipientGroup *RecipientGroup        `json:"recipient_group,omitempty"` // Recipient group if no email provided
	Role           string                 `json:"role,omitempty"`            // Role of the signer
	BirthDate      string                 `json:"birth_date,omitempty"`      // Birth date (YYYY-MM-DD format)
	Address        *Address               `json:"address,omitempty"`         // Signer's address
	PersonalInfo   map[string]interface{} `json:"personal_info,omitempty"`   // Additional personal information
	Order          int                    `json:"order,omitempty"`           // Order the signers will sign in
}

// RecipientGroup represents a group of recipients
type RecipientGroup struct {
	GroupName string   `json:"group_name"` // Name of the recipient group
	Emails    []string `json:"emails"`     // List of email addresses in the group
}

// Cosigner represents a cosigner for a transaction
type Cosigner struct {
	FirstName   string   `json:"first_name,omitempty"`   // First name, maxLength: 5000
	MiddleName  string   `json:"middle_name,omitempty"`  // Middle name, maxLength: 5000
	LastName    string   `json:"last_name,omitempty"`    // Last name, maxLength: 5000
	Email       string   `json:"email,omitempty"`        // Email address, maxLength: 5000
	PhoneNumber string   `json:"phone_number,omitempty"` // Phone number, maxLength: 15
	ExternalID  string   `json:"external_id,omitempty"`  // External ID, maxLength: 5000
	BirthDate   string   `json:"birth_date,omitempty"`   // Birth date (YYYY-MM-DD format)
	Address     *Address `json:"address,omitempty"`      // Cosigner's address
}

// BundleOrder represents document bundle ordering
type BundleOrder struct {
	ID    string `json:"id,omitempty"`    // Document ID
	Order int    `json:"order,omitempty"` // Order in the bundle
}

// Contact represents a contact for a transaction
type Contact struct {
	Email       string `json:"email,omitempty"`        // Email address, maxLength: 5000
	FirstName   string `json:"first_name,omitempty"`   // First name, maxLength: 5000
	LastName    string `json:"last_name,omitempty"`    // Last name, maxLength: 5000
	PhoneNumber string `json:"phone_number,omitempty"` // Phone number, maxLength: 15
	Role        string `json:"role,omitempty"`         // Contact role
}

// CCRecipientEmail represents a CC recipient
type CCRecipientEmail struct {
	Email string `json:"email,omitempty"` // Email address, maxLength: 5000
}

// Attachment represents a transaction attachment
type Attachment struct {
	ID          string `json:"id,omitempty"`           // Attachment ID
	FileName    string `json:"file_name,omitempty"`    // File name
	ContentType string `json:"content_type,omitempty"` // MIME type
	URL         string `json:"url,omitempty"`          // Download URL
}

// Transaction represents a real estate transaction
type Transaction struct {
	ID                      string              `json:"id,omitempty"`
	AllowSignerAnnotations  bool                `json:"allow_signer_annotations,omitempty"`
	Attachments             []Attachment        `json:"attachments,omitempty"`
	AuditTrailURL           string              `json:"audit_trail_url,omitempty"`
	CCRecipientEmails       []CCRecipientEmail  `json:"cc_recipient_emails,omitempty"`
	ConfigID                string              `json:"config_id,omitempty"`
	DateCreated             *time.Time          `json:"date_created,omitempty"`
	DateUpdated             *time.Time          `json:"date_updated,omitempty"`
	DetailedStatus          string              `json:"detailed_status,omitempty"`
	ExpirationTime          *time.Time          `json:"expiration_time,omitempty"`
	FileNumber              string              `json:"file_number,omitempty"`
	LoanNumber              string              `json:"loan_number,omitempty"`
	NotaryID                string              `json:"notary_id,omitempty"`
	PaperNoteConsent        bool                `json:"paper_note_consent,omitempty"`
	PropertyState           string              `json:"property_state,omitempty"`
	RecordingJurisdictionID string              `json:"recording_jurisdiction_id,omitempty"`
	RequireSecondaryPhotoID bool                `json:"require_secondary_photo_id,omitempty"`
	Signer                  *Signer             `json:"signer,omitempty"`
	Signers                 []Signer            `json:"signers,omitempty"`
	StreetAddress           *Address            `json:"street_address,omitempty"`
	TitleAgencyID           string              `json:"title_agency_id,omitempty"`
	TransactionType         string              `json:"transaction_type,omitempty"`
	UnderwriterName         string              `json:"underwriter_name,omitempty"`
	WebhookURL              string              `json:"webhook_url,omitempty"`
	Contacts                []Contact           `json:"contacts,omitempty"`
	Documents               []BundleOrder       `json:"documents,omitempty"`
	Cosigner                *Cosigner           `json:"cosigner,omitempty"`
	Draft                   bool                `json:"draft,omitempty"`
	AllowedPrimaryIDTypes   map[string][]string `json:"allowed_primary_id_types,omitempty"`
	AllowedSecondaryIDTypes map[string][]string `json:"allowed_secondary_id_types,omitempty"`
}

// CreateTransactionRequest represents a request to create a transaction
type CreateTransactionRequest struct {
	TransactionType         string              `json:"transaction_type,omitempty"`
	Signer                  *Signer             `json:"signer,omitempty"`
	Signers                 []Signer            `json:"signers,omitempty"`
	Documents               []BundleOrder       `json:"documents,omitempty"`
	Cosigner                *Cosigner           `json:"cosigner,omitempty"`
	RecordingJurisdictionID string              `json:"recording_jurisdiction_id,omitempty"`
	TitleAgencyID           string              `json:"title_agency_id,omitempty"`
	Draft                   bool                `json:"draft,omitempty"`
	StreetAddress           *Address            `json:"street_address,omitempty"`
	RequireSecondaryPhotoID bool                `json:"require_secondary_photo_id,omitempty"`
	FileNumber              string              `json:"file_number,omitempty"`
	LoanNumber              string              `json:"loan_number,omitempty"`
	PaperNoteConsent        bool                `json:"paper_note_consent,omitempty"`
	ExpirationTime          *time.Time          `json:"expiration_time,omitempty"`
	WebhookURL              string              `json:"webhook_url,omitempty"`
	AllowSignerAnnotations  bool                `json:"allow_signer_annotations,omitempty"`
	CCRecipientEmails       []CCRecipientEmail  `json:"cc_recipient_emails,omitempty"`
	Contacts                []Contact           `json:"contacts,omitempty"`
	AllowedPrimaryIDTypes   map[string][]string `json:"allowed_primary_id_types,omitempty"`
	AllowedSecondaryIDTypes map[string][]string `json:"allowed_secondary_id_types,omitempty"`
}

// ListTransactionsResponse represents the response from listing transactions
type ListTransactionsResponse struct {
	Count    int           `json:"count"`
	Next     string        `json:"next,omitempty"`
	Previous string        `json:"previous,omitempty"`
	Results  []Transaction `json:"results"`
}

// ListTransactions retrieves a list of transactions
func ListTransactions(client *utils.ProofClient, params *ListTransactionsParams) ([]byte, error) {
	path := TransactionsEndpoint

	// Build query parameters
	queryParams := utils.BuildQueryParams(params)
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	return client.Get(path)
}

// GetTransaction retrieves a specific transaction by ID
func GetTransaction(client *utils.ProofClient, transactionID string) ([]byte, error) {
	path := TransactionsEndpoint + "/" + transactionID
	return client.Get(path)
}

// CreateTransaction creates a new transaction
func CreateTransaction(client *utils.ProofClient, request *CreateTransactionRequest) ([]byte, error) {
	path := TransactionsEndpoint
	return client.Post(path, request)
}

// UpdateTransaction updates an existing transaction
func UpdateTransaction(client *utils.ProofClient, transactionID string, request *CreateTransactionRequest) ([]byte, error) {
	path := TransactionsEndpoint + "/" + transactionID
	return client.Put(path, request)
}

// DeleteTransaction deletes a transaction
func DeleteTransaction(client *utils.ProofClient, transactionID string) ([]byte, error) {
	path := TransactionsEndpoint + "/" + transactionID
	return client.Delete(path)
}

// PlaceOrder places an order for a transaction
func PlaceOrder(client *utils.ProofClient, transactionID string) ([]byte, error) {
	path := TransactionsEndpoint + "/" + transactionID + "/place_order"
	return client.Post(path, nil)
}

// RecallTransaction recalls a transaction
func RecallTransaction(client *utils.ProofClient, transactionID string) ([]byte, error) {
	path := TransactionsEndpoint + "/" + transactionID + "/recall"
	return client.Post(path, nil)
}
