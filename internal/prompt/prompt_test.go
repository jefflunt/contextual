package prompt

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPromptYesNo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Yes lowercase", "y\n", true},
		{"Yes spelled out", "yes\n", true},
		{"Yes uppercase", "Y\n", true},
		{"No lowercase", "n\n", false},
		{"No spelled out", "no\n", false},
		{"Random string", "maybe\n", false},
		{"Empty string", "\n", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Mock stdin and stdout
			inBuf := bytes.NewBufferString(tc.input)
			outBuf := &bytes.Buffer{}

			// Save and restore
			oldStdin := Stdin
			oldStderr := Stderr
			Stdin = inBuf
			Stderr = outBuf
			defer func() {
				Stdin = oldStdin
				Stderr = oldStderr
			}()

			res := PromptYesNo("Proceed? [y/N] ")
			if res != tc.expected {
				t.Errorf("PromptYesNo(%q) = %v, expected %v", tc.input, res, tc.expected)
			}

			if outBuf.String() != "Proceed? [y/N] " {
				t.Errorf("Expected question to be printed, got: %q", outBuf.String())
			}
		})
	}
}

func TestConfirmOverwrite(t *testing.T) {
	// Create a temp file to test overwrite confirmation
	tmpDir, err := os.MkdirTemp("", "prompt-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	existingFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to write existing file: %v", err)
	}

	nonExistingFile := filepath.Join(tmpDir, "does-not-exist.txt")

	t.Run("non-existing file returns true without prompting", func(t *testing.T) {
		res := ConfirmOverwrite(nonExistingFile)
		if !res {
			t.Errorf("expected ConfirmOverwrite to return true for non-existing file")
		}
	})

	t.Run("existing file prompts and returns true if user says yes", func(t *testing.T) {
		inBuf := bytes.NewBufferString("y\n")
		outBuf := &bytes.Buffer{}

		oldStdin := Stdin
		oldStderr := Stderr
		Stdin = inBuf
		Stderr = outBuf
		defer func() {
			Stdin = oldStdin
			Stderr = oldStderr
		}()

		res := ConfirmOverwrite(existingFile)
		if !res {
			t.Errorf("expected true when user confirms overwrite")
		}
	})

	t.Run("existing file prompts and returns false if user says no", func(t *testing.T) {
		inBuf := bytes.NewBufferString("n\n")
		outBuf := &bytes.Buffer{}

		oldStdin := Stdin
		oldStderr := Stderr
		Stdin = inBuf
		Stderr = outBuf
		defer func() {
			Stdin = oldStdin
			Stderr = oldStderr
		}()

		res := ConfirmOverwrite(existingFile)
		if res {
			t.Errorf("expected false when user declines overwrite")
		}
	})
}
