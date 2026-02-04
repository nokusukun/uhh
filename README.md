# UHH - AI-Powered Terminal Command Assistant

[![CI](https://github.com/nokusukun/uhh/actions/workflows/ci.yaml/badge.svg)](https://github.com/nokusukun/uhh/actions/workflows/ci.yaml)
[![Release](https://github.com/nokusukun/uhh/actions/workflows/release.yaml/badge.svg)](https://github.com/nokusukun/uhh/actions/workflows/release.yaml)
[![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

UHH is an AI-powered CLI tool that generates shell commands from natural language descriptions. It supports multiple LLM providers and automatically detects your shell environment.

## Features

- **Multi-Provider Support**: OpenAI, Google Gemini, DeepSeek, Kimi (Moonshot), GLM (Zhipu AI)
- **Smart Shell Detection**: Automatically detects PowerShell, CMD, Bash, Zsh, Fish
- **Clipboard Integration**: Generated commands are automatically copied to clipboard
- **Command History**: Tracks your prompts and generated commands
- **"Actually" Revision**: Refine previous commands with additional context
- **File Context**: Optionally include referenced file contents in prompts
- **Interactive Onboarding**: Easy setup wizard for first-time configuration
- **Cross-Platform**: Binaries available for Linux, macOS, and Windows

## Installation

### One-Line Install (Recommended)

**Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/nokusukun/uhh/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/nokusukun/uhh/main/install.ps1 | iex
```

### Manual Download

Download the latest release for your platform from the [Releases](https://github.com/nokusukun/uhh/releases) page.

#### Linux

```bash
# AMD64
curl -LO https://github.com/nokusukun/uhh/releases/latest/download/uhh-linux-amd64.tar.gz
tar -xzf uhh-linux-amd64.tar.gz
sudo mv uhh /usr/local/bin/

# ARM64
curl -LO https://github.com/nokusukun/uhh/releases/latest/download/uhh-linux-arm64.tar.gz
tar -xzf uhh-linux-arm64.tar.gz
sudo mv uhh /usr/local/bin/
```

#### macOS

```bash
# Intel Mac
curl -LO https://github.com/nokusukun/uhh/releases/latest/download/uhh-darwin-amd64.tar.gz
tar -xzf uhh-darwin-amd64.tar.gz
sudo mv uhh /usr/local/bin/

# Apple Silicon
curl -LO https://github.com/nokusukun/uhh/releases/latest/download/uhh-darwin-arm64.tar.gz
tar -xzf uhh-darwin-arm64.tar.gz
sudo mv uhh /usr/local/bin/
```

#### Windows

Download `uhh-windows-amd64.zip` from releases, extract, and add to your PATH.

### Build from Source

```bash
git clone https://github.com/nokusukun/uhh.git
cd uhh
go build -o uhh ./cmd/uhh

# Install (Unix)
sudo mv uhh /usr/local/bin/
```

## Quick Start

### First-Time Setup

Run the interactive setup wizard:

```bash
uhh init
```

This will guide you through:
1. Selecting LLM providers
2. Entering API keys
3. Choosing default settings

### Basic Usage

Simply describe what you want to do:

```bash
# No quotes needed for simple prompts
uhh list all go files recursively

# Use quotes for complex prompts
uhh "find files larger than 100MB modified in the last week"

# The generated command is printed and copied to clipboard
```

### Examples

```bash
# File operations
uhh find all pdf files in downloads folder
uhh delete all node_modules folders recursively
uhh compress this directory into a zip file

# Git operations
uhh show uncommitted changes
uhh create a new branch called feature-login
uhh squash last 3 commits

# System operations
uhh show disk usage by folder
uhh find process using port 8080
uhh list all running docker containers
```

## Configuration

### Config File

Configuration is stored in `~/.uhh/config.json`:

```json
{
  "default_provider": "openai",
  "providers": {
    "openai": {
      "enabled": true,
      "api_key": "sk-...",
      "model": "gpt-4o",
      "temperature": 0.7
    },
    "gemini": {
      "enabled": false,
      "api_key": "",
      "model": "gemini-2.0-flash"
    },
    "deepseek": {
      "enabled": false,
      "api_key": "",
      "model": "deepseek-chat",
      "base_url": "https://api.deepseek.com/v1"
    },
    "kimi": {
      "enabled": false,
      "api_key": "",
      "model": "kimi-coding/k2p5"
    },
    "glm": {
      "enabled": false,
      "api_key": "",
      "model": "glm-4",
      "base_url": "https://open.bigmodel.cn/api/paas/v4"
    }
  },
  "agent": {
    "auto_approve": false,
    "max_iterations": 10,
    "enabled_tools": []
  },
  "shell": {
    "override": "",
    "append_file_context": false,
    "max_context_tokens": 1000
  },
  "ui": {
    "no_color": false
  }
}
```

### Environment Variables

Environment variables override config file settings:

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `GOOGLE_API_KEY` | Google Gemini API key |
| `DEEPSEEK_API_KEY` | DeepSeek API key |
| `MOONSHOT_API_KEY` | Kimi/Moonshot API key |
| `GLM_API_KEY` | GLM/Zhipu AI API key |
| `UHH_PROVIDER` | Override default provider |
| `UHH_MODEL` | Override model for current provider |
| `UHH_SHELL` | Override shell detection |
| `UHH_NO_COLOR` | Disable colored output |
| `UHH_AUTO_APPROVE` | Auto-approve tool executions (1/true) |
| `UHH_APPEND_SMALL_CONTEXT` | Include file contents (true/1/token_limit) |

## CLI Reference

```
Usage:
  uhh [prompt] [flags]
  uhh [command]

Available Commands:
  init        Initialize or reconfigure UHH
  config      Show current configuration
  update      Check for and install updates
  version     Print version information
  completion  Generate shell completion scripts
  help        Help about any command

Flags:
  -p, --provider string   LLM provider (openai, gemini, deepseek, kimi, glm)
  -m, --model string      Model to use
  -s, --shell string      Override shell (powershell, cmd, bash, zsh, fish)
  -y, --auto-approve      Auto-approve tool executions
  -a, --agent             Run in agent mode with tool calling
  -h, --help              Help for uhh
```

### Provider Selection

```bash
# Use default provider (from config)
uhh list files

# Use specific provider
uhh --provider gemini list files
uhh -p deepseek list files

# Use specific model
uhh --provider openai --model gpt-4-turbo list files
```

### Shell Override

```bash
# Auto-detect (default)
uhh list files

# Force specific shell
uhh --shell bash list files
uhh -s powershell list files
```

### "Actually" Revision

Refine your previous command:

```bash
uhh list all files
# Output: ls -la

uhh actually only show hidden files
# Output: ls -la | grep '^\.'
```

## Supported Providers

| Provider | Models | Notes |
|----------|--------|-------|
| OpenAI | gpt-4o, gpt-4-turbo, gpt-3.5-turbo | Default provider |
| Google Gemini | gemini-2.0-flash, gemini-pro | Requires GOOGLE_API_KEY |
| DeepSeek | deepseek-chat, deepseek-coder | OpenAI-compatible API |
| Kimi (Moonshot) | kimi-coding/k2p5, moonshot-v1-8k | Supports KimiCoding API |
| GLM (Zhipu AI) | glm-4, glm-3-turbo | OpenAI-compatible API |

### Getting API Keys

- **OpenAI**: [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
- **Google Gemini**: [aistudio.google.com/apikey](https://aistudio.google.com/apikey)
- **DeepSeek**: [platform.deepseek.com](https://platform.deepseek.com)
- **Kimi**: [platform.moonshot.ai](https://platform.moonshot.ai) or [kimi.com](https://kimi.com) for KimiCoding
- **GLM**: [open.bigmodel.cn](https://open.bigmodel.cn)

## Shell Support

| Shell | Auto-Detection | Override Names |
|-------|----------------|----------------|
| PowerShell | Yes | `powershell`, `pwsh`, `ps` |
| Command Prompt | Yes | `cmd`, `command` |
| Bash | Yes | `bash` |
| Zsh | Yes | `zsh` |
| Fish | Yes | `fish` |

## Advanced Features

### Agent Mode (Experimental)

Run commands interactively with tool calling:

```bash
# Enable agent mode
uhh --agent "find large files and show their sizes"

# With auto-approve (use with caution)
uhh --agent --auto-approve "clean up temp files"
```

### File Context

Include referenced file contents in prompts:

```bash
# Enable in config or via environment
export UHH_APPEND_SMALL_CONTEXT=true
uhh "parse the package.json and list dependencies"
```

### Updating

UHH can update itself to the latest version:

```bash
# Check for updates and install
uhh update

# Check current version
uhh version
```

### Shell Completions

```bash
# Bash
uhh completion bash > /etc/bash_completion.d/uhh

# Zsh
uhh completion zsh > "${fpath[1]}/_uhh"

# Fish
uhh completion fish > ~/.config/fish/completions/uhh.fish

# PowerShell
uhh completion powershell > uhh.ps1
```

## Project Structure

```
uhh/
├── cmd/uhh/main.go           # CLI entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── provider/             # LLM provider implementations
│   ├── tools/                # Tool definitions (bash, file operations)
│   ├── agent/                # Agent loop for tool calling
│   ├── tui/                  # Terminal UI components
│   ├── shell/                # Shell detection and prompts
│   ├── history/              # Command history
│   ├── output/               # Colored output utilities
│   └── updater/              # Self-update functionality
├── .github/workflows/        # CI/CD pipelines
├── go.mod
└── README.md
```

## Troubleshooting

### Common Issues

**"No API key found"**
```bash
# Set via environment
export OPENAI_API_KEY="sk-..."

# Or run setup
uhh init
```

**"Unknown provider"**
```bash
# Check available providers
uhh config
```

**Wrong shell commands generated**
```bash
# Override shell detection
uhh --shell bash "your prompt"
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [langchaingo](https://github.com/tmc/langchaingo) - Go LLM framework
- [Charm](https://charm.sh) - Beautiful TUI components
- [Cobra](https://github.com/spf13/cobra) - CLI framework
