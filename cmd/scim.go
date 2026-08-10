package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/pkg/utils"
	"github.com/tsarlewey/proof-sdk-go/scim"
)

// scimCmd represents the scim command
var scimCmd = &cobra.Command{
	Use:     "scim",
	Aliases: []string{"s"},
	Short:   "SCIM (System for Cross-domain Identity Management) operations",
	Long:    `Commands for interacting with the Proof SCIM API for user and identity management`,
}

// scimSCIMJSON is the content type SCIM requires on request bodies.
const scimSCIMJSON = "application/scim+json"

// registerSCIMUserFlags declares the flag set shared by the create and update
// commands, which is also what buildSCIMUserBody reads.
func registerSCIMUserFlags(cmd *cobra.Command) {
	f := cmd.Flags()
	f.String("username", "", "Username (email address) - required")
	f.String("given-name", "", "First name")
	f.String("family-name", "", "Last name")
	f.String("email", "", "Email address")
	f.StringSlice("roles", []string{}, "User roles (can specify multiple)")
	f.String("external-id", "", "External ID from SAML provider")
	f.Bool("active", true, "Whether the user is active")
}

// buildSCIMUserBody assembles the user body shared by the create (POST) and
// update (PUT) commands, both of which send a full user representation.
func buildSCIMUserBody(cmd *cobra.Command) scim.UserCreationParams {
	userName, _ := cmd.Flags().GetString("username")
	givenName, _ := cmd.Flags().GetString("given-name")
	familyName, _ := cmd.Flags().GetString("family-name")
	email, _ := cmd.Flags().GetString("email")
	roles, _ := cmd.Flags().GetStringSlice("roles")
	externalID, _ := cmd.Flags().GetString("external-id")
	active, _ := cmd.Flags().GetBool("active")

	if userName == "" {
		fmt.Println("Error: username is required")
		os.Exit(1)
	}

	body := scim.UserCreationParams{UserName: userName, Active: &active}

	if givenName != "" || familyName != "" {
		body.Name = &struct {
			FamilyName *string `json:"familyName,omitempty"`
			GivenName  *string `json:"givenName,omitempty"`
		}{}
		if givenName != "" {
			body.Name.GivenName = &givenName
		}
		if familyName != "" {
			body.Name.FamilyName = &familyName
		}
	}

	if email != "" {
		body.Emails = &[]struct {
			Value *string `json:"value,omitempty"`
		}{{Value: &email}}
	}

	if len(roles) > 0 {
		values := make([]struct {
			Value *string `json:"value,omitempty"`
		}, len(roles))
		for i := range roles {
			values[i].Value = &roles[i]
		}
		body.Roles = &values
	}

	if externalID != "" {
		body.ExternalId = &externalID
	}

	return body
}

// scimPatchOperation is a single SCIM PATCH operation as sent to the server.
type scimPatchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path,omitempty"`
	Value any    `json:"value,omitempty"`
}

type scimPatchRequest struct {
	Operations []scimPatchOperation `json:"Operations"`
}

// parseSCIMPatchOp parses a single operation string in "op:path[:value]" format.
// The value segment is JSON-decoded when it parses as valid JSON, otherwise
// kept as a raw string so users can pass plain values like "admin".
func parseSCIMPatchOp(raw string) (scimPatchOperation, error) {
	parts := strings.SplitN(raw, ":", 3)
	if len(parts) < 2 {
		return scimPatchOperation{}, fmt.Errorf("invalid operation format: %q (expected op:path[:value])", raw)
	}

	op := scimPatchOperation{Op: parts[0], Path: parts[1]}
	if len(parts) == 3 {
		var jsonValue any
		if err := json.Unmarshal([]byte(parts[2]), &jsonValue); err == nil {
			op.Value = jsonValue
		} else {
			op.Value = parts[2]
		}
	}
	return op, nil
}

