package internal

import (
	"bytes"
	"strings"

	"github.com/rs/zerolog"

	"cosmossdk.io/log"
)

// NewBufferedLogger creates a new logger that writes to the provided buffer.
// Error log lines will start with "ERR ".
// Info log lines will start with "INF ".
// Debug log lines will start with "DBG ".
func NewBufferedLogger(buffer *bytes.Buffer, level zerolog.Level) log.Logger {
	lw := zerolog.ConsoleWriter{
		Out:          buffer,
		NoColor:      true,
		PartsExclude: []string{"time"}, // Without this, each line starts with "<nil> "
	}
	logger := zerolog.New(lw).Level(level)
	return log.NewCustomLogger(logger)
}

// NewBufferedInfoLogger creates a new logger with level info that writes to the provided buffer.
// Error log lines will start with "ERR ".
// Info log lines will start with "INF ".
// Debug log lines are omitted, but would start with "DBG ".
func NewBufferedInfoLogger(buffer *bytes.Buffer) log.Logger {
	return NewBufferedLogger(buffer, zerolog.InfoLevel)
}

// NewBufferedDebugLogger creates a new logger with level debug that writes to the provided buffer.
// Error log lines will start with "ERR ".
// Info log lines will start with "INF ".
// Debug log lines will start with "DBG ".
func NewBufferedDebugLogger(buffer *bytes.Buffer) log.Logger {
	return NewBufferedLogger(buffer, zerolog.DebugLevel)
}

// SplitLogLines splits the provided logs string into its individual lines.
func SplitLogLines(logs string) []string {
	lines := strings.Split(logs, "\n")
	// Trim spaces from each line.
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	// Remove empty lines from the end (at least one gets added due to a final newline in the logs).
	for len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	return lines
}
