package business

import (
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// ReferralsEndpoint is the base endpoint for referrals API
	ReferralsEndpoint = "/v1/referrals"
)

// CreateReferralParams represents the request body for creating a referral campaign
type CreateReferralParams struct {
	Name           string `json:"name"`                      // Required: Name of the new campaign
	CoverPayment   bool   `json:"cover_payment,omitempty"`   // Will the organization pay for these referred transactions? Default: false
	OrganizationID string `json:"organization_id,omitempty"` // ID of organization to create the campaign for (child orgs only), maxLength: 5000
	RedirectURL    string `json:"redirect_url,omitempty"`    // URL that customers will be sent to from the referral, maxLength: 5000
	UseBranding    bool   `json:"use_branding,omitempty"`    // Will the referred transactions display the orgs branding? Default: false
}

// CreateReferral creates a new referral campaign
func CreateReferral(client *utils.ProofClient, params *CreateReferralParams) ([]byte, error) {
	return client.Post(ReferralsEndpoint, params)
}

// GenerateReferralCodeParams represents the request body for generating a referral code
type GenerateReferralCodeParams struct {
	ExpiresAt string `json:"expires_at,omitempty"` // ISO 8601 timestamp for when the code should expire. Default: 3 months from creation, maxLength: 100
}

// GenerateReferralCode generates a single use referral link for a campaign
func GenerateReferralCode(client *utils.ProofClient, referralCampaignID string, params *GenerateReferralCodeParams) ([]byte, error) {
	path := ReferralsEndpoint + "/" + referralCampaignID + "/generate_link"
	return client.Post(path, params)
}
