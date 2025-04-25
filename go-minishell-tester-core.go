package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/fatih/color"
)

const (
	// ASCII art banner
	AsciiLogo = `
 @@@@@@ @@@  @@@ @@@@@@@@ @@@      @@@           @@@@@@@@@@  @@@@@@@@      @@@@@@@@@@   @@@@@@  @@@ @@@ @@@@@@@  @@@@@@@@
!@@     @@!  @@@ @@!      @@!      @@!           @@! @@! @@! @@!           @@! @@! @@! @@!  @@@ @@! !@@ @@!  @@@ @@!
 !@@!!  @!@!@!@! @!!!:!   @!!      @!!           @!! !!@ @!@ @!!!:!        @!! !!@ @!@ @!@!@!@!  !@!@!  @!@!@!@  @!!!:!
    !:! !!:  !!! !!:      !!:      !!:           !!:     !!: !!:           !!:     !!: !!:  !!!   !!:   !!:  !!! !!:
::.: :   :   : : : :: ::: : ::.: : : ::.: :       :      :   : :: :::       :      :    :   : :   .:    :: : ::  : :: :::
`
)

var (
	colorBoldBlue   = color.New(color.FgBlue, color.Bold)
	colorBoldRed    = color.New(color.FgRed, color.Bold)
	colorBoldYellow = color.New(color.FgYellow, color.Bold)
	colorGray       = color.RGB(127, 127, 127)
	colorGreen      = color.New(color.FgGreen)
	colorBold       = color.New(color.Bold)
)

// TestCase defines a single shell command test
type TestCase struct {
	Command     string // The shell command to test
	Description string // Optional description of what is being tested
	Skip        bool   // Whether to skip this test
}

// TestCategory groups related tests together
type TestCategory struct {
	Name        string     // Name of the category (builtins, pipes, etc.)
	Description string     // Description of this test category
	Tests       []TestCase // Tests in this category
}

// Configuration options
type Config struct {
	MinishellPath   string
	Categories      []string // Categories to test (empty means all)
	OutfilesDir     string
	MiniOutDir      string
	BashOutDir      string
	Verbose         bool
	SkipValgrind    bool
	ShowLeaks       bool
	ShowOpenFDs     bool
	Timeout         time.Duration
	ValgrindTimeout time.Duration
	TmpDir          string
	NoColor         bool
	MaxOutputLength int
	NoDetails       bool
}

// Results of a single test
type TestResult struct {
	Command      string
	Passed       bool
	MiniOutput   string
	BashOutput   string
	MiniExitCode int
	BashExitCode int
	MiniErrorMsg string
	BashErrorMsg string
	OutfilesDiff string
	HasLeaks     bool
	HasOpenFDs   bool
	TimeTaken    time.Duration
	Error        error
}

// Helper to remove ANSI color codes from output
func removeColors(s string) string {
	re := regexp.MustCompile("\x1B\\[[0-9;]{1,}[A-Za-z]")
	return re.ReplaceAllString(s, "")
}

// Get the minishell prompt string
func getPrompt(minishellPath string) (string, error) {
	// Run minishell and get the initial prompt before any commands
	cmd := exec.Command("bash", "-c", fmt.Sprintf("echo -e '\\nexit\\n' | %s | head -n 1", minishellPath))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get prompt: %w", err)
	}

	// Remove colors from the prompt
	cleanPrompt := removeColors(string(out))
	// Trim any whitespace
	cleanPrompt = strings.TrimSpace(cleanPrompt)

	// If the prompt is empty or just contains whitespace, try a fallback method
	if cleanPrompt == "" {
		// Try another approach - assuming the prompt ends with a space and a special character
		cmd = exec.Command("bash", "-c", fmt.Sprintf("echo -e '\\n' | %s | head -n 1", minishellPath))
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("failed to get prompt with fallback: %w", err)
		}

		cleanPrompt = removeColors(string(out))
		cleanPrompt = strings.TrimSpace(cleanPrompt)

		// If still empty, use a default fallback
		if cleanPrompt == "" {
			return "$", nil
		}
	}

	// Look for common prompt patterns ($ prompt)
	if !strings.Contains(cleanPrompt, "$") {
		// Search for the last non-alphanumeric character that might be the prompt symbol
		for i := len(cleanPrompt) - 1; i >= 0; i-- {
			if !isAlphaNumeric(rune(cleanPrompt[i])) && !unicode.IsSpace(rune(cleanPrompt[i])) {
				// Found a potential prompt character
				return cleanPrompt, nil
			}
		}
	}

	return cleanPrompt, nil
}

