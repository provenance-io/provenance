package cmd

import (
	"io"
	"strings"
	"testing"
	"unicode"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"
)

func TestIAVLConfig(t *testing.T) {
	require.Equal(t, getIAVLCacheSize(sdksim.EmptyAppOptions{}), cast.ToInt(serverconfig.DefaultConfig().IAVLCacheSize))
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		str string
		exp string
	}{
		{str: "", exp: ""},
		{str: "camelCase", exp: "CamelCase"},
		{str: "PascalCase", exp: "PascalCase"},
		{str: "Three WoRds Nochange", exp: "Three WoRds Nochange"},
		{str: "three WorDs CHAngeD", exp: "Three WorDs CHAngeD"},
		{str: "œnon-ascii š", exp: "Œnon-ascii š"},
		{str: "<something else>", exp: "<something else>"},
	}

	for _, tc := range tests {
		name := tc.str
		if name == "" {
			name = "empty"
		}
		t.Run(name, func(t *testing.T) {
			act := capitalizeFirst(tc.str)
			assert.Equal(t, tc.exp, act)
		})
	}
}

func TestAllUsesStartUpperCased(t *testing.T) {
	rootCmd, _ := NewRootCmd(false)
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	// We don't care about the error here.
	// This /should/ just output provenanced help.
	// But we need to call it so that a) all flags have been added,
	// and b) the capitilzation has occurred.
	_ = Execute(rootCmd)

	assertFirstIsNotLower := func(t *testing.T, str, msg string, args ...interface{}) bool {
		if len(str) == 0 {
			return true
		}
		var first rune
		for _, r := range str {
			first = r
			break
		}
		isUpper := unicode.IsLower(first)
		return assert.Falsef(t, isUpper, "unicode.IsLower(%+q), the first letter of "+msg+": %+q",
			append([]interface{}{first}, append(args, str)...)...)
	}

	// Set this to true to have the test log out ALL the flags and usages.
	verbose := false

	assertFlagUsageFirstIsLower := func(t *testing.T, getter func() *pflag.FlagSet, name string) {
		if verbose {
			t.Logf("%s()", name)
		}
		var s *pflag.FlagSet
		// A panic can occur here when a command has a flag with a shorthand
		// that conflicts with a persistent flag of an ancestor.
		ok := assert.NotPanics(t, func() { s = getter() }, "%s()", name)
		if !ok {
			return
		}
		s.VisitAll(func(f *pflag.Flag) {
			if verbose {
				t.Logf("  --%s\t%s", f.Name, f.Usage)
			}
			assertFirstIsNotLower(t, f.Usage, "--%s flag usage", f.Name)
		})
	}

	var checkCmd func(t *testing.T, parents []string, cmd *cobra.Command)
	checkCmd = func(t *testing.T, parents []string, cmd *cobra.Command) {
		parents = append(parents, cmd.Name())
		t.Run(strings.Join(parents, " "), func(t *testing.T) {
			if verbose {
				t.Logf("%s short usage: %q", cmd.Name(), cmd.Short)
			}
			assertFirstIsNotLower(t, cmd.Short, "cmd.Short")

			assertFlagUsageFirstIsLower(t, cmd.NonInheritedFlags, "NonInheritedFlags")
			assertFlagUsageFirstIsLower(t, cmd.InheritedFlags, "InheritedFlags")
		})
		for _, sc := range cmd.Commands() {
			checkCmd(t, parents, sc)
		}
	}

	checkCmd(t, nil, rootCmd)
}
