package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeCreateRootName defines the type for a CreateRootNameProposal
	ProposalTypeCreateRootName = "CreateRootName"
	ProposalTypeModifyName     = "ModifyName"
)

// Assert CreateRootNameProposal implements govtypesv1beta1.Content at compile-time
var _ govtypesv1beta1.Content = &CreateRootNameProposal{}

func init() {
	govtypesv1beta1.RegisterProposalType(ProposalTypeCreateRootName)
	govtypesv1beta1.RegisterProposalType(ProposalTypeModifyName)
}

// NewCreateRootNameProposal create a new governance proposal request to create a root name
//
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
	err := govtypesv1beta1.ValidateAbstract(crnp)
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
