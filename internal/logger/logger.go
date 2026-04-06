package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Mode controls what contextual prints to stderr while running.
type Mode int

const (
	// ModesilENT prints nothing to stderr until the program finishes.
	ModeSilent Mode = iota
	// ModeProgress prints "loading context " once, then a dot per success and
	// X per error, followed by a newline when done.
	ModeProgress
	// ModeVerbose prints a line for every fetch attempt and every error.
	ModeVerbose
)

// Logger handles per-item progress output to stderr and full structured
// output to ~/.contextual/log.log.
type Logger struct {
	mode       Mode
	fileLogger *log.Logger
	fileHandle io.Closer

	progressStarted bool // whether we have printed "loading context " yet
}

// New creates a Logger. It always opens (or creates) ~/.contextual/log.log.
// Callers must call Close() when done.
func New(mode Mode) (*Logger, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("finding home directory: %w", err)
	}

	dir := filepath.Join(home, ".contextual")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating ~/.contextual/: %w", err)
	}

	logPath := filepath.Join(dir, "log.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening log file %s: %w", logPath, err)
	}

	fileLogger := log.New(f, "", 0) // we format the prefix ourselves

	return &Logger{
		mode:       mode,
		fileLogger: fileLogger,
		fileHandle: f,
	}, nil
}

// Close flushes and closes the underlying log file.
// In progress mode it also prints a trailing newline to stderr.
func (l *Logger) Close() {
	if l.mode == ModeProgress && l.progressStarted {
		fmt.Fprintln(os.Stderr)
	}
	if l.fileHandle != nil {
		l.fileHandle.Close()
	}
}

// Info logs an informational fetch event.
func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.writeToFile("[INFO] " + msg)

	switch l.mode {
	case ModeVerbose:
		fmt.Fprintf(os.Stderr, "[INFO] %s\n", msg)
	case ModeProgress:
		l.ensureProgressHeader()
		fmt.Fprint(os.Stderr, ".")
	}
}

// Error logs a fetch error.
func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.writeToFile("[ERROR] " + msg)

	switch l.mode {
	case ModeVerbose:
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", msg)
	case ModeProgress:
		l.ensureProgressHeader()
		fmt.Fprint(os.Stderr, "X")
	}
}

// Warn logs a warning (configuration issues, missing credentials, etc.).
// Warnings are written to the log file and, in verbose mode, to stderr.
// In silent and progress modes they are suppressed from stderr.
func (l *Logger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.writeToFile("[WARN] " + msg)

	if l.mode == ModeVerbose {
		fmt.Fprintf(os.Stderr, "[WARN] %s\n", msg)
	}
}

func (l *Logger) ensureProgressHeader() {
	if !l.progressStarted {
		fmt.Fprint(os.Stderr, "loading context ")
		l.progressStarted = true
	}
}

func (l *Logger) writeToFile(msg string) {
	ts := time.Now().UTC().Format(time.RFC3339)
	l.fileLogger.Printf("%s %s", ts, msg)
}
