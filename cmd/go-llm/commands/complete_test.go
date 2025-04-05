package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/zerobang-dev/go-llm/pkg/llm"
)

// Mock execution of the command without actually calling external APIs
func mockCommand(t *testing.T, args []string, wantErr bool, contains []string) {
	// Store original models map to restore after tests
	origModelToProvider := llm.ModelToProvider
	defer func() { llm.ModelToProvider = origModelToProvider }()

	// Set up test models
	llm.ModelToProvider = map[string]string{
		"claude-3-7-sonnet-latest": "anthropic",
		"deepseek-chat":            "deepseek",
		"test-model":               "test",
	}

	// Create a fresh command for each test to avoid state between runs
	cmd := &cobra.Command{Use: "go-llm"}
	cmd.AddCommand(completeCmd)
	cmd.PersistentFlags().StringVarP(&modelFlag, "model", "m", "claude-3-7-sonnet-latest", "LLM model to use")

	// Modify completeCmd to just print the values for testing
	originalRunE := completeCmd.RunE
	defer func() { completeCmd.RunE = originalRunE }()

	completeCmd.RunE = func(cmd *cobra.Command, args []string) error {
		var prompt string
		if len(args) == 1 {
			prompt = args[0]
		} else {
			return nil // Skip stdin reading in tests
		}

		// Check for query all flag
		if queryAllFlag {
			fmt.Println("Querying all providers")
			fmt.Printf("Prompt: %s\n", prompt)
			fmt.Printf("Temperature: %.1f\n", temperatureFlag)
			if systemPromptFlag != "" {
				fmt.Printf("System: %s\n", systemPromptFlag)
			}

			// Mock table output
			fmt.Println("PROVIDER  MODEL                   RESPONSE")
			fmt.Println("--------  -----                   --------")
			fmt.Println("anthropic claude-3-7-sonnet-latest Mock Anthropic response")
			fmt.Println("deepseek  deepseek-chat           Mock Deepseek response")

			// Mock detailed responses
			fmt.Println("\nDetailed responses:")
			fmt.Println("------------------")
			fmt.Println("## anthropic (claude-3-7-sonnet-latest)")
			fmt.Println("This is a detailed mock response from Anthropic.")
			fmt.Println("## deepseek (deepseek-chat)")
			fmt.Println("This is a detailed mock response from Deepseek.")
		} else {
			// Just print the values for testing without calling APIs
			model := modelFlag
			fmt.Printf("Using model: %s\n", model)
			fmt.Printf("Prompt: %s\n", prompt)
			fmt.Printf("Temperature: %.1f\n", temperatureFlag)
			if systemPromptFlag != "" {
				fmt.Printf("System: %s\n", systemPromptFlag)
			}
			fmt.Println("Response: This is a simulated test response.")
		}

		return nil
	}

	// Capture stdout
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	// Execute command
	cmd.SetArgs(args)
	err := cmd.Execute()

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if (err != nil) != wantErr {
		t.Errorf("Execute() error = %v, wantErr %v", err, wantErr)
		return
	}

	// Skip content check for error cases
	if wantErr {
		return
	}

	// Check output contains expected strings
	for _, want := range contains {
		if !strings.Contains(output, want) {
			t.Errorf("Output should contain %q but got: %s", want, output)
		}
	}
}

func TestCompleteCommand(t *testing.T) {
	// Test cases
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		contains []string
	}{
		{
			name:    "Valid prompt",
			args:    []string{"complete", "This is a test prompt"},
			wantErr: false,
			contains: []string{
				"Using model: claude-3-7-sonnet-latest",
				"Prompt: This is a test prompt",
				"Response: This is a simulated test response.",
			},
		},
		{
			name:    "Valid prompt with model flag",
			args:    []string{"complete", "--model", "test-model", "This is a test prompt"},
			wantErr: false,
			contains: []string{
				"Using model: test-model",
				"Prompt: This is a test prompt",
			},
		},
		{
			name:    "Valid prompt with short model flag",
			args:    []string{"complete", "-m", "test-model", "This is a test prompt"},
			wantErr: false,
			contains: []string{
				"Using model: test-model",
				"Prompt: This is a test prompt",
			},
		},
		{
			name:    "With temperature flag",
			args:    []string{"complete", "-t", "0.9", "This is a test prompt"},
			wantErr: false,
			contains: []string{
				"Temperature: 0.9",
			},
		},
		{
			name:    "With system prompt",
			args:    []string{"complete", "-s", "You are a helpful assistant", "This is a test prompt"},
			wantErr: false,
			contains: []string{
				"System: You are a helpful assistant",
			},
		},
		{
			name:    "With query all flag",
			args:    []string{"complete", "--all", "This is a test prompt"},
			wantErr: false,
			contains: []string{
				"Querying all providers",
				"Prompt: This is a test prompt",
				"PROVIDER  MODEL                   RESPONSE",
				"anthropic claude-3-7-sonnet-latest Mock Anthropic response",
				"deepseek  deepseek-chat           Mock Deepseek response",
				"Detailed responses:",
			},
		},
		{
			name:    "With query all short flag",
			args:    []string{"complete", "-a", "This is a test prompt"},
			wantErr: false,
			contains: []string{
				"Querying all providers",
				"Prompt: This is a test prompt",
				"PROVIDER  MODEL                   RESPONSE",
			},
		},
		{
			name:    "With query all and system prompt",
			args:    []string{"complete", "-a", "-s", "You are a helpful assistant", "This is a test prompt"},
			wantErr: false,
			contains: []string{
				"Querying all providers",
				"System: You are a helpful assistant",
				"PROVIDER  MODEL                   RESPONSE",
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCommand(t, tt.args, tt.wantErr, tt.contains)
		})
	}
}
