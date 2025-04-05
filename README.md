# gollm

A command-line interface for interacting with Large Language Models (LLMs).

## Features

- Get completions from LLMs directly from your terminal
- Support for specifying different models via flags
- Configure temperature and system prompts
- Compare multiple providers side-by-side
- Support for multiple providers (Anthropic Claude, Deepseek, Google Gemini)
- Configuration management via config file

## Installation

```bash
go install github.com/zerobang-dev/gollm@latest
```

Or build from source:

```bash
git clone https://github.com/zerobang-dev/gollm.git
cd gollm
go build -o gollm ./cmd/gollm
```

## Configuration

You can set up your API keys in two ways:

1. Using the `set` command (recommended):
   ```bash
   # For Anthropic
   gollm set anthropic --api-key your_anthropic_api_key_here
   
   # For Deepseek
   gollm set deepseek --api-key your_deepseek_api_key_here
   
   # For Google
   gollm set google --api-key your_google_api_key_here
   ```

2. Using environment variables:
   ```bash
   # For Anthropic
   export ANTHROPIC_API_KEY=your_anthropic_api_key_here
   
   # For Deepseek
   export DEEPSEEK_API_KEY=your_deepseek_api_key_here
   
   # For Google
   export GOOGLE_API_KEY=your_google_api_key_here
   ```

The configuration is stored in `~/.config/gollm/config.yml`.

## Usage

```bash
# Get a completion with default model (claude-3-7-sonnet-latest)
gollm "Write a haiku about programming"

# Using Anthropic's Claude
gollm -m claude-3-7-sonnet-latest "Write a haiku about programming"

# Using Deepseek models
gollm -m deepseek-chat "Explain quantum computing"
gollm -m deepseek-coder "Write a function to sort an array in Go"

# Using Google Gemini models
gollm -m gemini-2.5-pro-exp-03-25 "Write a function to merge two sorted arrays"
gollm -m gemini-1.5-flash "Explain parallel computing"

# Adjust the temperature
gollm -t 0.9 "Write a creative story"

# Add a system prompt
gollm -s "You are a helpful coding assistant" "How do I implement a binary search in Go?"

# Query all configured providers and compare responses
gollm -a "Explain the difference between slices and arrays in Go"

# Combine flags
gollm -a -t 0.8 -s "You are a Go expert" "What are the best practices for error handling in Go?"
```

## Supported Models

### Anthropic
- `claude-3-7-sonnet-latest`

### Deepseek
- `deepseek-chat`
- `deepseek-coder`

### Google Gemini
- `gemini-2.5-pro-exp-03-25`
- `gemini-2.0-flash`
- `gemini-2.0-flash-lite`
- `gemini-1.5-flash`
- `gemini-1.5-flash-8b`

## Command-line Options

- `-m, --model`: Specify the model to use
- `-t, --temperature`: Set the temperature for response generation (0.0-1.0)
- `-s, --system`: Provide a system prompt for context
- `-a, --all`: Query all configured providers and compare responses side-by-side

## Future Features

- Interactive chat sessions
- Streaming responses in real-time
- Support for more LLM providers
- Conversation history and context management

## License

MIT