package business

import (
	"net/url"
	"time"

	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// TransactionsEndpoint is the base endpoint for transactions API
	TransactionsEndpoint = "/v1/transactions"
	DocumentURLVersion   = "v2"
)

// ListTransactionsParams represents query parameters for listing transactions
type ListTransactionsParams struct {
	Limit                int        `json:"limit,omitempty"`
	Offset               int        `json:"offset,omitempty"`
	CreatedDateStart     *time.Time `json:"created_date_start,omitempty"`
	CreatedDateEnd       *time.Time `json:"created_date_end,omitempty"`
	LastUpdatedDateStart *time.Time `json:"last_updated_date_start,omitempty"`
	LastUpdatedDateEnd   *time.Time `json:"last_updated_date_end,omitempty"`
	TransactionStatus    string     `json:"transaction_status,omitempty"`
	DocumentURLVersion   string     `json:"document_url_version,omitempty"`
}

// Address represents a physical address
type Address struct {
	Line1   string `json:"line1,omitempty"`   // House number and Street name, maxLength: 5000
	Line2   string `json:"line2,omitempty"`   // Unit number, maxLength: 5000
	City    string `json:"city,omitempty"`    // Town name, maxLength: 5000
	State   string `json:"state,omitempty"`   // Two character state abbreviation (e.g. "CO"), maxLength: 2
	Postal  string `json:"postal,omitempty"`  // Five or ten digit postal code, maxLength: 10
	Country string `json:"country,omitempty"` // Two character country code (e.g. "US"), maxLength: 2
}

// CredentialAsset represents a credential asset object
type CredentialAsset struct {
	// Add credential asset fields as needed
}

// ProofRequirementIDVerification represents ID verification requirements
type ProofRequirementIDVerification struct {
	Selfie bool `json:"selfie"` // Required: Signer identity verified through credential analysis and selfie comparison
}

// ProofRequirementMultiFactorAuthentication represents MFA requirements
type ProofRequirementMultiFactorAuthentication struct {
	// Add MFA fields as needed
}

// ProofRequirement represents proof requirements for a signer
type ProofRequirement struct {
	IDVerification               *ProofRequirementIDVerification            `json:"id_verification,omitempty"`
	KnowledgeBasedAuthentication bool                                       `json:"knowledge_based_authentication"` // Required: (KBA) Signer answers personal identity questions
	MultiFactorAuthentication    *ProofRequirementMultiFactorAuthentication `json:"multi_factor_authentication,omitempty"`
}

// RecipientGroup represents a group recipient configuration
type RecipientGroup struct {
	SharedInboxEmail string `json:"shared_inbox_email,omitempty"` // Email shared by a group, maxLength: 5000
}

// SigningGroup represents a signing group (deprecated, use RecipientGroup)
type SigningGroup struct {
	SharedInboxEmail string `json:"shared_inbox_email,omitempty"` // (Deprecated) use recipient_group, maxLength: 5000
}

// VerifyToolConfiguration represents verification tool configuration
type VerifyToolConfiguration struct {
	Requirement string `json:"requirement"` // Required: "run", "pass", or "none"
}

// VerifyToolsConfiguration represents all verification tool configurations
type VerifyToolsConfiguration struct {
	CreditCardConfig         *VerifyToolConfiguration `json:"credit_card_config,omitempty"`
	SelfieConfig             *VerifyToolConfiguration `json:"selfie_config,omitempty"`
	CredentialAnalysisConfig *VerifyToolConfiguration `json:"credential_analysis_config,omitempty"`
}

// Signer represents a transaction signer
type Signer struct {
	Email                         string                    `json:"email"`                  // Required unless recipient_group provided, maxLength: 5000
	FirstName                     string                    `json:"first_name,omitempty"`   // First name, maxLength: 5000
	MiddleName                    string                    `json:"middle_name,omitempty"`  // Middle name, maxLength: 5000
	LastName                      string                    `json:"last_name,omitempty"`    // Last name, maxLength: 5000
	PhoneNumber                   string                    `json:"phone_number,omitempty"` // Phone with optional country code (e.g. +11234567890), maxLength: 5000
	Address                       *Address                  `json:"address,omitempty"`
	ExternalID                    string                    `json:"external_id,omitempty"`                // External system ID, maxLength: 5000
	Entity                        string                    `json:"entity,omitempty"`                     // Entity signing on behalf of (required with capacity), maxLength: 5000
	Capacity                      string                    `json:"capacity,omitempty"`                   // Capacity when signing for entity (required with entity), maxLength: 5000
	CredentialAssets              []CredentialAsset         `json:"credential_assets,omitempty"`          // Recent credential assets
	PersonallyKnownToNotary       bool                      `json:"personally_known_to_notary,omitempty"` // If signer is personally known to notary
	ProofRequirement              *ProofRequirement         `json:"proof_requirement,omitempty"`
	RecipientGroup                *RecipientGroup           `json:"recipient_group,omitempty"`
	SigningGroup                  *SigningGroup             `json:"signing_group,omitempty"`       // Deprecated: use recipient_group
	SigningRequirement            *string                   `json:"signing_requirement,omitempty"` // Override document signing requirement
	VerifyToolsConfiguration      *VerifyToolsConfiguration `json:"verify_tools_configuration,omitempty"`
	PrimaryIDAllowListByCountry   map[string][]string       `json:"primary_id_allow_list_by_country,omitempty"`   // Government IDs by country
	SecondaryIDAllowListByCountry map[string][]string       `json:"secondary_id_allow_list_by_country,omitempty"` // IDs and supplemental docs by country
	SigningStatus                 string                    `json:"signing_status,omitempty"`                     // "incomplete", "in_progress", or "complete"
}

