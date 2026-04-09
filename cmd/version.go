package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the version of the CLI
	Version = "0.2.0-alpha"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of the CLI application`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("proof version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
