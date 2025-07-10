# UHH - Terminal Command Autocorrect

UHH is an intelligent terminal command autocorrect system that uses AI to fix and suggest compatible shell commands. It automatically detects your shell environment and provides context-aware command suggestions.

## Features

- **Multi-shell support**: Automatically detects and supports PowerShell, CMD, Bash, Zsh, and Fish
- **Shell override**: Override detected shell for specific commands or globally
- **Smart file context**: Automatically includes small file contents when referenced in prompts
- **Command history**: Maintains history with "actually" revision feature
- **Clipboard integration**: Automatically copies suggested commands to clipboard
- **Configurable models**: Support for different OpenAI models

## Warning
*UHH is a powerful tool that can generate and execute shell commands. Use it with caution, especially when running commands that modify or delete files. Always review generated commands before executing them.*

## Installation

1. Clone the repository:
```bash
git clone https://github.com/nokusukun/uhh
cd uhh
```

2. Build the application:
```bash
go build -o uhh.exe
```

3. Set up your OpenAI API key (see [Configuration](#configuration))

## Usage

### Basic Usage

```bash
# Basic command correction
uhh list files in current directory

# Interactive mode (if no arguments provided)
uhh
> What do you want?
> list files recursively
```

### Shell Override

Override the detected shell for specific commands:

```bash
# Use bash commands regardless of current shell
uhh !shell=bash list files recursively

# Use PowerShell commands
uhh !shell=powershell get running processes

# Use CMD commands
uhh !shell=cmd "show directory contents"

# Alternative syntax
uhh --shell bash "find files with extension .txt"
```

### Command Revision with "Actually"

Revise the previous command with additional context:

```bash
uhh copy file to backup
# Output: cp file.txt backup/

uhh actually make it recursive
# Output: cp -r file.txt backup/
```

### File Context Integration

When you reference files in your prompt, UHH can automatically include their contents for better context:

```bash
# Enable file context and reference a config file
export UHH_APPEND_SMALL_CONTEXT=true
uhh install dependencies from package.json
```

## Configuration

### Environment Variables

#### `OPENAI_API_KEY`
- **Description**: Your OpenAI API key for accessing GPT models
- **Required**: Yes
- **Alternative**: Store in `~/.openai.token.txt` file
- **Example**: `export OPENAI_API_KEY="sk-..."`

#### `UHH_MODEL`
- **Description**: OpenAI model to use for command generation
- **Default**: `gpt-4o`
- **Options**: Any valid OpenAI model (`gpt-4`, `gpt-3.5-turbo`, etc.)
- **Example**: `export UHH_MODEL="gpt-4"`

#### `UHH_SHELL`
- **Description**: Override shell detection globally
- **Default**: Auto-detected
- **Options**: `powershell`, `cmd`, `bash`, `zsh`, `fish`
- **Example**: `export UHH_SHELL="bash"`

#### `UHH_APPEND_SMALL_CONTEXT`
- **Description**: Enable automatic inclusion of small file contents when referenced
- **Default**: Disabled
- **Options**: 
  - `true` or `1`: Enable with default 1000 token limit
  - `false` or `0`: Disable
  - Number (e.g., `500`): Enable with custom token limit
- **Example**: `export UHH_APPEND_SMALL_CONTEXT="true"`

#### `UHH_NO_COLOR` / `NO_COLOR`
- **Description**: Disable colored terminal output
- **Default**: Colors enabled
- **Options**: Any non-empty value disables colors
- **Example**: `export UHH_NO_COLOR="1"` or `export NO_COLOR="1"`
- **Note**: Supports both `UHH_NO_COLOR` and the standard `NO_COLOR` environment variable

### API Key Setup

#### Option 1: Environment Variable
```bash
# Windows (PowerShell)
$env:OPENAI_API_KEY = "sk-your-api-key-here"

# Windows (CMD)
set OPENAI_API_KEY=sk-your-api-key-here

# Linux/macOS
export OPENAI_API_KEY="sk-your-api-key-here"
```

#### Option 2: Token File
Create a file at `~/.openai.token.txt` containing only your API key:
```
sk-your-api-key-here
```

## Color Scheme

UHH uses colored output to improve readability and user experience:

- **Generated Commands**: Bright green and bold - the main command output
- **Success Messages**: Green - confirmation messages like "✓ Copied to clipboard!"
- **Info Messages**: Cyan - informational messages like revision notifications
- **Warning Messages**: Yellow - non-critical warnings
- **Error Messages**: Red and bold - error messages
- **Prompts**: Blue - interactive prompts like "What do you want?"

### Disabling Colors

Colors can be disabled using environment variables:

```bash
# Using UHH_NO_COLOR
export UHH_NO_COLOR="1"
uhh "list files"

# Using standard NO_COLOR
export NO_COLOR="1"
uhh "list files"

# Temporary disable (PowerShell)
$env:UHH_NO_COLOR = "1"; .\uhh.exe "list files"

# Temporary disable (CMD)
set UHH_NO_COLOR=1 && uhh "list files"
```

## Shell Support

### Supported Shells

| Shell | Auto-Detection | Override Names |
|-------|----------------|----------------|
| PowerShell | ✅ | `powershell`, `pwsh`, `ps` |
| Command Prompt | ✅ | `cmd`, `command` |
| Bash | ✅ | `bash` |
| Zsh | ✅ | `zsh` |
| Fish | ✅ | `fish` |

### Shell Detection Priority

1. **Command line argument** (`!shell=` or `--shell`) - Highest priority
2. **Environment variable** (`UHH_SHELL`) - Medium priority
3. **Auto-detection** - Lowest priority (fallback)

## File Context Feature

When `UHH_APPEND_SMALL_CONTEXT` is enabled, UHH automatically detects file references in your prompts and includes their contents for better context.

### Supported File Patterns

- Files with extensions: `file.txt`, `config.yaml`, `script.py`
- Quoted file paths: `"path/to/file.json"`, `'config.ini'`
- Common config files: `package.json`, `go.mod`, `Dockerfile`, `Makefile`, `README.md`, `.gitignore`

### Example Usage

```bash
export UHH_APPEND_SMALL_CONTEXT="true"

# References package.json - content will be included automatically
uhh "install the dependencies listed in package.json"

# Custom token limit
export UHH_APPEND_SMALL_CONTEXT="500"
uhh "build the project using the configuration in go.mod"
```

## Command History

UHH maintains a command history file at `~/.uhh.history.txt` that includes:
- Timestamps
- Shell used
- Original prompt
- Generated command

### History Revision

Use the "actually" feature to revise previous commands:

```bash
uhh "find files"
# Output: find . -type f

uhh "actually only .txt files"
# Output: find . -type f -name "*.txt"
```

## Examples

### Basic Commands

```bash
# File operations
uhh "copy file.txt to backup folder"
uhh "delete all .log files"
uhh "create a new directory called projects"

# Process management
uhh "kill process running on port 8080"
uhh "show all running processes"
uhh "restart nginx service"

# Git operations
uhh "add all files and commit with message 'initial commit'"
uhh "push to origin main branch"
uhh "create new branch called feature/auth"
```

### Shell-Specific Examples

```bash
# Force PowerShell syntax
uhh !shell=powershell "get all files larger than 100MB"
# Output: Get-ChildItem -Recurse | Where-Object {$_.Length -gt 100MB}

# Force Bash syntax
uhh !shell=bash "get all files larger than 100MB"
# Output: find . -type f -size +100M

# Force CMD syntax
uhh !shell=cmd "get all files larger than 100MB"
# Output: forfiles /s /m *.* /c "cmd /c if @fsize gtr 104857600 echo @path"
```

### File Context Examples

```bash
# With package.json context
export UHH_APPEND_SMALL_CONTEXT="true"
uhh "run the start script defined in package.json"

# With Dockerfile context
uhh "build docker image using the Dockerfile configuration"

# With go.mod context
uhh "install dependencies specified in go.mod"
```

## Troubleshooting

### Common Issues

#### API Key Issues
```
Error: Failed to initialize OpenAI client
```
- Ensure `OPENAI_API_KEY` is set or `~/.openai.token.txt` exists
- Verify your API key is valid and has sufficient credits

#### Shell Detection Issues
```
Commands don't match my shell
```
- Use `!shell=<shell_name>` to override detection
- Set `UHH_SHELL` environment variable for persistent override
- Check that your shell is supported

#### File Context Issues
```
Referenced files not being included
```
- Ensure `UHH_APPEND_SMALL_CONTEXT` is set to `true` or a number
- Check that files exist and are under the token limit (default 1000 tokens ≈ 4000 characters)
- Verify file paths are correct (relative paths are converted to absolute)

#### Color Display Issues
```
Colors not displaying properly or garbled output
```
- Some terminals may not support all color features
- Disable colors using `UHH_NO_COLOR=1` or `NO_COLOR=1`
- Ensure your terminal supports ANSI color codes
- On Windows, use Windows Terminal or PowerShell 7+ for best color support

### Debug Mode

To enable debug output during shell parsing, you can temporarily modify the `ParseShellOverride` function to include debug prints.

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

[Add your license information here]

## Changelog

### Current Version
- ✅ Multi-shell support with auto-detection
- ✅ **Colored terminal output** with configurable disable options
- ✅ Shell override via `!shell=` and `--shell` syntax
- ✅ File context integration with `UHH_APPEND_SMALL_CONTEXT`
- ✅ Command history and "actually" revision feature
- ✅ Configurable OpenAI models via `UHH_MODEL`
- ✅ Clipboard integration
- ✅ Environment variable configuration

## Support

For issues, feature requests, or questions, please [create an issue](link-to-issues) in the repository.
