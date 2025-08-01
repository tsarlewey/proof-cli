package business

import (
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// WebhooksEndpoint is the base endpoint for webhooks API v2
	WebhooksEndpoint = "/v2/webhooks"
)

const (
	WebhookSubscriptionAll                                 = "*"
	WebhookSubscriptionTransactionAll                      = "transaction.*"
	WebhookSubscriptionTransactionCreated                  = "transaction.created"
	WebhookSubscriptionTransactionSent                     = "transaction.sent"
	WebhookSubscriptionTransactionUpdated                  = "transaction.updated"
	WebhookSubscriptionTransactionReceived                 = "transaction.received"
	WebhookSubscriptionTransactionCompleted                = "transaction.completed"
	WebhookSubscriptionTransactionCompletedWithRejections  = "transaction.completed_with_rejections"
	WebhookSubscriptionTransactionPartiallyCompleted       = "transaction.partially_completed"
	WebhookSubscriptionTransactionDeleted                  = "transaction.deleted"
	WebhookSubscriptionTransactionRecalled                 = "transaction.recalled"
	WebhookSubscriptionTransactionExpired                  = "transaction.expired"
	WebhookSubscriptionTransactionReleased                 = "transaction.released"
	WebhookSubscriptionTransactionReviewed                 = "transaction.reviewed"
	WebhookSubscriptionTransactionSentToClosingOps         = "transaction.sent_to_closing_ops"
	WebhookSubscriptionTransactionSentToSigner             = "transaction.sent_to_signer"
	WebhookSubscriptionTransactionSentToTitleAgency        = "transaction.sent_to_title_agency"
	WebhookSubscriptionTransactionDocumentUpload           = "transaction.document.upload"
	WebhookSubscriptionTransactionDocumentProcessed        = "transaction.document.processed"
	WebhookSubscriptionTransactionImportProcessed          = "transaction.import.processed"
	WebhookSubscriptionTransactionImportFailed             = "transaction.import.failed"
	WebhookSubscriptionTransactionMeetingCreated           = "transaction.meeting.created"
	WebhookSubscriptionTransactionMeetingFailed            = "transaction.meeting.failed"
	WebhookSubscriptionTransactionMeetingRequested         = "transaction.meeting.requested"
	WebhookSubscriptionTransactionMeetingVideoProcessed    = "transaction.meeting.video.processed"
	WebhookSubscriptionTransactionNotaryAssigned           = "transaction.notary.assigned"
	WebhookSubscriptionTransactionSignerKBAFailed          = "transaction.signer.kba_failed"
	WebhookSubscriptionTransactionSignerKBAPassed          = "transaction.signer.kba_passed"
	WebhookSubscriptionTransactionSignerHighRiskDetected   = "transaction.signer.high_risk_detected"
	WebhookSubscriptionTransactionSignerMediumRiskDetected = "transaction.signer.medium_risk_detected"
	WebhookSubscriptionTransactionIDVFailed                = "transaction.transaction.automated_identity_verification_by_idv_service_failed"
	WebhookSubscriptionTransactionIDVPassed                = "transaction.transaction.automated_identity_verification_by_idv_service_passed"
	WebhookSubscriptionTransactionUnderwriterNotAvailable  = "transaction.underwriter.not_available"
	WebhookSubscriptionNotaryAll                           = "notary.*"
	WebhookSubscriptionNotaryCreated                       = "notary.created"
	WebhookSubscriptionNotaryNeedsReview                   = "notary.needs_review"
	WebhookSubscriptionNotaryCompliant                     = "notary.compliant"
	WebhookSubscriptionNotaryNonCompliant                  = "notary.non_compliant"
	WebhookSubscriptionNotarySignerReady                   = "notary.signer_ready"
)

// CreateWebhookParams represents the request body for creating a webhook
type CreateWebhookParams struct {
	URL           string   `json:"url"`              // Required: Proof will make a POST request to this URL for every subscribed event, maxLength: 5000
	Header        *string  `json:"header,omitempty"` // Optional: Header value to pass through every request (e.g. "X-Custom-Header:X-Custom-Key"), maxLength: 5000
	Subscriptions []string `json:"subscriptions"`    // Required: Array of events to subscribe to (see WebhookSubscriptions constants)
}

// UpdateWebhookParams represents the request body for updating a webhook
type UpdateWebhookParams struct {
	URL           string   `json:"url"`              // Required: Proof will make a POST request to this URL for every subscribed event, maxLength: 5000
	Header        *string  `json:"header,omitempty"` // Optional: Header value to pass through every request (e.g. "X-Custom-Header:X-Custom-Key"), maxLength: 5000
	Subscriptions []string `json:"subscriptions"`    // Required: Array of events to subscribe to (see WebhookSubscriptions constants)
}

// GetAllWebhooks retrieves all webhooks
func GetAllWebhooks(client *utils.ProofClient) ([]byte, error) {
	path := WebhooksEndpoint
	return client.Get(path)
}

// CreateWebhook creates a new webhook
func CreateWebhook(client *utils.ProofClient, params *CreateWebhookParams) ([]byte, error) {
	return client.Post(WebhooksEndpoint, params)
}

// GetWebhook retrieves a specific webhook by ID
func GetWebhook(client *utils.ProofClient, webhookID string) ([]byte, error) {
	path := WebhooksEndpoint + "/" + webhookID
	return client.Get(path)
}

// UpdateWebhook updates an existing webhook
func UpdateWebhook(client *utils.ProofClient, webhookID string, params *UpdateWebhookParams) ([]byte, error) {
	path := WebhooksEndpoint + "/" + webhookID
	return client.Put(path, params)
}

// GetWebhookEventsParams represents query parameters for retrieving webhook events
type GetWebhookEventsParams struct {
	Limit  int `json:"limit,omitempty"`  // How many results to return. Default is 20; max is 100
	Offset int `json:"offset,omitempty"` // Offset request by given # of items. Default is 0
}

// DeleteWebhook deletes a webhook by ID
func DeleteWebhook(client *utils.ProofClient, webhookID string) ([]byte, error) {
	path := WebhooksEndpoint + "/" + webhookID
	return client.Delete(path)
}

// GetWebhookEvents retrieves events for a specific webhook
func GetWebhookEvents(client *utils.ProofClient, webhookID string, params *GetWebhookEventsParams) ([]byte, error) {
	queryParams := utils.BuildQueryParams(params)
	path := WebhooksEndpoint + "/" + webhookID + "/events"
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}
	return client.Get(path)
}
