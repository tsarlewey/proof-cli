package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/internal/business"
)

// businessCmd represents the business command
var businessCmd = &cobra.Command{
	Use:     "business",
	Aliases: []string{"biz", "b"},
	Short:   "Business API operations",
	Long:    `Commands for interacting with the Proof Business API`,
}

// Business Transactions Commands
var bizTransactionsCmd = &cobra.Command{
	Use:     "transactions",
	Aliases: []string{"txn", "tx", "trans"},
	Short:   "Business transaction operations",
	Long:    `Commands for managing business transactions`,
}

var bizListTransactionsCmd = &cobra.Command{
	Use:        "list",
	Aliases:    []string{"ls", "show"},
	SuggestFor: []string{"lst", "lsit", "lists"},
	Short:      "List business transactions",
	Long:       `List all transactions for your organization`,
	PreRun:     initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		status, _ := cmd.Flags().GetString("status")
		dateStart, _ := cmd.Flags().GetString("created-start")
		dateEnd, _ := cmd.Flags().GetString("created-end")

		// Build query parameters
		params := &business.ListTransactionsParams{
			Limit:              limit,
			Offset:             offset,
			TransactionStatus:  status,
			DocumentURLVersion: business.DocumentURLVersion,
		}

		// Parse date filters if provided
		if dateStart != "" {
			t, err := time.Parse("2006-01-02", dateStart)
			if err != nil {
				fmt.Printf("Error parsing created-start date: %v\n", err)
				os.Exit(1)
			}
			params.CreatedDateStart = &t
		}

		if dateEnd != "" {
			t, err := time.Parse("2006-01-02", dateEnd)
			if err != nil {
				fmt.Printf("Error parsing created-end date: %v\n", err)
				os.Exit(1)
			}
			params.CreatedDateEnd = &t
		}

		// Make API call using global client
		resp, err := business.GetAllTransactions(proofClient, params)
		if err != nil {
			fmt.Println("Error fetching transactions:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var bizGetTransactionCmd = &cobra.Command{
	Use:     "get <transaction-id>",
	Aliases: []string{"g"},
	Short:   "Get a business transaction",
	Long:    `Get details of a specific transaction`,
	Args:    cobra.ExactArgs(1),
	PreRun:  initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]

		// Make API call using global client
		resp, err := business.GetTransaction(proofClient, transactionID)
		if err != nil {
			fmt.Println("Error fetching transaction:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var bizCreateTransactionCmd = &cobra.Command{
	Use:        "create",
	Aliases:    []string{"c"},
	SuggestFor: []string{"creat", "craete", "make"},
	Short:      "Create a business transaction",
	Long:       `Create a new transaction with signers and documents`,
	PreRun:     initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")
		documentPath, _ := cmd.Flags().GetString("document")
		transactionName, _ := cmd.Flags().GetString("name")
		draft, _ := cmd.Flags().GetBool("draft")
		transactionType, _ := cmd.Flags().GetString("type")

		if email == "" || documentPath == "" {
			fmt.Println("Error: email and document are required")
			os.Exit(1)
		}

		// Read the document file
		documentData, err := os.ReadFile(documentPath)
		if err != nil {
			fmt.Println("Error reading document file:", err)
			os.Exit(1)
		}

		// Encode document to base64
		documentBase64 := base64.StdEncoding.EncodeToString(documentData)

		// Build transaction parameters
		params := &business.CreateTransactionParams{
			TransactionName: transactionName,
			TransactionType: transactionType,
			Draft:           draft,
			Documents:       []string{documentBase64},
			Signers: []business.Signer{
				{
					Email:     email,
					FirstName: firstName,
					LastName:  lastName,
				},
			},
		}

		// Make API call using global client
		resp, err := business.CreateTransaction(proofClient, params)
		if err != nil {
			fmt.Println("Error creating transaction:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var bizDeleteTransactionCmd = &cobra.Command{
	Use:    "delete <transaction-id>",
	Short:  "Delete a business transaction",
	Long:   `Delete a specific transaction`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]

		// Make API call using global client
		resp, err := business.DeleteTransaction(proofClient, transactionID)
		if err != nil {
			fmt.Println("Error deleting transaction:", err)
			os.Exit(1)
		}

		fmt.Println("Transaction deleted successfully")

		// Show response details if verbose
		PrintVerbose(string(resp))
	},
}

var bizActivateTransactionCmd = &cobra.Command{
	Use:    "activate <transaction-id>",
	Short:  "Activate a draft transaction",
	Long:   `Activate a draft transaction to send it to the signer`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]

		// Make API call using global client
		resp, err := business.ActivateDraftTransaction(proofClient, transactionID)
		if err != nil {
			fmt.Println("Error activating transaction:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var bizRecallTransactionCmd = &cobra.Command{
	Use:    "recall <transaction-id>",
	Short:  "Recall a transaction",
	Long:   `Recall a transaction with an optional reason`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]
		recallReason, _ := cmd.Flags().GetString("reason")

		// Make API call using global client
		resp, err := business.RecallTransaction(proofClient, transactionID, recallReason)
		if err != nil {
			fmt.Println("Error recalling transaction:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var bizResendEmailCmd = &cobra.Command{
	Use:    "resend-email <transaction-id>",
	Short:  "Resend transaction email",
	Long:   `Resend the transaction email with an optional message`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]
		messageToSigner, _ := cmd.Flags().GetString("message")

		// Make API call using global client
		resp, err := business.ResendTransactionEmail(proofClient, transactionID, messageToSigner)
		if err != nil {
			fmt.Println("Error resending email:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var bizResendSMSCmd = &cobra.Command{
	Use:    "resend-sms <transaction-id>",
	Short:  "Resend transaction SMS",
	Long:   `Resend the transaction SMS with optional phone parameters`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]
		phoneNumber, _ := cmd.Flags().GetString("phone-number")

		// Make API call using global client
		resp, err := business.ResendTransactionSMS(proofClient, transactionID, phoneNumber)
		if err != nil {
			fmt.Println("Error resending SMS:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var bizGetEligibleNotariesCmd = &cobra.Command{
	Use:    "eligible-notaries <transaction-id>",
	Short:  "Get eligible notaries for a transaction",
	Long:   `Get all eligible notaries for a specific transaction`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]

		// Make API call using global client
		resp, err := business.GetAllEligibleNotaries(proofClient, transactionID)
		if err != nil {
			fmt.Println("Error getting eligible notaries:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

// Business Documents Commands
var bizDocumentsCmd = &cobra.Command{
	Use:     "documents",
	Aliases: []string{"d", "docs"},
	Short:   "Business document operations",
	Long:    `Commands for managing business documents`,
}

var bizAddDocumentCmd = &cobra.Command{
	Use:    "add <transaction-id> <file-path>",
	Short:  "Add a document to a transaction",
	Long:   `Add a document to an existing transaction`,
	Args:   cobra.ExactArgs(2),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]
		filePath := args[1]

		// Get flag values
		filename, _ := cmd.Flags().GetString("filename")
		requirement, _ := cmd.Flags().GetString("requirement")
		notarizationRequired, _ := cmd.Flags().GetBool("notarization-required")
		witnessRequired, _ := cmd.Flags().GetBool("witness-required")
		bundlePosition, _ := cmd.Flags().GetInt("bundle-position")
		esignRequired, _ := cmd.Flags().GetBool("esign-required")
		identityConfirmationRequired, _ := cmd.Flags().GetBool("identity-confirmation-required")
		signingRequiresMeeting, _ := cmd.Flags().GetBool("signing-requires-meeting")
		vaulted, _ := cmd.Flags().GetBool("vaulted")
		authorizationHeader, _ := cmd.Flags().GetString("authorization-header")
		customerCanAnnotate, _ := cmd.Flags().GetBool("customer-can-annotate")
		pdfBookmarked, _ := cmd.Flags().GetBool("pdf-bookmarked")
		trackingID, _ := cmd.Flags().GetString("tracking-id")
		textTagSyntax, _ := cmd.Flags().GetString("text-tag-syntax")

		// Read the file
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}

		// Encode document to base64
		documentBase64 := base64.StdEncoding.EncodeToString(fileContent)

		// Build document parametersI
		params := &business.AddDocumentParams{
			Resource:                     documentBase64,
			Filename:                     filename,
			Requirement:                  requirement,
			NotarizationRequired:         notarizationRequired,
			WitnessRequired:              witnessRequired,
			BundlePosition:               bundlePosition,
			EsignRequired:                esignRequired,
			IdentityConfirmationRequired: identityConfirmationRequired,
			SigningRequiresMeeting:       signingRequiresMeeting,
			Vaulted:                      vaulted,
			AuthorizationHeader:          authorizationHeader,
			CustomerCanAnnotate:          customerCanAnnotate,
			PDFBookmarked:                pdfBookmarked,
			TrackingID:                   trackingID,
			TextTagSyntax:                textTagSyntax,
		}

		// Make API call using global client
		resp, err := business.AddDocument(proofClient, transactionID, params)
		if err != nil {
			fmt.Println("Error adding document:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var bizGetDocumentCmd = &cobra.Command{
	Use:    "get <transaction-id> <document-id>",
	Short:  "Get a document",
	Long:   `Get a document from a transaction`,
	Args:   cobra.ExactArgs(2),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		transactionID := args[0]
		documentID := args[1]

		// Get flag values
		encoding, _ := cmd.Flags().GetString("encoding")

		// Build query parameters
		params := &business.GetDocumentParams{
			Encoding: encoding,
		}

		// Make API call using global client
		resp, err := business.GetDocument(proofClient, transactionID, documentID, params)
		if err != nil {
			fmt.Println("Error fetching document:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var bizDeleteDocumentCmd = &cobra.Command{
	Use:    "delete <document-id>",
	Short:  "Delete a document",
	Long:   `Delete a document from a transaction`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		documentID := args[0]

		// Make API call using global client
		resp, err := business.DeleteDocument(proofClient, documentID)
		if err != nil {
			fmt.Println("Error deleting document:", err)
			os.Exit(1)
		}

		fmt.Println("Document deleted successfully")

		// Show response details if verbose
		PrintVerbose(string(resp))
	},
}

// Business Webhooks Commands
var bizWebhooksCmd = &cobra.Command{
	Use:     "webhooks",
	Aliases: []string{"w", "wh"},
	Short:   "Business webhook operations",
	Long:    `Commands for managing business webhooks`,
}

var bizGetWebhookCmd = &cobra.Command{
	Use:    "get",
	Short:  "Get webhook URL",
	Long:   `Retrieve the webhook URL for your organization`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		PrintVerbose("Fetching webhook configuration")

		body, err := proofClient.Get("/v1/webhooks")
		if err != nil {
			fmt.Println("Error getting webhook:", err)
			os.Exit(1)
		}

		PrintResponse(body)
	},
}

var bizListWebhooksCmd = &cobra.Command{
	Use:    "list",
	Short:  "List webhooks v2",
	Long:   `List all webhooks v2 for your organization`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		PrintVerbose("Fetching webhooks v2 list")

		body, err := proofClient.Get("/v2/webhooks")
		if err != nil {
			fmt.Println("Error listing webhooks:", err)
			os.Exit(1)
		}

		PrintResponse(body)
	},
}

var bizGetWebhookV2Cmd = &cobra.Command{
	Use:    "get-v2 <webhook-id>",
	Short:  "Get webhook v2 details",
	Long:   `Get details of a specific webhook v2`,
	PreRun: initializeForAPICall,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		webhookID := args[0]

		path := "/v2/webhooks/" + webhookID
		PrintVerbose("Fetching webhook v2 from: " + path)

		body, err := proofClient.Get(path)
		if err != nil {
			fmt.Println("Error getting webhook v2:", err)
			os.Exit(1)
		}

		PrintResponse(body)
	},
}

var bizCreateWebhookCmd = &cobra.Command{
	Use:    "create",
	Short:  "Create webhook v2",
	Long:   `Create a new webhook v2`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		name, _ := cmd.Flags().GetString("name")
		events, _ := cmd.Flags().GetStringSlice("events")

		if url == "" || name == "" {
			fmt.Println("Error: url and name are required")
			os.Exit(1)
		}

		requestBody := map[string]any{
			"url":    url,
			"name":   name,
			"events": events,
		}

		PrintVerbose("Creating webhook v2 with URL: " + url)

		body, err := proofClient.Post("/v2/webhooks", requestBody)
		if err != nil {
			fmt.Println("Error creating webhook v2:", err)
			os.Exit(1)
		}

		PrintResponse(body)
	},
}

var bizUpdateWebhookCmd = &cobra.Command{
	Use:    "update <webhook-id>",
	Short:  "Update webhook v2",
	Long:   `Update an existing webhook v2`,
	PreRun: initializeForAPICall,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		webhookID := args[0]
		url, _ := cmd.Flags().GetString("url")
		name, _ := cmd.Flags().GetString("name")
		events, _ := cmd.Flags().GetStringSlice("events")

		requestBody := make(map[string]any)
		if url != "" {
			requestBody["url"] = url
		}
		if name != "" {
			requestBody["name"] = name
		}
		if len(events) > 0 {
			requestBody["events"] = events
		}

		if len(requestBody) == 0 {
			fmt.Println("Error: at least one field to update is required")
			os.Exit(1)
		}

		path := "/v2/webhooks/" + webhookID
		PrintVerbose("Updating webhook v2: " + path)

		body, err := proofClient.Put(path, requestBody)
		if err != nil {
			fmt.Println("Error updating webhook v2:", err)
			os.Exit(1)
		}

		PrintResponse(body)
	},
}

var bizDeleteWebhookCmd = &cobra.Command{
	Use:    "delete <webhook-id>",
	Short:  "Delete webhook v2",
	Long:   `Delete a webhook v2`,
	PreRun: initializeForAPICall,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		webhookID := args[0]

		path := "/v2/webhooks/" + webhookID
		PrintVerbose("Deleting webhook v2: " + path)

		_, err := proofClient.Delete(path)
		if err != nil {
			fmt.Println("Error deleting webhook v2:", err)
			os.Exit(1)
		}

		fmt.Println("Webhook v2 deleted successfully")
	},
}

var bizGetWebhookEventsCmd = &cobra.Command{
	Use:    "events <webhook-id>",
	Short:  "Get webhook v2 events",
	Long:   `Get events for a specific webhook v2`,
	PreRun: initializeForAPICall,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		webhookID := args[0]

		path := "/v2/webhooks/" + webhookID + "/events"
		PrintVerbose("Fetching webhook v2 events from: " + path)

		body, err := proofClient.Get(path)
		if err != nil {
			fmt.Println("Error getting webhook v2 events:", err)
			os.Exit(1)
		}

		PrintResponse(body)
	},
}

var bizGetWebhookSubscriptionsCmd = &cobra.Command{
	Use:    "subscriptions",
	Short:  "Get webhook v2 subscriptions",
	Long:   `Get available webhook v2 event subscriptions`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		PrintVerbose("Fetching webhook v2 subscriptions")

		body, err := proofClient.Get("/v2/webhooks/subscriptions")
		if err != nil {
			fmt.Println("Error getting webhook v2 subscriptions:", err)
			os.Exit(1)
		}

		PrintResponse(body)
	},
}

// Business Notaries Commands
var bizNotariesCmd = &cobra.Command{
	Use:   "notaries",
	Short: "Business notary operations",
	Long:  `Commands for managing business notaries`,
}

var bizListNotariesCmd = &cobra.Command{
	Use:    "list",
	Short:  "List notaries",
	Long:   `List all notaries for your organization`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		orgID, _ := cmd.Flags().GetString("org-id")
		state, _ := cmd.Flags().GetString("state")

		path := "/v1/notaries/"
		if orgID != "" || state != "" {
			path += "?"
			if orgID != "" {
				path += fmt.Sprintf("organization_id=%s", orgID)
				if state != "" {
					path += "&"
				}
			}
			if state != "" {
				path += fmt.Sprintf("us_state_abbr=%s", state)
			}
		}

		PrintVerbose(fmt.Sprintf("Fetching notaries from: %s", path))

		body, err := proofClient.Get(path)
		if err != nil {
			fmt.Println("Error listing notaries:", err)
			os.Exit(1)
		}

		PrintResponse(body)
	},
}

var bizGetNotaryCmd = &cobra.Command{
	Use:    "get <notary-id>",
	Short:  "Get a notary",
	Long:   `Get details of a specific notary`,
	PreRun: initializeForAPICall,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		notaryID := args[0]

		path := fmt.Sprintf("/v1/notaries/%s", notaryID)
		PrintVerbose(fmt.Sprintf("Fetching notary from: %s", path))

		body, err := proofClient.Get(path)
		if err != nil {
			fmt.Println("Error getting notary:", err)
			os.Exit(1)
		}

		PrintResponse(body)
	},
}