// Helper function to check if a character is alphanumeric
func isAlphaNumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// Clean a directory by removing all files
func cleanDir(dir string) error {
	// Ensure directory exists first
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Read all entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Remove each entry
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}

	return nil
}

// Copy all files from one directory to another
func copyFiles(srcDir, dstDir string) error {
	// Ensure destination exists
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	// Get all files in source
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	// Copy each file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		src := filepath.Join(srcDir, entry.Name())
		dst := filepath.Join(dstDir, entry.Name())

		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}

		if err := os.WriteFile(dst, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

// Compare two directories and return the differences
func compareDirs(dir1, dir2 string) (string, error) {
	cmd := exec.Command("diff", "--brief", dir1, dir2)
	output, err := cmd.CombinedOutput()

	// diff returns exit code 1 when differences are found
	if err != nil && err.(*exec.ExitError).ExitCode() != 1 {
		return "", fmt.Errorf("diff command failed: %w", err)
	}

	return string(output), nil
}

// Run valgrind to check for memory leaks and open file descriptors
func runValgrindCheck(config *Config, command string) (bool, bool, error) {
	if config.SkipValgrind {
		return false, false, nil
	}

	// Create valgrind command with appropriate options
	valgrindCmd := []string{
		"valgrind",
		"--leak-check=full",
		"--show-leak-kinds=all",
		"--track-fds=yes",
		"--track-origins=yes",
		"--errors-for-leak-kinds=all",
		"--suppression=readline.supp",
		config.MinishellPath,
	}

	cmd := exec.Command(valgrindCmd[0], valgrindCmd[1:]...)

	// Setup stdin for input
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false, false, err
	}

	// Capture stderr for analysis
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		return false, false, err
	}

	// Write command and exit
	if _, err := io.WriteString(stdin, command+"\nexit\n"); err != nil {
		// Try to kill the process if writing fails
		cmd.Process.Kill()
		return false, false, err
	}
	stdin.Close()

	// Use the separate valgrind timeout from config
	timeout := config.ValgrindTimeout
	if timeout == 0 {
		// If not set, use double the regular timeout as a fallback
		timeout = config.Timeout * 2
	}

	// Wait with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		// Try to kill the process gracefully first
		cmd.Process.Signal(os.Interrupt)

		// Give it a brief moment to terminate
		select {
		case <-done:
			// Process exited after SIGINT
		case <-time.After(500 * time.Millisecond):
			// Force kill if still running
			cmd.Process.Kill()
		}

		return false, false, fmt.Errorf("valgrind timed out after %s", timeout)
	case err := <-done:
		if err != nil && !strings.Contains(err.Error(), "exit status") {
			return false, false, err
		}
	}

	// Check for memory leaks
	valgrindOutput := stderr.String()
	hasLeaks := strings.Contains(valgrindOutput, "definitely lost") ||
		strings.Contains(valgrindOutput, "indirectly lost") ||
		strings.Contains(valgrindOutput, "possibly lost") ||
		strings.Contains(valgrindOutput, "still reachable")

	// Check for open file descriptors
	hasOpenFDs := strings.Contains(valgrindOutput, "file descriptors are left open")

	// Save detailed valgrind output if requested
	if config.Verbose && (hasLeaks || hasOpenFDs) {
		logDir := filepath.Join(config.TmpDir, "valgrind_logs")
		if err := os.MkdirAll(logDir, 0755); err == nil {
			// Create a safe filename from the command
			safeFilename := strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
					return r
				}
				return '_'
			}, command)

			if len(safeFilename) > 50 {
				safeFilename = safeFilename[:50]
			}

			logFile := filepath.Join(logDir, safeFilename+".log")
			os.WriteFile(logFile, []byte(valgrindOutput), 0644)
		}
	}

	return hasLeaks, hasOpenFDs, nil
}

