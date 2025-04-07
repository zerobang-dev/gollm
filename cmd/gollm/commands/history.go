package commands

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/zerobang-dev/gollm/pkg/config"
	"github.com/zerobang-dev/gollm/pkg/logger"
)

var (
	historyLimitFlag  int
	historyDetailFlag bool
	historySearchFlag string
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View query history",
	Long:  `Display the history of your recent queries.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir := config.GetConfigDir()

		// Initialize logger
		queryLogger, err := logger.NewLogger(configDir)
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		defer func() {
			if err := queryLogger.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing logger: %v\n", err)
			}
		}()

		var queries []logger.Query

		// Get queries - either search or recent
		if historySearchFlag != "" {
			queries, err = queryLogger.SearchQueries(historySearchFlag, historyLimitFlag)
			if err != nil {
				return fmt.Errorf("failed to search query history: %w", err)
			}
		} else {
			queries, err = queryLogger.GetRecentQueries(historyLimitFlag)
			if err != nil {
				return fmt.Errorf("failed to retrieve query history: %w", err)
			}
		}

		if len(queries) == 0 {
			if historySearchFlag != "" {
				fmt.Println("No queries found matching your search.")
			} else {
				fmt.Println("No query history found.")
			}
			return nil
		}

		// Display queries in a table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		if _, err := fmt.Fprintln(w, "TIME\tMODEL\tDURATION\tPROMPT"); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
		if _, err := fmt.Fprintln(w, "----\t-----\t--------\t------"); err != nil {
			return fmt.Errorf("failed to write separator: %w", err)
		}

		for _, q := range queries {
			// Format timestamp
			timeStr := q.Timestamp.Format("2006-01-02 15:04:05")

			// Format duration
			durationStr := fmt.Sprintf("%dms", q.Duration)

			// Truncate prompt
			promptPreview := truncateString(q.Prompt, 40)

			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				timeStr, q.Model, durationStr, promptPreview); err != nil {
				return fmt.Errorf("failed to write row: %w", err)
			}
		}
		if err := w.Flush(); err != nil {
			return fmt.Errorf("failed to flush writer: %w", err)
		}

		// Show detailed view of the most recent query if requested
		if historyDetailFlag && len(queries) > 0 {
			latest := queries[0]

			fmt.Println("\nLatest Query Details:")
			fmt.Println("--------------------")
			fmt.Printf("Time: %s\n", latest.Timestamp.Format("2006-01-02 15:04:05"))
			fmt.Printf("Model: %s\n", latest.Model)
			fmt.Printf("Duration: %dms\n", latest.Duration)
			fmt.Printf("Temperature: %.2f\n", latest.Temperature)

			fmt.Println("\nPrompt:")
			fmt.Println(latest.Prompt)

			fmt.Println("\nResponse:")
			fmt.Println(latest.Response)
		}

		return nil
	},
}

func init() {
	historyCmd.Flags().IntVarP(&historyLimitFlag, "limit", "l", 10, "Number of queries to show")
	historyCmd.Flags().BoolVarP(&historyDetailFlag, "detail", "d", false, "Show detailed view of the most recent query")
	historyCmd.Flags().StringVarP(&historySearchFlag, "search", "s", "", "Search for queries containing text")

	rootCmd.AddCommand(historyCmd)
}

// Helper function to truncate strings
func truncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
