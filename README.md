# Minishell Tester with External Test Files

A comprehensive testing framework for Minishell implementations, using external test files for better organization and maintainability.

## Features

- **External Test Files**: Tests are defined in separate text or JSON files
- **Memory Leak Detection**: Valgrind integration for memory leak and unclosed file descriptor detection
- **Comprehensive Comparison**: Compares stdout, stderr, and exit codes with bash
- **File Redirection Testing**: Tests file input/output redirection handling
- **Detailed Reporting**: Clear reporting of test failures with color-coded output

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/minishell-tester.git
cd minishell-tester

# Build the tester
make build
```

## Usage

```
./maybe [options]
```

### Options

| Option | Description |
|--------|-------------|
| `--minishell <path>` | Path to the minishell executable (default: "../minishell") |
| `--categories <list>` | Comma-separated list of test categories to run |
| `--verbose` | Enable verbose output |
| `--skip-valgrind` | Skip valgrind checks |
| `--show-leaks` | Show memory leak details (default: true) |
| `--show-fds` | Show unclosed file descriptors (default: true) |
| `--timeout <seconds>` | Timeout in seconds for each test (default: 10) |
| `--no-color` | Disable colored output |
| `--no-details` | Don't display detailed test failure information |
| `--list` | List available test categories |
| `--create-tests` | Create default test files in ./tests directory |
| `--version` | Show version information |
| `--help` | Show help message |

### Examples

```bash
# Run all tests
./maybe

# Run only builtins and pipes tests
./maybe --categories builtins,pipes

# Run tests without displaying detailed failure information
./maybe --no-details

# List available test categories
./maybe --list

# Create default test files
./maybe --create-tests
```

## Test Files

Tests are defined in the `./tests` directory. The tester supports two formats:

### Simple Text Files

Simple `.txt` files contain one shell command per line:

```
echo hello world
pwd
ls -la
```

The file name becomes the category name (e.g., `builtins.txt` becomes the "builtins" category).

### JSON Files

JSON files provide more control with descriptions and the ability to skip tests:

```json
{
  "Name": "quoting",
  "Description": "Tests for shell quoting behavior",
  "Tests": [
    {
      "Command": "echo \"Double $USER quotes\"",
      "Description": "Double quotes with expansion",
      "Skip": false
    },
    {
      "Command": "echo 'Single $USER quotes'",
      "Description": "Single quotes prevent expansion",
      "Skip": false
    }
  ]
}
```

## Creating Custom Test Categories

1. Create a new file in the `./tests` directory with either `.txt` or `.json` extension
2. For text files, add one shell command per line
3. For JSON files, follow the structure shown above
4. Run the tester with `--list` to verify your new category is recognized

## Makefile Commands

- `make build`: Build the tester
- `make clean`: Clean up build artifacts
- `make test`: Build and run all tests
- `make list`: Build and list available test categories
- `make create-tests`: Create default test files

## Troubleshooting

- If tests fail with "command not found" errors, check if your minishell binary is correctly located at "../minishell"
- For valgrind-related errors, ensure valgrind is installed on your system
- If no test categories are found, try running `./maybe --create-tests` to create default test files