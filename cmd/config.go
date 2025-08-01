package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and modify configuration settings for the CLI.`,
}

// configGetCmd represents the config get command
var configGetCmd = &cobra.Command{
	Use:    "get",
	Short:  "Get configuration",
	Long:   `Get the current configuration settings.`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			os.Exit(1)
		}
		fmt.Println("API Endpoint:", config.APIEndpoint)
		fmt.Println("Timeout:", config.Timeout)
		
		// Show API Key status
		if config.APIKey != "" {
			fmt.Println("API Key: configured")
		} else {
			fmt.Println("API Key: not configured")
		}
		
		// Show OAuth configuration
		if config.OAuth != nil {
			fmt.Println("OAuth Enabled:", config.OAuth.Enabled)
			if config.OAuth.Enabled {
				fmt.Println("OAuth Client ID:", config.OAuth.ClientID)
				if config.OAuth.Scope != "" {
					fmt.Println("OAuth Scope:", config.OAuth.Scope)
				}
				// Show OAuth token status
				if config.OAuthToken != nil {
					if config.OAuthToken.IsExpired() {
						fmt.Println("OAuth Token: expired")
					} else {
						fmt.Println("OAuth Token: valid until", config.OAuthToken.ExpiresAt.Format(time.RFC3339))
					}
				} else {
					fmt.Println("OAuth Token: not present")
				}
			}
		} else {
			fmt.Println("OAuth Enabled: false")
		}
	},
}

// configSetEndpointCmd represents the config set-endpoint command
var configSetEndpointCmd = &cobra.Command{
	Use:    "set-endpoint [endpoint]",
	Short:  "Set API endpoint",
	Long:   `Set the API endpoint for the CLI.`,
	PreRun: initializeForAPICall,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		endpoint := args[0]

		config, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			os.Exit(1)
		}

		config.APIEndpoint = endpoint

		if err := utils.SaveConfig(config); err != nil {
			fmt.Println("Error saving config:", err)
			os.Exit(1)
		}

		fmt.Println("API endpoint set to:", endpoint)
	},
}

// configSetTimeoutCmd represents the config set-timeout command
var configSetTimeoutCmd = &cobra.Command{
	Use:    "set-timeout [timeout]",
	Short:  "Set timeout",
	Long:   `Set the timeout for API requests in seconds.`,
	PreRun: initializeForAPICall,
	Args:   cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var timeout int
		if _, err := fmt.Sscanf(args[0], "%d", &timeout); err != nil {
			fmt.Println("Error: timeout must be a number")
			os.Exit(1)
		}

		config, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			os.Exit(1)
		}

		config.Timeout = time.Duration(timeout) * time.Second

		if err := utils.SaveConfig(config); err != nil {
			fmt.Println("Error saving config:", err)
			os.Exit(1)
		}

		fmt.Println("Timeout set to:", timeout, "seconds")
	},
}

// configSetAPIKeyCmd represents the config set-api-key command
var configSetAPIKeyCmd = &cobra.Command{
	Use:    "set-api-key [api_key]",
	Short:  "Set API key",
	Long:   `Set the API key for the CLI.`,
	Args:   cobra.ExactArgs(1),
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := args[0]

		if err := utils.SaveAPIKey(apiKey); err != nil {
			fmt.Println("Error saving API key:", err)
			os.Exit(1)
		}

		fmt.Println("API key set successfully")
	},
}

// configSetOAuthCmd represents the config set-oauth command
var configSetOAuthCmd = &cobra.Command{
	Use:   "set-oauth <client-id> <client-secret>",
	Short: "Set OAuth credentials",
	Long:  `Set OAuth client credentials for authentication`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		clientID := args[0]
		clientSecret := args[1]
		scope, _ := cmd.Flags().GetString("scope")

		config, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			os.Exit(1)
		}

		// Initialize OAuth config if it doesn't exist
		if config.OAuth == nil {
			config.OAuth = &utils.OAuthConfig{}
		}

		config.OAuth.Enabled = true
		config.OAuth.ClientID = clientID
		config.OAuth.ClientSecret = clientSecret
		config.OAuth.Scope = scope

		if err := utils.SaveConfig(config); err != nil {
			fmt.Println("Error saving config:", err)
			os.Exit(1)
		}

		fmt.Println("OAuth credentials configured successfully")
		fmt.Println("Client ID:", clientID)
		if scope != "" {
			fmt.Println("Scope:", scope)
		}
	},
}

// configDisableOAuthCmd represents the config disable-oauth command
var configDisableOAuthCmd = &cobra.Command{
	Use:   "disable-oauth",
	Short: "Disable OAuth authentication",
	Long:  `Disable OAuth authentication and fall back to API key`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			os.Exit(1)
		}

		if config.OAuth != nil {
			config.OAuth.Enabled = false
		}

		if err := utils.SaveConfig(config); err != nil {
			fmt.Println("Error saving config:", err)
			os.Exit(1)
		}

		fmt.Println("OAuth authentication disabled")
	},
}

// configTestOAuthCmd represents the config test-oauth command
var configTestOAuthCmd = &cobra.Command{
	Use:   "test-oauth",
	Short: "Test OAuth authentication",
	Long:  `Test OAuth authentication by getting a token`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := utils.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			os.Exit(1)
		}

		if config.OAuth == nil || !config.OAuth.Enabled {
			fmt.Println("OAuth is not enabled. Use 'proof config set-oauth' to configure OAuth credentials.")
			os.Exit(1)
		}

		// Create a client to test OAuth (this will automatically attempt OAuth)
		client, err := utils.NewProofClient()
		if err != nil {
			fmt.Println("OAuth authentication failed:", err)
			os.Exit(1)
		}

		// Test OAuth by forcing a token refresh through the client
		token, err := client.TestOAuthAuthentication()
		if err != nil {
			fmt.Println("OAuth authentication failed:", err)
			os.Exit(1)
		}

		fmt.Println("OAuth authentication successful!")
		fmt.Println("Token Type:", token.TokenType)
		fmt.Println("Expires At:", token.ExpiresAt.Format(time.RFC3339))
		if token.Scope != "" {
			fmt.Println("Scope:", token.Scope)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetEndpointCmd)
	configCmd.AddCommand(configSetTimeoutCmd)
	configCmd.AddCommand(configSetAPIKeyCmd)
	configCmd.AddCommand(configSetOAuthCmd)
	configCmd.AddCommand(configDisableOAuthCmd)
	configCmd.AddCommand(configTestOAuthCmd)

	configSetOAuthCmd.Flags().String("scope", "", "OAuth scope (optional)")
}