// buildSCIMPatchBody parses a slice of operation strings and returns the
// SCIM PATCH request body as JSON bytes. The body always has Operations as
// an array to satisfy SCIM, even though the upstream OpenAPI spec models it
// as a single object.
func buildSCIMPatchBody(rawOps []string) ([]byte, error) {
	if len(rawOps) == 0 {
		return nil, fmt.Errorf("at least one operation is required")
	}
	ops := make([]scimPatchOperation, 0, len(rawOps))
	for _, raw := range rawOps {
		op, err := parseSCIMPatchOp(raw)
		if err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return json.Marshal(scimPatchRequest{Operations: ops})
}

// SCIM Users Commands
var scimUsersCmd = &cobra.Command{
	Use:   "users",
	Short: "SCIM user operations",
	Long:  `Commands for managing SCIM users`,
}

var scimListUsersCmd = &cobra.Command{
	Use:    "list <organization-id>",
	Short:  "List SCIM users",
	Long:   `List SCIM users in an organization with optional pagination`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		organizationID := args[0]

		// Get command line flags
		startIndex, _ := cmd.Flags().GetInt("start-index")
		count, _ := cmd.Flags().GetInt("count")

		params := &scim.ListUsersParams{}
		if startIndex > 0 {
			params.StartIndex = &startIndex
		}
		if count > 0 {
			params.Count = &count
		}

		// Make API call using SDK client
		client := getSCIMClient()
		resp, err := client.ListUsersWithResponse(context.Background(), organizationID, params)
		utils.HandleError(err, "listing users")
		checkAPIStatus(resp.StatusCode(), resp.Body, "listing users")

		// Use global helper to print response
		PrintResponse(resp.Body)
	},
}

var scimGetUserCmd = &cobra.Command{
	Use:    "get <organization-id> <user-id>",
	Short:  "Get a SCIM user",
	Long:   `Get details of a specific SCIM user`,
	Args:   cobra.ExactArgs(2),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		organizationID := args[0]
		userID := args[1]

		// Make API call using SDK client
		client := getSCIMClient()
		resp, err := client.GetUserWithResponse(context.Background(), organizationID, userID)
		utils.HandleError(err, "getting user")
		checkAPIStatus(resp.StatusCode(), resp.Body, "getting user")

		// Use global helper to print response
		PrintResponse(resp.Body)
	},
}

var scimCreateUserCmd = &cobra.Command{
	Use:    "create <organization-id>",
	Short:  "Create a SCIM user",
	Long:   `Create a new SCIM user in an organization`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		organizationID := args[0]

		body := buildSCIMUserBody(cmd)

		// Make API call using SDK client
		client := getSCIMClient()
		resp, err := client.CreateUserWithApplicationScimPlusJSONBodyWithResponse(
			context.Background(), organizationID, body)
		utils.HandleError(err, "creating user")
		checkAPIStatus(resp.StatusCode(), resp.Body, "creating user")

		// Use global helper to print response
		PrintResponse(resp.Body)
	},
}

var scimUpdateUserCmd = &cobra.Command{
	Use:    "update <organization-id> <user-id>",
	Short:  "Update a SCIM user",
	Long:   `Update a SCIM user (replaces all fields)`,
	Args:   cobra.ExactArgs(2),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		organizationID := args[0]
		userID := args[1]

		body := buildSCIMUserBody(cmd)

		// Make API call using SDK client
		client := getSCIMClient()
		resp, err := client.ReplaceUserWithApplicationScimPlusJSONBodyWithResponse(
			context.Background(), organizationID, userID, body)
		utils.HandleError(err, "updating user")
		checkAPIStatus(resp.StatusCode(), resp.Body, "updating user")

		// Use global helper to print response
		PrintResponse(resp.Body)
	},
}

var scimPatchUserCmd = &cobra.Command{
	Use:    "patch <organization-id> <user-id>",
	Short:  "Patch a SCIM user",
	Long:   `Partially update a SCIM user using PATCH operations`,
	Args:   cobra.ExactArgs(2),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		organizationID := args[0]
		userID := args[1]

		operations, _ := cmd.Flags().GetStringSlice("operation")
		bodyBytes, err := buildSCIMPatchBody(operations)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		client := getSCIMClient()
		resp, err := client.PatchUserWithBodyWithResponse(
			context.Background(),
			organizationID,
			userID,
			scimSCIMJSON,
			bytes.NewReader(bodyBytes),
		)
		utils.HandleError(err, "patching user")
		checkAPIStatus(resp.StatusCode(), resp.Body, "patching user")

		PrintResponse(resp.Body)
	},
}

