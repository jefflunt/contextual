package planner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jefflunt/contextual/internal/logger"
)

const promptFilePlaceholder = "<promptFile>"

// RunPlanner writes the prompt to a temp file, substitutes <promptFile> in
// plannerCmd with the temp file path, then executes the resulting command.
//
// plannerCmd must contain the <promptFile> placeholder — for example:
//
//	copilot -p @<promptFile> --allow-all-tools --allow-all-paths --autopilot -s
//
// outputDir is the directory where the planner should write plan.md. The
// prompt is augmented with the exact output path so the planner knows where
// to write.
func RunPlanner(plannerCmd string, promptText string, outputDir string, lg *logger.Logger) error {
	if plannerCmd == "" {
		return fmt.Errorf(
			"no planner configured\n" +
				"  set the `planner` key in ~/.contextual/config.yml\n" +
				"  see contextual.config.example.yml in the repo for a full example",
		)
	}

	if !strings.Contains(plannerCmd, promptFilePlaceholder) {
		return fmt.Errorf(
			"planner command is missing the %s placeholder\n"+
				"  contextual substitutes %s with the path to a temp file containing the prompt\n"+
				"  example: copilot -p @%s --allow-all-tools --allow-all-paths --autopilot -s\n"+
				"  see contextual.config.example.yml in the repo for more examples",
			promptFilePlaceholder, promptFilePlaceholder, promptFilePlaceholder,
		)
	}

	planPath := filepath.Join(outputDir, "plan.md")

	// Augment the prompt with the exact output path.
	promptWithPath := promptText + fmt.Sprintf(
		"\n\n---\n\n## Output\n\nWrite the completed plan file to this exact path: `%s`\n"+
			"Do not ask for confirmation. Create any necessary parent directories.\n",
		planPath,
	)

	// Write prompt to a temp file to avoid shell argument length limits.
	tmp, err := os.CreateTemp("", "contextual-prompt-*.txt")
	if err != nil {
		return fmt.Errorf("creating prompt temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(promptWithPath); err != nil {
		tmp.Close()
		return fmt.Errorf("writing prompt temp file: %w", err)
	}
	tmp.Close()

	// Substitute the placeholder with the actual temp file path.
	resolved := strings.ReplaceAll(plannerCmd, promptFilePlaceholder, tmp.Name())

	// Execute via shell so the user's command string is interpreted literally,
	// preserving quoting, backticks, and other shell constructs.
	cmd := exec.Command("sh", "-c", resolved)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	lg.Info("Invoking planner: %s", resolved)

	if err := cmd.Run(); err != nil {
		exitCode := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		lg.Error("Planner exited with code %d: %s", exitCode, resolved)
		return fmt.Errorf("planner exited with code %d", exitCode)
	}
	return nil
}
