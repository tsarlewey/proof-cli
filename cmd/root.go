package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/pkg/utils"
	"github.com/tsarlewey/proof-sdk-go/business"
	"github.com/tsarlewey/proof-sdk-go/certificates"
	"github.com/tsarlewey/proof-sdk-go/common"
	"github.com/tsarlewey/proof-sdk-go/logs"
	"github.com/tsarlewey/proof-sdk-go/realestate"
	"github.com/tsarlewey/proof-sdk-go/scim"
)

var (
	prettyPrint bool
	verbose     bool
	proofClient *utils.ProofClient

	// SDK clients - lazily initialized
	businessClient     *business.ClientWithResponses
	realestateClient   *realestate.ClientWithResponses
	scimClient         *scim.ClientWithResponses
	logsClient         *logs.ClientWithResponses
	certificatesClient *certificates.ClientWithResponses
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "proof",
	Short: "A CLI for interacting with the Proof API",
	Long: `A command-line interface for interacting with the Proof API.
This CLI allows you to manage transactions, documents, notaries, and webhooks.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize global settings that apply to all commands
		if verbose {
			fmt.Fprintf(os.Stderr, "Running command: %s\n", cmd.CommandPath())
		}
	},
	SilenceUsage:  true,  // Don't show usage on errors
	SilenceErrors: false, // Show errors but don't duplicate with usage
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initializeForAPICall ensures client is ready
func initializeForAPICall(cmd *cobra.Command, args []string) {
	if proofClient == nil {
		client, err := utils.NewProofClient()
		utils.HandleError(err, "Failed to create client")
		proofClient = client
	}
}

// getBusinessClient returns a lazily-initialized Business SDK client
func getBusinessClient() *business.ClientWithResponses {
	if businessClient == nil {
		authDoer := common.NewAuthenticatedDoer(proofClient)
		client, err := business.NewClientWithResponses(
			proofClient.GetConfig().APIEndpoint,
			business.WithHTTPClient(authDoer),
		)
		utils.HandleError(err, "Failed to create Business SDK client")
		businessClient = client
	}
	return businessClient
}

// getRealEstateClient returns a lazily-initialized Real Estate SDK client
func getRealEstateClient() *realestate.ClientWithResponses {
	if realestateClient == nil {
		authDoer := common.NewAuthenticatedDoer(proofClient)
		client, err := realestate.NewClientWithResponses(
			proofClient.GetConfig().APIEndpoint,
			realestate.WithHTTPClient(authDoer),
		)
		utils.HandleError(err, "Failed to create Real Estate SDK client")
		realestateClient = client
	}
	return realestateClient
}

// getSCIMClient returns a lazily-initialized SCIM SDK client
func getSCIMClient() *scim.ClientWithResponses {
	if scimClient == nil {
		authDoer := common.NewAuthenticatedDoer(proofClient)
		client, err := scim.NewClientWithResponses(
			proofClient.GetConfig().APIEndpoint,
			scim.WithHTTPClient(authDoer),
		)
		utils.HandleError(err, "Failed to create SCIM SDK client")
		scimClient = client
	}
	return scimClient
}

// getLogsClient returns a lazily-initialized Logs SDK client
func getLogsClient() *logs.ClientWithResponses {
	if logsClient == nil {
		authDoer := common.NewAuthenticatedDoer(proofClient)
		client, err := logs.NewClientWithResponses(
			proofClient.GetConfig().APIEndpoint,
			logs.WithHTTPClient(authDoer),
		)
		utils.HandleError(err, "Failed to create Logs SDK client")
		logsClient = client
	}
	return logsClient
}

// getCertificatesClient returns a lazily-initialized Certificates SDK client
func getCertificatesClient() *certificates.ClientWithResponses {
	if certificatesClient == nil {
		authDoer := common.NewAuthenticatedDoer(proofClient)
		client, err := certificates.NewClientWithResponses(
			proofClient.GetConfig().APIEndpoint,
			certificates.WithHTTPClient(authDoer),
		)
		utils.HandleError(err, "Failed to create Certificates SDK client")
		certificatesClient = client
	}
	return certificatesClient
}

// PrintResponse handles response output with optional pretty printing
func PrintResponse(resp []byte, prefix ...string) {
	// Print prefix if provided
	if len(prefix) > 0 {
		fmt.Println(prefix[0])
	}

	if prettyPrint {
		var result any
		if err := json.Unmarshal(resp, &result); err != nil {
			// If JSON parsing fails, just print raw response
			fmt.Println(string(resp))
			return
		}

		// Use json.MarshalIndent for proper formatting
		prettyJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			// If formatting fails, print raw response
			fmt.Println(string(resp))
			return
		}

		// Apply colors to the properly formatted JSON
		fmt.Println(colorizeJSON(string(prettyJSON)))
	} else {
		fmt.Println(string(resp))
	}
}

// colorizeJSON adds color to JSON output using proper patterns
func colorizeJSON(jsonStr string) string {
	// Define colors using fatih/color
	keyColor := color.New(color.FgCyan).SprintFunc()
	stringColor := color.New(color.FgGreen).SprintFunc()
	numberColor := color.New(color.FgYellow).SprintFunc()
	boolColor := color.New(color.FgMagenta).SprintFunc()
	nullColor := color.New(color.FgRed).SprintFunc()

	// More precise regular expressions for JSON elements
	// Match keys: "key":
	keyPattern := regexp.MustCompile(`"([^"\\]*(\\.[^"\\]*)*)"(\s*):`)
	// Match string values: : "value"
	stringValuePattern := regexp.MustCompile(`(:\s*)"([^"\\]*(\\.[^"\\]*)*)"`)
	// Match numbers: : 123, : -45.67, : 1.23e-4
	numberPattern := regexp.MustCompile(`(:\s*)(-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)`)
	// Match booleans: : true, : false
	boolPattern := regexp.MustCompile(`(:\s*)(true|false)`)
	// Match null: : null
	nullPattern := regexp.MustCompile(`(:\s*)(null)`)

	result := jsonStr

	// Apply colors in the correct order to avoid conflicts
	// 1. Color keys (property names)
	result = keyPattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := keyPattern.FindStringSubmatch(match)
		if len(parts) >= 4 {
			key := parts[1]
			colon := parts[3]
			return keyColor(`"`+key+`"`) + colon + ":"
		}
		return match
	})

	// 2. Color string values
	result = stringValuePattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := stringValuePattern.FindStringSubmatch(match)
		if len(parts) >= 3 {
			prefix := parts[1]
			value := parts[2]
			return prefix + stringColor(`"`+value+`"`)
		}
		return match
	})

	// 3. Color numbers
	result = numberPattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := numberPattern.FindStringSubmatch(match)
		if len(parts) >= 3 {
			prefix := parts[1]
			number := parts[2]
			return prefix + numberColor(number)
		}
		return match
	})

	// 4. Color booleans
	result = boolPattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := boolPattern.FindStringSubmatch(match)
		if len(parts) >= 3 {
			prefix := parts[1]
			boolVal := parts[2]
			return prefix + boolColor(boolVal)
		}
		return match
	})

	// 5. Color null values
	result = nullPattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := nullPattern.FindStringSubmatch(match)
		if len(parts) >= 3 {
			prefix := parts[1]
			nullVal := parts[2]
			return prefix + nullColor(nullVal)
		}
		return match
	})

	return result
}

// PrintVerbose prints additional information when verbose flag is set
func PrintVerbose(message string) {
	if verbose {
		fmt.Println(message)
	}
}

// isSuccess reports whether an HTTP status code is in the 2xx range.
func isSuccess(code int) bool {
	return code >= 200 && code < 300
}

// checkAPIStatus exits 1 if the HTTP status is non-2xx, printing the response
// body to stderr so the caller can see the API's error detail. SDK calls
// return (resp, err) where err is only set on transport errors — this helper
// closes the gap so application-level 4xx/5xx responses also fail fast.
func checkAPIStatus(statusCode int, body []byte, action string) {
	if isSuccess(statusCode) {
		return
	}
	fmt.Fprintf(os.Stderr, "Error %s: API returned status %d\n", action, statusCode)
	if len(body) > 0 {
		fmt.Fprintln(os.Stderr, string(body))
	}
	os.Exit(1)
}

// parseDateFlag parses an optional date flag value. Returns nil when value is
// empty. On parse failure it prints an error and exits.
func parseDateFlag(flagName, value, layout string) *time.Time {
	if value == "" {
		return nil
	}
	t, err := time.Parse(layout, value)
	if err != nil {
		fmt.Printf("Error parsing %s: %v\n", flagName, err)
		os.Exit(1)
	}
	return &t
}

func init() {
	// Configure command behavior
	rootCmd.CompletionOptions.DisableDefaultCmd = false
	rootCmd.CompletionOptions.DisableDescriptions = false

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&prettyPrint, "pretty", "p", true, "pretty print JSON output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show additional output")

	// Make --debug a global flag since it's already handled globally
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug output")
}
