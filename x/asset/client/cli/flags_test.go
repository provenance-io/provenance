package cli_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/x/asset/client/cli"
)

func TestAddFlagsURI(t *testing.T) {
	tests := []struct {
		name            string
		typeName        string
		expURIUsage     string
		expURIHashUsage string
	}{
		{
			name:            "asset",
			typeName:        "asset",
			expURIUsage:     "URI of the asset",
			expURIHashUsage: "hash of the content of the asset URI",
		},
		{
			name:            "metadata",
			typeName:        "metadata",
			expURIUsage:     "URI of the metadata",
			expURIHashUsage: "hash of the content of the metadata URI",
		},
		{
			name:            "empty type name",
			typeName:        "",
			expURIUsage:     "URI of the ",
			expURIHashUsage: "hash of the content of the  URI",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "testing",
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			}

			cli.AddFlagsURI(cmd, tc.typeName)

			uriFlag := cmd.Flags().Lookup(cli.FlagURI)
			if assert.NotNil(t, uriFlag, "The --%s flag", cli.FlagURI) {
				assert.Equal(t, tc.expURIUsage, uriFlag.Usage, "The --%s flag usage", cli.FlagURI)
			}

			uriHashFlag := cmd.Flags().Lookup(cli.FlagURIHash)
			if assert.NotNil(t, uriHashFlag, "The --%s flag", cli.FlagURIHash) {
				assert.Equal(t, tc.expURIHashUsage, uriHashFlag.Usage, "The --%s flag usage", cli.FlagURIHash)
			}
		})
	}
}

func TestReadFlagURI(t *testing.T) {
	tests := []struct {
		name    string
		flagSet func() *pflag.FlagSet
		flags   []string
		expVal  string
	}{
		{
			name: "flag not defined",
			flagSet: func() *pflag.FlagSet {
				return pflag.NewFlagSet("", pflag.ContinueOnError)
			},
			expVal: "",
		},
		{
			name: "flag provided with value",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagURI, "", "The URI")
				return flagSet
			},
			flags:  []string{"--" + cli.FlagURI, "https://example.com/uri"},
			expVal: "https://example.com/uri",
		},
		{
			name: "flag not provided",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagURI, "", "The URI")
				return flagSet
			},
			expVal: "",
		},
		{
			name: "flag provided with empty value",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagURI, "", "The URI")
				return flagSet
			},
			flags:  []string{"--" + cli.FlagURI, ""},
			expVal: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flagSet := tc.flagSet()
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var val string
			testFunc := func() {
				val = cli.ReadFlagURI(flagSet)
			}
			require.NotPanics(t, testFunc, "ReadFlagURI")
			assert.Equal(t, tc.expVal, val, "ReadFlagURI value")
		})
	}
}

func TestReadFlagURIHash(t *testing.T) {
	tests := []struct {
		name    string
		flagSet func() *pflag.FlagSet
		flags   []string
		expVal  string
	}{
		{
			name: "flag not defined",
			flagSet: func() *pflag.FlagSet {
				return pflag.NewFlagSet("", pflag.ContinueOnError)
			},
			expVal: "",
		},
		{
			name: "flag provided with value",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagURIHash, "", "The URI hash")
				return flagSet
			},
			flags:  []string{"--" + cli.FlagURIHash, "abc123def456"},
			expVal: "abc123def456",
		},
		{
			name: "flag not provided",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagURIHash, "", "The URI hash")
				return flagSet
			},
			expVal: "",
		},
		{
			name: "flag provided with empty value",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagURIHash, "", "The URI hash")
				return flagSet
			},
			flags:  []string{"--" + cli.FlagURIHash, ""},
			expVal: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flagSet := tc.flagSet()
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var val string
			testFunc := func() {
				val = cli.ReadFlagURIHash(flagSet)
			}
			require.NotPanics(t, testFunc, "ReadFlagURIHash")
			assert.Equal(t, tc.expVal, val, "ReadFlagURIHash value")
		})
	}
}

