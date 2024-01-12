package internal

import (
	"bytes"
	"strings"

	"github.com/rs/zerolog"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/server"
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
	return server.ZeroLogWrapper{Logger: logger}
}

// NewBufferedInfoLogger creates a new logger with level info that writes to the provided buffer.
// Error log lines will start with "ERR ".
// Info log lines will start with "INF ".
// Debug log lines are omitted, but would start with "DBG ".
func NewBufferedInfoLogger(buffer *bytes.Buffer) log.Logger {
	return NewBufferedLogger(buffer, zerolog.InfoLevel)
}

// SplitLogLines splits the provided logs string into its individual lines.
func SplitLogLines(logs string) []string {
	rv := strings.Split(logs, "\n")
	// Trim spaces from each line.
	for i, line := range rv {
		rv[i] = strings.TrimSpace(line)
	}
	// Remove empty lines from the end (at least one gets added due to a final newline in the logs).
	for len(rv) > 0 && len(rv[len(rv)-1]) == 0 {
		rv = rv[:len(rv)-1]
	}
	return rv
}
