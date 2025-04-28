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
		"env|\"wc \"-l",
		"expr 1 + 1",
		"expr $? + $?",
		"\"env -i ./minishell",
		"env\"",
		"\"env -i ./minishell",
		"export\"",
		"\"env -i ./minishell",
		"cd\"",
		"\"env -i ./minishell",
		"cd ~\"",
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

	// Create pwd.txt
	pwdTests := []string{
		"pwd",
		"pwd hola",
		"pwd ./hola",
		"pwd hola que tal",
		"pwd -p",
		"pwd --p",
		"pwd ---p",
		"pwd -- p",
		"pwd pwd pwd",
		"pwd ls",
		"pwd ls env",
	}

	if err := createTestFile(testsDir, "pwd.txt", pwdTests); err != nil {
		return err
	}

	// Create path.txt
	cdTests := []string{
		"cd",
		"cd .",
		"cd ./",
		"cd ./././.",
		"cd ././././",
		"cd ..",
		"cd ../",
		"cd ../..",
		"cd ../.",
		"cd .././././.",
		"cd srcs",
		"cd srcs objs",
		"cd 'srcs'",
		"cd \"srcs\"",
		"cd '/etc'",
		"cd /e'tc'",
		"cd /e\"tc\"",
		"cd sr",
		"cd Makefile",
		"cd ../minishell",
		"cd ../../../../../../..",
		"cd .././../.././../bin/ls",
		"cd /",
		"cd '/'",
		"\"cd //",
		"pwd\"",
		"\"cd '//'",
		"pwd\"",
		"\"cd ///",
		"pwd\"",
		"\"cd ////////",
		"pwd\"",
		"\"cd '////////'",
		"pwd\"",
		"cd /minishell",
		"\"cd /",
		"cd ..\"",
		"cd _",
		"cd -",
		"cd --",
		"cd ---",
		"cd $HOME",
		"cd $HOME $HOME",
		"cd $HOME/42_works",
		"cd \"$PWD/srcs\"",
		"cd '$PWD/srcs'",
		"\"unset HOME",
		"cd $HOME\"",
		"\"unset HOME",
		"export HOME=",
		"cd\"",
		"\"unset HOME",
		"export HOME",
		"cd\"",
		"cd minishell Docs crashtest.c",
		"\"   cd / | echo $?",
		"pwd\"",
		"cd ~",
	}

	if err := createTestFile(testsDir, "cd.txt", cdTests); err != nil {
		return err
	}

	// Create path.txt
	pathTests := []string{
		"\"mkdir a",
		"mkdir a/b",
		"cd a/b",
		"rm -r ../../a",
		"cd ..\"",
		"\"mkdir a",
		"mkdir a/b",
		"cd a/b",
		"rm -r ../../a",
		"pwd\"",
		"\"mkdir a",
		"mkdir a/b",
		"cd a/b",
		"rm -r ../../a",
		"echo $PWD",
		"echo $OLDPWD\"",
		"\"mkdir a",
		"mkdir a/b",
		"cd a/b",
		"rm -r ../../a",
		"cd",
		"echo $PWD",
		"echo $OLDPWD\"",
		"\"mkdir a",
		"cd a",
		"rm -r ../a",
		"echo $PWD",
		"echo $OLDPWD\"",
		"\"export CDPATH=/",
		"cd $HOME/..\"",
		"\"export CDPATH=/",
		"cd home/vietdu91\"",
		"\"export CDPATH=./",
		"cd .\"",
		"\"export CDPATH=./",
		"cd ..\"",
		"\"chmod 000 minishell",
		"./minishell\"",
		"ls hola",
		"./Makefile",
		"./minishell",
		"\"env | grep SHLVL",
		"./minishell",
		"env | grep SHLVL",
		"exit",
		"env | grep SHLVL\"",
		"\"touch hola",
		"./hola\"",
	}

	if err := createTestFile(testsDir, "path.txt", pathTests); err != nil {
		return err
	}

	// Create exit.txt
	exitTests := []string{
		"exit",
		"exit exit",
		"exit hola",
		"exit hola que tal",
		"exit 42",
		"exit 000042",
		"exit 666",
		"exit 666 666",
		"exit -666 666",
		"exit hola 666",
		"exit 666 666 666 666",
		"exit 666 hola 666",
		"exit hola 666 666",
		"exit 259",
		"exit -4",
		"exit -42",
		"exit -0000042",
		"exit -259",
		"exit -666",
		"exit +666",
		"exit 0",
		"exit +0",
		"exit -0",
		"exit +42",
		"exit -69 -96",
		"exit --666",
		"exit ++++666",
		"exit ++++++0",
		"exit ------0",
		"exit \"666\"",
		"exit '666'",
		"exit '-666'",
		"exit '+666'",
		"exit '----666'",
		"exit '++++666'",
		"exit '6'66",
		"exit '2'66'32'",
		"exit \"'666'\"",
		"exit '\"666\"'",
		"exit '666'\"666\"666",
		"exit +'666'\"666\"666",
		"exit -'666'\"666\"666",
		"exit 9223372036854775807",
		"exit 9223372036854775808",
		"exit -9223372036854775808",
		"exit -9223372036854775809",
	}

	if err := createTestFile(testsDir, "exit.txt", exitTests); err != nil {
		return err
	}

	// Create pipes.txt
	pipesTests := []string{
		"echo hello | cat",
		"echo hello | cat | grep hello",
		"ls | wc -l",
		"cat /etc/passwd | grep root | wc -l",
		"cat | cat | cat | ls",
		"ls | exit",
		"ls | exit 42",
		"exit | ls",
		"\"echo hola > bonjour",
		"exit | cat -e bonjour\"",
		"\"echo hola > bonjour",
		"cat -e bonjour | exit\"",
		"echo | echo",
		"echo hola | echo que tal",
		"pwd | echo hola",
		"env | echo hola",
		"echo oui | cat -e",
		"echo oui | echo non | echo hola | grep oui",
		"echo oui | echo non | echo hola | grep non",
		"echo oui | echo non | echo hola | grep hola",
		"echo hola | cat -e | cat -e | cat -e",
		"cd .. | echo \"hola\"",
		"cd / | echo \"hola\"",
		"cd .. | pwd",
		"ifconfig | grep \":\"",
		"ifconfig | grep hola",
		"whoami | grep $USER",
		"\"whoami | grep $USER > /tmp/bonjour",
		"cat /tmp/bonjour\"",
		"\"whoami | cat -e | cat -e > /tmp/bonjour",
		"cat /tmp/bonjour\"",
		"\"whereis ls | cat -e | cat -e > /tmp/bonjour",
		"cat /tmp/bonjour\"",
		"ls | hola",
		"ls | ls hola",
		"ls | ls | hola",
		"ls | hola | ls",
		"ls | ls | hola | rev",
		"ls | ls | echo hola | rev",
		"ls -la | grep \".\"",
		"ls -la | grep \"'.'\"",
		"echo test.c | cat -e| cat -e| cat -e| cat -e| cat -e| cat -e| cat -e| cat -e|cat -e|cat -e|cat -e",
		"\"ls|ls|ls|ls|ls|ls|ls|ls|ls|ls|ls|ls",
		"|ls|ls|ls|ls|ls|ls|ls|ls|ls|ls|ls|ls|ls|ls|ls|ls\"",
		"echo hola | cat | cat | cat | cat | cat | grep hola",
		"echo hola | cat",
		"echo hola| cat",
		"echo hola |cat",
		"echo hola|cat",
		"echo hola || cat",
		"echo hola ||| cat",
		"ech|o hola | cat",
		"cat Makefile | cat -e | cat -e",
		"cat Makefile | grep srcs | cat -e",
		"cat Makefile | grep srcs | grep srcs | cat -e",
		"cat Makefile | grep pr | head -n 5 | cd file_not_exist",
		"cat Makefile | grep pr | head -n 5 | hello",
		"export HOLA=bonjour | cat -e | cat -e",
		"unset HOLA | cat -e",
		"\"export HOLA | echo hola",
		"env | grep PROUT\"",
		"export | echo hola",
		"sleep 3 | sleep 3",
		"time sleep 3 | sleep 3",
		"sleep 3 | exit",
		"exit | sleep 3",
		"\"echo hola > a",
		">>b echo que tal",
		"cat a | <b cat | cat > c | cat\"",
	}

	if err := createTestFile(testsDir, "pipes.txt", pipesTests); err != nil {
		return err
	}

	// Create redirects.txt
	redirectsTests := []string{
		"\"echo hola > bonjour",
		"cat bonjour\"",
		"\"echo que tal >> bonjour",
		"cat bonjour\"",
		"\"echo hola > bonjour",
		"echo que tal >> bonjour",
		"cat < bonjour\"",
		"\"echo hola > bonjour",
		"rm bonjour",
		"echo que tal >> bonjour",
		"cat < bonjour\"",
		"\"echo hola que tal > bonjour",
		"cat bonjour\"",
		"\"echo hola que tal > /tmp/bonjour",
		"cat -e /tmp/bonjour\"",
		"\"export HOLA=hey",
		"echo bonjour > $HOLA",
		"echo $HOLA\"",
		"\"whereis grep > Docs/bonjour",
		"cat Docs/bonjour\"",
		"\"ls -la > Docs/bonjour",
		"cat Docs/bonjour\"",
		"\"pwd>bonjour",
		"cat bonjour\"",
		"\"pwd >                     bonjour",
		"cat bonjour\"",
		"echo hola > > bonjour",
		"echo hola < < bonjour",
		"echo hola >>> bonjour",
		"\"> bonjour echo hola",
		"cat bonjour\"",
		"\"> bonjour | echo hola",
		"cat bonjour\"",
		"\"prout hola > bonjour",
		"ls\"",
		"\"echo hola > hello >> hello >> hello",
		"ls",
		"cat hello\"",
		"\"echo hola > hello >> hello >> hello",
		"echo hola >> hello",
		"cat < hello\"",
		"\"echo hola > hello >> hello >> hello",
		"echo hola >> hello",
		"echo hola > hello >> hello >> hello",
		"cat < hello\"",
		"\"echo hola >> hello >> hello > hello",
		"echo hola >> hello",
		"cat < hello\"",
		"\"echo hola >> hello >> hello > hello",
		"echo hola >> hello",
		"echo hola >> hello >> hello > hello",
		"cat < hello\"",
		"\"echo hola > hello",
		"echo hola >> hello >> hello >> hello",
		"echo hola >> hello",
		"cat < hello\"",
		"\"echo hola > hello",
		"echo hey > bonjour",
		"echo <bonjour <hello\"",
		"\"echo hola > hello",
		"echo hey > bonjour",
		"echo <hello <bonjour\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"rm bonjour hello",
		"echo hola > bonjour > hello > bonjour",
		"cat bonjour",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"echo hola > bonjour > hello > bonjour",
		"cat bonjour",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"rm bonjour hello",
		"echo hola > bonjour >> hello > bonjour",
		"cat bonjour",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"echo hola > bonjour > hello > bonjour",
		"cat bonjour",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"rm bonjour hello",
		"echo hola > bonjour > hello >> bonjour",
		"cat bonjour",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"echo hola > bonjour > hello >> bonjour",
		"cat bonjour",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"rm bonjour hello",
		"echo hola >> bonjour > hello > bonjour",
		"cat bonjour",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"echo hola >> bonjour > hello > bonjour",
		"cat bonjour",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"rm bonjour hello",
		"echo hola >> bonjour >> hello >> bonjour",
		"cat bonjour",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"echo hola >> bonjour >> hello >> bonjour",
		"cat bonjour",
		"cat hello\"",
		"\"> bonjour echo hola bonjour",
		"cat bonjour\"",
		"\">bonjour echo > hola>bonjour>hola>>bonjour>hola hey >bonjour hola >hola",
		"cat bonjour",
		"cat hola\"",
		"\"echo bonjour > hola1",
		"echo hello > hola2",
		"echo 2 >hola1 >> hola2",
		"ls",
		"cat hola1",
		"cat hola2\"",
		"\"echo bonjour > hola1",
		"echo hello > hola2",
		"echo 2 >>hola1 > hola2",
		"ls",
		"cat hola1",
		"cat hola2\"",
		"\"> pwd",
		"ls\"",
		"< pwd",
		"< Makefile .",
		"cat <pwd",
		"cat <srcs/pwd",
		"cat <../pwd",
		"cat >>",
		"cat >>>",
		"cat >> <<",
		"cat >> > >> << >>",
		"cat < ls",
		"cat < ls > ls",
		"\"cat > ls1 < ls2",
		"ls\"",
		"\">>hola",
		"cat hola\"",
		"\"echo hola > bonjour",
		"cat < bonjour\"",
		"\"echo hola >bonjour",
		"cat <bonjour\"",
		"\"echo hola>bonjour",
		"cat<bonjour\"",
		"\"echo hola> bonjour",
		"cat< bonjour\"",
		"\"echo hola               >bonjour",
		"cat<                     bonjour\"",
		"\"echo hola          >     bonjour",
		"cat            <         bonjour\"",
		"\"echo hola > srcs/bonjour",
		"cat < srcs/bonjour\"",
		"\"echo hola >srcs/bonjour",
		"cat <srcs/bonjour\"",
		"\"echo hola > bonjour",
		"echo que tal >> bonjour",
		"cat < bonjour\"",
		"\"echo hola > bonjour",
		"rm bonjour",
		"echo que tal >> bonjour",
		"cat < bonjour\"",
		"\"e'c'\"\"h\"\"o hola > bonjour",
		"cat 'bo'\"\"n\"\"jour\"",
		"\"echo hola > bonjour\\ 1",
		"ls",
		"cat bonjour\\ 1\"",
		"\"echo hola > bonjour hey",
		"ls",
		"cat bonjour",
		"cat hey\"",
		"\"echo hola > srcs/bonjour",
		"echo hey > srcs/hello",
		">srcs/bonjour >srcs/hello <prout",
		"cat srcs/bonjour srcs/hello\"",
		"\"echo hola > srcs/bonjour",
		"echo hey > srcs/hello",
		"rm srcs/bonjour srcs/hello",
		">srcs/bonjour >srcs/hello <prout",
		"ls srcs",
		"cat srcs/bonjour srcs/hello\"",
		"\"echo hola > srcs/bonjour",
		"echo hey > srcs/hello",
		">srcs/bonjour <prout >srcs/hello ",
		"cat srcs/bonjour ",
		"cat srcs/hello\"",
		"\"echo hola > srcs/bonjour",
		"echo hey > srcs/hello",
		"rm srcs/bonjour srcs/hello",
		">srcs/bonjour <prout >srcs/hello ",
		"ls srcs",
		"cat srcs/bonjour\"",
		"\"echo hola > ../bonjour",
		"echo hey > ../hello",
		">../bonjour >../hello <prout",
		"cat ../bonjour ../hello\"",
		"\"echo hola > ../bonjour",
		"echo hey > ../hello",
		"rm ../bonjour ../hello",
		">../bonjour >../hello <prout",
		"ls ..",
		"cat ../bonjour ../hello\"",
		"\"echo hola > ../bonjour",
		"echo hey > ../hello",
		">../bonjour <prout >../hello ",
		"cat ../bonjour ",
		"cat ../hello\"",
		"\"echo hola > ../bonjour",
		"echo hey > ../hello",
		"rm ../bonjour ../hello",
		">../bonjour <prout >../hello ",
		"ls ..",
		"cat ../bonjour\"",
		"\"echo hola > srcs/bonjour",
		"echo hey > srcs/hello",
		">srcs/bonjour >>srcs/hello <prout",
		"cat srcs/bonjour ",
		"cat srcs/hello\"",
		"\"echo hola > srcs/bonjour",
		"echo hey > srcs/hello",
		">>srcs/bonjour >srcs/hello <prout",
		"cat srcs/bonjour ",
		"cat srcs/hello\"",
		"\"echo hola > srcs/bonjour",
		"echo hey > srcs/hello",
		">>srcs/bonjour >>srcs/hello <prout",
		"cat srcs/bonjour ",
		"cat srcs/hello\"",
		"\"echo hola > srcs/bonjour",
		"echo hey > srcs/hello",
		">srcs/bonjour <prout >>srcs/hello",
		"cat srcs/bonjour ",
		"cat srcs/hello\"",
		"\"echo hola > srcs/bonjour",
		"echo hey > srcs/hello",
		">>srcs/bonjour <prout >srcs/hello",
		"cat srcs/bonjour ",
		"cat srcs/hello\"",
		"\"echo hola > srcs/bonjour",
		"echo hey > srcs/hello",
		">>srcs/bonjour <prout >>srcs/hello",
		"cat srcs/bonjour ",
		"cat srcs/hello\"",
		"\"echo hola > srcs/bonjour",
		"echo hey > srcs/hello",
		"<prout >>srcs/bonjour >>srcs/hello",
		"cat srcs/bonjour ",
		"cat srcs/hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"<bonjour >hello",
		"cat bonjour ",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		">bonjour >hello < prout",
		"cat bonjour ",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"rm bonjour hello",
		">bonjour >hello < prout",
		"cat bonjour ",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		">bonjour <prout hello",
		"cat bonjour ",
		"cat hello\"",
		"\"echo hola > bonjour",
		"echo hey > hello",
		"rm bonjour hello",
		">bonjour <prout hello",
		"cat bonjour \"",
		"\"echo hola > bonjour",
		"<bonjour cat | wc > bonjour",
		"cat bonjour\"",
		"\"rm -f bonjour",
		"rm bonjour > bonjour",
		"ls -l bonjour\"",
		"\"export HOLA=\"\"bonjour hello\"\"",
		">$HOLA",
		"ls\"",
		"\"export HOLA=\"\"bonjour hello\"\"",
		">\"\"$HOLA\"\"",
		"ls\"",
		"\"export HOLA=\"\"bonjour hello\"\"",
		">$\"\"HOLA\"\"",
		"ls\"",
		"\"export HOLA=\"\"bonjour hello\"\"",
		">$HOLA>hey",
		"ls\"",
		"\"export HOLA=\"\"bonjour hello\"\"",
		">hey>$HOLA",
		"ls\"",
		"\"export HOLA=\"\"bonjour hello\"\"",
		">hey>$HOLA>hey>hey",
		"ls\"",
		"\"export A=hey",
		"export A B=Hola D E C=\"\"Que Tal\"\"",
		"echo $PROUT$B$C > /tmp/a > /tmp/b > /tmp/c",
		"cat /tmp/a",
		"cat /tmp/b",
		"cat /tmp/c\"",
		"<a cat <b <c",
		"\"<a cat <b <c",
		"cat a",
		"cat b",
		"cat c\"",
		"\">a ls >b >>c >d",
		"cat a",
		"cat b",
		"cat c",
		"cat d\"",
		"\">a ls >b >>c >d",
		"cat a",
		"cat b",
		"cat c",
		"cat d\"",
		"\"echo hola > a > b > c",
		"cat a",
		"cat b",
		"cat c\"",
		"\"mkdir dir",
		"ls -la > dir/bonjour",
		"cat dir/bonjour\"",
		"\"<a",
		"cat a\"",
		"\">d cat <a >>e",
		"cat a",
		"cat d",
		"cat e\"",
		"\"< a > b cat > hey >> d",
		"cat d",
		"ls\"",
		"cat << hola",
		"cat << 'hola'",
		"cat << \"hola\"",
		"cat << ho\"la\"",
		"cat << $HOME",
		"\"cat << hola > bonjour",
		"cat bonjour\"",
		"cat << hola | rev",
		"<< hola",
		"<<hola",
		"cat <<",
		"cat << prout << lol << koala",
		"prout << lol << cat << koala",
		"<< $hola",
		"<< $\"hola\"$\"b\"",
		"<< $\"$hola\"$$\"b\"",
		"<< ho$la$\"$a\"$$\"b\"",
		"echo hola <<< bonjour",
		"echo hola <<<< bonjour",
		"echo hola <<<<< bonjour",
		"cat <<a >>>out | <<b",
	}

	if err := createTestFile(testsDir, "redirects.txt", redirectsTests); err != nil {
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
