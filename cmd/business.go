package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/pkg/utils"
	"github.com/tsarlewey/proof-sdk-go/business"
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
		params := &business.GetAllTransactionsParams{
			Limit:              utils.Ptr(limit),
			Offset:             utils.Ptr(offset),
			DocumentUrlVersion: utils.Ptr(business.GetAllTransactionsParamsDocumentUrlVersionV2),
		}

		// Parse status if provided
		if status != "" {
			params.TransactionStatus = utils.Ptr(business.GetAllTransactionsParamsTransactionStatus(status))
		}

		params.CreatedDateStart = parseDateFlag("created-start date", dateStart, "2006-01-02")
		params.CreatedDateEnd = parseDateFlag("created-end date", dateEnd, "2006-01-02")

		// Make API call using SDK
		client := getBusinessClient()
		resp, err := client.GetAllTransactionsWithResponse(context.Background(), params)
		utils.HandleError(err, "fetching transactions")
		checkAPIStatus(resp.StatusCode(), resp.Body, "fetching transactions")

		PrintResponse(resp.Body)
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

		params := &business.GetTransactionParams{
			DocumentUrlVersion: utils.Ptr(business.GetTransactionParamsDocumentUrlVersionV2),
		}

		// Make API call using SDK
		client := getBusinessClient()
		resp, err := client.GetTransactionWithResponse(context.Background(), transactionID, params)
		utils.HandleError(err, "fetching transaction")
		checkAPIStatus(resp.StatusCode(), resp.Body, "fetching transaction")

		PrintResponse(resp.Body)
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
		utils.HandleError(err, "reading document file")

		// Encode document to base64
		documentBase64 := base64.StdEncoding.EncodeToString(documentData)

		// Build query parameters
		queryParams := &business.CreateTransactionParams{
			DocumentUrlVersion: utils.Ptr(business.CreateTransactionParamsDocumentUrlVersionV2),
		}

		// Build request body
		body := business.CreateTransactionJSONRequestBody{
			Signer: business.Signer{
				Email:     email,
				FirstName: utils.PtrIfNotEmpty(firstName),
				LastName:  utils.PtrIfNotEmpty(lastName),
			},
			Documents:       utils.Ptr([]string{documentBase64}),
			Draft:           utils.Ptr(draft),
			TransactionName: utils.PtrIfNotEmpty(transactionName),
			TransactionType: utils.PtrIfNotEmpty(transactionType),
		}

		// Make API call using SDK
		client := getBusinessClient()
		resp, err := client.CreateTransactionWithResponse(context.Background(), queryParams, body)
		utils.HandleError(err, "creating transaction")
		checkAPIStatus(resp.StatusCode(), resp.Body, "creating transaction")

		PrintResponse(resp.Body)
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

		// Make API call using SDK
		client := getBusinessClient()
		resp, err := client.DeleteTransactionWithResponse(context.Background(), transactionID)
		utils.HandleError(err, "deleting transaction")
		checkAPIStatus(resp.StatusCode(), resp.Body, "deleting transaction")

		if isSuccess(resp.StatusCode()) {
			fmt.Println("Transaction deleted successfully")
		}
		PrintVerbose(string(resp.Body))
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

		suppressEmail, _ := cmd.Flags().GetBool("suppress-email")
		requireVerification, _ := cmd.Flags().GetBool("require-new-signer-verification")

		params := &business.ActivateDraftTransactionParams{
			DocumentUrlVersion: utils.Ptr(business.ActivateDraftTransactionParamsDocumentUrlVersionV2),
		}
		body := business.ActivateDraftTransactionJSONRequestBody{
			SuppressEmail:                &suppressEmail,
			RequireNewSignerVerification: &requireVerification,
		}

		// Make API call using SDK
		client := getBusinessClient()
		resp, err := client.ActivateDraftTransactionWithResponse(context.Background(), transactionID, params, body)
		utils.HandleError(err, "activating transaction")
		checkAPIStatus(resp.StatusCode(), resp.Body, "activating transaction")

		PrintResponse(resp.Body)
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

		params := &business.RecallTransactionParams{
			DocumentUrlVersion: utils.Ptr(business.RecallTransactionParamsDocumentUrlVersionV2),
		}
		if recallReason != "" {
			params.RecallReason = utils.Ptr(recallReason)
		}

		// Make API call using SDK
		client := getBusinessClient()
		resp, err := client.RecallTransactionWithResponse(context.Background(), transactionID, params)
		utils.HandleError(err, "recalling transaction")
		checkAPIStatus(resp.StatusCode(), resp.Body, "recalling transaction")

		PrintResponse(resp.Body)
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

		params := &business.ResendTransactionEmailParams{
			DocumentUrlVersion: utils.Ptr(business.ResendTransactionEmailParamsDocumentUrlVersionV2),
		}
		if messageToSigner != "" {
			params.MessageToSigner = utils.Ptr(messageToSigner)
		}

		// Make API call using SDK
		client := getBusinessClient()
		resp, err := client.ResendTransactionEmailWithResponse(context.Background(), transactionID, params)
		utils.HandleError(err, "resending email")
		checkAPIStatus(resp.StatusCode(), resp.Body, "resending email")

		PrintResponse(resp.Body)
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

		params := &business.ResendTransactionSMSParams{
			DocumentUrlVersion: utils.Ptr(business.ResendTransactionSMSParamsDocumentUrlVersionV2),
		}
		if phoneNumber != "" {
			params.PhoneNumber = utils.Ptr(phoneNumber)
		}

		// Make API call using SDK
		client := getBusinessClient()
		resp, err := client.ResendTransactionSMSWithResponse(context.Background(), transactionID, params, business.ResendTransactionSMSJSONRequestBody{})
		utils.HandleError(err, "resending SMS")
		checkAPIStatus(resp.StatusCode(), resp.Body, "resending SMS")

		PrintResponse(resp.Body)
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

		// Make API call using SDK
		client := getBusinessClient()
		resp, err := client.GetAllEligibleNotariesWithResponse(context.Background(), transactionID)
		utils.HandleError(err, "getting eligible notaries")
		checkAPIStatus(resp.StatusCode(), resp.Body, "getting eligible notaries")

		PrintResponse(resp.Body)
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
		utils.HandleError(err, "reading file")

		// Encode document to base64
		documentBase64 := base64.StdEncoding.EncodeToString(fileContent)

		// Build query parameters
		queryParams := &business.AddDocumentParams{
			DocumentUrlVersion: utils.Ptr(business.AddDocumentParamsDocumentUrlVersionV2),
		}

		// Build request body
		body := business.AddDocumentJSONRequestBody{
			Resource:                     utils.Ptr(documentBase64),
			Filename:                     utils.PtrIfNotEmpty(filename),
			NotarizationRequired:         utils.Ptr(notarizationRequired),
			WitnessRequired:              utils.Ptr(witnessRequired),
			EsignRequired:                utils.Ptr(esignRequired),
			IdentityConfirmationRequired: utils.Ptr(identityConfirmationRequired),
			SigningRequiresMeeting:       utils.Ptr(signingRequiresMeeting),
			Vaulted:                      utils.Ptr(vaulted),
			CustomerCanAnnotate:          utils.Ptr(customerCanAnnotate),
			PdfBookmarked:                utils.Ptr(pdfBookmarked),
			TrackingId:                   utils.PtrIfNotEmpty(trackingID),
			TextTagSyntax:                utils.PtrIfNotEmpty(textTagSyntax),
			Requirement:                  utils.PtrIfNotEmpty(requirement),
			AuthorizationHeader:          utils.PtrIfNotEmpty(authorizationHeader),
		}

		if bundlePosition > 0 {
			body.BundlePosition = utils.Ptr(bundlePosition)
		}

		// Make API call using SDK
		client := getBusinessClient()
		resp, err := client.AddDocumentWithResponse(context.Background(), transactionID, queryParams, body)
		utils.HandleError(err, "adding document")
		checkAPIStatus(resp.StatusCode(), resp.Body, "adding document")

		PrintResponse(resp.Body)
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

		encoding, _ := cmd.Flags().GetString("encoding")

		params := &business.GetDocumentParams{
			DocumentUrlVersion: utils.Ptr(business.GetDocumentParamsDocumentUrlVersionV2),
		}
		if encoding != "" {
			params.Encoding = utils.Ptr(encoding)
		}

		// Make API call using SDK
		client := getBusinessClient()
		resp, err := client.GetDocumentWithResponse(context.Background(), transactionID, documentID, params)
		utils.HandleError(err, "fetching document")
		checkAPIStatus(resp.StatusCode(), resp.Body, "fetching document")

		PrintResponse(resp.Body)
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

		// Make API call using SDK
		client := getBusinessClient()
		resp, err := client.DeleteDocumentWithResponse(context.Background(), documentID)
		utils.HandleError(err, "deleting document")
		checkAPIStatus(resp.StatusCode(), resp.Body, "deleting document")

		if isSuccess(resp.StatusCode()) {
			fmt.Println("Document deleted successfully")
		}
		PrintVerbose(string(resp.Body))
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

		client := getBusinessClient()
		resp, err := client.GetWebhookURLWithResponse(context.Background())
		utils.HandleError(err, "getting webhook")
		checkAPIStatus(resp.StatusCode(), resp.Body, "getting webhook")

		PrintResponse(resp.Body)
	},
}

