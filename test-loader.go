package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadTestsFromFile loads tests from a text file containing shell commands
func LoadTestsFromFile(filename string) (TestCategory, error) {
	// Extract category name from filename
	base := filepath.Base(filename)
	categoryName := strings.TrimSuffix(base, filepath.Ext(base))

	file, err := os.Open(filename)
	if err != nil {
		return TestCategory{}, fmt.Errorf("failed to open test file %s: %w", filename, err)
	}
	defer file.Close()

	// Create test category with default description
	category := TestCategory{
		Name:        categoryName,
		Description: fmt.Sprintf("Tests for %s commands", categoryName),
		Tests:       []TestCase{},
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue // Skip empty lines
		}

		// Create test case
		testCase := TestCase{
			Command:     line,
			Description: "", // No description for simple text files
			Skip:        false,
		}

		category.Tests = append(category.Tests, testCase)
	}

	if err := scanner.Err(); err != nil {
		return TestCategory{}, fmt.Errorf("error reading test file: %w", err)
	}

	return category, nil
}

// LoadTestsFromJSON loads tests from a JSON file with more metadata
func LoadTestsFromJSON(filename string) (TestCategory, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return TestCategory{}, fmt.Errorf("failed to read JSON file %s: %w", filename, err)
	}

	var category TestCategory
	if err := json.Unmarshal(file, &category); err != nil {
		return TestCategory{}, fmt.Errorf("failed to parse JSON file %s: %w", filename, err)
	}

	return category, nil
}

// LoadAllTestCategories loads all test categories from the tests directory
func LoadAllTestCategories() ([]TestCategory, error) {
	var categories []TestCategory

	// Define the tests directory
	testsDir := "./tests"

	// Check if directory exists
	if _, err := os.Stat(testsDir); os.IsNotExist(err) {
		// Create tests directory if it doesn't exist
		if err := os.MkdirAll(testsDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create tests directory: %w", err)
		}

		// Create default test files if directory was just created
		if err := createDefaultTestFiles(testsDir); err != nil {
			return nil, fmt.Errorf("failed to create default test files: %w", err)
		}
	}

	// Walk through the tests directory
	err := filepath.Walk(testsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		var category TestCategory
		var loadErr error

		// Load test file based on extension
		ext := filepath.Ext(path)
		switch ext {
		case ".json":
			category, loadErr = LoadTestsFromJSON(path)
		case ".txt", "":
			category, loadErr = LoadTestsFromFile(path)
		default:
			// Skip files with unknown extensions
			return nil
		}

		if loadErr != nil {
			fmt.Printf("Warning: Failed to load test file %s: %v\n", path, loadErr)
			return nil // Continue with other files
		}

		// Add category to the list
		categories = append(categories, category)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking tests directory: %w", err)
	}

	return categories, nil
}