// Cosigner represents a transaction cosigner
type Cosigner struct {
	FirstName          string  `json:"first_name,omitempty"`          // First name, maxLength: 5000
	LastName           string  `json:"last_name,omitempty"`           // Last name, maxLength: 5000
	SigningRequirement *string `json:"signing_requirement,omitempty"` // Override document signing requirement
}

// NotaryInstruction represents instructions for the notary
type NotaryInstruction struct {
	NotaryNote string `json:"notary_note,omitempty"` // 500 character limit
}

// Redirect represents post-completion redirect configuration
type Redirect struct {
	URL     string `json:"url,omitempty"`     // Valid URL to redirect after meeting, maxLength: 5000
	Message string `json:"message,omitempty"` // Message to show with redirect, maxLength: 5000
}

// TransactionParams represents the common request body structure for creating, updating, or patching a transaction
type TransactionParams struct {
	ActivationTime               string              `json:"activation_time,omitempty"`                 // ISO-8601 datetime when signer can connect with notary, maxLength: 100
	Signer                       *Signer             `json:"signer,omitempty"`                          // Single signer (deprecated, use signers array)
	Signers                      []Signer            `json:"signers,omitempty"`                         // Array of signers (limit 10), each needs at minimum email
	Cosigner                     *Cosigner           `json:"cosigner,omitempty"`                        // Optional cosigner
	Draft                        bool                `json:"draft,omitempty"`                           // Create transaction in draft state
	Expiry                       string              `json:"expiry,omitempty"`                          // ISO-8601 datetime after which transaction expires, maxLength: 100
	Payer                        string              `json:"payer,omitempty"`                           // "signer" or "sender"
	ExternalID                   string              `json:"external_id,omitempty"`                     // External system ID (e.g. Order ID), maxLength: 5000
	TransactionName              string              `json:"transaction_name,omitempty"`                // Human-readable name, defaults to "Untitled Transaction", maxLength: 5000
	MessageToSigner              string              `json:"message_to_signer,omitempty"`               // GitHub Flavored Markdown message, maxLength: 30000
	TransactionType              string              `json:"transaction_type,omitempty"`                // Category (e.g. "Account Opening"), maxLength: 5000
	MessageSubject               string              `json:"message_subject,omitempty"`                 // Email subject line, maxLength: 30000
	MessageSignature             string              `json:"message_signature,omitempty"`               // Email signature, maxLength: 30000
	PDFBookmarked                bool                `json:"pdf_bookmarked,omitempty"`                  // Split PDF by bookmarks
	RequireSecondaryPhotoID      bool                `json:"require_secondary_photo_id,omitempty"`      // Require two forms of photo ID
	SuppressEmail                bool                `json:"suppress_email,omitempty"`                  // Don't send notification email on activation
	AuthenticationRequirement    string              `json:"authentication_requirement,omitempty"`      // "sms" or "none"
	RequireNewSignerVerification bool                `json:"require_new_signer_verification,omitempty"` // Require email verification for new signers
	NotaryInstructions           []NotaryInstruction `json:"notary_instructions,omitempty"`             // Instructions for notary
	Redirect                     *Redirect           `json:"redirect,omitempty"`                        // Post-completion redirect
	NotaryID                     string              `json:"notary_id,omitempty"`                       // Assigned notary user ID, maxLength: 5000
	NotaryMeetingTime            string              `json:"notary_meeting_time,omitempty"`             // ISO-8601 meeting time, maxLength: 100
	DocumentURLVersion           string              `json:"document_url_version,omitempty"`            // "v1" or "v2"
	Documents                    []string            `json:"documents,omitempty"`                       // Array of document resources, maxLength per doc: 42000000
	Document                     string              `json:"document,omitempty"`                        // Single document (deprecated), maxLength: 42000000
	CCRecipientEmails            []string            `json:"cc_recipient_emails,omitempty"`             // CC email addresses
	ConfigID                     string              `json:"config_id,omitempty"`                       // Configuration ID, maxLength: 5000
}