// Run a single test and return the results
func runTest(config *Config, prompt string, test TestCase) TestResult {
	startTime := time.Now()
	result := TestResult{
		Command: test.Command,
	}

	// Skip test if marked
	if test.Skip {
		result.Error = fmt.Errorf("test skipped")
		return result
	}

	// Clean output directories
	if err := cleanDir(config.OutfilesDir); err != nil {
		result.Error = fmt.Errorf("failed to clean outfiles dir: %w", err)
		return result
	}

	if err := cleanDir(config.MiniOutDir); err != nil {
		result.Error = fmt.Errorf("failed to clean mini outfiles dir: %w", err)
		return result
	}

	if err := cleanDir(config.BashOutDir); err != nil {
		result.Error = fmt.Errorf("failed to clean bash outfiles dir: %w", err)
		return result
	}

	// Run minishell command with timeout protection
	miniCmd := exec.Command("bash", "-c", fmt.Sprintf("echo -e \"%s\" | %s 2>/tmp/mini_stderr.txt",
		strings.ReplaceAll(test.Command, "\"", "\\\""),
		config.MinishellPath))

	// Create a channel to signal command completion
	miniDone := make(chan error, 1)
	var miniOutput []byte

	// Run command in goroutine
	go func() {
		var err error
		miniOutput, err = miniCmd.Output()
		miniDone <- err
	}()

	// Wait for command or timeout
	var miniErr error
	select {
	case miniErr = <-miniDone:
		// Command completed normally
		if miniErr != nil {
			// Store exit code if available
			if exitErr, ok := miniErr.(*exec.ExitError); ok {
				result.MiniExitCode = exitErr.ExitCode()
			}
		} else {
			result.MiniExitCode = 0
		}
	case <-time.After(config.Timeout):
		// Command timed out, kill it
		if miniCmd.Process != nil {
			miniCmd.Process.Kill()
		}
		result.Error = fmt.Errorf("minishell command timed out after %s", config.Timeout)
		result.MiniOutput = "COMMAND TIMED OUT"
		result.MiniExitCode = -1 // Use -1 to indicate timeout
		return result
	}

	// Process minishell output
	miniOutputStr := removeColors(string(miniOutput))

	// Improved prompt handling - remove all lines with the prompt
	if prompt != "" {
		// Split into lines, filter out prompt lines and exit lines
		lines := strings.Split(miniOutputStr, "\n")
		var filteredLines []string

		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			// Skip lines that only contain the prompt or exit
			if !strings.HasPrefix(trimmedLine, prompt) &&
				!strings.Contains(trimmedLine, "$ exit") &&
				trimmedLine != "exit" {
				filteredLines = append(filteredLines, line)
			}
		}

		miniOutputStr = strings.Join(filteredLines, "\n")
	}

	result.MiniOutput = strings.TrimSpace(miniOutputStr)

	// Copy minishell outfiles
	if err := copyFiles(config.OutfilesDir, config.MiniOutDir); err != nil {
		result.Error = fmt.Errorf("failed to copy mini outfiles: %w", err)
		return result
	}

	// Get minishell error message
	miniErrorBytes, err := os.ReadFile("/tmp/mini_stderr.txt")
	if err == nil {
		// Extract relevant part of error message
		miniErrorMsg := string(miniErrorBytes)
		if len(miniErrorMsg) > 0 {
			parts := strings.Split(miniErrorMsg, ":")
			if len(parts) > 1 {
				result.MiniErrorMsg = strings.TrimSpace(parts[len(parts)-1])
			} else {
				result.MiniErrorMsg = strings.TrimSpace(miniErrorMsg)
			}
		}
	}

	// Clean outfiles directory for bash test
	if err := cleanDir(config.OutfilesDir); err != nil {
		result.Error = fmt.Errorf("failed to clean outfiles dir: %w", err)
		return result
	}

	// Run bash command with timeout protection
	bashCmd := exec.Command("bash", "-c", fmt.Sprintf("echo -e \"%s\" | bash 2>/tmp/bash_stderr.txt",
		strings.ReplaceAll(test.Command, "\"", "\\\"")))

	// Create a channel to signal bash command completion
	bashDone := make(chan error, 1)
	var bashOutput []byte

	// Run bash command in goroutine
	go func() {
		var err error
		bashOutput, err = bashCmd.Output()
		bashDone <- err
	}()

	// Wait for bash command or timeout
	var bashErr error
	select {
	case bashErr = <-bashDone:
		// Command completed normally
		if bashErr != nil {
			// Store exit code if available
			if exitErr, ok := bashErr.(*exec.ExitError); ok {
				result.BashExitCode = exitErr.ExitCode()
			}
		} else {
			result.BashExitCode = 0
		}
	case <-time.After(config.Timeout):
		// Bash command timed out, kill it
		if bashCmd.Process != nil {
			bashCmd.Process.Kill()
		}
		result.Error = fmt.Errorf("bash command timed out after %s", config.Timeout)
		result.BashOutput = "COMMAND TIMED OUT"
		result.BashExitCode = -1 // Use -1 to indicate timeout
		return result
	}

	result.BashOutput = strings.TrimSpace(string(bashOutput))

	// Copy bash outfiles
	if err := copyFiles(config.OutfilesDir, config.BashOutDir); err != nil {
		result.Error = fmt.Errorf("failed to copy bash outfiles: %w", err)
		return result
	}

	// Get bash error message
	bashErrorBytes, err := os.ReadFile("/tmp/bash_stderr.txt")
	if err == nil {
		// Extract relevant part of error message
		bashErrorMsg := string(bashErrorBytes)
		if len(bashErrorMsg) > 0 {
			parts := strings.Split(bashErrorMsg, ":")
			if len(parts) > 1 {
				result.BashErrorMsg = strings.TrimSpace(parts[len(parts)-1])
			} else {
				result.BashErrorMsg = strings.TrimSpace(bashErrorMsg)
			}
		}
	}

	// Compare outfiles
	outfilesDiff, err := compareDirs(config.MiniOutDir, config.BashOutDir)
	if err != nil {
		result.Error = fmt.Errorf("failed to compare outfiles: %w", err)
		return result
	}
	result.OutfilesDiff = outfilesDiff

	// Check for memory leaks and open file descriptors with timeout handling
	hasLeaks, hasOpenFDs, err := runValgrindCheck(config, test.Command)
	if err != nil && !config.SkipValgrind {
		result.Error = fmt.Errorf("valgrind check failed: %w", err)
		return result
	}
	result.HasLeaks = hasLeaks
	result.HasOpenFDs = hasOpenFDs

	// Determine if test passed
	outputMatches := result.MiniOutput == result.BashOutput
	exitCodeMatches := result.MiniExitCode == result.BashExitCode
	noOutfileDiff := result.OutfilesDiff == ""
	noMemoryIssues := !result.HasLeaks && !result.HasOpenFDs

	if config.SkipValgrind {
		result.Passed = outputMatches && exitCodeMatches && noOutfileDiff
	} else {
		result.Passed = outputMatches && exitCodeMatches && noOutfileDiff && noMemoryIssues
	}

	// Record time taken
	result.TimeTaken = time.Since(startTime)

	return result
}