var bizListWebhooksCmd = &cobra.Command{
	Use:    "list",
	Short:  "List webhooks v2",
	Long:   `List all webhooks v2 for your organization`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		PrintVerbose("Fetching webhooks v2 list")

		client := getBusinessClient()
		resp, err := client.GetAllWebhooksV2WithResponse(context.Background())
		utils.HandleError(err, "listing webhooks")
		checkAPIStatus(resp.StatusCode(), resp.Body, "listing webhooks")

		PrintResponse(resp.Body)
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

		PrintVerbose("Fetching webhook v2: " + webhookID)

		client := getBusinessClient()
		resp, err := client.GetWebhookV2WithResponse(context.Background(), webhookID)
		utils.HandleError(err, "getting webhook v2")
		checkAPIStatus(resp.StatusCode(), resp.Body, "getting webhook v2")

		PrintResponse(resp.Body)
	},
}

var bizCreateWebhookCmd = &cobra.Command{
	Use:    "create",
	Short:  "Create webhook v2",
	Long:   `Create a new webhook v2`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		events, _ := cmd.Flags().GetStringSlice("events")
		header, _ := cmd.Flags().GetString("header")

		if url == "" {
			fmt.Println("Error: url is required")
			os.Exit(1)
		}

		body := business.CreateWebhookV2JSONRequestBody{
			Url:           url,
			Subscriptions: events,
		}
		if header != "" {
			body.Header = utils.Ptr(header)
		}

		PrintVerbose("Creating webhook v2 with URL: " + url)

		client := getBusinessClient()
		resp, err := client.CreateWebhookV2WithResponse(context.Background(), body)
		utils.HandleError(err, "creating webhook v2")
		checkAPIStatus(resp.StatusCode(), resp.Body, "creating webhook v2")

		PrintResponse(resp.Body)
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
		events, _ := cmd.Flags().GetStringSlice("events")
		header, _ := cmd.Flags().GetString("header")

		// For updates, we need to provide the full body
		// URL and Subscriptions are required fields
		body := business.UpdateWebhookV2JSONRequestBody{
			Url:           url,
			Subscriptions: events,
		}
		if header != "" {
			body.Header = utils.Ptr(header)
		}

		PrintVerbose("Updating webhook v2: " + webhookID)

		client := getBusinessClient()
		resp, err := client.UpdateWebhookV2WithResponse(context.Background(), webhookID, body)
		utils.HandleError(err, "updating webhook v2")
		checkAPIStatus(resp.StatusCode(), resp.Body, "updating webhook v2")

		PrintResponse(resp.Body)
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

		PrintVerbose("Deleting webhook v2: " + webhookID)

		client := getBusinessClient()
		resp, err := client.DeleteWebhookV2WithResponse(context.Background(), webhookID)
		utils.HandleError(err, "deleting webhook v2")
		checkAPIStatus(resp.StatusCode(), resp.Body, "deleting webhook v2")

		if isSuccess(resp.StatusCode()) {
			fmt.Println("Webhook v2 deleted successfully")
		}
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

		PrintVerbose("Fetching webhook v2 events: " + webhookID)

		client := getBusinessClient()
		resp, err := client.GetWebhookEventsV2WithResponse(context.Background(), webhookID, nil)
		utils.HandleError(err, "getting webhook v2 events")
		checkAPIStatus(resp.StatusCode(), resp.Body, "getting webhook v2 events")

		PrintResponse(resp.Body)
	},
}

