package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	// Stdin allows overriding standard input for unit testing.
	Stdin io.Reader = os.Stdin
	// Stderr allows overriding standard error for unit testing.
	Stderr io.Writer = os.Stderr
)

// ConfirmOverwrite checks if a file exists at path and, if so, asks the user
// whether to overwrite it. Returns true if it is safe to write (either the
// file does not exist or the user confirmed).
func ConfirmOverwrite(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true
	}
	return PromptYesNo(fmt.Sprintf("%s already exists. Overwrite? [y/N] ", path))
}

// PromptYesNo prints a question to Stderr and reads a y/yes answer from Stdin.
func PromptYesNo(question string) bool {
	fmt.Fprint(Stderr, question)
	scanner := bufio.NewScanner(Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes"
	}
	return false
}

