package cli_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/x/attribute/client/cli"
)

func TestAccountDataFlagsUse(t *testing.T) {
	// This checks that the AccountDataFlagsUse is as expected because
	// if it changes, the other tests in here will probably also fail.
	expected := "{--value <value>|--file <file>|--delete}"
	assert.Equal(t, expected, cli.AccountDataFlagsUse, "AccountDataFlagsUse")
}

func TestAddAccountDataFlagsToCmd(t *testing.T) {
	// create a dummy command and add the flags to it.
	dummyCmd := &cobra.Command{
		Use: "dummy " + cli.AccountDataFlagsUse,
		Run: func(cmd *cobra.Command, args []string) {
			panic("this dummy command should not be executed")
		},
	}
	cli.AddAccountDataFlagsToCmd(dummyCmd)

	// Now get the flags back out so we can make sure they're all defined.
	flags := dummyCmd.Flags()

	// Get the flag use string, split it on each line.
	flagUseStr := flags.FlagUsages()
	t.Logf("Flag Usages:\n%s", flagUseStr)
	// Split them into individual lines.
	// Trim leading/trailing spaces and replace repeated spaces with a single space
	// so that I don't have to worry about columnar spacing applied to the strings.
	flagUses := strings.Split(flagUseStr, "\n")
	spaceRx := regexp.MustCompile(`\s{2,}`)
	for i := range flagUses {
		flagUses[i] = strings.TrimSpace(flagUses[i])
		flagUses[i] = spaceRx.ReplaceAllString(flagUses[i], " ")
	}

	tests := []struct {
		name  string
		usage string
		arg   string
	}{
		{
			name:  "value",
			usage: "The value to set the account data to",
			arg:   "string",
		},
		{
			name:  "file",
			usage: "A file containing the value to set the account data to",
			arg:   "string",
		},
		{
			name:  "delete",
			usage: "The account data should be deleted",
			arg:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := flags.Lookup(tc.name)
			if assert.NotNilf(t, f, "flags.Lookup(%q)", tc.name) {
				assert.Equal(t, tc.name, f.Name)
				assert.Equal(t, tc.usage, f.Usage, "flag usage")
			}

			flagUse := "--" + tc.name
			if len(tc.arg) > 0 {
				flagUse = flagUse + " " + tc.arg
			}
			useLine := fmt.Sprintf("%s %s", flagUse, tc.usage)
			assert.Contains(t, flagUses, useLine, "cleaned flag usage lines")
		})
	}
}

