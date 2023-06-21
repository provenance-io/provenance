package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"

	"github.com/provenance-io/provenance/cmd/provenanced/config"
)

func TestIAVLConfig(t *testing.T) {
	require.Equal(t, getIAVLCacheSize(sdksim.EmptyAppOptions{}), cast.ToInt(serverconfig.DefaultConfig().IAVLCacheSize))
}

type testOpts map[string]interface{}

var _ servertypes.AppOptions = (*testOpts)(nil)

func (o testOpts) Get(key string) interface{} {
	return o[key]
}

type panicOpts struct{}

var _ servertypes.AppOptions = (*panicOpts)(nil)

func (o panicOpts) Get(key string) interface{} {
	panic(fmt.Errorf("panic forced while getting option %q", key))
}

func TestWarnAboutSettings(t *testing.T) {
	keyChainID := flags.FlagChainID
	keySkipTimeoutCommit := "consensus.skip_timeout_commit"
	keyTimeoutCommit := "consensus.timeout_commit"

	upperLimit := config.DefaultConsensusTimeoutCommit + 2*time.Second
	overUpperLimit := upperLimit + 1*time.Millisecond
	mainnetChainID := "pio-mainnet-1"
	notMainnetChainID := "not-" + mainnetChainID

	tooHighMsg := func(timeoutCommit time.Duration) string {
		return fmt.Sprintf("ERR Your consensus.timeout_commit config value is too high and should be decreased to at most %q. The recommended value is %q. Your current value is %q.",
			upperLimit, config.DefaultConsensusTimeoutCommit, timeoutCommit)
	}

	mainnetOpts := func(skipTimeoutCommit bool, timeoutCommit time.Duration) testOpts {
		return testOpts{
			keyChainID:           mainnetChainID,
			keySkipTimeoutCommit: skipTimeoutCommit,
			keyTimeoutCommit:     timeoutCommit,
		}
	}
	notMainnetOpts := func(skipTimeoutCommit bool, timeoutCommit time.Duration) testOpts {
		return testOpts{
			keyChainID:           notMainnetChainID,
			keySkipTimeoutCommit: skipTimeoutCommit,
			keyTimeoutCommit:     timeoutCommit,
		}
	}

	tests := []struct {
		name      string
		appOpts   servertypes.AppOptions
		expLogged []string
	}{
		{
			name:      "mainnet not skipped value over upper limit",
			appOpts:   mainnetOpts(false, overUpperLimit),
			expLogged: []string{tooHighMsg(overUpperLimit)},
		},
		{
			name:    "mainnet not skipped value at limit",
			appOpts: mainnetOpts(false, upperLimit),
		},
		{
			name:    "mainnet not skipped value at default",
			appOpts: mainnetOpts(false, config.DefaultConsensusTimeoutCommit),
		},
		{
			name:    "mainnet not skipped value at zero",
			appOpts: mainnetOpts(false, 0),
		},
		{
			name:    "mainnet skipped value over upper limit",
			appOpts: mainnetOpts(true, overUpperLimit),
		},
		{
			name:    "mainnet skipped value at upper limit",
			appOpts: mainnetOpts(true, upperLimit),
		},
		{
			name:    "mainnet skipped value at default",
			appOpts: mainnetOpts(true, config.DefaultConsensusTimeoutCommit),
		},
		{
			name:    "mainnet skipped value at zero",
			appOpts: mainnetOpts(true, 0),
		},
		{
			name:    "not mainnet not skipped value over upper limit",
			appOpts: notMainnetOpts(false, overUpperLimit),
		},
		{
			name:    "not mainnet not skipped value at upper limit",
			appOpts: notMainnetOpts(false, upperLimit),
		},
		{
			name:    "not mainnet not skipped value at default",
			appOpts: notMainnetOpts(false, config.DefaultConsensusTimeoutCommit),
		},
		{
			name:    "not mainnet not skipped value at zero",
			appOpts: notMainnetOpts(false, 0),
		},
		{
			name:    "not mainnet skipped value over upper limit",
			appOpts: notMainnetOpts(true, overUpperLimit),
		},
		{
			name:    "not mainnet skipped value at upper limit",
			appOpts: notMainnetOpts(true, upperLimit),
		},
		{
			name:    "not mainnet skipped value at default",
			appOpts: notMainnetOpts(true, config.DefaultConsensusTimeoutCommit),
		},
		{
			name:    "not mainnet skipped value at zero",
			appOpts: notMainnetOpts(true, 0),
		},
		{
			name:    "timeout commit opt not a duration",
			appOpts: testOpts{keyChainID: mainnetChainID, keySkipTimeoutCommit: false, keyTimeoutCommit: "nope"},
		},
		{
			name:    "empty opts",
			appOpts: testOpts{},
		},
		{
			name:    "only chain-id mainnet",
			appOpts: testOpts{keyChainID: mainnetChainID},
		},
		{
			name:    "only chain-id not mainnet",
			appOpts: testOpts{keyChainID: mainnetChainID},
		},
		{
			name:    "mainnet not skipped no timeout commit opt",
			appOpts: testOpts{keyChainID: mainnetChainID, keySkipTimeoutCommit: false},
		},
		{
			name:    "panic from getter",
			appOpts: panicOpts{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// If expLogged isn't defined, we're expecting zero logged lines.
			if tc.expLogged == nil {
				tc.expLogged = make([]string, 0)
			}

			// Create a logger wrapped around our own buffer so we can see what was logged.
			var buf bytes.Buffer
			lw := zerolog.ConsoleWriter{
				Out:          &buf,
				NoColor:      true,
				PartsExclude: []string{"time"}, // Without this, each line starts with "<nil> "
			}
			// Error log lines will start with "ERR ".
			// Info log lines will start with "INF ".
			// Debug log lines are omitted, but would start with "DBG ".
			logger := server.ZeroLogWrapper{Logger: zerolog.New(lw).Level(zerolog.InfoLevel)}

			// Make sure the function never panics.
			require.NotPanics(t, func() { warnAboutSettings(logger, tc.appOpts) }, "warnAboutSettings")

			// Make sure the logged output is as expected.
			out := buf.String()
			// Splitting it into lines and comparing the string slices makes test failure output nicer than
			// if we just compare two multi-line strings.
			loggedLines := strings.Split(out, "\n")
			// Get rid of last "line" if it's just an empty string. Happens when out ends with "\n".
			if len(loggedLines[len(loggedLines)-1]) == 0 {
				loggedLines = loggedLines[:len(loggedLines)-1]
			}
			assert.Equal(t, tc.expLogged, loggedLines, "lines logged during warnAboutSettings")
		})
	}
}
