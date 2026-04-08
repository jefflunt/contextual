package planner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jluntpcty/contextual/internal/types"
)

var slugSafeRe = regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)

// ResolveOutputDir determines where to write output files for plan mode.
// It returns the path of a slug-named subdirectory that has been created and
// is ready to receive context.md and plan.md.
//
// Rules (checked against the current working directory):
//  1. If ./agent_docs/plans/ exists → use ./agent_docs/plans/<slug>/
//  2. If ./agent_docs/ exists but ./agent_docs/plans/ does not → ask the user
//     whether to create it; if yes, use ./agent_docs/plans/<slug>/
//  3. Otherwise → use ./<slug>/
func ResolveOutputDir(primaryItem types.Item) (string, error) {
	slug := ItemSlug(primaryItem)

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	plansDir := filepath.Join(cwd, "agent_docs", "plans")
	agentDocsDir := filepath.Join(cwd, "agent_docs")

	var baseDir string
	switch {
	case dirExists(plansDir):
		baseDir = plansDir

	case dirExists(agentDocsDir):
		if PromptYesNo("agent_docs/plans/ does not exist. Create it? [y/N] ") {
			if err := os.Mkdir(plansDir, 0755); err != nil {
				return "", fmt.Errorf("creating agent_docs/plans/: %w", err)
			}
			baseDir = plansDir
		} else {
			baseDir = cwd
		}

	default:
		baseDir = cwd
	}

	dir := filepath.Join(baseDir, slug)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating output directory %s: %w", dir, err)
	}
	return dir, nil
}

// ConfirmOverwrite checks if a file exists at path and, if so, asks the user
// whether to overwrite it. Returns true if it is safe to write (either the
// file does not exist or the user confirmed).
func ConfirmOverwrite(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true
	}
	return PromptYesNo(fmt.Sprintf("%s already exists. Overwrite? [y/N] ", path))
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// PromptYesNo prints question to stderr and reads a y/yes answer from stdin.
func PromptYesNo(question string) bool {
	fmt.Fprint(os.Stderr, question)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return answer == "y" || answer == "yes"
	}
	return false
}

// ItemSlug returns a filesystem-safe identifier for the primary item.
func ItemSlug(item types.Item) string {
	switch item.Type {
	case types.ItemTypeJira:
		return item.ID // already safe, e.g. CTX-1234
	case types.ItemTypeConfluence:
		if item.Title != "" {
			return slugify(item.Title)
		}
		return "confluence-" + item.ID
	case types.ItemTypeWeb:
		if item.Title != "" {
			return slugify(item.Title)
		}
		s := strings.TrimPrefix(item.URL, "https://")
		s = strings.TrimPrefix(s, "http://")
		return slugify(s)
	}
	return "context"
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = slugSafeRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 80 {
		s = s[:80]
		s = strings.TrimRight(s, "-")
	}
	return s
}