// Run tests for a category
func runCategoryTests(config *Config, prompt string, category TestCategory) ([]TestResult, error) {
	var results []TestResult

	fmt.Printf("Running %s: %s\n",
		colorBoldBlue.Sprint(category.Name),
		colorGray.Sprint(category.Description),
	)

	dotsPerLine := 50 // Number of progress dots per line
	currentDots := 0  // Counter for dots on current line
	totalTests := len(category.Tests)

	for i, test := range category.Tests {
		if config.Verbose {
			fmt.Printf("  Running test %d/%d: %s\n", i+1, totalTests, test.Command)
		}

		result := runTest(config, prompt, test)
		results = append(results, result)

		// Show progress in non-verbose mode
		if !config.Verbose {
			if result.Passed {
				colorGreen.Print(".")
			} else if result.Error != nil && strings.Contains(result.Error.Error(), "skipped") {
				colorBoldYellow.Print("s")
			} else {
				colorBoldRed.Print("F")
			}

			currentDots++

			// Line break after dotsPerLine dots or on the last test
			if currentDots >= dotsPerLine && i+1 < totalTests {
				// Just print a newline, no count yet
				fmt.Println()
				currentDots = 0 // Reset dot counter
			}
		} else if !result.Passed && !config.NoDetails {
			// In verbose mode, print failures immediately unless NoDetails is set
			printTestFailure(config, &result, i+1, category.Name)
		}
	}

	// Only print the final count after all tests have completed
	if !config.Verbose {
		// Count passed tests
		passed := 0
		for _, r := range results {
			if r.Passed {
				passed++
			}
		}

		// Calculate how many spaces we need for alignment
		spacesNeeded := 0
		if currentDots < dotsPerLine {
			spacesNeeded = dotsPerLine - currentDots
		}

		// Print the final pass count aligned to the right
		colorGray.Printf("%s %d/%d\n",
			strings.Repeat(" ", spacesNeeded),
			passed,
			totalTests)
	}

	return results, nil
}

