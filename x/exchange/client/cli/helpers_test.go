package cli_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/exchange/client/cli"
)

func TestAddUseArgs(t *testing.T) {
	tests := []struct {
		name   string
		use    string
		args   []string
		expUse string
	}{
		{
			name:   "new, one arg",
			use:    "unit-test",
			args:   []string{"arg1"},
			expUse: "unit-test arg1",
		},
		{
			name:   "new, three args",
			use:    "testing",
			args:   []string{"{--yes|--no}", "--id <id>", "[--thing <thing>]"},
			expUse: "testing {--yes|--no} --id <id> [--thing <thing>]",
		},
		{
			name:   "already has stuff, one arg",
			use:    "do-thing <id> <name>",
			args:   []string{"[--foo]"},
			expUse: "do-thing <id> <name> [--foo]",
		},
		{
			name:   "already has stuff, three args",
			use:    "complex <name> <value>",
			args:   []string{"--opt1 <val1>", "[--nope]", "[--yup]"},
			expUse: "complex <name> <value> --opt1 <val1> [--nope] [--yup]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: tc.use,
				RunE: func(cmd *cobra.Command, args []string) error {
					return errors.New("the command should not have been run")
				},
			}

			testFunc := func() {
				cli.AddUseArgs(cmd, tc.args...)
			}
			require.NotPanics(t, testFunc, "AddUseArgs")
			actUse := cmd.Use
			assert.Equal(t, tc.expUse, actUse, "cmd.Use string after AddUseArgs")
		})
	}
}

func TestAddUseDetails(t *testing.T) {
	tests := []struct {
		name     string
		use      string
		sections []string
		expUse   string
	}{
		{
			name:     "no sections",
			use:      "some-command {<id>|--id <id>} [flags]",
			sections: []string{},
			expUse:   "some-command {<id>|--id <id>} [flags]",
		},
		{
			name:     "one section",
			use:      "testing <stuff> [flags]",
			sections: []string{"Section 1, Line 1\nSection 1, Line 2\nSection 1, Line 3"},
			expUse: `testing <stuff> [flags]

Section 1, Line 1
Section 1, Line 2
Section 1, Line 3`,
		},
		{
			name: "",
			use:  "longer <stuff> <more stuff> [flags]",
			sections: []string{
				"Section 1, Line 1\nSection 1, Line 2\nSection 1, Line 3",
				"Section 2, Line 1\nSection 2, Line 2\nSection 2, Line 3",
				"Section 3, Line 1\nSection 3, Line 2\nSection 3, Line 3",
			},
			expUse: `longer <stuff> <more stuff> [flags]

Section 1, Line 1
Section 1, Line 2
Section 1, Line 3

Section 2, Line 1
Section 2, Line 2
Section 2, Line 3

Section 3, Line 1
Section 3, Line 2
Section 3, Line 3`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: tc.use,
				RunE: func(cmd *cobra.Command, args []string) error {
					return errors.New("the command should not have been run")
				},
			}

			testFunc := func() {
				cli.AddUseDetails(cmd, tc.sections...)
			}
			require.NotPanics(t, testFunc, "AddUseDetails")
			actUse := cmd.Use
			assert.Equal(t, tc.expUse, actUse, "cmd.Use string after AddUseDetails")
			assert.True(t, cmd.DisableFlagsInUseLine, "cmd.DisableFlagsInUseLine")
		})
	}
}

func TestAddQueryExample(t *testing.T) {
	tests := []struct {
		name       string
		use        string
		example    string
		args       []string
		expExample string
	}{
		{
			name:       "first, no args",
			use:        "mycmd",
			expExample: version.AppName + " query exchange mycmd",
		},
		{
			name:       "first, one arg",
			use:        "yourcmd",
			args:       []string{"--dance"},
			expExample: version.AppName + " query exchange yourcmd --dance",
		},
		{
			name:       "first, three args",
			use:        "theircmd",
			args:       []string{"party", "someaddr", "--lights=off"},
			expExample: version.AppName + " query exchange theircmd party someaddr --lights=off",
		},
		{
			name: "third, no args",
			use:  "mycmd",
			example: version.AppName + " query exchange mycmd --opt1 party\n" +
				version.AppName + " query exchange mycmd --opt2 sleep",
			expExample: version.AppName + " query exchange mycmd --opt1 party\n" +
				version.AppName + " query exchange mycmd --opt2 sleep\n" +
				version.AppName + " query exchange mycmd",
		},
		{
			name: "third, one arg",
			use:  "yourcmd",
			example: version.AppName + " query exchange yourcmd --opt1 party\n" +
				version.AppName + " query exchange yourcmd --opt2 sleep",
			args: []string{"--no-pants"},
			expExample: version.AppName + " query exchange yourcmd --opt1 party\n" +
				version.AppName + " query exchange yourcmd --opt2 sleep\n" +
				version.AppName + " query exchange yourcmd --no-pants",
		},
		{
			name: "third, three args",
			use:  "theircmd",
			example: version.AppName + " query exchange theircmd --opt1 party\n" +
				version.AppName + " query exchange theircmd --opt2 sleep",
			args: []string{"--no-shirt", "--no-shoes", "--no-service"},
			expExample: version.AppName + " query exchange theircmd --opt1 party\n" +
				version.AppName + " query exchange theircmd --opt2 sleep\n" +
				version.AppName + " query exchange theircmd --no-shirt --no-shoes --no-service",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:     tc.use,
				Example: tc.example,
				RunE: func(cmd *cobra.Command, args []string) error {
					return errors.New("the command should not have been run")
				},
			}

			testFunc := func() {
				cli.AddQueryExample(cmd, tc.args...)
			}
			require.NotPanics(t, testFunc, "AddQueryExample")
			actExample := cmd.Example
			assert.Equal(t, tc.expExample, actExample, "cmd.Example string after AddQueryExample")
		})
	}
}

