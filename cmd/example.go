package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/internal/business"
	"github.com/tsarlewey/proof-cli/internal/real_estate"
	"github.com/tsarlewey/proof-cli/internal/scim"
)

// exampleCmd represents the example command group
var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Example commands demonstrating API usage",
	Long:  `Example commands that demonstrate how to use the Proof CLI to interact with various API endpoints across Business, Real Estate, and SCIM APIs.`,
}

// Business API Examples

// exampleListBusinessTransactionsCmd demonstrates listing business transactions
var exampleListBusinessTransactionsCmd = &cobra.Command{
	Use:    "business-transactions",
	Short:  "List business transactions",
	Long:   `Example command that demonstrates how to list business transactions with filtering.`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		// Build query parameters
		params := &business.ListTransactionsParams{
			Limit:              10,
			DocumentURLVersion: business.DocumentURLVersion,
		}

		// Make API call using global client
		fmt.Println("Fetching business transactions...")
		resp, err := business.GetAllTransactions(proofClient, params)
		if err != nil {
			fmt.Println("Error fetching transactions:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

// exampleCreateNotaryCmd demonstrates creating a notary
var exampleCreateNotaryCmd = &cobra.Command{
	Use:   "business-notary",
	Short: "Create a business notary",
	Long: `Example command that demonstrates how to create a new notary.
This example creates a notary for demonstration purposes.`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		// Create the request body
		requestBody := map[string]string{
			"email":         "notary@example.com",
			"first_name":    "Jane",
			"last_name":     "Smith",
			"us_state_abbr": "CA",
		}

		// Make API call using global client
		fmt.Println("Creating notary...")
		body, err := proofClient.Post("/v1/notaries/", requestBody)
		if err != nil {
			fmt.Println("Error creating notary:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(body)
	},
}

// Real Estate API Examples

// exampleListRealEstateTransactionsCmd demonstrates listing real estate transactions
var exampleListRealEstateTransactionsCmd = &cobra.Command{
	Use:    "real-estate-transactions",
	Short:  "List real estate transactions",
	Long:   `Example command that demonstrates how to list real estate transactions.`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		// Build query parameters
		params := &real_estate.ListTransactionsParams{
			Limit:  5,
			Offset: 0,
		}

		// Make API call using global client
		fmt.Println("Fetching real estate transactions...")
		resp, err := real_estate.ListTransactions(proofClient, params)
		if err != nil {
			fmt.Println("Error fetching real estate transactions:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

// exampleVerifyAddressCmd demonstrates address verification
var exampleVerifyAddressCmd = &cobra.Command{
	Use:    "verify-address",
	Short:  "Verify a street address",
	Long:   `Example command that demonstrates how to verify an address and get recording jurisdiction info.`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		request := &real_estate.VerifyAddressRequest{
			StreetAddress: &real_estate.Address{
				Line1:      "123 Main St",
				City:       "San Francisco",
				State:      "CA",
				PostalCode: "94102",
			},
		}

		// Make API call using global client
		fmt.Println("Verifying address...")
		resp, err := real_estate.VerifyAddress(proofClient, request)
		if err != nil {
			fmt.Println("Error verifying address:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

// SCIM API Examples

// exampleListSCIMUsersCmd demonstrates listing SCIM users
var exampleListSCIMUsersCmd = &cobra.Command{
	Use:    "scim-users <organization-id>",
	Short:  "List SCIM users",
	Long:   `Example command that demonstrates how to list users via SCIM API.`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		organizationID := args[0]

		// Build query parameters
		params := &scim.ListUsersParams{
			StartIndex: 1,
			Count:      10,
		}

		// Make API call using global client
		fmt.Printf("Fetching SCIM users for organization %s...\n", organizationID)
		resp, err := scim.ListUsers(proofClient, organizationID, params)
		if err != nil {
			fmt.Println("Error listing users:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

// exampleGetSCIMSchemaCmd demonstrates getting SCIM user schema
var exampleGetSCIMSchemaCmd = &cobra.Command{
	Use:    "scim-schema <organization-id>",
	Short:  "Get SCIM user schema",
	Long:   `Example command that demonstrates how to get the SCIM user schema.`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		organizationID := args[0]

		// Make API call using global client
		fmt.Printf("Fetching SCIM user schema for organization %s...\n", organizationID)
		resp, err := scim.GetUserSchema(proofClient, organizationID)
		if err != nil {
			fmt.Println("Error getting user schema:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

func init() {
	rootCmd.AddCommand(exampleCmd)

	// Add Business API examples
	exampleCmd.AddCommand(exampleListBusinessTransactionsCmd)
	exampleCmd.AddCommand(exampleCreateNotaryCmd)

	// Add Real Estate API examples
	exampleCmd.AddCommand(exampleListRealEstateTransactionsCmd)
	exampleCmd.AddCommand(exampleVerifyAddressCmd)

	// Add SCIM API examples
	exampleCmd.AddCommand(exampleListSCIMUsersCmd)
	exampleCmd.AddCommand(exampleGetSCIMSchemaCmd)
}
