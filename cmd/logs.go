package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tsarlewey/proof-cli/pkg/utils"
	"github.com/tsarlewey/proof-sdk-go/logs"
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:     "logs",
	Aliases: []string{"l"},
	Short:   "Security event log operations",
	Long:    `Commands for reading the Proof security event log (OCSF-formatted audit events)`,
}

var logsListEventsCmd = &cobra.Command{
	Use:    "list",
	Short:  "List security events",
	Long:   `List security events with optional filtering and cursor-based pagination`,
	PreRun: initializeForAPICall,
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		since, _ := cmd.Flags().GetString("since")
		classUID, _ := cmd.Flags().GetInt("class-uid")
		severityID, _ := cmd.Flags().GetInt("severity-id")

		params := &logs.ListSecurityEventsParams{}
		if limit > 0 {
			params.Limit = &limit
		}
		if cursor != "" {
			params.Cursor = &cursor
		}
		if since != "" {
			params.Since = &since
		}
		if classUID > 0 {
			params.ClassUid = &classUID
		}
		if severityID > 0 {
			params.SeverityId = &severityID
		}

		client := getLogsClient()
		resp, err := client.ListSecurityEventsWithResponse(context.Background(), params)
		utils.HandleError(err, "listing security events")
		checkAPIStatus(resp.StatusCode(), resp.Body, "listing security events")

		PrintResponse(resp.Body)
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.AddCommand(logsListEventsCmd)

	logsListEventsCmd.Flags().Int("limit", 0, "Number of events per page (1-1000, default 100)")
	logsListEventsCmd.Flags().String("cursor", "", "Pagination cursor from a previous response's next_cursor")
	logsListEventsCmd.Flags().String("since", "", "ISO 8601 timestamp; only events at or after this time (max 90 days back)")
	logsListEventsCmd.Flags().Int("class-uid", 0, "Filter by OCSF event class identifier")
	logsListEventsCmd.Flags().Int("severity-id", 0, "Filter by OCSF severity identifier")
}