func TestReadAccountDataFlags(t *testing.T) {
	tmpDir := t.TempDir()
	fileTXT := filepath.Join(tmpDir, "file1.txt")
	fileJSON := filepath.Join(tmpDir, "file2.json")
	fileBIN := filepath.Join(tmpDir, "file3.whatever")
	fileDNE := filepath.Join(tmpDir, "file4.does.not.exist")

	fileTXTContent := "This is the content of a text file.\nIt has two lines.\n"
	fileJSONContent := `{"field1":"value1","field2":true}`
	fileBINContentBz := []byte{0x01, 0x02, 0x03, 0x04}
	fileBINContent := string(fileBINContentBz)

	require.NoError(t, os.WriteFile(fileTXT, []byte(fileTXTContent), 0644), "writing %s", fileTXT)
	require.NoError(t, os.WriteFile(fileJSON, []byte(fileJSONContent), 0644), "writing %s", fileJSON)
	require.NoError(t, os.WriteFile(fileBIN, fileBINContentBz, 0644), "writing %s", fileBINContent)

	someValue := "This is really some value."

	fValue := "--value"
	fFile := "--file"
	fDelete := "--delete"

	noFlagsErr := "exactly one of these must be provided: " + cli.AccountDataFlagsUse
	bothValueFileErr := "cannot provide more than one of these: {--value <value>|--file <file>}"
	bothValueDelErr := "cannot provide more than one of these: {--value <value>|--delete}"
	bothFileDelErr := "cannot provide more than one of these: {--file <file>|--delete}"
	allThreeErr := "cannot provide more than one of these: " + cli.AccountDataFlagsUse

	// newDummyCMD creates a new dummy command without any flags defined on it.
	newDummyCMD := func() *cobra.Command {
		return &cobra.Command{
			Use: "dummy",
			Run: func(cmd *cobra.Command, args []string) {
				panic("this dummy command should not be executed")
			},
		}
	}

	tests := []struct {
		name   string
		args   []string
		maker  func() *cobra.Command
		expVal string
		expErr string
	}{
		{
			name:   "no flags",
			args:   []string{},
			expErr: noFlagsErr,
		},
		{
			name:   "value flag",
			args:   []string{fValue, "This is a test value."},
			expVal: "This is a test value.",
		},
		{
			name:   "file flag with text file",
			args:   []string{fFile, fileTXT},
			expVal: fileTXTContent,
		},
		{
			name:   "file flag with json file",
			args:   []string{fFile, fileJSON},
			expVal: fileJSONContent,
		},
		{
			name:   "file flag with binary file",
			args:   []string{fFile, fileBIN},
			expVal: fileBINContent,
		},
		{
			name:   "file flag with file that does not exist",
			args:   []string{fFile, fileDNE},
			expErr: "failed to read value from --file: open " + fileDNE + `: no such file or directory`,
		},
		{
			name:   "file flag with directory",
			args:   []string{fFile, tmpDir},
			expErr: "failed to read value from --file: read " + tmpDir + `: is a directory`,
		},
		{
			name:   "delete flag",
			args:   []string{fDelete},
			expVal: "",
			expErr: "",
		},
		{
			name:   "both value and file flags",
			args:   []string{fValue, someValue, fFile, fileTXT},
			expErr: bothValueFileErr,
		},
		{
			name:   "both file and value flags",
			args:   []string{fFile, fileTXT, fValue, someValue},
			expErr: bothValueFileErr,
		},
		{
			name:   "both value and delete flags",
			args:   []string{fValue, someValue, fDelete},
			expErr: bothValueDelErr,
		},
		{
			name:   "both delete and value flags",
			args:   []string{fDelete, fValue, someValue},
			expErr: bothValueDelErr,
		},
		{
			name:   "both file and delete flags",
			args:   []string{fFile, fileJSON, fDelete},
			expErr: bothFileDelErr,
		},
		{
			name:   "both delete and file flags",
			args:   []string{fDelete, fFile, fileJSON},
			expErr: bothFileDelErr,
		},
		{
			name:   "all three flags value file delete",
			args:   []string{fValue, someValue, fFile, fileDNE, fDelete},
			expErr: allThreeErr,
		},
		{
			name:   "all three flags value delete file",
			args:   []string{fValue, someValue, fDelete, fFile, fileDNE},
			expErr: allThreeErr,
		},
		{
			name:   "all three flags file value delete",
			args:   []string{fFile, fileDNE, fValue, someValue, fDelete},
			expErr: allThreeErr,
		},
		{
			name:   "all three flags file delete value",
			args:   []string{fFile, fileDNE, fDelete, fValue, someValue},
			expErr: allThreeErr,
		},
		{
			name:   "all three flags delete value file",
			args:   []string{fDelete, fValue, someValue, fFile, fileDNE},
			expErr: allThreeErr,
		},
		{
			name:   "all three flags delete file value",
			args:   []string{fDelete, fFile, fileDNE, fValue, someValue},
			expErr: allThreeErr,
		},
		{
			name: "value flag not defined",
			args: []string{},
			maker: func() *cobra.Command {
				// Define all applicable flags except the value flag.
				cmd := newDummyCMD()
				cmd.Flags().String("file", "", "a file")
				cmd.Flags().Bool("delete", false, "delete")
				return cmd
			},
			expErr: "failed to read --value <value> flag: flag accessed but not defined: value",
		},
		{
			name: "file flag not defined",
			args: []string{},
			maker: func() *cobra.Command {
				// Define all applicable flags except the file flag.
				cmd := newDummyCMD()
				cmd.Flags().String("value", "", "a value")
				cmd.Flags().Bool("delete", false, "delete")
				return cmd
			},
			expErr: "failed to read --file <file> flag: flag accessed but not defined: file",
		},
		{
			name: "delete flag not defined",
			args: []string{},
			maker: func() *cobra.Command {
				// Define all applicable flags except the delete flag.
				cmd := newDummyCMD()
				cmd.Flags().String("value", "", "a value")
				cmd.Flags().String("file", "", "a file")
				return cmd
			},
			expErr: "failed to read --delete flag: flag accessed but not defined: delete",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.maker == nil {
				tc.maker = func() *cobra.Command {
					// Dummy command with all applicable flags added as (hopefully) normal.
					dummyCmd := newDummyCMD()
					cli.AddAccountDataFlagsToCmd(dummyCmd)
					return dummyCmd
				}
			}
			cmd := tc.maker()
			// Flag parsing happens before there's a chance to call ReadAccountDataFlags.
			// Here, I assume that cobra is doing things normally, so test cases that would cause
			// a flag parsing error are not covered in these unit tests.
			// Example failures: providing "--value" without a value, or providing "--value" and
			// "<value>" without it being defined on in the command.
			require.NoError(t, cmd.ParseFlags(tc.args), "ParseFlags")

			flags := cmd.Flags()
			value, err := cli.ReadAccountDataFlags(flags)
			if len(tc.expErr) > 0 {
				assert.EqualError(t, err, tc.expErr, "ReadAccountDataFlags error")
			} else {
				assert.NoError(t, err, tc.expErr, "ReadAccountDataFlags error")
			}
			assert.Equal(t, tc.expVal, value, "ReadAccountDataFlags value")
		})
	}
}