var bizCreateNotaryCmd = &cobra.Command{
	Use:    "create",
	Short:  "Create a notary",
	Long:   `Create a new notary for your organization`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		email, _ := cmd.Flags().GetString("email")
		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")
		middleName, _ := cmd.Flags().GetString("middle-name")
		state, _ := cmd.Flags().GetString("state")

		if email == "" || firstName == "" || lastName == "" || state == "" {
			fmt.Println("Error: email, first-name, last-name, and state are required")
			os.Exit(1)
		}

		// Create the request body
		requestBody := map[string]string{
			"email":         email,
			"first_name":    firstName,
			"last_name":     lastName,
			"us_state_abbr": state,
		}

		if middleName != "" {
			requestBody["middle_name"] = middleName
		}

		PrintVerbose(fmt.Sprintf("Creating notary with email: %s", email))

		body, err := proofClient.Post("/v1/notaries/", requestBody)
		if err != nil {
			fmt.Println("Error creating notary:", err)
			os.Exit(1)
		}

		PrintResponse(body)
	},
}

var bizDeleteNotaryCmd = &cobra.Command{
	Use:    "delete <notary-id>",
	Short:  "Delete a notary",
	Long:   `Delete a specific notary from your organization`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		notaryID := args[0]

		path := fmt.Sprintf("/v1/notaries/%s", notaryID)
		PrintVerbose(fmt.Sprintf("Deleting notary: %s", path))

		_, err := proofClient.Delete(path)
		if err != nil {
			fmt.Println("Error deleting notary:", err)
			os.Exit(1)
		}

		fmt.Println("Notary deleted successfully")
	},
}

