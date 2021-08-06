package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeCreateRootName defines the type for a CreateRootNameProposal
	ProposalTypeCreateRootName = "CreateRootName"
)

// Assert CreateRootNameProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &CreateRootNameProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeCreateRootName)
	govtypes.RegisterProposalTypeCodec(&CreateRootNameProposal{}, "provenance/CreateRootNameProposal")
}

// NewCreateRootNameProposal create a new governance proposal request to create a root name
//nolint:interfacer
func NewCreateRootNameProposal(title, description, name string, owner sdk.AccAddress, restricted bool) *CreateRootNameProposal {
	return &CreateRootNameProposal{
		Title:       title,
		Description: description,
		Name:        name,
		Owner:       owner.String(),
		Restricted:  restricted,
	}
}

// GetTitle returns the title of a community pool spend proposal.
func (crnp CreateRootNameProposal) GetTitle() string { return crnp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (crnp CreateRootNameProposal) GetDescription() string { return crnp.Description }

// ProposalRoute returns the routing key of a community pool spend proposal.
func (crnp CreateRootNameProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (crnp CreateRootNameProposal) ProposalType() string { return ProposalTypeCreateRootName }

// ValidateBasic runs basic stateless validity checks
func (crnp CreateRootNameProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(crnp)
	if err != nil {
		return err
	}
	if strings.TrimSpace(crnp.Owner) != "" {
		if _, err := sdk.AccAddressFromBech32(crnp.Owner); err != nil {
			return ErrInvalidAddress
		}
	}
	if strings.TrimSpace(crnp.Name) == "" {
		return ErrInvalidLengthName
	}
	if strings.Contains(crnp.Name, ".") {
		return ErrNameContainsSegments
	}

	return nil
}

// String implements the Stringer interface.
func (crnp CreateRootNameProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Create Root Name Proposal:
  Title:       %s
  Description: %s
  Owner:       %s
  Name:        %s
  Restricted:  %v
`, crnp.Title, crnp.Description, crnp.Owner, crnp.Name, crnp.Restricted))
	return b.String()
}
