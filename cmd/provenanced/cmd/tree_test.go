package cmd_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil/assertions"
)

func executeTreeCmd(t *testing.T, cmdArgs []string) *cmdResult {
	return executeRootCmd(t, "", append([]string{"tree"}, cmdArgs...)...)
}

func TestGetTreeCmd(t *testing.T) {
	origCache := sdk.IsAddrCacheEnabled()
	defer sdk.SetAddrCacheEnabled(origCache)
	sdk.SetAddrCacheEnabled(false)

	testDefs := []struct {
		name          string
		args          []string
		expErr        bool
		expInOutLines []string // each entry must match an entire line in stdout.
	}{
		{
			name: "no args",
			args: nil,
			expInOutLines: []string{
				"provenanced help",
				"provenanced tree",
				"provenanced query bank balance",
				"provenanced tx exchange commit",
			},
		},
		{
			name: "just aliases flag",
			args: []string{"--aliases"},
			expInOutLines: []string{
				"provenanced help",
				"provenanced tree",
				"provenanced [query q] bank balance",
				"provenanced tx [exchange ex] [commit commit-funds]",
			},
		},
		{
			name:   "unknown sub-command",
			args:   []string{"bananas"},
			expErr: true,
		},
		{
			name:   "partially known sub-command",
			args:   []string{"config", "bananas"},
			expErr: true,
		},
		{
			name: "known sub-command",
			args: []string{"config"},
			expInOutLines: []string{
				"provenanced config changed",
				"provenanced config get",
				"provenanced config home",
				"provenanced config pack",
				"provenanced config set",
				"provenanced config unpack",
			},
		},
		{
			name: "known sub-command with aliases",
			args: []string{"config", "--aliases"},
			expInOutLines: []string{
				"provenanced [config conf] changed",
				"provenanced [config conf] get",
				"provenanced [config conf] home",
				"provenanced [config conf] pack",
				"provenanced [config conf] set",
				"provenanced [config conf] [unpack update]",
			},
		},
		{
			name: "known sub-command provided as alias",
			args: []string{"conf"},
			expInOutLines: []string{
				"provenanced config changed",
				"provenanced config get",
				"provenanced config home",
				"provenanced config pack",
				"provenanced config set",
				"provenanced config unpack",
			},
		},
		{
			name: "known sub-command provided as alias with aliases",
			args: []string{"conf", "--aliases"},
			expInOutLines: []string{
				"provenanced [config conf] changed",
				"provenanced [config conf] get",
				"provenanced [config conf] home",
				"provenanced [config conf] pack",
				"provenanced [config conf] set",
				"provenanced [config conf] [unpack update]",
			},
		},
	}

	type testCase struct {
		name          string
		args          []string
		expErr        string
		expInOutLines []string
		expExitCode   int
	}

	tests := make([]testCase, 0, len(testDefs)*2)
	for _, tc := range testDefs {
		tcNoProv := testCase{
			name:          tc.name + " without provenanced",
			args:          tc.args,
			expInOutLines: tc.expInOutLines,
		}
		tcWProv := testCase{
			name:          tc.name + " with provenanced",
			args:          append([]string{"provenanced"}, tc.args...),
			expInOutLines: tc.expInOutLines,
		}

		if tc.expErr {
			tcWProv.expErr = fmt.Sprintf("command not found: %q", tcWProv.args)
			tcWProv.expExitCode = 1
			tcNoProv.expErr = tcWProv.expErr
			tcNoProv.expExitCode = 1
		}

		tests = append(tests, tcNoProv, tcWProv)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := executeTreeCmd(t, tc.args)

			assert.Equal(t, tc.expExitCode, res.ExitCode, "exit code")
			assertions.AssertErrorValue(t, res.Result, tc.expErr, "error from command")
			if len(tc.expErr) > 0 {
				assert.Equal(t, res.Stderr, fmt.Sprintf("Error: %s\n", tc.expErr), "stderr")
			}

			if len(tc.expInOutLines) > 0 {
				outLines := strings.Split(res.Stdout, "\n")
				assert.Subset(t, outLines, tc.expInOutLines, "stdout")
			}
		})
	}
}
