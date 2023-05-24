package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

const (
	// flagValue is a flag name for defining a value.
	flagValue = "value"
	// flagValueUse is a use string for the value flag.
	flagValueUse = "--" + flagValue + " <value>"
	// flagFile is a flag name for defining a file.
	flagFile = "file"
	// flagFileUse is a use string for the file flag.
	flagFileUse = "--" + flagFile + " <file>"
	// flagDelete is a flag name for deleting something.
	flagDelete = "delete"
	// flagDeleteUse is a use string for the delete flag.
	flagDeleteUse = "--" + flagDelete

	// AccountDataFlagsUse is a use string for the mutually exclusive account data flags.
	AccountDataFlagsUse = "{" + flagValueUse + "|" + flagFileUse + "|" + flagDeleteUse + "}"
)

// AddAccountDataFlagsToCmd adds flags to a command for providing account data.
// See also: ReadAccountDataFlags
func AddAccountDataFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().String(flagValue, "", "The value to set the account data to")
	cmd.Flags().String(flagFile, "", "A file containing the value to set the account data to")
	cmd.Flags().Bool(flagDelete, false, "The account data should be deleted")
}

// ReadAccountDataFlags parses the account data flags and returns the desired account data value.
// See also: AddAccountDataFlagsToCmd
func ReadAccountDataFlags(flagSet *flag.FlagSet) (string, error) {
	// Read all the flag values.
	value, err := flagSet.GetString(flagValue)
	if err != nil {
		return "", fmt.Errorf("failed to read %s flag: %w", flagValueUse, err)
	}
	file, err := flagSet.GetString(flagFile)
	if err != nil {
		return "", fmt.Errorf("failed to read %s flag: %w", flagFileUse, err)
	}
	deleteValue, err := flagSet.GetBool(flagDelete)
	if err != nil {
		return "", fmt.Errorf("failed to read %s flag: %w", flagDeleteUse, err)
	}

	// Make sure that exactly one of them was provided.
	provided := make([]string, 0, 1)
	if len(value) > 0 {
		provided = append(provided, flagValueUse)
	}
	if len(file) > 0 {
		provided = append(provided, flagFileUse)
	}
	if deleteValue {
		provided = append(provided, flagDeleteUse)
	}

	if len(provided) == 0 {
		return "", fmt.Errorf("exactly one of these must be provided: %s", AccountDataFlagsUse)
	}
	if len(provided) > 1 {
		return "", fmt.Errorf("cannot provide more than one of these: {%s}", strings.Join(provided, "|"))
	}

	// Wooo! Exactly one was provided.

	if deleteValue {
		return "", nil
	}

	if len(value) > 0 {
		return value, nil
	}

	bz, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("failed to read value from --%s: %w", flagFile, err)
	}
	return string(bz), nil
}
