package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/pkg/utils"
)

var (
	prettyPrint bool
	verbose     bool
	proofClient *utils.ProofClient
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

// initializeForAPICall sets up debug logging and ensures client is ready
func initializeForAPICall(cmd *cobra.Command, args []string) {
	toggleDebug(cmd, args)
	if proofClient == nil {
		client, err := utils.NewProofClient()
		utils.HandleError(err, "Failed to create client")
		proofClient = client
	}
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
