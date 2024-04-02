package app

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/provenance-io/provenance/internal"
	"github.com/provenance-io/provenance/testutil/assertions"
)

// StoreLoaderMocker is a struct with a StoreLoader method that records that the store loader was called and returns a pre-determined error message.
type StoreLoaderMocker struct {
	Called bool
	ErrMsg string
}

func NewStoreLoaderMocker(errMsg string) *StoreLoaderMocker {
	return &StoreLoaderMocker{
		ErrMsg: errMsg,
	}
}

func (s *StoreLoaderMocker) StoreLoader(_ storetypes.CommitMultiStore) error {
	s.Called = true
	if len(s.ErrMsg) > 0 {
		return errors.New(s.ErrMsg)
	}
	return nil
}

// MockSleeper is a Sleeper that only records what sleep was requested (instead of actually sleeping).
type MockSleeper struct {
	LastSleep time.Duration
}

var _ Sleeper = (*MockSleeper)(nil)

func NewMockSleeper() *MockSleeper {
	return &MockSleeper{}
}

func (s *MockSleeper) Sleep(d time.Duration) {
	s.LastSleep = d
}

// MockAppOptions is a mocked version of AppOpts that allows the developer to provide the pruning attribute.
type MockAppOptions struct {
	pruning string
	indexer string
	db      string
}

// Get returns the value for the provided option.
func (m MockAppOptions) Get(opt string) interface{} {
	switch opt {
	case "pruning-interval":
		return m.pruning
	case "tx_index":
		return map[string]interface{}{
			"indexer": m.indexer,
		}
	case "app-db-backend":
		return m.db
	case "db-backend":
		return m.db
	}

	return nil
}

func TestValidateWrapper(t *testing.T) {
	defer func() {
		ValidateWrapperSleeper = &DefaultSleeper{}
	}()

	recAppOpts := MockAppOptions{
		pruning: "10",
		db:      "goleveldb",
		indexer: "null",
	}

	tests := []struct {
		name       string
		appOpts    MockAppOptions
		pioAckWarn bool
		expErr     string
		expLogMsgs bool
		expSleep   bool
	}{
		{
			name:       "empty opts",
			appOpts:    MockAppOptions{},
			expErr:     "",
			expLogMsgs: false,
			expSleep:   false,
		},
		{
			name: "bad config",
			appOpts: MockAppOptions{
				db: "cleveldb",
			},
			expLogMsgs: true,
			expSleep:   true,
		},
		{
			name: "bad config no sleep",
			appOpts: MockAppOptions{
				db: "cleveldb",
			},
			pioAckWarn: true,
			expLogMsgs: true,
			expSleep:   false,
		},
		{
			name:    "err from store loader",
			appOpts: recAppOpts,
			expErr:  "injected test error",
		},
		{
			name: "bad config and err from store loader",
			appOpts: MockAppOptions{
				db: "somethingelse",
			},
			expErr:     "another injected error for testing",
			expLogMsgs: true,
			expSleep:   true,
		},
		{
			name: "bad config and err from store loader no sleep",
			appOpts: MockAppOptions{
				db: "somethingelse",
			},
			pioAckWarn: true,
			expErr:     "another injected error for testing",
			expLogMsgs: true,
			expSleep:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sleeper := NewMockSleeper()
			ValidateWrapperSleeper = sleeper

			if tc.pioAckWarn {
				t.Setenv("PIO_ACKWARN", "1")
			}

			var buffer bytes.Buffer
			logger := internal.NewBufferedInfoLogger(&buffer)
			slMocker := NewStoreLoaderMocker(tc.expErr)
			storeLoader := ValidateWrapper(logger, tc.appOpts, slMocker.StoreLoader)

			var err error
			testFunc := func() {
				err = storeLoader(nil)
			}
			require.NotPanics(t, testFunc, "calling the storeLoader that was returned by ValidateWrapper")
			assertions.AssertErrorValue(t, err, tc.expErr, "error from storeLoader")

			logMsgs := buffer.String()
			if tc.expLogMsgs {
				assert.NotEmpty(t, logMsgs, "messages logged during storeLoader")
			} else {
				assert.Empty(t, logMsgs, "messages logged during storeLoader")
			}

			didSleep := sleeper.LastSleep != 0
			assert.Equal(t, tc.expSleep, didSleep, "whether sleep was called")
		})
	}
}

