package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/internal/scim"
)

// scimCmd represents the scim command
var scimCmd = &cobra.Command{
	Use:     "scim",
	Aliases: []string{"s"},
	Short:   "SCIM (System for Cross-domain Identity Management) operations",
	Long:    `Commands for interacting with the Proof SCIM API for user and identity management`,
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

		params := &scim.ListUsersParams{
			StartIndex: startIndex,
			Count:      count,
		}

		// Make API call using global proofClient
		resp, err := scim.ListUsers(proofClient, organizationID, params)
		if err != nil {
			fmt.Println("Error listing users:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
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

		// Make API call
		resp, err := scim.GetUser(proofClient, organizationID, userID)
		if err != nil {
			fmt.Println("Error getting user:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
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

		// Get command line flags
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

		params := &scim.CreateUserParams{
			UserName: userName,
			Active:   active,
		}

		// Add name if provided
		if givenName != "" || familyName != "" {
			params.Name = &scim.UserName{
				GivenName:  givenName,
				FamilyName: familyName,
			}
		}

		// Add email if provided
		if email != "" {
			params.Emails = []scim.UserEmail{
				{
					Value:   email,
					Primary: true,
				},
			}
		}

		// Add roles if provided
		if len(roles) > 0 {
			for i, role := range roles {
				params.Roles = append(params.Roles, scim.UserRole{
					Value:   role,
					Primary: i == 0, // First role is primary
				})
			}
		}

		// Add external ID if provided
		if externalID != "" {
			params.ExternalID = externalID
		}

		// Make API call
		resp, err := scim.CreateUser(proofClient, organizationID, params)
		if err != nil {
			fmt.Println("Error creating user:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
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

		// Get command line flags
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

		params := &scim.UpdateUserParams{
			UserName: userName,
			Active:   active,
		}

		// Add name if provided
		if givenName != "" || familyName != "" {
			params.Name = &scim.UserName{
				GivenName:  givenName,
				FamilyName: familyName,
			}
		}

		// Add email if provided
		if email != "" {
			params.Emails = []scim.UserEmail{
				{
					Value:   email,
					Primary: true,
				},
			}
		}

		// Add roles if provided
		if len(roles) > 0 {
			for i, role := range roles {
				params.Roles = append(params.Roles, scim.UserRole{
					Value:   role,
					Primary: i == 0, // First role is primary
				})
			}
		}

		// Add external ID if provided
		if externalID != "" {
			params.ExternalID = externalID
		}

		// Make API call
		resp, err := scim.UpdateUser(proofClient, organizationID, userID, params)
		if err != nil {
			fmt.Println("Error updating user:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
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

		// Get patch operations from flags
		operations, _ := cmd.Flags().GetStringSlice("operation")

		if len(operations) == 0 {
			fmt.Println("Error: at least one operation is required. Use --operation flag")
			os.Exit(1)
		}

		var patchOps []scim.PatchOperation
		for _, op := range operations {
			// Parse operation string in format "op:path:value"
			parts := strings.SplitN(op, ":", 3)
			if len(parts) < 2 {
				fmt.Printf("Error: invalid operation format: %s. Expected format: op:path[:value]\n", op)
				os.Exit(1)
			}

			patchOp := scim.PatchOperation{
				Op:   parts[0],
				Path: parts[1],
			}

			// Add value if provided
			if len(parts) == 3 {
				value := parts[2]
				// Try to parse as JSON, fallback to string
				var jsonValue interface{}
				if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
					patchOp.Value = jsonValue
				} else {
					patchOp.Value = value
				}
			}

			patchOps = append(patchOps, patchOp)
		}

		params := &scim.PatchUserParams{
			Operations: patchOps,
		}

		// Make API call
		resp, err := scim.PatchUser(proofClient, organizationID, userID, params)
		if err != nil {
			fmt.Println("Error patching user:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
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

		// Make API call
		resp, err := scim.DeleteUser(proofClient, organizationID, userID)
		if err != nil {
			fmt.Println("Error deleting user:", err)
			os.Exit(1)
		}

		if len(resp) > 0 {
			PrintResponse(resp)
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

		// Make API call
		resp, err := scim.GetUserSchema(proofClient, organizationID)
		if err != nil {
			fmt.Println("Error getting user schema:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
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

		// Make API call
		resp, err := scim.GetServiceProviderConfig(proofClient, organizationID)
		if err != nil {
			fmt.Println("Error getting service provider config:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
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

		// Make API call
		resp, err := scim.GetResourceTypes(proofClient, organizationID)
		if err != nil {
			fmt.Println("Error getting resource types:", err)
			os.Exit(1)
		}

		// Use global helper to print response
		PrintResponse(resp)
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

	// Add flags for user create
	scimCreateUserCmd.Flags().String("username", "", "Username (email address) - required")
	scimCreateUserCmd.Flags().String("given-name", "", "First name")
	scimCreateUserCmd.Flags().String("family-name", "", "Last name")
	scimCreateUserCmd.Flags().String("email", "", "Email address")
	scimCreateUserCmd.Flags().StringSlice("roles", []string{}, "User roles (can specify multiple)")
	scimCreateUserCmd.Flags().String("external-id", "", "External ID from SAML provider")
	scimCreateUserCmd.Flags().Bool("active", true, "Whether the user is active")
	scimCreateUserCmd.MarkFlagRequired("username")

	// Add flags for user update
	scimUpdateUserCmd.Flags().String("username", "", "Username (email address) - required")
	scimUpdateUserCmd.Flags().String("given-name", "", "First name")
	scimUpdateUserCmd.Flags().String("family-name", "", "Last name")
	scimUpdateUserCmd.Flags().String("email", "", "Email address")
	scimUpdateUserCmd.Flags().StringSlice("roles", []string{}, "User roles (can specify multiple)")
	scimUpdateUserCmd.Flags().String("external-id", "", "External ID from SAML provider")
	scimUpdateUserCmd.Flags().Bool("active", true, "Whether the user is active")
	scimUpdateUserCmd.MarkFlagRequired("username")

	// Add flags for user patch
	scimPatchUserCmd.Flags().StringSlice("operation", []string{}, "PATCH operations in format 'op:path[:value]' (e.g., 'replace:active:false')")
	scimPatchUserCmd.MarkFlagRequired("operation")
}
