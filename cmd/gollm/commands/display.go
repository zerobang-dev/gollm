package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
	"github.com/zerobang-dev/gollm/pkg/llm"
)

// displayProviderResults displays the results from multiple providers
func displayProviderResults(results map[string]llm.ProviderResponse) error {
	// Create a new tabwriter for formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Add header
	if _, err := fmt.Fprintln(w, "PROVIDER\tMODEL\tTIME\tRESPONSE"); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}
	if _, err := fmt.Fprintln(w, "--------\t-----\t----\t--------"); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Sort providers for consistent output
	sortedProviders := make([]string, 0, len(results))
	for provider := range results {
		sortedProviders = append(sortedProviders, provider)
	}
	sort.Strings(sortedProviders)

	// Print each result
	for _, provider := range sortedProviders {
		result := results[provider]
		var responseText string

		if result.Error != nil {
			// Format error
			responseText = fmt.Sprintf("ERROR: %v", result.Error)
		} else {
			// Get response and handle multi-line responses
			responseLines := strings.Split(result.Response, "\n")
			if len(responseLines) > 1 {
				// For multi-line responses, just show first line with indication
				responseText = responseLines[0] + " [...]"
			} else {
				responseText = result.Response
			}

			// Truncate long responses
			if len(responseText) > 60 {
				responseText = responseText[:57] + "..."
			}
		}

		// Format elapsed time
		timeStr := fmt.Sprintf("%dms", result.ElapsedTime.Milliseconds())

		// Print provider, model, time, and response
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", result.Provider, result.Model, timeStr, responseText); err != nil {
			return fmt.Errorf("error writing result: %w", err)
		}
	}

	// Flush the tabwriter
	if err := w.Flush(); err != nil {
		return fmt.Errorf("error flushing tabwriter: %w", err)
	}

	// For each provider, print the full response
	fmt.Println("\nDetailed responses:")
	fmt.Println("------------------")
	for _, provider := range sortedProviders {
		result := results[provider]

		fmt.Printf("\n## %s (%s)\n\n", result.Provider, result.Model)
		if result.Error != nil {
			fmt.Printf("ERROR: %v\n", result.Error)
		} else {
			fmt.Println(result.Response)
		}
	}

	return nil
}

// displayVerboseResult displays a verbose result for a single provider
func displayVerboseResult(prompt string, modelFlag string, response string, elapsedTime time.Duration) error {
	// Create a new tabwriter for formatted output with colors
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Define colors
	headerColor := color.New(color.FgCyan, color.Bold)
	modelColor := color.New(color.FgGreen)
	timeColor := color.New(color.FgYellow)

	// Add header with colors
	if _, err := headerColor.Fprintln(w, "PROMPT\tMODEL\tTIME\tRESPONSE"); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}
	if _, err := headerColor.Fprintln(w, "------\t-----\t----\t--------"); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Format prompt (truncate if too long)
	promptText := prompt
	if len(promptText) > 40 {
		promptText = promptText[:37] + "..."
	}

	// Format response for display (handle multiline)
	responseText := response
	responseLines := strings.Split(responseText, "\n")
	if len(responseLines) > 1 {
		responseText = responseLines[0] + " [...]"
	}

	// Truncate long responses for table
	if len(responseText) > 60 {
		responseText = responseText[:57] + "..."
	}

	// Print row with colored model and time
	if _, err := fmt.Fprintf(w, "%s\t", promptText); err != nil {
		return fmt.Errorf("error writing prompt: %w", err)
	}
	if _, err := modelColor.Fprintf(w, "%s\t", modelFlag); err != nil {
		return fmt.Errorf("error writing model: %w", err)
	}
	if _, err := timeColor.Fprintf(w, "%dms\t", elapsedTime.Milliseconds()); err != nil {
		return fmt.Errorf("error writing time: %w", err)
	}
	if _, err := fmt.Fprintf(w, "%s\n", responseText); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	// Flush the tabwriter
	if err := w.Flush(); err != nil {
		return fmt.Errorf("error flushing tabwriter: %w", err)
	}

	// Print full response after the table
	fmt.Println("\nFull response:")
	fmt.Println("-------------")
	fmt.Println(response)

	return nil
}

// displaySimpleResult displays a simple result for a single provider
func displaySimpleResult(response string, elapsedTime time.Duration) {
	// Print timing information
	fmt.Printf("Time: %dms\n\n", elapsedTime.Milliseconds())

	// Print response
	fmt.Println(response)
}