// CreateDefaultTestFiles creates default test files in the tests directory
func createDefaultTestFiles(testsDir string) error {
	// Create empty_prompt.txt
	emptyPromptTests := []string{
		"",
		" ",
		"                                          ",
		":", // return code without errors 1 for some reason
		"!", // return code 0 ?????????
	}

	if err := createTestFile(testsDir, "empty_prompt.txt", emptyPromptTests); err != nil {
		return err
	}

	// Create parsing_errors.txt
	parsingErrorsTests := []string{
		">",
		"<",
		">>",
		"<<",
		"<>",
		">>>>>",
		">>>>>>>>>>>>>>>",
		"<<<<<",
		"<<<<<<<<<<<<<<<",
		"> > > >",
		">> >> >> >>",
		">>>> >> >> >>",
		"|",
		"| bonjour",
		"| | |",
		"||",
		"|||||",
		"|||||||||||||",
		">>|><",
		"&&",
		"&&&&&",
		"&&&&&&&&&&&&&&",
	}

	if err := createTestFile(testsDir, "parsing_errors.txt", parsingErrorsTests); err != nil {
		return err
	}

	// Create path_dir.txt
	pathDirTests := []string{
		"/",
		"//",
		"/.",
		"/./../../../../..",
		"///////",
	}

	if err := createTestFile(testsDir, "path_dir.txt", pathDirTests); err != nil {
		return err
	}

	// Create cmd_not_found.txt
	cmdNotFoundTests := []string{
		"\"bonjour\"",
		"bonjour'",
		"bonjour",
		"bonjour comment va",
		"Makefile",
	}

	if err := createTestFile(testsDir, "cmd_not_found.txt", cmdNotFoundTests); err != nil {
		return err
	}

	// Create echo.txt
	echoTests := []string{
		"echo",
		"echo -n",
		"echo Hola",
		"echoHola",
		"echo-nHola",
		"echo -n Hola",
		"echo \"-n\" Hola",
		"echo -nHola",
		"echo Hola -n",
		"echo Hola Que Tal",
		"echo         Hola",
		"echo    Hola     Que    Tal",
		"echo      \\n hola",
		"echo \"         \" | cat -e",
		"echo           | cat -e",
		"\"\"''echo hola\"\"'''' que\"\"'' tal\"\"''",
		"echo -n -n",
		"echo -n -n Hola Que",
		"echo -p",
		"echo -nnnnn",
		"echo -n -nnn -nnnn",
		"echo -n-nnn -nnnn",
		"echo -n -nnn hola -nnnn",
		"echo -n -nnn-nnnn",
		"echo --------n",
		"echo -nnn --------n",
		"echo -nnn -----nn---nnnn",
		"echo -nnn --------nnnn",
		"echo $",
		"echo $?",
		"echo $?$",
		"echo $? | echo $? | echo $?",
		"echo $:$= | cat -e",
		"echo \" $ \" | cat -e",
		"echo ' $ ' | cat -e",
		"echo $HOME",
		"echo \\$HOME",
		"echo my shit terminal is [$TERM]",
		"echo my shit terminal is [$TERM4",
		"echo my shit terminal is [$TERM4]",
		"echo $UID",
		"echo $HOME9",
		"echo $9HOME",
		"echo $HOME%",
		"echo $UID$HOME",
		"echo Le path de mon HOME est $HOME",
		"echo $USER$var\\$USER$USER\\$USERtest$USER",
		"echo $hola*",
		"echo -nnnn $hola",
		"echo > <",
		"echo | |",
		"EechoE",
		".echo.",
		">echo>",
		"<echo<",
		">>echo>>",
		"|echo|",
		"|echo -n hola",
		"echo *",
		"echo '*'",
		"echo D*",
		"echo *Z",
		"echo *t hola",
		"echo *t",
		"echo $*",
		"echo hola*hola *",
		"echo $hola*",
		"echo $HOME*",
		"echo $\"\"",
		"echo \"$\"\"\"",
		"echo '$'''",
		"echo $\"HOME\"",
		"echo $''HOME",
		"echo $\"\"HOME",
		"echo \"$HO\"ME",
		"echo '$HO'ME",
		"echo \"$HO\"\"ME\"",
		"echo '$HO''ME'",
		"echo \"'$HO''ME'\"",
		"echo \"\"$HOME",
		"echo \"\" $HOME",
		"echo ''$HOME",
		"echo '' $HOME",
		"echo $\"HO\"\"ME\"",
		"echo $'HO''ME'",
		"echo $'HOME'",
		"echo \"$\"HOME",
		"echo $=HOME",
		"echo $\"HOLA\"",
		"echo $'HOLA'",
		"echo $DONTEXIST Hola",
		"echo \"hola\"",
		"echo 'hola'",
		"echo ''hola''",
		"echo ''h'o'la''",
		"echo \"''h'o'la''\"",
		"echo \"'\"h'o'la\"'\"",
		"echo\"'hola'\"",
		"echo \"'hola'\"",
		"echo '\"hola\"'",
		"echo '''ho\"''''l\"a'''",
		"echo hola\"\"\"\"\"\"\"\"\"\"\"\"",
		"echo hola\"''''''''''\"",
		"echo hola''''''''''''",
		"echo hola'\"\"\"\"\"\"\"\"\"\"'",
		"e\"cho hola\"",
		"e'cho hola'",
		"echo \"hola     \" | cat -e",
		"echo \"\"hola",
		"echo \"\" hola",
		"echo \"\"             hola",
		"echo \"\"hola",
		"echo \"\" hola",
		"echo hola\"\"bonjour",
		"\"e\"'c'ho 'b'\"o\"nj\"o\"'u'r",
		"\"\"e\"'c'ho 'b'\"o\"nj\"o\"'u'r\"",
		"echo \"$DONTEXIST\"Makefile",
		"echo \"$DONTEXIST\"\"Makefile\"",
		"echo \"$DONTEXIST\" \"Makefile\"",
	}

	if err := createTestFile(testsDir, "echo.txt", echoTests); err != nil {
		return err
	}

	// Create env.txt
	envTests := []string{
		"$?",
		"$?$?",
		"?$HOME",
		"$",
		"$HOME",
		"$HOMEdskjhfkdshfsd",
		"\"$HOMEdskjhfkdshfsd\"",
		"$HOMEdskjhfkdshfsd'",
		"$DONTEXIST",
		"$LESS$VAR",
		"env",
	}

	if err := createTestFile(testsDir, "env.txt", envTests); err != nil {
		return err
	}

	// Create export_unset.txt
	exportUnsetTests := []string{
		"\"export HOLA=bonjour",
		"env\"",
		"\"export       HOLA=bonjour",
		"env\"",
		"export",
		"\"export Hola",
		"export\"",
		"\"export Hola9hey",
		"export\"",
		"export $DONTEXIST",
		"export | grep \"HOME\"",
		"export \"\"",
		"export =",
		"export %",
		"export $?",
		"export ?=2",
		"export 9HOLA=",
		"\"export HOLA9=bonjour",
		"env\"",
		"\"export _HOLA=bonjour",
		"env\"",
		"\"export ___HOLA=bonjour",
		"env\"",
		"\"export _HO_LA_=bonjour",
		"env\"",
		"export HOL@=bonjour",
		"export HOL\\~A=bonjour",
		"export -HOLA=bonjour",
		"export --HOLA=bonjour",
		"export HOLA-=bonjour",
		"export HO-LA=bonjour",
		"export HOL.A=bonjour",
		"export HOL\\\\\\$A=bonjour",
		"export HO\\\\\\\\LA=bonjour",
		"export HOL}A=bonjour",
		"export HOL{A=bonjour",
		"export HO*LA=bonjour",
		"export HO#LA=bonjour",
		"export HO@LA=bonjour",
		"export HO!LA=bonjour",
		"\"export HO$?LA=bonjour",
		"env\"",
		"export +HOLA=bonjour",
		"export HOL+A=bonjour",
		"\"export HOLA+=bonjour",
		"env\"",
		"\"export HOLA=bonjour",
		"export HOLA+=bonjour",
		"env\"",
		"\"exportHOLA=bonjour",
		"env\"",
		"export HOLA =bonjour",
		"export HOLA = bonjour",
		"\"export HOLA=bon jour",
		"env\"",
		"\"export HOLA= bonjour",
		"env\"",
		"\"export HOLA=bonsoir",
		"export HOLA=bonretour",
		"export HOLA=bonjour",
		"env\"",
		"\"export HOLA=$HOME",
		"env\"",
		"\"export HOLA=bonjour$HOME",
		"env\"",
		"\"export HOLA=$HOMEbonjour",
		"env\"",
		"\"export HOLA=bon$jour",
		"env\"",
		"\"export HOLA=bon\\jour",
		"env\"",
		"\"export HOLA=bon\\\\jour",
		"env\"",
		"export HOLA=bon(jour",
		"export HOLA=bon()jour",
		"export HOLA=bon&jour",
		"\"export HOLA=bon@jour",
		"env\"",
		"\"export HOLA=bon;jour",
		"env\"",
		"export HOLA=bon!jour",
		"\"export HOLA=bon\"\"jour\"\"",
		"env\"",
		"\"export HOLA$USER=bonjour",
		"env\"",
		"\"export HOLA=bonjour=casse-toi",
		"echo $HOLA\"",
		"\"export \"\"HOLA=bonjour\"\"=casse-toi",
		"echo $HOLA\"",
		"\"export HOLA=bonjour",
		"export BYE=casse-toi",
		"echo $HOLA et $BYE\"",
		"\"export HOLA=bonjour BYE=casse-toi",
		"echo $HOLA et $BYE\"",
		"\"export A=a B=b C=c",
		"echo $A $B $C\"",
		"\"export $HOLA=bonjour",
		"env\"",
		"\"export HOLA=\"\"bonjour      \"\"  ",
		"echo $HOLA | cat -e\"",
		"\"export HOLA=\"\"   -n bonjour   \"\"  ",
		"echo $HOLA\"",
		"\"export HOLA=\"\"bonjour   \"\"/",
		"echo $HOLA\"",
		"\"export HOLA='\"\"'",
		"echo \"\" $HOLA \"\" | cat -e\"",
		"\"export HOLA=at",
		"c$HOLA Makefile\"",
		"\"export \"\"\"\" HOLA=bonjour",
		"env\"",
		"\"export HOLA=\"\"cat Makefile | grep NAME\"\"  ",
		"echo $HOLA\"",
		"\"export HOLA=hey ",
		"echo $HOLA$HOLA$HOLA=hey$HOLA\"",
		"\"export HOLA=\"\"  bonjour  hey  \"\"  ",
		"echo $HOLA | cat -e\"",
		"\"export HOLA=\"\"  bonjour  hey  \"\"  ",
		"echo \"\"\"\"\"\"$HOLA\"\"\"\"\"\" | cat -e\"",
		"\"export HOLA=\"\"  bonjour  hey  \"\"  ",
		"echo wesh\"\"$HOLA\"\" | cat -e\"",
		"\"export HOLA=\"\"  bonjour  hey  \"\"  ",
		"echo wesh\"\"\"\"$HOLA.\"",
		"\"export HOLA=\"\"  bonjour  hey  \"\"  ",
		"echo wesh$\"\"\"\"HOLA.\"",
		"\"export HOLA=\"\"  bonjour  hey  \"\"  ",
		"echo wesh$\"\"HOLA HOLA\"\".\"",
		"\"export HOLA=bonjour",
		"export HOLA=\"\" hola et $HOLA\"\"",
		"echo $HOLA\"",
		"\"export HOLA=bonjour",
		"export HOLA=' hola et $HOLA'",
		"echo $HOLA\"",
		"\"export HOLA=bonjour",
		"export HOLA=\"\" hola et $HOLA\"\"$HOLA",
		"echo $HOLA\"",
		"\"export HOLA=\"\"ls        -l    - a\"\"",
		"echo $HOLA\"",
		"\"export HOLA=\"\"s -la\"\" ",
		"l$HOLA\"",
		"\"export HOLA=\"\"s -la\"\" ",
		"l\"\"$HOLA\"\"\"",
		"\"export HOLA=\"\"s -la\"\" ",
		"l'$HOLA'\"",
		"\"export HOLA=\"\"l\"\" ",
		"$HOLAs\"",
		"\"export HOLA=\"\"l\"\" ",
		"\"\"$HOLA\"\"s\"",
		"\"export HOL=A=bonjour",
		"env\"",
		"\"export HOLA=\"\"l\"\" ",
		"'$HOLA's\"",
		"\"export HOL=A=\"\"\"\"",
		"env\"",
		"\"export TE+S=T",
		"env\"",
		"export \"\"=\"\"",
		"export ''=''",
		"export \"=\"=\"=\"",
		"export '='='='",
		"\"export HOLA=p",
		"export BYE=w",
		"$HOLA\"\"BYE\"\"d\"",
		"\"export HOLA=p",
		"export BYE=w",
		"\"\"$HOLA\"\"'$BYE'd\"",
		"\"export HOLA=p",
		"export BYE=w",
		"\"\"$HOLA\"\"\"\"$BYE\"\"d\"",
		"\"export HOLA=p",
		"export BYE=w",
		"$\"\"HOLA\"\"$\"\"BYE\"\"d\"",
		"\"export HOLA=p",
		"export BYE=w",
		"$'HOLA'$'BYE'd\"",
		"\"export HOLA=-n",
		"\"\"echo $HOLA\"\" hey\"",
		"\"export A=1 B=2 C=3 D=4 E=5 F=6 G=7 H=8",
		"echo \"\"$A'$B\"\"'$C\"\"$D'$E'\"\"$F'\"\"'$G'$H\"\"\"",
		"\"export HOLA=bonjour",
		"env",
		"unset HOLA",
		"env\"",
		"\"export HOLA=bonjour",
		"env",
		"unset HOLA",
		"unset HOLA",
		"env\"",
		"\"unset PATH",
		"echo $PATH\"",
		"\"unset PATH",
		"ls\"",
		"unset \"\"",
		"unset INEXISTANT",
		"\"unset PWD",
		"env | grep PWD",
		"pwd\"",
		"\"pwd",
		"unset PWD",
		"env | grep PWD",
		"cd $PWD",
		"pwd\"",
		"\"unset OLDPWD",
		"env | grep OLDPWD\"",
		"unset 9HOLA",
		"unset HOLA9",
		"unset HOL?A",
		"unset HOLA HOL?A",
		"unset HOL?A HOLA",
		"unset HOL?A HOL.A",
		"unset HOLA=",
		"unset HOL\\\\\\\\A",
		"unset HOL;A",
		"unset HOL.A",
		"unset HOL+A",
		"unset HOL=A",
		"unset HOL{A",
		"unset HOL}A",
		"unset HOL-A",
		"unset -HOLA",
		"unset _HOLA",
		"unset HOL_A",
		"unset HOLA_",
		"unset HOL*A",
		"unset HOL#A",
		"unset $HOLA",
		"unset $PWD",
		"unset HOL@",
		"unset HOL!A",
		"unset HOL^A",
		"unset HOL$?A",
		"unset HOL\\~A",
		"\"unset \"\"\"\" HOLA",
		"env | grep HOLA\"",
		"\"unset PATH",
		"echo $PATH\"",
		"\"unset PATH",
		"cat Makefile\"",
		"unset =",
		"unset ======",
		"unset ++++++",
		"unset _______",
		"unset export",
		"unset echo",
		"unset pwd",
		"unset cd",
		"unset unset",
		"unset sudo",
		"export hola | unset hola | echo $?",
	}

	if err := createTestFile(testsDir, "export_unset.txt", exportUnsetTests); err != nil {
		return err
	}

	// Create pipes.txt
	pipesTests := []string{
		"echo hello | cat",
		"echo hello | cat | grep hello",
		"ls | wc -l",
		"cat /etc/passwd | grep root | wc -l",
	}

	if err := createTestFile(testsDir, "pipes.txt", pipesTests); err != nil {
		return err
	}

	// Create redirects.txt
	redirectsTests := []string{
		"echo hello > ./outfiles/out1",
		"cat < ./test_files/infile",
		"ls >> ./outfiles/out2",
		"cat < ./test_files/infile > ./outfiles/out3",
	}

	if err := createTestFile(testsDir, "redirects.txt", redirectsTests); err != nil {
		return err
	}

	// Create syntax.txt
	syntaxTests := []string{
		"|",
		"| echo oi",
		"| |",
		">",
		">>",
		"<",
		"echo hi <",
		"echo hi | >",
	}

	if err := createTestFile(testsDir, "syntax.txt", syntaxTests); err != nil {
		return err
	}

	// Create example JSON file
	quotingCategory := TestCategory{
		Name:        "quoting",
		Description: "Tests for shell quoting behavior",
		Tests: []TestCase{
			{Command: "echo \"Double $USER quotes\"", Description: "Double quotes with expansion"},
			{Command: "echo 'Single $USER quotes'", Description: "Single quotes prevent expansion"},
			{Command: "echo \"Nested 'quotes'\"", Description: "Nested quotes"},
			{Command: "echo 'Nested \"quotes\"'", Description: "Nested quotes reversed"},
			{Command: "echo \"$HOME\"'$HOME'", Description: "Adjacent different quotes"},
		},
	}

	jsonData, err := json.MarshalIndent(quotingCategory, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filepath.Join(testsDir, "quoting.json"), jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

// CreateTestFile creates a test file with the given tests
func createTestFile(testsDir, filename string, tests []string) error {
	filePath := filepath.Join(testsDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, test := range tests {
		_, err := writer.WriteString(test + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to file %s: %w", filePath, err)
		}
	}

	return writer.Flush()
}
