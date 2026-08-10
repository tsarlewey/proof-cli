package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/pkg/utils"
	"github.com/tsarlewey/proof-sdk-go/credentials"
)

// credentialsCmd represents the credentials command
var credentialsCmd = &cobra.Command{
	Use:     "credentials",
	Aliases: []string{"vc"},
	Short:   "Verifiable Credentials operations",
	Long:    `Commands for the Proof Verifiable Credentials presentation flow`,
}

// buildAuthorizeParams validates the response-mode/URI pairing and assembles
// the authorize request params. The API requires redirect_uri with fragment
// mode and response_uri with direct_post mode, and rejects the other one.
func buildAuthorizeParams(cmd *cobra.Command) (*credentials.AuthorizeVerifiableCredentialPresentationParams, error) {
	clientID, _ := cmd.Flags().GetString("client-id")
	responseMode, _ := cmd.Flags().GetString("response-mode")
	redirectURI, _ := cmd.Flags().GetString("redirect-uri")
	responseURI, _ := cmd.Flags().GetString("response-uri")
	scope, _ := cmd.Flags().GetString("scope")
	loginHint, _ := cmd.Flags().GetString("login-hint")
	nonce, _ := cmd.Flags().GetString("nonce")
	state, _ := cmd.Flags().GetString("state")

	params := &credentials.AuthorizeVerifiableCredentialPresentationParams{
		ClientId:     clientID,
		ResponseType: "vp_token",
		ResponseMode: credentials.AuthorizeVerifiableCredentialPresentationParamsResponseMode(responseMode),
		Scope:        credentials.AuthorizeVerifiableCredentialPresentationParamsScope(scope),
		LoginHint:    loginHint,
		Nonce:        nonce,
	}
	if state != "" {
		params.State = &state
	}

	switch responseMode {
	case "fragment":
		if redirectURI == "" {
			return nil, fmt.Errorf("--redirect-uri is required when --response-mode is fragment")
		}
		if responseURI != "" {
			return nil, fmt.Errorf("--response-uri is not permitted when --response-mode is fragment")
		}
		params.RedirectUri = &redirectURI
	case "direct_post":
		if responseURI == "" {
			return nil, fmt.Errorf("--response-uri is required when --response-mode is direct_post")
		}
		if redirectURI != "" {
			return nil, fmt.Errorf("--redirect-uri is not permitted when --response-mode is direct_post")
		}
		params.ResponseUri = &responseURI
	default:
		return nil, fmt.Errorf("invalid --response-mode %q (expected fragment or direct_post)", responseMode)
	}

	return params, nil
}

var credentialsAuthorizeURLCmd = &cobra.Command{
	Use:   "authorize-url",
	Short: "Build the Verifiable Credential presentation authorize URL",
	Long: `Build the URL that requests a Verifiable Credential presentation from an End-User.

This endpoint is a browser redirect, not a server-to-server API call — the CLI
prints the URL rather than following it. Send the End-User to the printed URL;
Proof returns the vp_token to your redirect_uri or response_uri.`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		params, err := buildAuthorizeParams(cmd)
		utils.HandleError(err, "building authorize URL")

		req, err := credentials.NewAuthorizeVerifiableCredentialPresentationRequest(
			proofClient.GetConfig().APIEndpoint, params)
		utils.HandleError(err, "building authorize URL")

		fmt.Println(req.URL.String())
	},
}

func init() {
	rootCmd.AddCommand(credentialsCmd)

	credentialsCmd.AddCommand(credentialsAuthorizeURLCmd)

	registerAuthorizeFlags(credentialsAuthorizeURLCmd)
	credentialsAuthorizeURLCmd.MarkFlagRequired("client-id")
	credentialsAuthorizeURLCmd.MarkFlagRequired("scope")
	credentialsAuthorizeURLCmd.MarkFlagRequired("login-hint")
	credentialsAuthorizeURLCmd.MarkFlagRequired("nonce")
}

// registerAuthorizeFlags declares the authorize-url flag set. Split out so
// tests can build the same flags without going through the root command.
func registerAuthorizeFlags(cmd *cobra.Command) {
	f := cmd.Flags()
	f.String("client-id", "", "Your Proof OAuth Application Client ID - required")
	f.String("response-mode", "fragment", "How the presentation is returned: fragment or direct_post")
	f.String("redirect-uri", "", "Redirect URI receiving the vp_token fragment (required for fragment mode)")
	f.String("response-uri", "", "URI Proof POSTs the vp_token to (required for direct_post mode)")
	f.String("scope", "", "Space-separated scopes that translate to DCQL queries - required")
	f.String("login-hint", "", "Email address of the End-User to request a presentation from - required")
	f.String("nonce", "", "Opaque client-generated value bound to the response to prevent replay - required")
	f.String("state", "", "Opaque value echoed back on the callback")
}
