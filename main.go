package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	appName    = "Shell Me Maybe"
	appVersion = "v1.0.0"
	appAuthor  = "Erwann Lagouche"
	appYear    = "2025"
)

func main() {
	// Command line flags
	var (
		minishellPath       = flag.String("minishell", "./minishell", "Path to the minishell executable")
		categoriesFlag      = flag.String("categories", "", "Comma-separated list of test categories to run")
		verbose             = flag.Bool("verbose", false, "Enable verbose output")
		skipValgrind        = flag.Bool("skip-valgrind", false, "Skip valgrind checks")
		showLeaks           = flag.Bool("show-leaks", true, "Show memory leak details")
		showOpenFDs         = flag.Bool("show-fds", true, "Show unclosed file descriptors")
		timeoutSecs         = flag.Int("timeout", 5, "Timeout in seconds for each test")
		valgrindTimeoutSecs = flag.Int("valgrind-timeout", 10, "Timeout in seconds for valgrind tests")
		version             = flag.Bool("version", false, "Show version information")
		listCategories      = flag.Bool("list", false, "List available test categories and exit")
		createTestsOnly     = flag.Bool("create-tests", false, "Create default test files and exit")
		maxOutputLength     = flag.Int("max-output", 1000, "Maximum length for displayed command outputs")
		noDetails           = flag.Bool("no-details", false, "Don't display detailed test failure information")
	)

	flag.Parse()

	if *version {
		fmt.Printf("%s %s\n© %s %s\n", appName, appVersion, appAuthor, appYear)
		os.Exit(0)
	}

	// Create tests directory and default test files if requested
	if *createTestsOnly {
		testsDir := "./tests"
		if err := os.MkdirAll(testsDir, 0755); err != nil {
			fmt.Printf("Error creating tests directory: %v\n", err)
			os.Exit(1)
		}

		if err := createDefaultTestFiles(testsDir); err != nil {
			fmt.Printf("Error creating default test files: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Default test files created in ./tests directory")
		os.Exit(0)
	}

	// Load all test categories
	allCategories, err := LoadAllTestCategories()
	if err != nil {
		fmt.Printf("Error loading test categories: %v\n", err)
		os.Exit(1)
	}

	if *listCategories {
		fmt.Println("Available test categories:")
		for _, category := range allCategories {
			fmt.Printf("  %s - %s (%d tests)\n",
				category.Name,
				category.Description,
				len(category.Tests))
		}
		os.Exit(0)
	}

	// Parse categories to run
	var requestedCategories []string
	if *categoriesFlag != "" {
		requestedCategories = strings.Split(*categoriesFlag, ",")
	}

	// Create configuration
	config := &Config{
		MinishellPath:   *minishellPath,
		Categories:      requestedCategories,
		OutfilesDir:     "./outfiles",
		MiniOutDir:      "./mini_outfiles",
		BashOutDir:      "./bash_outfiles",
		Verbose:         *verbose,
		SkipValgrind:    *skipValgrind,
		ShowLeaks:       *showLeaks,
		ShowOpenFDs:     *showOpenFDs,
		Timeout:         time.Duration(*timeoutSecs) * time.Second,
		ValgrindTimeout: time.Duration(*valgrindTimeoutSecs) * time.Second,
		TmpDir:          os.TempDir(),
		MaxOutputLength: *maxOutputLength,
		NoDetails:       *noDetails,
	}

	// Support for bonus tests if the first category is "bonus" or "wildcards"
	if len(requestedCategories) > 0 && (requestedCategories[0] == "bonus" || requestedCategories[0] == "wildcards") {
		config.MinishellPath = "../minishell_bonus"
	}

	color.Magenta(AsciiLogo)
	color.Magenta("%s%s (%s)\n\n", strings.Repeat(" ", 48), appName, appVersion)

	// Setup test environment
	if err := setupTestEnvironment(config); err != nil {
		color.Red("Error setting up test environment: %v\n", err)
		os.Exit(1)
	}
	defer cleanupTestEnvironment(config)

	// Get minishell prompt
	prompt, err := getPrompt(config.MinishellPath)
	if err != nil {
		fmt.Printf("Error getting minishell prompt: %v\n", err)
		// Continue with empty prompt - this is not a fatal error
	}

	// Filter test categories based on user selection
	var categoriesToRun []TestCategory
	if len(config.Categories) == 0 {
		categoriesToRun = allCategories
	} else {
		for _, category := range allCategories {
			for _, requestedName := range config.Categories {
				if category.Name == requestedName {
					categoriesToRun = append(categoriesToRun, category)
					break
				}
			}
		}
	}

	if len(categoriesToRun) == 0 {
		fmt.Println("No test categories found matching the specified criteria")
		os.Exit(1)
	}

	// Run tests for each category
	categoryResults := make(map[string][]TestResult)

	for _, category := range categoriesToRun {
		results, err := runCategoryTests(config, prompt, category)
		if err != nil {
			fmt.Printf("Error running tests for category %s: %v\n", category.Name, err)
			continue
		}

		categoryResults[category.Name] = results
	}

	// Print summary and exit with appropriate code
	exitCode := printSummary(config, categoryResults)
	os.Exit(exitCode)
}