// Print the details of a failed test
func printTestFailure(config *Config, result *TestResult, testNum int, categoryName string) {
	// Maximum length for displayed outputs
	const maxOutputLength = 1000
	const maxErrorLength = 500

	fmt.Printf("%s %s%s %s %s\n",
		colorBoldYellow.Sprint("Test"),
		colorBoldBlue.Sprint(categoryName),
		colorGray.Sprintf("#%d:", testNum),
		colorBoldRed.Sprint("✗"),
		colorGray.Sprint(result.Command))

	if result.Error != nil {
		fmt.Printf("Error: %s\n", truncateString(result.Error.Error(), maxErrorLength))
		// Add a separator line for better readability when showing multiple failures
		colorGray.Println(strings.Repeat("─", 54))
		return
	}

	// Display output mismatch in a more readable format
	if result.MiniOutput != result.BashOutput {
		colorBold.Println("Output mismatch:")

		// Count lines in both outputs
		miniLines := 0
		if result.MiniOutput != "" {
			miniLines = len(strings.Split(result.MiniOutput, "\n"))
		}

		bashLines := 0
		if result.BashOutput != "" {
			bashLines = len(strings.Split(result.BashOutput, "\n"))
		}

		// Use a different format for longer outputs
		if miniLines > 3 || bashLines > 3 {
			// Format and possibly truncate minishell output
			miniFormatted := formatOutputForDisplay(result.MiniOutput, maxOutputLength,
				colorBold.Sprint("minishell output"))

			// Format and possibly truncate bash output
			bashFormatted := formatOutputForDisplay(result.BashOutput, maxOutputLength,
				colorBold.Sprint("bash output"))

			// Display both outputs
			fmt.Printf("  %s\n", miniFormatted)
			fmt.Printf("  %s\n", bashFormatted)
		} else {
			// Simple format for shorter outputs
			fmt.Printf("  minishell: %s\n", result.MiniOutput)
			fmt.Printf("  bash:      %s\n", result.BashOutput)
		}
	}

	if result.MiniExitCode != result.BashExitCode {
		colorBold.Println("Exit code mismatch:")
		fmt.Printf("  minishell: %d\n", result.MiniExitCode)
		fmt.Printf("  bash:      %d\n", result.BashExitCode)
	}

	if result.MiniErrorMsg != result.BashErrorMsg {
		colorBold.Println("Exit message mismatch:")
		fmt.Printf("  minishell: %s\n", truncateString(result.MiniErrorMsg, maxErrorLength))
		fmt.Printf("  bash:      %s\n", truncateString(result.BashErrorMsg, maxErrorLength))
	}

	if result.OutfilesDiff != "" {
		colorBold.Printf("Outfiles difference:\n%s\n", truncateString(result.OutfilesDiff, maxOutputLength))
	}

	if result.HasLeaks && config.ShowLeaks {
		fmt.Printf("%s %s Memory leaks detected %s\n",
			colorBold.Sprint("❗"),
			colorBoldRed.Sprint("Memory leaks detected"),
			colorGray.Sprint(""))
	}

	if result.HasOpenFDs && config.ShowOpenFDs {
		fmt.Printf("%s %s Unclosed file descriptors detected %s\n",
			colorBold.Sprint("❗"),
			colorBoldRed.Sprint("Unclosed file descriptors detected"),
			colorGray.Sprint(""))
	}

	// Add a separator line using the box-drawing character
	fmt.Printf("%s\n", colorGray.Sprint(strings.Repeat("─", 50)))
}