func TestSimplePerms(t *testing.T) {
	var actual string
	testFunc := func() {
		actual = cli.SimplePerms()
	}
	require.NotPanics(t, testFunc, "SimplePerms()")
	for _, perm := range exchange.AllPermissions() {
		t.Run(perm.String(), func(t *testing.T) {
			exp := perm.SimpleString()
			assert.Contains(t, actual, exp, "SimplePerms()")
		})
	}
}

func TestReqSignerDesc(t *testing.T) {
	for _, name := range []string{cli.FlagBuyer, cli.FlagSeller, cli.FlagSigner, "whatever"} {
		t.Run(name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = cli.ReqSignerDesc(name)
			}
			require.NotPanics(t, testFunc, "ReqSignerDesc(%q)", name)
			assert.Contains(t, actual, "--"+name, "ReqSignerDesc(%q)", name)
			assert.Contains(t, actual, "<"+name+">", "ReqSignerDesc(%q)", name)
			assert.NotContains(t, actual, " "+name, "ReqSignerDesc(%q)", name)
			assert.Contains(t, actual, "--"+flags.FlagFrom, "ReqSignerDesc(%q)", name)
		})
	}
}

func TestReqSignerUse(t *testing.T) {
	for _, name := range []string{cli.FlagBuyer, cli.FlagSeller, cli.FlagSigner, "whatever"} {
		t.Run(name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = cli.ReqSignerUse(name)
			}
			require.NotPanics(t, testFunc, "ReqSignerUse(%q)", name)
			assert.Contains(t, actual, "--"+name, "ReqSignerUse(%q)", name)
			assert.Contains(t, actual, "<"+name+">", "ReqSignerUse(%q)", name)
			assert.Contains(t, actual, "--"+flags.FlagFrom, "ReqSignerUse(%q)", name)
		})
	}
}

func TestReqFlagUse(t *testing.T) {
	tests := []struct {
		name string
		opt  string
		exp  string
	}{
		{name: cli.FlagMarket, opt: "market id", exp: "--market <market id>"},
		{name: cli.FlagOrder, opt: "order id", exp: "--order <order id>"},
		{name: cli.FlagPrice, opt: "price", exp: "--price <price>"},
		{name: "whatever", opt: "stuff", exp: "--whatever <stuff>"},
		{name: cli.FlagAuthority, opt: "", exp: "--authority"},
		{name: cli.FlagEnable, opt: "", exp: "--enable"},
		{name: cli.FlagDisable, opt: "", exp: "--disable"},
		{name: "dance", opt: "", exp: "--dance"},
	}

	for _, tc := range tests {
		t.Run(tc.exp, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = cli.ReqFlagUse(tc.name, tc.opt)
			}
			require.NotPanics(t, testFunc, "ReqFlagUse(%q, %q)", tc.name, tc.opt)
			assert.Equal(t, tc.exp, actual, "ReqFlagUse(%q, %q)", tc.name, tc.opt)
		})
	}
}

func TestOptFlagUse(t *testing.T) {
	tests := []struct {
		name string
		opt  string
		exp  string
	}{
		{name: cli.FlagMarket, opt: "market id", exp: "[--market <market id>]"},
		{name: cli.FlagOrder, opt: "order id", exp: "[--order <order id>]"},
		{name: cli.FlagPrice, opt: "price", exp: "[--price <price>]"},
		{name: "whatever", opt: "stuff", exp: "[--whatever <stuff>]"},
		{name: cli.FlagAuthority, opt: "", exp: "[--authority]"},
		{name: cli.FlagEnable, opt: "", exp: "[--enable]"},
		{name: cli.FlagDisable, opt: "", exp: "[--disable]"},
		{name: "dance", opt: "", exp: "[--dance]"},
	}

	for _, tc := range tests {
		t.Run(tc.exp, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = cli.OptFlagUse(tc.name, tc.opt)
			}
			require.NotPanics(t, testFunc, "OptFlagUse(%q, %q)", tc.name, tc.opt)
			assert.Equal(t, tc.exp, actual, "OptFlagUse(%q, %q)", tc.name, tc.opt)
		})
	}
}

func TestProposalFileDesc(t *testing.T) {
	msgTypes := []sdk.Msg{
		&exchange.MsgGovCreateMarketRequest{},
		&exchange.MsgGovManageFeesRequest{},
	}

	for _, msgType := range msgTypes {
		t.Run(fmt.Sprintf("%T", msgType), func(t *testing.T) {
			expInRes := []string{
				"--proposal", "json", "Tx", "--generate-only",
				`{
  "body": {
    "messages": [
      {
        "@type": "` + sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}) + `",
        "messages": [
          {
            "@type": "` + sdk.MsgTypeURL(msgType) + `",
`,
			}

			var actual string
			testFunc := func() {
				actual = cli.ProposalFileDesc(msgType)
			}
			require.NotPanics(t, testFunc, "ProposalFileDesc(%T)", msgType)

			defer func() {
				if t.Failed() {
					t.Logf("Result:\n%s", actual)
				}
			}()

			for _, exp := range expInRes {
				assert.Contains(t, actual, exp, "ProposalFileDesc(%T) result. Should contain:\n%s", msgType, exp)
			}
		})
	}
}