func TestIssueConfigWarnings(t *testing.T) {
	sleepErr1 := "ERR NODE WILL CONTINUE AFTER 30 SECONDS."
	sleepErr2 := "ERR This wait can be bypassed by fixing the above warnings or setting the PIO_ACKWARN environment variable to \"1\"."

	tests := []struct {
		name        string
		appOpts     MockAppOptions
		pioAckWarn  string
		expLogLines []string // can be in any order, but all must be there.
		expSleep    bool
	}{
		{
			name:        "no app opts",
			appOpts:     MockAppOptions{},
			expLogLines: nil,
		},
		{
			name: "recommended app opts",
			appOpts: MockAppOptions{
				pruning: "10",
				db:      "goleveldb",
				indexer: "null",
			},
			expLogLines: nil,
		},
		{
			name: "bad pruning interval",
			appOpts: MockAppOptions{
				pruning: "1000",
				db:      "goleveldb",
				indexer: "null",
			},
			expLogLines: []string{
				"ERR pruning-interval 1000 EXCEEDS 999 AND IS NOT RECOMMENDED, AS IT CAN LEAD TO MISSED BLOCKS ON VALIDATORS.",
				sleepErr1,
				sleepErr2,
			},
			expSleep: true,
		},
		{
			name: "bad indexer",
			appOpts: MockAppOptions{
				pruning: "10",
				db:      "goleveldb",
				indexer: "kv",
			},
			expLogLines: []string{
				"ERR indexer \"kv\" IS NOT RECOMMENDED, AND IT IS RECOMMENDED TO USE \"null\".",
				sleepErr1,
				sleepErr2,
			},
			expSleep: true,
		},
		{
			name: "bad db",
			appOpts: MockAppOptions{
				pruning: "10",
				db:      "cleveldb",
				indexer: "null",
			},
			expLogLines: []string{
				"ERR cleveldb IS NO LONGER SUPPORTED. MIGRATE TO goleveldb.",
				sleepErr1,
				sleepErr2,
			},
			expSleep: true,
		},
		{
			name: "all bad with sleep",
			appOpts: MockAppOptions{
				pruning: "1001",
				db:      "badgerdb",
				indexer: "psql",
			},
			expLogLines: []string{
				"ERR pruning-interval 1001 EXCEEDS 999 AND IS NOT RECOMMENDED, AS IT CAN LEAD TO MISSED BLOCKS ON VALIDATORS.",
				"ERR indexer \"psql\" IS NOT RECOMMENDED, AND IT IS RECOMMENDED TO USE \"null\".",
				"ERR badgerdb IS NO LONGER SUPPORTED. MIGRATE TO goleveldb.",
				sleepErr1,
				sleepErr2,
			},
			expSleep: true,
		},
		{
			name: "all bad no sleep",
			appOpts: MockAppOptions{
				pruning: "1001",
				db:      "badgerdb",
				indexer: "psql",
			},
			pioAckWarn: "1",
			expLogLines: []string{
				"ERR pruning-interval 1001 EXCEEDS 999 AND IS NOT RECOMMENDED, AS IT CAN LEAD TO MISSED BLOCKS ON VALIDATORS.",
				"ERR indexer \"psql\" IS NOT RECOMMENDED, AND IT IS RECOMMENDED TO USE \"null\".",
				"ERR badgerdb IS NO LONGER SUPPORTED. MIGRATE TO goleveldb.",
			},
			expSleep: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var expSleepDur time.Duration
			if tc.expSleep {
				expSleepDur = 30 * time.Second
			}

			if len(tc.pioAckWarn) > 0 {
				t.Setenv("PIO_ACKWARN", tc.pioAckWarn)
			}
			var buffer bytes.Buffer
			logger := internal.NewBufferedInfoLogger(&buffer)
			sleeper := NewMockSleeper()

			testFunc := func() {
				IssueConfigWarnings(logger, tc.appOpts, sleeper)
			}
			require.NotPanics(t, testFunc, "IssueConfigWarnings")

			loggedLines := internal.SplitLogLines(buffer.String())
			assert.ElementsMatch(t, tc.expLogLines, loggedLines, "Lines logged during IssueConfigWarnings. List A is the expected lines.")
			actSleepDur := sleeper.LastSleep
			assert.Equal(t, expSleepDur.String(), actSleepDur.String(), "sleep duration during IssueConfigWarnings")
		})
	}
}

func TestHaveAckWarn(t *testing.T) {
	tests := []struct {
		pioAckWarn   string
		noPioAckWarn bool
		expected     bool
	}{
		{noPioAckWarn: true, expected: false},
		{pioAckWarn: "", expected: false},
		{pioAckWarn: "   ", expected: false},
		{pioAckWarn: "0", expected: false},
		{pioAckWarn: " 0 ", expected: false},
		{pioAckWarn: "false", expected: false},
		{pioAckWarn: " False", expected: false},
		{pioAckWarn: "FALSE ", expected: false},
		{pioAckWarn: "f", expected: false},
		{pioAckWarn: "   F   ", expected: false},

		{pioAckWarn: "1", expected: true},
		{pioAckWarn: "yes", expected: true},
		{pioAckWarn: "t", expected: true},
		{pioAckWarn: "true", expected: true},
		{pioAckWarn: "  T", expected: true},
		{pioAckWarn: "TRUE  ", expected: true},
		{pioAckWarn: " True  ", expected: true},
		{pioAckWarn: "whatever", expected: true},
		{pioAckWarn: "X", expected: true},
		{pioAckWarn: "ff", expected: true},
		{pioAckWarn: "farse", expected: true},
	}

	for _, tc := range tests {
		name := tc.pioAckWarn
		if tc.noPioAckWarn {
			name = "no PIO_ACKWARN set"
		}
		if len(name) == 0 {
			name = "empty string"
		}

		t.Run(name, func(t *testing.T) {
			if !tc.noPioAckWarn {
				t.Setenv("PIO_ACKWARN", tc.pioAckWarn)
			}
			var actual bool
			testFunc := func() {
				actual = HaveAckWarn()
			}
			require.NotPanics(t, testFunc, "HaveAckWarn")
			assert.Equal(t, tc.expected, actual, "HaveAckWarn result")
		})
	}
}