// Business Templates Commands
var bizTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Business template operations",
	Long:  `Commands for managing business templates`,
}

var bizListTemplatesCmd = &cobra.Command{
	Use:    "list",
	Short:  "List templates",
	Long:   `List all document templates for your organization`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		// Build query parameters
		path := "/v1/templates"
		var queryParams []string

		if limit > 0 {
			queryParams = append(queryParams, "limit="+strconv.Itoa(limit))
		}
		if offset > 0 {
			queryParams = append(queryParams, "offset="+strconv.Itoa(offset))
		}

		if len(queryParams) > 0 {
			result := queryParams[0]
			for i := 1; i < len(queryParams); i++ {
				result += "&" + queryParams[i]
			}
			path += "?" + result
		}

		PrintVerbose(fmt.Sprintf("Fetching templates from: %s", path))

		body, err := proofClient.Get(path)
		if err != nil {
			fmt.Println("Error listing templates:", err)
			os.Exit(1)
		}

		PrintResponse(body)
	},
}

// Business Referrals Commands
var bizReferralsCmd = &cobra.Command{
	Use:   "referrals",
	Short: "Business referral operations",
	Long:  `Commands for managing business referral campaigns`,
}

var bizCreateReferralCmd = &cobra.Command{
	Use:    "create",
	Short:  "Create a referral campaign",
	Long:   `Create a new referral campaign for your organization`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		coverPayment, _ := cmd.Flags().GetBool("cover-payment")
		organizationID, _ := cmd.Flags().GetString("organization-id")
		redirectURL, _ := cmd.Flags().GetString("redirect-url")
		useBranding, _ := cmd.Flags().GetBool("use-branding")

		if name == "" {
			fmt.Println("Error: name is required")
			os.Exit(1)
		}

		// Build referral parameters
		params := &business.CreateReferralParams{
			Name:           name,
			CoverPayment:   coverPayment,
			OrganizationID: organizationID,
			RedirectURL:    redirectURL,
			UseBranding:    useBranding,
		}

		// Make API call using global client
		resp, err := business.CreateReferral(proofClient, params)
		if err != nil {
			fmt.Println("Error creating referral campaign:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

var bizGenerateReferralCodeCmd = &cobra.Command{
	Use:    "generate-code <referral-campaign-id>",
	Short:  "Generate a referral code",
	Long:   `Generate a single use referral link for a campaign`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		referralCampaignID := args[0]
		expiresAt, _ := cmd.Flags().GetString("expires-at")

		// Build parameters
		params := &business.GenerateReferralCodeParams{
			ExpiresAt: expiresAt,
		}

		// Make API call using global client
		resp, err := business.GenerateReferralCode(proofClient, referralCampaignID, params)
		if err != nil {
			fmt.Println("Error generating referral code:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
	},
}

// Business Integrations Commands
var bizIntegrationsCmd = &cobra.Command{
	Use:   "integrations",
	Short: "Business integration operations",
	Long:  `Commands for managing business integrations`,
}

var bizCreateIntegrationCmd = &cobra.Command{
	Use:    "create",
	Short:  "Create an integration",
	Long:   `Create a new integration for your organization`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		orgID, _ := cmd.Flags().GetString("org-id")
		accountID, _ := cmd.Flags().GetString("account-id")
		environment, _ := cmd.Flags().GetString("environment")

		if name == "" || orgID == "" {
			fmt.Println("Error: name and org-id are required")
			os.Exit(1)
		}

		// Validate integration name
		switch name {
		case "ADOBE", "DOCUTECH":
			// Valid integration name
		default:
			fmt.Println("Error: name must be one of: ADOBE, DOCUTECH")
			os.Exit(1)
		}

		// Create the request params
		params := &business.CreateIntegrationParams{
			Name:           name,
			OrganizationID: orgID,
		}

		// Add configuration if provided
		if accountID != "" || environment != "" {
			params.Configuration = &business.IntegrationConfiguration{
				AccountID:   accountID,
				Environment: environment,
			}
		}

		PrintVerbose(fmt.Sprintf("Creating %s integration for organization %s", name, orgID))

		body, err := business.CreateIntegration(proofClient, params)
		if err != nil {
			fmt.Println("Error creating integration:", err)
			os.Exit(1)
		}

		PrintResponse(body)
	},
}

func init() {
	rootCmd.AddCommand(businessCmd)

	// Add subcommands
	businessCmd.AddCommand(bizTransactionsCmd)
	businessCmd.AddCommand(bizDocumentsCmd)
	businessCmd.AddCommand(bizWebhooksCmd)
	businessCmd.AddCommand(bizNotariesCmd)
	businessCmd.AddCommand(bizTemplatesCmd)
	businessCmd.AddCommand(bizReferralsCmd)
	businessCmd.AddCommand(bizIntegrationsCmd)

	// Transaction subcommands
	bizTransactionsCmd.AddCommand(bizListTransactionsCmd)
	bizTransactionsCmd.AddCommand(bizGetTransactionCmd)
	bizTransactionsCmd.AddCommand(bizCreateTransactionCmd)
	bizTransactionsCmd.AddCommand(bizDeleteTransactionCmd)
	bizTransactionsCmd.AddCommand(bizActivateTransactionCmd)
	bizTransactionsCmd.AddCommand(bizRecallTransactionCmd)
	bizTransactionsCmd.AddCommand(bizResendEmailCmd)
	bizTransactionsCmd.AddCommand(bizResendSMSCmd)
	bizTransactionsCmd.AddCommand(bizGetEligibleNotariesCmd)

	// Document subcommands
	bizDocumentsCmd.AddCommand(bizAddDocumentCmd)
	bizDocumentsCmd.AddCommand(bizGetDocumentCmd)
	bizDocumentsCmd.AddCommand(bizDeleteDocumentCmd)

	// Webhook subcommands
	bizWebhooksCmd.AddCommand(bizGetWebhookCmd)
	bizWebhooksCmd.AddCommand(bizListWebhooksCmd)
	bizWebhooksCmd.AddCommand(bizGetWebhookV2Cmd)
	bizWebhooksCmd.AddCommand(bizCreateWebhookCmd)
	bizWebhooksCmd.AddCommand(bizUpdateWebhookCmd)
	bizWebhooksCmd.AddCommand(bizDeleteWebhookCmd)
	bizWebhooksCmd.AddCommand(bizGetWebhookEventsCmd)
	bizWebhooksCmd.AddCommand(bizGetWebhookSubscriptionsCmd)

	// Notary subcommands
	bizNotariesCmd.AddCommand(bizListNotariesCmd)
	bizNotariesCmd.AddCommand(bizGetNotaryCmd)
	bizNotariesCmd.AddCommand(bizCreateNotaryCmd)
	bizNotariesCmd.AddCommand(bizDeleteNotaryCmd)

	// Template subcommands
	bizTemplatesCmd.AddCommand(bizListTemplatesCmd)

	// Referral subcommands
	bizReferralsCmd.AddCommand(bizCreateReferralCmd)
	bizReferralsCmd.AddCommand(bizGenerateReferralCodeCmd)

	// Integration subcommands
	bizIntegrationsCmd.AddCommand(bizCreateIntegrationCmd)

	// Add flags for transaction commands
	bizListTransactionsCmd.Flags().Int("limit", 10, "Number of transactions to return")
	bizListTransactionsCmd.Flags().Int("offset", 0, "Offset for pagination")
	bizListTransactionsCmd.Flags().String("status", "", "Filter by transaction status")
	bizListTransactionsCmd.Flags().String("created-start", "", "Filter by created date start (YYYY-MM-DD)")
	bizListTransactionsCmd.Flags().String("created-end", "", "Filter by created date end (YYYY-MM-DD)")
	bizListTransactionsCmd.Flags().String("last-updated-start", "", "Filter by last updated date start (YYYY-MM-DD)")
	bizListTransactionsCmd.Flags().String("last-updated-end", "", "Filter by last updated date end (YYYY-MM-DD)")

	bizCreateTransactionCmd.Flags().String("email", "", "Signer's email address (required)")
	bizCreateTransactionCmd.Flags().String("first-name", "", "Signer's first name")
	bizCreateTransactionCmd.Flags().String("last-name", "", "Signer's last name")
	bizCreateTransactionCmd.Flags().String("document", "", "Path to document file (required)")
	bizCreateTransactionCmd.Flags().String("name", "", "Transaction name")
	bizCreateTransactionCmd.Flags().String("type", "", "Transaction type")
	bizCreateTransactionCmd.Flags().Bool("draft", false, "Create transaction as draft")
	bizCreateTransactionCmd.Flags().String("middle-name", "", "Signer's middle name")
	bizCreateTransactionCmd.Flags().String("phone-number", "", "Signer's phone number")
	bizCreateTransactionCmd.Flags().String("message-to-signer", "", "Message to signer (GitHub Flavored Markdown)")
	bizCreateTransactionCmd.Flags().String("message-subject", "", "Email subject line")
	bizCreateTransactionCmd.Flags().String("activation-time", "", "ISO-8601 datetime when signer can connect with notary")
	bizCreateTransactionCmd.Flags().String("expiry", "", "ISO-8601 datetime after which transaction expires")
	bizCreateTransactionCmd.Flags().Bool("suppress-email", false, "Don't send notification email on activation")
	bizCreateTransactionCmd.Flags().String("auth-requirement", "", "Authentication requirement (sms or none)")
	bizCreateTransactionCmd.Flags().Bool("require-secondary-photo-id", false, "Require two forms of photo ID")
	bizCreateTransactionCmd.Flags().String("payer", "", "Who pays for the transaction (signer or sender)")
	bizCreateTransactionCmd.Flags().String("external-id", "", "External system ID")

	// Add flags for document commands
	bizAddDocumentCmd.Flags().String("filename", "", "Plain language name for the document")
	bizAddDocumentCmd.Flags().String("requirement", "", "Completion requirement (notarization, esign, identity_confirmation, readonly, non_essential)")
	bizAddDocumentCmd.Flags().Bool("notarization-required", false, "Whether notarization is required")
	bizAddDocumentCmd.Flags().Bool("witness-required", false, "Whether additional witness must be present")
	bizAddDocumentCmd.Flags().Bool("esign-required", false, "Whether e-signature is required")
	bizAddDocumentCmd.Flags().Bool("identity-confirmation-required", false, "Whether identity confirmation is required")
	bizAddDocumentCmd.Flags().Bool("vaulted", false, "Whether to store authoritative copy in eVault")
	bizAddDocumentCmd.Flags().Bool("customer-can-annotate", false, "Whether signer can add annotations")
	bizAddDocumentCmd.Flags().String("tracking-id", "", "External tracking identifier")
	bizAddDocumentCmd.Flags().Int("bundle-position", 0, "Position in document bundle")
	bizAddDocumentCmd.Flags().Bool("signing-requires-meeting", false, "Whether signing requires a meeting")
	bizAddDocumentCmd.Flags().String("authorization-header", "", "Header for fetching doc URLs (format: header_name:header_value)")
	bizAddDocumentCmd.Flags().Bool("pdf-bookmarked", false, "Whether document is bookmarked PDF (splits by bookmarks)")
	bizAddDocumentCmd.Flags().String("text-tag-syntax", "", "Syntax used by text tags")

	bizGetDocumentCmd.Flags().String("encoding", "", "Can be 'base64' or 'uri'. 'uri' returns hosted URL (only after transaction completion)")

	// Add flags for webhook commands
	bizCreateWebhookCmd.Flags().String("url", "", "Webhook URL")
	bizCreateWebhookCmd.Flags().String("name", "", "Webhook name")
	bizCreateWebhookCmd.Flags().StringSlice("events", []string{}, "Event types to subscribe to")
	bizCreateWebhookCmd.Flags().String("header", "", "Header value to pass through every request (e.g. X-Custom-Header:X-Custom-Key)")

	bizUpdateWebhookCmd.Flags().String("url", "", "Webhook URL")
	bizUpdateWebhookCmd.Flags().String("name", "", "Webhook name")
	bizUpdateWebhookCmd.Flags().StringSlice("events", []string{}, "Event types to subscribe to")
	bizUpdateWebhookCmd.Flags().String("header", "", "Header value to pass through every request (e.g. X-Custom-Header:X-Custom-Key)")

	// Add flags for notary commands
	bizListNotariesCmd.Flags().String("org-id", "", "Organization ID")
	bizListNotariesCmd.Flags().String("state", "", "Two-letter state abbreviation")

	bizCreateNotaryCmd.Flags().String("email", "", "Notary's email address")
	bizCreateNotaryCmd.Flags().String("first-name", "", "Notary's first name")
	bizCreateNotaryCmd.Flags().String("last-name", "", "Notary's last name")
	bizCreateNotaryCmd.Flags().String("middle-name", "", "Notary's middle name")
	bizCreateNotaryCmd.Flags().String("state", "", "Two-letter state abbreviation")

	// Add flags for template commands
	bizListTemplatesCmd.Flags().Int("limit", 0, "How many results to return (default: 100, max: 1000)")
	bizListTemplatesCmd.Flags().Int("offset", 0, "Number of results to skip for pagination")

	// Add flags for referral commands
	bizCreateReferralCmd.Flags().String("name", "", "Name of the new campaign (required)")
	bizCreateReferralCmd.Flags().Bool("cover-payment", false, "Will the organization pay for these referred transactions?")
	bizCreateReferralCmd.Flags().String("organization-id", "", "ID of organization to create the campaign for (child orgs only)")
	bizCreateReferralCmd.Flags().String("redirect-url", "", "URL that customers will be sent to from the referral")
	bizCreateReferralCmd.Flags().Bool("use-branding", false, "Will the referred transactions display the orgs branding?")

	bizGenerateReferralCodeCmd.Flags().String("expires-at", "", "ISO 8601 timestamp for when the code should expire (default: 3 months from creation)")

	// Add flags for integration commands
	bizCreateIntegrationCmd.Flags().String("name", "", "Integration name (ADOBE or DOCUTECH)")
	bizCreateIntegrationCmd.Flags().String("org-id", "", "Organization ID")
	bizCreateIntegrationCmd.Flags().String("account-id", "", "Integration account ID")
	bizCreateIntegrationCmd.Flags().String("environment", "", "Integration environment")

	// Add flags for additional transaction commands
	bizRecallTransactionCmd.Flags().String("reason", "", "Optional reason for recalling the transaction")
	bizResendEmailCmd.Flags().String("message", "", "Optional message to signer")
	bizResendSMSCmd.Flags().String("phone-number", "", "Optional phone number to send SMS to")
}