// CreateTransactionParams is an alias for TransactionParams used when creating a transaction
type CreateTransactionParams = TransactionParams

// UpdateTransactionParams is an alias for TransactionParams used when updating a transaction
type UpdateTransactionParams = TransactionParams

// PatchTransactionParams is an alias for TransactionParams used when patching a transaction
type PatchTransactionParams = TransactionParams

// GetAllTransactions lists transactions with optional filtering
func GetAllTransactions(client *utils.ProofClient, params *ListTransactionsParams) ([]byte, error) {
	// Build query parameters
	queryParams := utils.BuildQueryParams(params)

	path := TransactionsEndpoint
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	return client.Get(path)
}

// GetTransaction retrieves a specific transaction by ID
func GetTransaction(client *utils.ProofClient, id string) ([]byte, error) {
	path := TransactionsEndpoint + "/" + id

	// Add default document URL version
	queryParams := url.Values{}
	queryParams.Set("document_url_version", DocumentURLVersion)
	path += "?" + queryParams.Encode()

	return client.Get(path)
}

// GetAllEligibleNotaries gets eligible notaries for a transaction
func GetAllEligibleNotaries(client *utils.ProofClient, id string) ([]byte, error) {
	path := TransactionsEndpoint + "/" + id + "/notaries"
	return client.Get(path)
}

// CreateTransaction creates a new transaction
func CreateTransaction(client *utils.ProofClient, params *CreateTransactionParams) ([]byte, error) {
	path := TransactionsEndpoint

	// Add default document URL version
	queryParams := url.Values{}
	queryParams.Set("document_url_version", DocumentURLVersion)
	path += "?" + queryParams.Encode()

	return client.Post(path, params)
}

// UpdateDraftTransaction updates a draft transaction
func UpdateDraftTransaction(client *utils.ProofClient, id string, params *UpdateTransactionParams) ([]byte, error) {
	path := TransactionsEndpoint + "/" + id

	// Add default document URL version
	queryParams := url.Values{}
	queryParams.Set("document_url_version", DocumentURLVersion)
	path += "?" + queryParams.Encode()

	return client.Put(path, params)
}

// PatchDraftTransaction performs a partial update on a draft transaction
func PatchDraftTransaction(client *utils.ProofClient, id string, params *PatchTransactionParams) ([]byte, error) {
	path := TransactionsEndpoint + "/" + id

	// Add default document URL version
	queryParams := url.Values{}
	queryParams.Set("document_url_version", DocumentURLVersion)
	path += "?" + queryParams.Encode()

	return client.Patch(path, params)
}

// DeleteTransaction deletes a transaction by ID
func DeleteTransaction(client *utils.ProofClient, id string) ([]byte, error) {
	path := TransactionsEndpoint + "/" + id
	return client.Delete(path)
}

// ActivateDraftTransaction activates a draft transaction (notarization ready)
func ActivateDraftTransaction(client *utils.ProofClient, id string) ([]byte, error) {
	path := TransactionsEndpoint + "/" + id + "/notarization_ready"

	// Add default document URL version
	queryParams := url.Values{}
	queryParams.Set("document_url_version", DocumentURLVersion)
	path += "?" + queryParams.Encode()

	return client.Post(path, nil)
}

// RecallTransaction recalls a transaction with an optional reason
func RecallTransaction(client *utils.ProofClient, id string, recallReason string) ([]byte, error) {
	path := TransactionsEndpoint + "/" + id + "/recall"

	// Build query parameters
	queryParams := url.Values{}
	queryParams.Set("document_url_version", DocumentURLVersion)

	// Add optional recall reason
	if recallReason != "" {
		queryParams.Set("recall_reason", recallReason)
	}

	path += "?" + queryParams.Encode()

	return client.Post(path, nil)
}

// ResendTransactionEmail resends the transaction email with an optional message
func ResendTransactionEmail(client *utils.ProofClient, id string, messageToSigner string) ([]byte, error) {
	path := TransactionsEndpoint + "/" + id + "/send_email"

	// Build query parameters
	queryParams := url.Values{}
	queryParams.Set("document_url_version", DocumentURLVersion)

	// Add optional message to signer
	if messageToSigner != "" {
		queryParams.Set("message_to_signer", messageToSigner)
	}

	path += "?" + queryParams.Encode()

	return client.Post(path, nil)
}

// ResendTransactionSMS resends the transaction SMS with optional phone parameters
func ResendTransactionSMS(client *utils.ProofClient, id string, phoneNumber string) ([]byte, error) {
	path := TransactionsEndpoint + "/" + id + "/send_sms"

	// Build query parameters
	queryParams := url.Values{}
	queryParams.Set("document_url_version", DocumentURLVersion)

	if phoneNumber != "" {
		queryParams.Set("phone_number", phoneNumber)
	}

	path += "?" + queryParams.Encode()

	return client.Post(path, nil)
}
