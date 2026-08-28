package cmd

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsarlewey/proof-sdk-go/business"
)

// applyForFlags runs the create-transaction flag set through
// applyTransactionParams the way cobra would, and returns the resulting body.
func applyForFlags(t *testing.T, args ...string) business.TransactionCreateParams {
	t.Helper()
	cmd := &cobra.Command{Use: "create", Run: func(*cobra.Command, []string) {}}
	registerTransactionParamFlags(cmd)
	cmd.SetArgs(args)
	cmd.SetOut(nil)
	require.NoError(t, cmd.Execute())

	body := business.TransactionCreateParams{Signer: business.Signer{Email: "signer@example.com"}}
	applyTransactionParams(cmd, &body)
	return body
}

// Every flag registered on create-transaction must actually reach the request
// body. An earlier version advertised eleven flags that Run never read, so
// they were silently dropped; this asserts the wiring end to end.
func TestApplyTransactionParams_WiresEveryFlag(t *testing.T) {
	body := applyForFlags(t,
		"--first-name", "Ada", "--last-name", "Lovelace", "--middle-name", "Byron",
		"--phone-number", "+15555550123",
		"--name", "Series A", "--type", "esign", "--draft",
		"--activation-time", "2026-09-01T10:00:00Z", "--expiry", "2026-09-08T10:00:00Z",
		"--external-id", "ext-42", "--config-id", "cfg-1", "--organization-id", "org-9",
		"--payer", "sender", "--pdf-bookmarked",
		"--auth-requirement", "sms", "--idv-use-case", "ACCOUNT_RECOVERY",
		"--require-secondary-photo-id", "--require-new-signer-verification",
		"--notary-id", "notary-7", "--notary-meeting-time", "2026-09-02T15:30:00Z",
		"--notary-note", "Check the seal", "--notary-note", "Verify address",
		"--allowed-notary-states", "CA,NY",
		"--suppress-email", "--message-subject", "Please sign",
		"--message-to-signer", "Hello", "--message-signature", "— Acme",
		"--cc-recipient-emails", "ops@example.com,legal@example.com",
		"--redirect-url", "https://example.com/done", "--redirect-message", "Redirecting",
		"--recipient-details-config", "name=locked",
		"--cosigner-first-name", "Grace", "--cosigner-last-name", "Hopper",
		"--cosigner-signing-requirement", "verify",
	)

	assert.Equal(t, "Ada", *body.Signer.FirstName)
	assert.Equal(t, "Lovelace", *body.Signer.LastName)
	assert.Equal(t, "Byron", *body.Signer.MiddleName)
	assert.Equal(t, "+15555550123", *body.Signer.PhoneNumber)

	assert.Equal(t, "Series A", *body.TransactionName)
	assert.Equal(t, "esign", *body.TransactionType)
	assert.True(t, *body.Draft)
	assert.Equal(t, "2026-09-01T10:00:00Z", *body.ActivationTime)
	assert.Equal(t, "2026-09-08T10:00:00Z", *body.Expiry)
	assert.Equal(t, "ext-42", *body.ExternalId)
	assert.Equal(t, "cfg-1", *body.ConfigId)
	assert.Equal(t, "org-9", *body.OrganizationId)
	assert.Equal(t, business.TransactionCreateParamsPayer("sender"), *body.Payer)
	assert.True(t, *body.PdfBookmarked)

	assert.Equal(t, business.TransactionCreateParamsAuthenticationRequirement("sms"), *body.AuthenticationRequirement)
	assert.Equal(t, business.TransactionCreateParamsIdvUseCase("ACCOUNT_RECOVERY"), *body.IdvUseCase)
	assert.True(t, *body.RequireSecondaryPhotoId)
	assert.True(t, *body.RequireNewSignerVerification)

	assert.Equal(t, "notary-7", *body.NotaryId)
	assert.Equal(t, "2026-09-02T15:30:00Z", body.NotaryMeetingTime.Format("2006-01-02T15:04:05Z"))
	require.Len(t, *body.NotaryInstructions, 2)
	assert.Equal(t, "Check the seal", *(*body.NotaryInstructions)[0].NotaryNote)
	assert.Equal(t, "Verify address", *(*body.NotaryInstructions)[1].NotaryNote)
	assert.Equal(t, []string{"CA", "NY"}, *body.AllowedNotaryStates)

	assert.True(t, *body.SuppressEmail)
	assert.Equal(t, "Please sign", *body.MessageSubject)
	assert.Equal(t, "Hello", *body.MessageToSigner)
	assert.Equal(t, "— Acme", *body.MessageSignature)
	assert.Equal(t, []string{"ops@example.com", "legal@example.com"}, *body.CcRecipientEmails)

	assert.Equal(t, "https://example.com/done", *body.Redirect.Url)
	assert.Equal(t, "Redirecting", *body.Redirect.Message)

	require.Len(t, *body.RecipientDetailsConfig, 1)
	assert.Equal(t, business.RecipientDetailsConfigField("name"), *(*body.RecipientDetailsConfig)[0].Field)
	assert.Equal(t, business.RecipientDetailsConfigDisplay("locked"), *(*body.RecipientDetailsConfig)[0].Display)

	assert.Equal(t, "Grace", *body.Cosigner.FirstName)
	assert.Equal(t, "Hopper", *body.Cosigner.LastName)
	assert.Equal(t, business.CosignerSigningRequirement("verify"), *body.Cosigner.SigningRequirement)
}

// Unpassed optional flags must be absent from the JSON entirely. Sending an
// explicit false or "" would override the organization's own defaults.
func TestApplyTransactionParams_OmitsUnsetFlags(t *testing.T) {
	body := applyForFlags(t)

	assert.Nil(t, body.SuppressEmail)
	assert.Nil(t, body.PdfBookmarked)
	assert.Nil(t, body.RequireSecondaryPhotoId)
	assert.Nil(t, body.RequireNewSignerVerification)
	assert.Nil(t, body.Payer)
	assert.Nil(t, body.Redirect)
	assert.Nil(t, body.Cosigner)
	assert.Nil(t, body.NotaryInstructions)
	assert.Nil(t, body.RecipientDetailsConfig)
	assert.Nil(t, body.NotaryMeetingTime)

	raw, err := json.Marshal(body)
	require.NoError(t, err)
	for _, key := range []string{
		"suppress_email", "pdf_bookmarked", "require_secondary_photo_id",
		"require_new_signer_verification", "payer", "redirect", "cosigner",
		"notary_instructions", "recipient_details_config", "notary_meeting_time",
		"authentication_requirement", "idv_use_case", "allowed_notary_states",
	} {
		assert.NotContains(t, string(raw), key)
	}

	// draft is the deliberate exception: always sent.
	require.NotNil(t, body.Draft)
	assert.False(t, *body.Draft)
	assert.Contains(t, string(raw), "draft")
}

func TestParseRecipientDetailsConfig(t *testing.T) {
	body := applyForFlags(t, "--recipient-details-config", "name=hidden")
	require.Len(t, *body.RecipientDetailsConfig, 1)
	assert.Equal(t, business.RecipientDetailsConfigDisplay("hidden"), *(*body.RecipientDetailsConfig)[0].Display)
}
