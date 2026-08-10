package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/pkg/utils"
	"github.com/tsarlewey/proof-sdk-go/certificates"
)

// certificatesCmd represents the certificates command
var certificatesCmd = &cobra.Command{
	Use:     "certificates",
	Aliases: []string{"certs"},
	Short:   "Organization certificate operations",
	Long:    `Commands for issuing, listing, signing with, and revoking organization certificates`,
}

// certCreateBody builds the JSON body for the certificate create endpoints.
// The upstream spec names these properties "common_name *required" and
// "csr *required", so the generated structs carry those literal keys and
// cannot be used — see the Known Issues section of CLAUDE.md.
func certCreateBody(profile, field, value string) ([]byte, error) {
	body := map[string]string{field: value}
	if profile != "" {
		body["certificate_profile"] = profile
	}
	return json.Marshal(body)
}

var certListCmd = &cobra.Command{
	Use:    "list",
	Short:  "List certificates",
	Long:   `List the organization's certificates with optional pagination`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		// The spec types both pagination params as strings.
		params := &certificates.GetV1CertificatesParams{}
		if limit > 0 {
			s := strconv.Itoa(limit)
			params.Limit = &s
		}
		if offset > 0 {
			s := strconv.Itoa(offset)
			params.Offset = &s
		}

		client := getCertificatesClient()
		resp, err := client.GetV1CertificatesWithResponse(context.Background(), params)
		utils.HandleError(err, "listing certificates")
		checkAPIStatus(resp.StatusCode(), resp.Body, "listing certificates")

		PrintResponse(resp.Body)
	},
}

var certGetCmd = &cobra.Command{
	Use:    "get <certificate-id>",
	Short:  "Get a certificate",
	Long:   `Get details of a specific certificate`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		client := getCertificatesClient()
		resp, err := client.GetV1CertificatesIdWithResponse(context.Background(), args[0])
		utils.HandleError(err, "getting certificate")
		checkAPIStatus(resp.StatusCode(), resp.Body, "getting certificate")

		PrintResponse(resp.Body)
	},
}

var certCreateCmd = &cobra.Command{
	Use:    "create",
	Short:  "Create a certificate (Proof-generated key)",
	Long:   `Create a certificate where Proof generates the key pair, identified by a common name`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		commonName, _ := cmd.Flags().GetString("common-name")
		profile, _ := cmd.Flags().GetString("profile")

		body, err := certCreateBody(profile, "common_name", commonName)
		utils.HandleError(err, "building certificate request")

		client := getCertificatesClient()
		resp, err := client.PostV1CertificatesWithBodyWithResponse(
			context.Background(), "application/json", bytes.NewReader(body))
		utils.HandleError(err, "creating certificate")
		checkAPIStatus(resp.StatusCode(), resp.Body, "creating certificate")

		PrintResponse(resp.Body)
	},
}

var certCreateFromCSRCmd = &cobra.Command{
	Use:    "create-from-csr",
	Short:  "Create a certificate from a CSR (bring your own key)",
	Long:   `Create a certificate from a PEM-encoded Certificate Signing Request you generated yourself`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		csr, _ := cmd.Flags().GetString("csr")
		profile, _ := cmd.Flags().GetString("profile")

		body, err := certCreateBody(profile, "csr", csr)
		utils.HandleError(err, "building certificate request")

		client := getCertificatesClient()
		resp, err := client.PostV2CertificatesWithBodyWithResponse(
			context.Background(), "application/json", bytes.NewReader(body))
		utils.HandleError(err, "creating certificate from CSR")
		checkAPIStatus(resp.StatusCode(), resp.Body, "creating certificate from CSR")

		PrintResponse(resp.Body)
	},
}

var certSignCmd = &cobra.Command{
	Use:    "sign <certificate-id>",
	Short:  "Sign digests with a certificate",
	Long:   `Sign up to 25 base64-encoded SHA256 digests with the given certificate`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		digests, _ := cmd.Flags().GetStringSlice("digest")
		algorithmOID, _ := cmd.Flags().GetString("algorithm-oid")

		oid := certificates.PostV1CertificatesIdSignJSONBodyAlgorithmOid(algorithmOID)
		body := certificates.PostV1CertificatesIdSignJSONRequestBody{
			Digests:      &digests,
			AlgorithmOid: &oid,
		}

		client := getCertificatesClient()
		resp, err := client.PostV1CertificatesIdSignWithResponse(context.Background(), args[0], body)
		utils.HandleError(err, "signing digests")
		checkAPIStatus(resp.StatusCode(), resp.Body, "signing digests")

		PrintResponse(resp.Body)
	},
}

var certRevokeCmd = &cobra.Command{
	Use:    "revoke <certificate-id>",
	Short:  "Revoke a certificate",
	Long:   `Revoke a certificate, optionally recording a reason`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		reason, _ := cmd.Flags().GetString("reason")

		params := &certificates.DeleteV1CertificatesIdParams{}
		if reason != "" {
			params.Reason = &reason
		}

		client := getCertificatesClient()
		resp, err := client.DeleteV1CertificatesIdWithResponse(context.Background(), args[0], params)
		utils.HandleError(err, "revoking certificate")
		checkAPIStatus(resp.StatusCode(), resp.Body, "revoking certificate")

		if isSuccess(resp.StatusCode()) {
			fmt.Println("Certificate revoked successfully")
		}
		PrintVerbose(string(resp.Body))
	},
}

func init() {
	rootCmd.AddCommand(certificatesCmd)

	certificatesCmd.AddCommand(certListCmd)
	certificatesCmd.AddCommand(certGetCmd)
	certificatesCmd.AddCommand(certCreateCmd)
	certificatesCmd.AddCommand(certCreateFromCSRCmd)
	certificatesCmd.AddCommand(certSignCmd)
	certificatesCmd.AddCommand(certRevokeCmd)

	certListCmd.Flags().Int("limit", 0, "Page size")
	certListCmd.Flags().Int("offset", 0, "0-based offset")

	const profileHelp = "Certificate profile: organization_authenticity_al1 through _al4"
	certCreateCmd.Flags().String("common-name", "", "CN field of the subject - required")
	certCreateCmd.Flags().String("profile", "", profileHelp)
	certCreateCmd.MarkFlagRequired("common-name")

	certCreateFromCSRCmd.Flags().String("csr", "", "PEM-encoded Certificate Signing Request - required")
	certCreateFromCSRCmd.Flags().String("profile", "", profileHelp)
	certCreateFromCSRCmd.MarkFlagRequired("csr")

	certSignCmd.Flags().StringSlice("digest", []string{}, "Base64-encoded SHA256 digest (repeatable, max 25) - required")
	certSignCmd.Flags().String("algorithm-oid", "1.2.840.10045.4.3.2", "Signing algorithm OID (only ECDSA with SHA256 is supported)")
	certSignCmd.MarkFlagRequired("digest")

	certRevokeCmd.Flags().String("reason", "", "Reason for revocation")
}
