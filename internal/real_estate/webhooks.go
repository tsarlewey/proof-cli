package real_estate

import (
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// WebhooksEndpoint is the base endpoint for webhooks API
	WebhooksEndpoint = "/mortgage/v2/webhooks"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID            string   `json:"id,omitempty"`
	URL           string   `json:"url"`                     // Proof will make a POST request to this URL for every event, maxLength: 5000
	Header        *string  `json:"header,omitempty"`        // Proof will add this header to every POST request, maxLength: 5000
	Subscriptions []string `json:"subscriptions,omitempty"` // Array of event subscriptions
	CreatedAt     string   `json:"created_at,omitempty"`    // ISO 8601 creation timestamp
	UpdatedAt     string   `json:"updated_at,omitempty"`    // ISO 8601 last update timestamp
}

// CreateWebhookRequest represents a request to create a webhook
type CreateWebhookRequest struct {
	URL           string   `json:"url"`                     // Webhook URL, maxLength: 5000
	Header        *string  `json:"header,omitempty"`        // Optional header, maxLength: 5000
	Subscriptions []string `json:"subscriptions,omitempty"` // Event subscriptions
}

// UpdateWebhookRequest represents a request to update a webhook
type UpdateWebhookRequest struct {
	URL           string   `json:"url,omitempty"`           // Webhook URL, maxLength: 5000
	Header        *string  `json:"header,omitempty"`        // Optional header, maxLength: 5000
	Subscriptions []string `json:"subscriptions,omitempty"` // Event subscriptions
}

// ListWebhooksParams represents query parameters for listing webhooks
type ListWebhooksParams struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// ListWebhooksResponse represents the response from listing webhooks
type ListWebhooksResponse struct {
	Count    int       `json:"count"`
	Next     string    `json:"next,omitempty"`
	Previous string    `json:"previous,omitempty"`
	Results  []Webhook `json:"results"`
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	ID          string `json:"id,omitempty"`
	WebhookID   string `json:"webhook_id,omitempty"`
	EventType   string `json:"event_type,omitempty"`
	Payload     string `json:"payload,omitempty"`
	Status      string `json:"status,omitempty"`
	DeliveredAt string `json:"delivered_at,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

// ListWebhookEventsResponse represents the response from listing webhook events
type ListWebhookEventsResponse struct {
	Count    int            `json:"count"`
	Next     string         `json:"next,omitempty"`
	Previous string         `json:"previous,omitempty"`
	Results  []WebhookEvent `json:"results"`
}

// WebhookSubscription represents available webhook subscriptions
type WebhookSubscription struct {
	EventType   string `json:"event_type"`
	Description string `json:"description,omitempty"`
}

// ListWebhooks retrieves a list of webhooks
func ListWebhooks(client *utils.ProofClient, params *ListWebhooksParams) ([]byte, error) {
	path := WebhooksEndpoint

	// Build query parameters
	queryParams := utils.BuildQueryParams(params)
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	return client.Get(path)
}

// GetWebhook retrieves a specific webhook by ID
func GetWebhook(client *utils.ProofClient, webhookID string) ([]byte, error) {
	path := WebhooksEndpoint + "/" + webhookID
	return client.Get(path)
}

// CreateWebhook creates a new webhook
func CreateWebhook(client *utils.ProofClient, request *CreateWebhookRequest) ([]byte, error) {
	path := WebhooksEndpoint
	return client.Post(path, request)
}

// UpdateWebhook updates an existing webhook
func UpdateWebhook(client *utils.ProofClient, webhookID string, request *UpdateWebhookRequest) ([]byte, error) {
	path := WebhooksEndpoint + "/" + webhookID
	return client.Put(path, request)
}

// DeleteWebhook deletes a webhook
func DeleteWebhook(client *utils.ProofClient, webhookID string) ([]byte, error) {
	path := WebhooksEndpoint + "/" + webhookID
	return client.Delete(path)
}

// ListWebhookEvents retrieves events for a specific webhook
func ListWebhookEvents(client *utils.ProofClient, webhookID string, params *ListWebhooksParams) ([]byte, error) {
	path := WebhooksEndpoint + "/" + webhookID + "/events"

	// Build query parameters
	queryParams := utils.BuildQueryParams(params)
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	return client.Get(path)
}

// ListWebhookSubscriptions retrieves available webhook subscriptions
func ListWebhookSubscriptions(client *utils.ProofClient) ([]byte, error) {
	path := WebhooksEndpoint + "/subscriptions"
	return client.Get(path)
}