var bizGetWebhookSubscriptionsCmd = &cobra.Command{
	Use:    "subscriptions",
	Short:  "Get webhook v2 subscriptions",
	Long:   `Get available webhook v2 event subscriptions`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		PrintVerbose("Fetching webhook v2 subscriptions")

		client := getBusinessClient()
		resp, err := client.GetWebhookSubscriptionsV2WithResponse(context.Background())
		utils.HandleError(err, "getting webhook v2 subscriptions")
		checkAPIStatus(resp.StatusCode(), resp.Body, "getting webhook v2 subscriptions")

		PrintResponse(resp.Body)
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

		params := &business.GetAllNotariesParams{}
		if orgID != "" {
			params.OrganizationId = utils.Ptr(orgID)
		}
		if state != "" {
			params.UsStateAbbr = utils.Ptr(state)
		}

		PrintVerbose("Fetching notaries")

		client := getBusinessClient()
		resp, err := client.GetAllNotariesWithResponse(context.Background(), params)
		utils.HandleError(err, "listing notaries")
		checkAPIStatus(resp.StatusCode(), resp.Body, "listing notaries")

		PrintResponse(resp.Body)
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

		PrintVerbose("Fetching notary: " + notaryID)

		client := getBusinessClient()
		resp, err := client.GetNotaryWithResponse(context.Background(), notaryID)
		utils.HandleError(err, "getting notary")
		checkAPIStatus(resp.StatusCode(), resp.Body, "getting notary")

		PrintResponse(resp.Body)
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

		body := business.CreateNotaryJSONRequestBody{
			Email:       email,
			FirstName:   firstName,
			LastName:    lastName,
			UsStateAbbr: state,
		}
		if middleName != "" {
			body.MiddleName = utils.Ptr(middleName)
		}

		PrintVerbose(fmt.Sprintf("Creating notary with email: %s", email))

		client := getBusinessClient()
		resp, err := client.CreateNotaryWithResponse(context.Background(), body)
		utils.HandleError(err, "creating notary")
		checkAPIStatus(resp.StatusCode(), resp.Body, "creating notary")

		PrintResponse(resp.Body)
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

		PrintVerbose("Deleting notary: " + notaryID)

		client := getBusinessClient()
		resp, err := client.DeleteNotaryWithResponse(context.Background(), notaryID)
		utils.HandleError(err, "deleting notary")
		checkAPIStatus(resp.StatusCode(), resp.Body, "deleting notary")

		if isSuccess(resp.StatusCode()) {
			fmt.Println("Notary deleted successfully")
		}
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

		params := &business.GetAllTemplatesParams{}
		if limit > 0 {
			params.Limit = utils.Ptr(limit)
		}
		if offset > 0 {
			params.Offset = utils.Ptr(offset)
		}

		PrintVerbose("Fetching templates")

		client := getBusinessClient()
		resp, err := client.GetAllTemplatesWithResponse(context.Background(), params)
		utils.HandleError(err, "listing templates")
		checkAPIStatus(resp.StatusCode(), resp.Body, "listing templates")

		PrintResponse(resp.Body)
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

		body := business.CreateReferralJSONRequestBody{
			Name:           name,
			CoverPayment:   utils.Ptr(coverPayment),
			OrganizationId: utils.PtrIfNotEmpty(organizationID),
			RedirectUrl:    utils.PtrIfNotEmpty(redirectURL),
			UseBranding:    utils.Ptr(useBranding),
		}

		client := getBusinessClient()
		resp, err := client.CreateReferralWithResponse(context.Background(), body)
		utils.HandleError(err, "creating referral campaign")
		checkAPIStatus(resp.StatusCode(), resp.Body, "creating referral campaign")

		PrintResponse(resp.Body)
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

		body := business.GenerateReferralCodeJSONRequestBody{}
		if expiresAt != "" {
			body.ExpiresAt = utils.Ptr(expiresAt)
		}

		client := getBusinessClient()
		resp, err := client.GenerateReferralCodeWithResponse(context.Background(), referralCampaignID, body)
		utils.HandleError(err, "generating referral code")
		checkAPIStatus(resp.StatusCode(), resp.Body, "generating referral code")

		PrintResponse(resp.Body)
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
		var integrationName business.IntegrationParamsName
		switch name {
		case "ADOBE":
			integrationName = business.ADOBE
		case "DOCUTECH":
			integrationName = business.DOCUTECH
		default:
			fmt.Println("Error: name must be one of: ADOBE, DOCUTECH")
			os.Exit(1)
		}

		body := business.CreateIntegrationJSONRequestBody{
			Name:           utils.Ptr(integrationName),
			OrganizationId: utils.Ptr(orgID),
		}

		// Add configuration if provided
		if accountID != "" || environment != "" {
			body.Configuration = &struct {
				AccountId   *string `json:"account_id,omitempty"`
				Environment *string `json:"environment,omitempty"`
			}{
				AccountId:   utils.PtrIfNotEmpty(accountID),
				Environment: utils.PtrIfNotEmpty(environment),
			}
		}

		PrintVerbose(fmt.Sprintf("Creating %s integration for organization %s", name, orgID))

		client := getBusinessClient()
		resp, err := client.CreateIntegrationWithResponse(context.Background(), body)
		utils.HandleError(err, "creating integration")
		checkAPIStatus(resp.StatusCode(), resp.Body, "creating integration")

		PrintResponse(resp.Body)
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
	bizActivateTransactionCmd.Flags().Bool("suppress-email", false, "Don't email the signer on activation (you must supply transaction_access_link yourself)")
	bizActivateTransactionCmd.Flags().Bool("require-new-signer-verification", false, "Require the signer to verify email ownership")
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
	bizCreateWebhookCmd.Flags().String("name", "", "Webhook name (deprecated, use subscriptions)")
	bizCreateWebhookCmd.Flags().StringSlice("events", []string{}, "Event subscriptions to subscribe to")
	bizCreateWebhookCmd.Flags().String("header", "", "Header value to pass through every request (e.g. X-Custom-Header:X-Custom-Key)")

	bizUpdateWebhookCmd.Flags().String("url", "", "Webhook URL")
	bizUpdateWebhookCmd.Flags().String("name", "", "Webhook name (deprecated)")
	bizUpdateWebhookCmd.Flags().StringSlice("events", []string{}, "Event subscriptions to subscribe to")
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
