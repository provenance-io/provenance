package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/version"
)

const (
	// FlagOwner is the flag for the owner address of a name record.
	FlagOwner = "owner"
	// FlagDescription is the flag for a description.
	FlagDescription = "description"
	// FlagTitle is the flag for a title.
	FlagTitle = "title"
	// FlagDeposit is the flag for a deposit.
	FlagDeposit string = "deposit"
	// FlagMetadata is the flag for a deposit.
	FlagMetadata string = "metadata"
)

// GetRootNameProposalCmd returns a command for registration with the gov module
func GetRootNameProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "root-name-proposal [name] (--owner [address]) (--unrestrict) [flags]",
		Short: "Submit a root name creation governance proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a root name creation governance proposal along with an initial deposit.
The proposal title and description must be provided through their respective flags.

IMPORTANT: The created root name will restrict the creation of sub-names by default unless the
unrestrict flag is included. The proposer will be the default owner that must approve
all child name creation unless an alterate owner is provided.

Example:
$ %s tx gov submit-legacy-proposal \
    root-name-proposal \
	<root name> \
	--unrestrict  \ 
	--owner <key_or_address> \
	--title "Proposal title" \
	--description "Description of proposal" 
	--from <key_or_address>
			`,
				version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("this command has been deprecated, and is no longer functional. Please use 'gov proposal submit-proposal' instead")
		},
	}

	cmd.Flags().String(FlagOwner, "", "The owner of the new name, optional (defaults to from address)")
	cmd.Flags().BoolP(FlagUnrestricted, "u", false, "Allow child name creation by everyone")
	// proposal flags
	cmd.Flags().String(FlagTitle, "", "Title of proposal")
	cmd.Flags().String(FlagDescription, "", "Description of proposal")
	cmd.Flags().String(FlagDeposit, "", "Deposit of proposal")
	return cmd
}