func TestAddFlagSymbol(t *testing.T) {
	cmd := &cobra.Command{
		Use: "testing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cli.AddFlagSymbol(cmd)

	symbolFlag := cmd.Flags().Lookup(cli.FlagSymbol)
	if assert.NotNil(t, symbolFlag, "The --%s flag", cli.FlagSymbol) {
		expUsage := "symbol of the asset class"
		assert.Equal(t, expUsage, symbolFlag.Usage, "The --%s flag usage", cli.FlagSymbol)
	}
}

func TestReadFlagSymbol(t *testing.T) {
	tests := []struct {
		name    string
		flagSet func() *pflag.FlagSet
		flags   []string
		expVal  string
	}{
		{
			name: "flag not defined",
			flagSet: func() *pflag.FlagSet {
				return pflag.NewFlagSet("", pflag.ContinueOnError)
			},
			expVal: "",
		},
		{
			name: "flag provided with value",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagSymbol, "", "The symbol")
				return flagSet
			},
			flags:  []string{"--" + cli.FlagSymbol, "HASH"},
			expVal: "HASH",
		},
		{
			name: "flag not provided",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagSymbol, "", "The symbol")
				return flagSet
			},
			expVal: "",
		},
		{
			name: "flag provided with empty value",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagSymbol, "", "The symbol")
				return flagSet
			},
			flags:  []string{"--" + cli.FlagSymbol, ""},
			expVal: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flagSet := tc.flagSet()
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var val string
			testFunc := func() {
				val = cli.ReadFlagSymbol(flagSet)
			}
			require.NotPanics(t, testFunc, "ReadFlagSymbol")
			assert.Equal(t, tc.expVal, val, "ReadFlagSymbol value")
		})
	}
}

func TestAddFlagDescription(t *testing.T) {
	cmd := &cobra.Command{
		Use: "testing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cli.AddFlagDescription(cmd)

	descFlag := cmd.Flags().Lookup(cli.FlagDescription)
	if assert.NotNil(t, descFlag, "The --%s flag", cli.FlagDescription) {
		expUsage := "description of the asset class"
		assert.Equal(t, expUsage, descFlag.Usage, "The --%s flag usage", cli.FlagDescription)
	}
}

func TestReadFlagDescription(t *testing.T) {
	tests := []struct {
		name    string
		flagSet func() *pflag.FlagSet
		flags   []string
		expVal  string
	}{
		{
			name: "flag not defined",
			flagSet: func() *pflag.FlagSet {
				return pflag.NewFlagSet("", pflag.ContinueOnError)
			},
			expVal: "",
		},
		{
			name: "flag provided with value",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagDescription, "", "The description")
				return flagSet
			},
			flags:  []string{"--" + cli.FlagDescription, "A detailed description"},
			expVal: "A detailed description",
		},
		{
			name: "flag not provided",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagDescription, "", "The description")
				return flagSet
			},
			expVal: "",
		},
		{
			name: "flag provided with empty value",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagDescription, "", "The description")
				return flagSet
			},
			flags:  []string{"--" + cli.FlagDescription, ""},
			expVal: "",
		},
		{
			name: "flag provided with multi-word value",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagDescription, "", "The description")
				return flagSet
			},
			flags:  []string{"--" + cli.FlagDescription, "This is a long description with many words"},
			expVal: "This is a long description with many words",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flagSet := tc.flagSet()
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var val string
			testFunc := func() {
				val = cli.ReadFlagDescription(flagSet)
			}
			require.NotPanics(t, testFunc, "ReadFlagDescription")
			assert.Equal(t, tc.expVal, val, "ReadFlagDescription value")
		})
	}
}

func TestAddFlagOwner(t *testing.T) {
	cmd := &cobra.Command{
		Use: "testing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cli.AddFlagOwner(cmd)

	ownerFlag := cmd.Flags().Lookup(cli.FlagOwner)
	if assert.NotNil(t, ownerFlag, "The --%s flag", cli.FlagOwner) {
		expUsage := "owner address"
		assert.Equal(t, expUsage, ownerFlag.Usage, "The --%s flag usage", cli.FlagOwner)
	}
}

func TestReadFlagOwner(t *testing.T) {
	tests := []struct {
		name    string
		flagSet func() *pflag.FlagSet
		flags   []string
		expVal  string
	}{
		{
			name: "flag not defined",
			flagSet: func() *pflag.FlagSet {
				return pflag.NewFlagSet("", pflag.ContinueOnError)
			},
			expVal: "",
		},
		{
			name: "flag provided with value",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagOwner, "", "The owner")
				return flagSet
			},
			flags:  []string{"--" + cli.FlagOwner, "pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk"},
			expVal: "pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk",
		},
		{
			name: "flag not provided",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagOwner, "", "The owner")
				return flagSet
			},
			expVal: "",
		},
		{
			name: "flag provided with empty value",
			flagSet: func() *pflag.FlagSet {
				flagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
				flagSet.String(cli.FlagOwner, "", "The owner")
				return flagSet
			},
			flags:  []string{"--" + cli.FlagOwner, ""},
			expVal: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flagSet := tc.flagSet()
			err := flagSet.Parse(tc.flags)
			require.NoError(t, err, "flagSet.Parse(%q)", tc.flags)

			var val string
			testFunc := func() {
				val = cli.ReadFlagOwner(flagSet)
			}
			require.NotPanics(t, testFunc, "ReadFlagOwner")
			assert.Equal(t, tc.expVal, val, "ReadFlagOwner value")
		})
	}
}
