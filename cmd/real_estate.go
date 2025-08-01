package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/internal/real_estate"
)

// realEstateCmd represents the real-estate command
var realEstateCmd = &cobra.Command{
	Use:     "real-estate",
	Aliases: []string{"r", "real"},
	Short:   "Real estate mortgage API operations",
	Long:    `Commands for interacting with the Proof Real Estate/Mortgage API`,
}

// Real Estate Transactions Commands
var reTransactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Real estate transaction operations",
	Long:  `Commands for managing real estate transactions`,
}

var reListTransactionsCmd = &cobra.Command{
	Use:    "list",
	Short:  "List real estate transactions",
	Long:   `List real estate transactions with optional filtering`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		// Get command line flags
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		status, _ := cmd.Flags().GetString("status")
		organizationID, _ := cmd.Flags().GetString("organization-id")
		loanNumber, _ := cmd.Flags().GetString("loan-number")
		documentURLVersion, _ := cmd.Flags().GetString("document-url-version")
		createdDateStart, _ := cmd.Flags().GetString("created-date-start")
		createdDateEnd, _ := cmd.Flags().GetString("created-date-end")
		lastUpdatedDateStart, _ := cmd.Flags().GetString("last-updated-date-start")
		lastUpdatedDateEnd, _ := cmd.Flags().GetString("last-updated-date-end")

		params := &real_estate.ListTransactionsParams{
			Limit:              limit,
			Offset:             offset,
			TransactionStatus:  status,
			OrganizationID:     organizationID,
			LoanNumber:         loanNumber,
			DocumentURLVersion: documentURLVersion,
		}

		// Parse date filters if provided
		if createdDateStart != "" {
			t, err := time.Parse(time.RFC3339, createdDateStart)
			if err != nil {
				fmt.Printf("Error parsing created-date-start: %v\n", err)
				os.Exit(1)
			}
			params.CreatedDateStart = &t
		}

		if createdDateEnd != "" {
			t, err := time.Parse(time.RFC3339, createdDateEnd)
			if err != nil {
				fmt.Printf("Error parsing created-date-end: %v\n", err)
				os.Exit(1)
			}
			params.CreatedDateEnd = &t
		}

		if lastUpdatedDateStart != "" {
			t, err := time.Parse(time.RFC3339, lastUpdatedDateStart)
			if err != nil {
				fmt.Printf("Error parsing last-updated-date-start: %v\n", err)
				os.Exit(1)
			}
			params.LastUpdatedDateStart = &t
		}

		if lastUpdatedDateEnd != "" {
			t, err := time.Parse(time.RFC3339, lastUpdatedDateEnd)
			if err != nil {
				fmt.Printf("Error parsing last-updated-date-end: %v\n", err)
				os.Exit(1)
			}
			params.LastUpdatedDateEnd = &t
		}

		// Make API call using global client
		resp, err := real_estate.ListTransactions(proofClient, params)
		if err != nil {
			fmt.Println("Error listing transactions:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var reGetTransactionCmd = &cobra.Command{
	Use:    "get <transaction-id>",
	Short:  "Get a real estate transaction",
	Long:   `Get details of a specific real estate transaction`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]

		// Make API call using global client
		resp, err := real_estate.GetTransaction(proofClient, transactionID)
		if err != nil {
			fmt.Println("Error getting transaction:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var reCreateTransactionCmd = &cobra.Command{
	Use:    "create",
	Short:  "Create a real estate transaction",
	Long:   `Create a new real estate transaction`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		// Get command line flags for basic transaction creation
		transactionType, _ := cmd.Flags().GetString("type")
		draft, _ := cmd.Flags().GetBool("draft")
		fileNumber, _ := cmd.Flags().GetString("file-number")
		loanNumber, _ := cmd.Flags().GetString("loan-number")

		request := &real_estate.CreateTransactionRequest{
			TransactionType: transactionType,
			Draft:           draft,
			FileNumber:      fileNumber,
			LoanNumber:      loanNumber,
		}

		// Make API call using global client
		resp, err := real_estate.CreateTransaction(proofClient, request)
		if err != nil {
			fmt.Println("Error creating transaction:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var rePlaceOrderCmd = &cobra.Command{
	Use:    "place-order <transaction-id>",
	Short:  "Place order for a real estate transaction",
	Long:   `Place an order for a real estate transaction`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]

		// Make API call using global client
		resp, err := real_estate.PlaceOrder(proofClient, transactionID)
		if err != nil {
			fmt.Println("Error placing order:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

// Real Estate Documents Commands
var reDocumentsCmd = &cobra.Command{
	Use:   "documents",
	Short: "Real estate document operations",
	Long:  `Commands for managing real estate documents`,
}

var reListDocumentsCmd = &cobra.Command{
	Use:    "list",
	Short:  "List real estate documents",
	Long:   `List real estate documents with optional filtering`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		// Get command line flags
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		transactionID, _ := cmd.Flags().GetString("transaction-id")
		documentType, _ := cmd.Flags().GetString("type")
		status, _ := cmd.Flags().GetString("status")

		params := &real_estate.ListDocumentsParams{
			Limit:         limit,
			Offset:        offset,
			TransactionID: transactionID,
			DocumentType:  documentType,
			Status:        status,
		}

		// Make API call using global client
		resp, err := real_estate.ListDocuments(proofClient, params)
		if err != nil {
			fmt.Println("Error listing documents:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var reGetDocumentCmd = &cobra.Command{
	Use:    "get <document-id>",
	Short:  "Get a real estate document",
	Long:   `Get details of a specific real estate document`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		documentID := args[0]

		// Make API call using global client
		resp, err := real_estate.GetDocument(proofClient, documentID)
		if err != nil {
			fmt.Println("Error getting document:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var reUploadDocumentCmd = &cobra.Command{
	Use:    "upload <transaction-id> <file-path>",
	Short:  "Upload a document to a real estate transaction",
	Long:   `Upload a document file to a real estate transaction`,
	Args:   cobra.ExactArgs(2),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]
		filePath := args[1]

		// Get optional flags
		documentType, _ := cmd.Flags().GetString("type")
		externalID, _ := cmd.Flags().GetString("external-id")

		request := &real_estate.AddDocumentRequest{
			DocumentType: documentType,
			ExternalID:   externalID,
		}

		// Make API call using global client
		resp, err := real_estate.AddDocument(proofClient, transactionID, filePath, request)
		if err != nil {
			fmt.Println("Error uploading document:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

// Real Estate Webhooks Commands
var reWebhooksCmd = &cobra.Command{
	Use:   "webhooks",
	Short: "Real estate webhook operations",
	Long:  `Commands for managing real estate webhooks`,
}

var reListWebhooksCmd = &cobra.Command{
	Use:    "list",
	Short:  "List real estate webhooks",
	Long:   `List real estate webhooks`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		// Get command line flags
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		params := &real_estate.ListWebhooksParams{
			Limit:  limit,
			Offset: offset,
		}

		// Make API call using global client
		resp, err := real_estate.ListWebhooks(proofClient, params)
		if err != nil {
			fmt.Println("Error listing webhooks:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var reCreateWebhookCmd = &cobra.Command{
	Use:    "create <url>",
	Short:  "Create a real estate webhook",
	Long:   `Create a new real estate webhook`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		header, _ := cmd.Flags().GetString("header")
		subscriptions, _ := cmd.Flags().GetStringSlice("subscriptions")

		request := &real_estate.CreateWebhookRequest{
			URL:           url,
			Subscriptions: subscriptions,
		}

		if header != "" {
			request.Header = &header
		}

		// Make API call using global client
		resp, err := real_estate.CreateWebhook(proofClient, request)
		if err != nil {
			fmt.Println("Error creating webhook:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

// Utility Commands
var reVerifyAddressCmd = &cobra.Command{
	Use:    "verify-address",
	Short:  "Verify a street address",
	Long:   `Verify a street address and get recording jurisdiction information`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		// Get address components from flags
		line1, _ := cmd.Flags().GetString("line1")
		city, _ := cmd.Flags().GetString("city")
		state, _ := cmd.Flags().GetString("state")
		postalCode, _ := cmd.Flags().GetString("postal-code")

		if line1 == "" || city == "" || state == "" {
			fmt.Println("Error: line1, city, and state are required")
			os.Exit(1)
		}

		request := &real_estate.VerifyAddressRequest{
			StreetAddress: &real_estate.Address{
				Line1:      line1,
				City:       city,
				State:      state,
				PostalCode: postalCode,
			},
		}

		// Make API call using global client
		resp, err := real_estate.VerifyAddress(proofClient, request)
		if err != nil {
			fmt.Println("Error verifying address:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

func init() {
	rootCmd.AddCommand(realEstateCmd)

	// Add subcommands
	realEstateCmd.AddCommand(reTransactionsCmd)
	realEstateCmd.AddCommand(reDocumentsCmd)
	realEstateCmd.AddCommand(reWebhooksCmd)
	realEstateCmd.AddCommand(reVerifyAddressCmd)

	// Transaction subcommands
	reTransactionsCmd.AddCommand(reListTransactionsCmd)
	reTransactionsCmd.AddCommand(reGetTransactionCmd)
	reTransactionsCmd.AddCommand(reCreateTransactionCmd)
	reTransactionsCmd.AddCommand(rePlaceOrderCmd)

	// Document subcommands
	reDocumentsCmd.AddCommand(reListDocumentsCmd)
	reDocumentsCmd.AddCommand(reGetDocumentCmd)
	reDocumentsCmd.AddCommand(reUploadDocumentCmd)

	// Webhook subcommands
	reWebhooksCmd.AddCommand(reListWebhooksCmd)
	reWebhooksCmd.AddCommand(reCreateWebhookCmd)

	// Add flags for transactions
	reListTransactionsCmd.Flags().Int("limit", 0, "Limit number of results")
	reListTransactionsCmd.Flags().Int("offset", 0, "Offset for pagination")
	reListTransactionsCmd.Flags().String("status", "", "Filter by transaction status")
	reListTransactionsCmd.Flags().String("organization-id", "", "Organization ID of child account")
	reListTransactionsCmd.Flags().String("loan-number", "", "Find transactions associated with loan number")
	reListTransactionsCmd.Flags().String("created-date-start", "", "ISO-8601 DateTime - transactions created after this time")
	reListTransactionsCmd.Flags().String("created-date-end", "", "ISO-8601 DateTime - transactions created before this time")
	reListTransactionsCmd.Flags().String("last-updated-date-start", "", "ISO-8601 DateTime - transactions updated after this time")
	reListTransactionsCmd.Flags().String("last-updated-date-end", "", "ISO-8601 DateTime - transactions updated before this time")
	reListTransactionsCmd.Flags().String("document-url-version", "v1", "Document URL version (v1 or v2)")

	reCreateTransactionCmd.Flags().String("type", "purchase", "Transaction type")
	reCreateTransactionCmd.Flags().Bool("draft", true, "Create as draft")
	reCreateTransactionCmd.Flags().String("file-number", "", "File number")
	reCreateTransactionCmd.Flags().String("loan-number", "", "Loan number")

	// Add flags for documents
	reListDocumentsCmd.Flags().Int("limit", 0, "Limit number of results")
	reListDocumentsCmd.Flags().Int("offset", 0, "Offset for pagination")
	reListDocumentsCmd.Flags().String("transaction-id", "", "Filter by transaction ID")
	reListDocumentsCmd.Flags().String("type", "", "Filter by document type")
	reListDocumentsCmd.Flags().String("status", "", "Filter by document status")

	reUploadDocumentCmd.Flags().String("type", "", "Document type")
	reUploadDocumentCmd.Flags().String("external-id", "", "External ID for the document")

	// Add flags for webhooks
	reListWebhooksCmd.Flags().Int("limit", 0, "Limit number of results")
	reListWebhooksCmd.Flags().Int("offset", 0, "Offset for pagination")

	reCreateWebhookCmd.Flags().String("header", "", "Custom header to include in webhook requests")
	reCreateWebhookCmd.Flags().StringSlice("subscriptions", []string{"*"}, "Webhook event subscriptions")

	// Add flags for address verification
	reVerifyAddressCmd.Flags().String("line1", "", "Street address line 1 (required)")
	reVerifyAddressCmd.Flags().String("city", "", "City (required)")
	reVerifyAddressCmd.Flags().String("state", "", "State abbreviation (required)")
	reVerifyAddressCmd.Flags().String("postal-code", "", "Postal/ZIP code")
	reVerifyAddressCmd.MarkFlagRequired("line1")
	reVerifyAddressCmd.MarkFlagRequired("city")
	reVerifyAddressCmd.MarkFlagRequired("state")
}