// Print summary of test results
func printSummary(config *Config, categoryResults map[string][]TestResult) int {
	var allResults []TestResult
	var failedResults []struct {
		CategoryName string
		TestIndex    int
		Result       TestResult
	}

	// Collect all results and track failed tests
	for categoryName, results := range categoryResults {
		allResults = append(allResults, results...)

		// Track failed tests with their category name and index
		for i, result := range results {
			if !result.Passed && (result.Error == nil || !strings.Contains(result.Error.Error(), "skipped")) {
				failedResults = append(failedResults, struct {
					CategoryName string
					TestIndex    int
					Result       TestResult
				}{
					CategoryName: categoryName,
					TestIndex:    i + 1,
					Result:       result,
				})
			}
		}
	}

	// Count passed, failed, and skipped tests
	total := len(allResults)
	passed := 0
	failed := 0
	skipped := 0

	for _, result := range allResults {
		if result.Passed {
			passed++
		} else if result.Error != nil && strings.Contains(result.Error.Error(), "skipped") {
			skipped++
		} else {
			failed++
		}
	}

	// Print summary header
	colorBold.Println("\nTEST SUMMARY")
	fmt.Printf("%s\n", colorGray.Sprint(strings.Repeat("─", 50)))

	// Print category breakdown
	fmt.Println("Category Results:")
	for category, results := range categoryResults {
		catPassed := 0
		catFailed := 0
		catSkipped := 0

		for _, r := range results {
			if r.Passed {
				catPassed++
			} else if r.Error != nil && strings.Contains(r.Error.Error(), "skipped") {
				catSkipped++
			} else {
				catFailed++
			}
		}

		statusColor := colorGreen
		if catFailed > 0 {
			statusColor = colorBoldRed
		} else if catSkipped > 0 {
			statusColor = colorBoldYellow
		}

		fmt.Printf("  %s: %s%d passed%s",
			colorBoldBlue.Sprint(category),
			statusColor.Sprint(""),
			catPassed,
			colorGray.Sprint(""))

		if catFailed > 0 {
			fmt.Printf(", %s%d failed%s",
				colorBoldRed.Sprint(""),
				catFailed,
				colorGray.Sprint(""))
		}

		if catSkipped > 0 {
			fmt.Printf(", %s%d skipped%s",
				colorBoldYellow.Sprint(""),
				catSkipped,
				colorGray.Sprint(""))
		}

		colorGray.Printf(" (total: %d)\n", len(results))
	}

	var myColor *color.Color
	if passed == total {
		myColor = colorGreen
	} else if passed > 0 {
		myColor = colorBoldYellow
	} else {
		myColor = colorBoldRed
	}

	// Print overall result
	passRate := float64(passed) / float64(total) * 100
	fmt.Printf("\n%s: %s%d/%d tests passed (%.2f%%)%s\n",
		colorBold.Sprint("Overall"),
		myColor.Sprintf(""),
		passed,
		total,
		passRate,
		colorGray.Sprint(""))

	if skipped > 0 {
		colorBoldYellow.Printf("%d tests skipped\n", skipped)
	}

	if failed > 0 {
		colorBoldRed.Printf("%d tests failed\n", failed)

		// Print details of failed tests when not in verbose mode and NoDetails is not set
		if !config.Verbose && !config.NoDetails && len(failedResults) > 0 {
			colorBoldRed.Println("\nFAILED TESTS DETAILS")
			fmt.Printf("%s\n", colorGray.Sprint(strings.Repeat("─", 50)))

			// Sort failedResults by category for better organization
			sort.Slice(failedResults, func(i, j int) bool {
				if failedResults[i].CategoryName == failedResults[j].CategoryName {
					return failedResults[i].TestIndex < failedResults[j].TestIndex
				}
				return failedResults[i].CategoryName < failedResults[j].CategoryName
			})

			// Display details for each failed test
			for _, failedTest := range failedResults {
				printTestFailure(config, &failedTest.Result, failedTest.TestIndex, failedTest.CategoryName)
			}
		} else if config.NoDetails && failed > 0 {
			// When NoDetails is set, just print a message that details are being suppressed
			colorBoldYellow.Println("\nTest failure details are suppressed (--no-details flag is set)")
			fmt.Printf("Re-run without the --no-details flag to see detailed failure information\n")
		}

		return 1 // Failure
	} else {
		fmt.Println("All tests passed successfully!")
		return 0 // Success
	}
}