var scimDeleteUserCmd = &cobra.Command{
	Use:    "delete <organization-id> <user-id>",
	Short:  "Delete a SCIM user",
	Long:   `Delete a SCIM user from an organization`,
	Args:   cobra.ExactArgs(2),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		organizationID := args[0]
		userID := args[1]

		// Make API call using SDK client
		client := getSCIMClient()
		resp, err := client.DeleteUserWithResponse(context.Background(), organizationID, userID)
		utils.HandleError(err, "deleting user")
		checkAPIStatus(resp.StatusCode(), resp.Body, "deleting user")

		if len(resp.Body) > 0 {
			PrintResponse(resp.Body)
		} else {
			fmt.Println("SCIM user deleted successfully")
		}
	},
}

// SCIM Schema Commands
var scimSchemasCmd = &cobra.Command{
	Use:   "schemas",
	Short: "SCIM schema operations",
	Long:  `Commands for retrieving SCIM schemas and configuration`,
}

var scimGetUserSchemaCmd = &cobra.Command{
	Use:    "user <organization-id>",
	Short:  "Get user schema",
	Long:   `Get the SCIM user schema for an organization`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		organizationID := args[0]

		// Make API call using SDK client
		client := getSCIMClient()
		resp, err := client.RetrieveUsersSchemaWithResponse(context.Background(), organizationID)
		utils.HandleError(err, "getting user schema")
		checkAPIStatus(resp.StatusCode(), resp.Body, "getting user schema")

		// Use global helper to print response
		PrintResponse(resp.Body)
	},
}

var scimGetServiceProviderConfigCmd = &cobra.Command{
	Use:    "service-provider-config <organization-id>",
	Short:  "Get service provider configuration",
	Long:   `Get the SCIM service provider configuration for an organization`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		organizationID := args[0]

		// Make API call using SDK client
		client := getSCIMClient()
		resp, err := client.GetServiceProviderConfigWithResponse(context.Background(), organizationID)
		utils.HandleError(err, "getting service provider config")
		checkAPIStatus(resp.StatusCode(), resp.Body, "getting service provider config")

		// Use global helper to print response
		PrintResponse(resp.Body)
	},
}

var scimGetResourceTypesCmd = &cobra.Command{
	Use:    "resource-types <organization-id>",
	Short:  "Get resource types",
	Long:   `Get the supported SCIM resource types for an organization`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		organizationID := args[0]

		// Make API call using SDK client
		client := getSCIMClient()
		resp, err := client.GetResourceTypesWithResponse(context.Background(), organizationID)
		utils.HandleError(err, "getting resource types")
		checkAPIStatus(resp.StatusCode(), resp.Body, "getting resource types")

		// Use global helper to print response
		PrintResponse(resp.Body)
	},
}

func init() {
	rootCmd.AddCommand(scimCmd)

	// Add subcommands
	scimCmd.AddCommand(scimUsersCmd)
	scimCmd.AddCommand(scimSchemasCmd)

	// User subcommands
	scimUsersCmd.AddCommand(scimListUsersCmd)
	scimUsersCmd.AddCommand(scimGetUserCmd)
	scimUsersCmd.AddCommand(scimCreateUserCmd)
	scimUsersCmd.AddCommand(scimUpdateUserCmd)
	scimUsersCmd.AddCommand(scimPatchUserCmd)
	scimUsersCmd.AddCommand(scimDeleteUserCmd)

	// Schema subcommands
	scimSchemasCmd.AddCommand(scimGetUserSchemaCmd)
	scimSchemasCmd.AddCommand(scimGetServiceProviderConfigCmd)
	scimSchemasCmd.AddCommand(scimGetResourceTypesCmd)

	// Add flags for user list
	scimListUsersCmd.Flags().Int("start-index", 1, "1-based index of first result")
	scimListUsersCmd.Flags().Int("count", 50, "Maximum number of results per page")

	// Create and update both send a full user representation, so they take the
	// same flag set — see buildSCIMUserBody.
	for _, c := range []*cobra.Command{scimCreateUserCmd, scimUpdateUserCmd} {
		registerSCIMUserFlags(c)
		c.MarkFlagRequired("username")
	}

	// Add flags for user patch
	scimPatchUserCmd.Flags().StringSlice("operation", []string{}, "PATCH operations in format 'op:path[:value]' (e.g., 'replace:active:false')")
	scimPatchUserCmd.MarkFlagRequired("operation")
}
