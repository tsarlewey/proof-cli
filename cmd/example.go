package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/pkg/sdk/business"
	"github.com/tsarlewey/proof-cli/pkg/sdk/realestate"
	"github.com/tsarlewey/proof-cli/pkg/sdk/scim"
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
		limit := 10
		docUrlVersion := business.GetAllTransactionsParamsDocumentUrlVersionV2
		params := &business.GetAllTransactionsParams{
			Limit:              &limit,
			DocumentUrlVersion: &docUrlVersion,
		}

		// Make API call using SDK client
		fmt.Println("Fetching business transactions...")
		client := getBusinessClient()
		resp, err := client.GetAllTransactionsWithResponse(context.Background(), params)
		if err != nil {
			fmt.Println("Error fetching transactions:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp.Body)
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
		body := business.CreateNotaryJSONRequestBody{
			Email:       "notary@example.com",
			FirstName:   "Jane",
			LastName:    "Smith",
			UsStateAbbr: "CA",
		}

		// Make API call using SDK client
		fmt.Println("Creating notary...")
		client := getBusinessClient()
		resp, err := client.CreateNotaryWithResponse(context.Background(), body)
		if err != nil {
			fmt.Println("Error creating notary:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp.Body)
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
		limit := 5
		offset := 0
		params := &realestate.GetAllMortgageTransactionsParams{
			Limit:  &limit,
			Offset: &offset,
		}

		// Make API call using SDK client
		fmt.Println("Fetching real estate transactions...")
		client := getRealEstateClient()
		resp, err := client.GetAllMortgageTransactionsWithResponse(context.Background(), params)
		if err != nil {
			fmt.Println("Error fetching real estate transactions:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp.Body)
	},
}

// exampleVerifyAddressCmd demonstrates address verification
var exampleVerifyAddressCmd = &cobra.Command{
	Use:    "verify-address",
	Short:  "Verify a street address",
	Long:   `Example command that demonstrates how to verify an address and get recording jurisdiction info.`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		// Build params for recording locations
		txnType := realestate.GetRecordingLocationsParamsTransactionTypeRefinance
		line1 := "123 Main St"
		city := "San Francisco"
		state := "CA"
		postal := "94102"

		params := &realestate.GetRecordingLocationsParams{
			TransactionType:     txnType,
			StreetAddressLine1:  &line1,
			StreetAddressCity:   &city,
			StreetAddressState:  &state,
			StreetAddressPostal: &postal,
		}

		// Make API call using SDK client
		fmt.Println("Verifying address...")
		client := getRealEstateClient()
		resp, err := client.GetRecordingLocationsWithResponse(context.Background(), params)
		if err != nil {
			fmt.Println("Error verifying address:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp.Body)
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
		startIndex := int32(1)
		count := int32(10)
		params := &scim.RetrieveResourceTypesCopyParams{
			StartIndex: &startIndex,
			Count:      &count,
		}

		// Make API call using SDK client
		fmt.Printf("Fetching SCIM users for organization %s...\n", organizationID)
		client := getSCIMClient()
		resp, err := client.RetrieveResourceTypesCopyWithResponse(context.Background(), organizationID, params)
		if err != nil {
			fmt.Println("Error listing users:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp.Body)
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

		// Make API call using SDK client
		fmt.Printf("Fetching SCIM user schema for organization %s...\n", organizationID)
		client := getSCIMClient()
		resp, err := client.RetrieveUsersSchemaWithResponse(context.Background(), organizationID)
		if err != nil {
			fmt.Println("Error getting user schema:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp.Body)
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