// Setup test environment
func setupTestEnvironment(config *Config) error {
	// Create test files directory if it doesn't exist
	testFilesDir := filepath.Join(".", "test_files")
	if err := os.MkdirAll(testFilesDir, 0755); err != nil {
		return fmt.Errorf("failed to create test_files directory: %w", err)
	}

	// Create invalid permission file for tests
	invalidPermFile := filepath.Join(testFilesDir, "invalid_permission")
	if _, err := os.Stat(invalidPermFile); os.IsNotExist(err) {
		if err := os.WriteFile(invalidPermFile, []byte("test"), 0644); err != nil {
			return fmt.Errorf("failed to create invalid_permission file: %w", err)
		}
	}

	// Set strict permissions
	if err := os.Chmod(invalidPermFile, 0000); err != nil {
		return fmt.Errorf("failed to set permissions on invalid_permission file: %w", err)
	}

	// Create infile for redirect tests
	infile := filepath.Join(testFilesDir, "infile")
	if _, err := os.Stat(infile); os.IsNotExist(err) {
		content := "hi\nhello\nworld\n42\n"
		if err := os.WriteFile(infile, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create infile: %w", err)
		}
	}

	// Create larger file for big redirects
	infileBig := filepath.Join(testFilesDir, "infile_big")
	if _, err := os.Stat(infileBig); os.IsNotExist(err) {
		// Use a paragraph of lorem ipsum as content
		content := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed non risus. Suspendisse lectus tortor, dignissim sit amet, adipiscing nec, ultricies sed, dolor. Cras elementum ultrices diam. Maecenas ligula massa, varius a, semper congue, euismod non, mi. Proin porttitor, orci nec nonummy molestie, enim est eleifend mi, non fermentum diam nisl sit amet erat. Duis semper. Duis arcu massa, scelerisque vitae, consequat in, pretium a, enim. Pellentesque congue. Ut in risus volutpat libero pharetra tempor. Cras vestibulum bibendum augue. Praesent egestas leo in pede. Praesent blandit odio eu enim. Pellentesque sed dui ut augue blandit sodales. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Aliquam nibh. Mauris ac mauris sed pede pellentesque fermentum. Maecenas adipiscing ante non diam sodales hendrerit.`
		if err := os.WriteFile(infileBig, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create infile_big: %w", err)
		}
	}

	// Create output directories
	for _, dir := range []string{config.OutfilesDir, config.MiniOutDir, config.BashOutDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// Cleanup test environment
func cleanupTestEnvironment(config *Config) {
	// Restore permissions on invalid_permission file
	invalidPermFile := filepath.Join(".", "test_files", "invalid_permission")
	if err := os.Chmod(invalidPermFile, 0666); err != nil {
		fmt.Printf("Warning: Failed to restore permissions on %s: %v\n", invalidPermFile, err)
	}

	// Remove output directories
	for _, dir := range []string{config.OutfilesDir, config.MiniOutDir, config.BashOutDir} {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Printf("Warning: Failed to clean up directory %s: %v\n", dir, err)
		}
	}
}

// Truncate a string to a maximum length, adding "..." if truncated
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	// For very short strings, just truncate with "..."
	if maxLength <= 10 {
		return s[:maxLength-3] + "..."
	}

	// For longer strings, try to truncate at a line boundary if possible
	lines := strings.Split(s, "\n")
	var result strings.Builder
	length := 0

	for i, line := range lines {
		// Check if adding this line would exceed the max length
		if length+len(line)+1 > maxLength-5 { // Account for "...\n..."
			// We've reached our limit
			if i == 0 {
				// If even the first line is too long, truncate it
				result.WriteString(line[:maxLength-5])
				result.WriteString("...")
			} else {
				// Otherwise, add "..." to indicate there's more
				result.WriteString("\n...")
			}
			break
		}

		// Add line to the result
		if i > 0 {
			result.WriteString("\n")
			length++
		}
		result.WriteString(line)
		length += len(line)
	}

	return result.String()
}

// Format and potentially truncate output for display
func formatOutputForDisplay(output string, maxLength int, prefix string) string {
	// Remove trailing newlines for cleaner display
	output = strings.TrimRight(output, "\n")

	// Count lines in the output
	lineCount := 0
	if output != "" {
		lineCount = len(strings.Split(output, "\n"))
	}

	// If it's a single line or empty, no need for special formatting
	if lineCount <= 1 {
		return output
	}

	// For multiple lines, format with line count and possible truncation
	truncated := truncateString(output, maxLength)

	// If truncation happened, indicate the original line count
	if truncated != output {
		return fmt.Sprintf("%s (%d lines, truncated):\n%s",
			prefix, lineCount, truncated)
	}

	return fmt.Sprintf("%s (%d lines):\n%s", prefix, lineCount, truncated)
}
