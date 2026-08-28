package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsarlewey/proof-sdk-go/business"
)

// allTransactionFlags is every flag the create command accepts, exercised
// together so the wiring test covers the whole surface at once.
var allTransactionFlags = []string{
	"--email", "signer@example.com",
	"--first-name", "Ada", "--last-name", "Lovelace", "--middle-name", "Byron",
	"--phone-number", "+15555550123",
	"--name", "Series A", "--type", "esign",
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
}

// buildForFlags runs the shared flag set through buildTransactionParams the
// way cobra would, and returns the resulting body.
func buildForFlags(t *testing.T, args ...string) business.TransactionParams {
	t.Helper()
	cmd := &cobra.Command{Use: "tx", Run: func(*cobra.Command, []string) {}}
	registerTransactionParamFlags(cmd)
	cmd.Flags().StringSlice("document-order", nil, "")
	cmd.SetArgs(args)
	require.NoError(t, cmd.Execute())
	return buildTransactionParams(cmd)
}

// Every flag registered on the transaction commands must actually reach the
// request body. An earlier version advertised eleven flags that Run never
// read, so they were silently dropped; this asserts the wiring end to end.
func TestBuildTransactionParams_WiresEveryFlag(t *testing.T) {
	body := buildForFlags(t, allTransactionFlags...)

	require.NotNil(t, body.Signer)
	assert.Equal(t, "signer@example.com", body.Signer.Email)
	assert.Equal(t, "Ada", *body.Signer.FirstName)
	assert.Equal(t, "Lovelace", *body.Signer.LastName)
	assert.Equal(t, "Byron", *body.Signer.MiddleName)
	assert.Equal(t, "+15555550123", *body.Signer.PhoneNumber)

	assert.Equal(t, "Series A", *body.TransactionName)
	assert.Equal(t, "esign", *body.TransactionType)
	assert.Equal(t, "2026-09-01T10:00:00Z", *body.ActivationTime)
	assert.Equal(t, "2026-09-08T10:00:00Z", *body.Expiry)
	assert.Equal(t, "ext-42", *body.ExternalId)
	assert.Equal(t, "cfg-1", *body.ConfigId)
	assert.Equal(t, "org-9", *body.OrganizationId)
	assert.Equal(t, business.TransactionParamsPayer("sender"), *body.Payer)
	assert.True(t, *body.PdfBookmarked)

	assert.Equal(t, business.TransactionParamsAuthenticationRequirement("sms"), *body.AuthenticationRequirement)
	assert.Equal(t, business.TransactionParamsIdvUseCase("ACCOUNT_RECOVERY"), *body.IdvUseCase)
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

// toCreateParams converts between two generated structs by JSON round-trip.
// If a future spec change makes a shared field's type diverge, the value would
// silently vanish from create requests — so assert every key the shared body
// produces survives with an identical value.
func TestBuildTransactionParams_SurvivesCreateConversion(t *testing.T) {
	params := buildForFlags(t, allTransactionFlags...)
	create := toCreateParams(params)

	var before, after map[string]any
	raw, err := json.Marshal(params)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &before))
	raw, err = json.Marshal(create)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &after))

	require.NotEmpty(t, before)
	for key, want := range before {
		assert.Equal(t, want, after[key], "field %q did not survive conversion to TransactionCreateParams", key)
	}
}

// Unpassed optional flags must be absent from the JSON entirely. Sending an
// explicit false or "" would override the organization's own defaults — which
// matters most for patch, where the whole point is to touch nothing else.
func TestBuildTransactionParams_OmitsUnsetFlags(t *testing.T) {
	body := buildForFlags(t)

	assert.Nil(t, body.Signer)
	assert.Nil(t, body.Signers)
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
	assert.Equal(t, "{}", string(raw), "an empty patch must send an empty body")
}

// The primary signer object is only sent when a signer flag was passed, so a
// patch that touches only, say, the expiry doesn't blank out the signer.
func TestBuildTransactionParams_SignerOnlyWhenFlagged(t *testing.T) {
	assert.Nil(t, buildForFlags(t, "--expiry", "2026-09-08T10:00:00Z").Signer)

	body := buildForFlags(t, "--first-name", "Ada")
	require.NotNil(t, body.Signer)
	assert.Equal(t, "Ada", *body.Signer.FirstName)
	assert.Empty(t, body.Signer.Email)
}

func TestParseSignersFile(t *testing.T) {
	write := func(t *testing.T, contents string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "signers.json")
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
		return path
	}

	t.Run("reads an array of signers", func(t *testing.T) {
		path := write(t, `[{"email":"a@example.com","first_name":"Ada"},{"email":"b@example.com"}]`)
		body := buildForFlags(t, "--signers-file", path)

		require.NotNil(t, body.Signers)
		require.Len(t, *body.Signers, 2)
		assert.Equal(t, "a@example.com", *(*body.Signers)[0].Email)
		assert.Equal(t, "Ada", *(*body.Signers)[0].FirstName)
		assert.Equal(t, "b@example.com", *(*body.Signers)[1].Email)
	})

	t.Run("an empty array sends no signers", func(t *testing.T) {
		assert.Nil(t, buildForFlags(t, "--signers-file", write(t, `[]`)).Signers)
	})
}

func TestParseDocumentOrder(t *testing.T) {
	cmd := &cobra.Command{Use: "tx", Run: func(*cobra.Command, []string) {}}
	registerTransactionParamFlags(cmd)
	cmd.Flags().StringSlice("document-order", nil, "")
	cmd.SetArgs([]string{"--document-order", "doc_abc=1", "--document-order", "doc_def=2"})
	require.NoError(t, cmd.Execute())

	order := parseDocumentOrder(cmd)
	require.Len(t, order, 2)
	assert.Equal(t, "doc_abc", *order[0].Id)
	assert.Equal(t, 1, *order[0].BundlePosition)
	assert.Equal(t, "doc_def", *order[1].Id)
	assert.Equal(t, 2, *order[1].BundlePosition)
}

func TestParseRecipientDetailsConfig(t *testing.T) {
	body := buildForFlags(t, "--recipient-details-config", "name=hidden")
	require.Len(t, *body.RecipientDetailsConfig, 1)
	assert.Equal(t, business.RecipientDetailsConfigDisplay("hidden"), *(*body.RecipientDetailsConfig)[0].Display)
}
