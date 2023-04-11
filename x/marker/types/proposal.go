package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeIncreaseSupply to mint coins
	ProposalTypeIncreaseSupply string = "IncreaseSupply"
	// ProposalTypeDecreaseSupply to burn coins
	ProposalTypeDecreaseSupply string = "DecreaseSupply"
	// ProposalTypeSetAdministrator to set permissions for an account address on marker account
	ProposalTypeSetAdministrator string = "SetAdministrator"
	// ProposalTypeRemoveAdministrator to remove an existing address and all permissions from marker account
	ProposalTypeRemoveAdministrator string = "RemoveAdministrator"
	// ProposalTypeChangeStatus to transition the status of a marker account.
	ProposalTypeChangeStatus string = "ChangeStatus"
	// ProposalTypeWithdrawEscrow is a proposal to withdraw coins from marker escrow and transfer to a specified account
	ProposalTypeWithdrawEscrow string = "WithdrawEscrow"
	// ProposalTypeSetDenomMetadata is a proposal to set denom metatdata.
	ProposalTypeSetDenomMetadata string = "SetDenomMetadata"
)

var (
	_ govtypesv1beta1.Content = &SupplyIncreaseProposal{}
	_ govtypesv1beta1.Content = &SupplyDecreaseProposal{}
	_ govtypesv1beta1.Content = &SetAdministratorProposal{}
	_ govtypesv1beta1.Content = &RemoveAdministratorProposal{}
	_ govtypesv1beta1.Content = &ChangeStatusProposal{}
	_ govtypesv1beta1.Content = &WithdrawEscrowProposal{}
	_ govtypesv1beta1.Content = &SetDenomMetadataProposal{}
)

func init() {
	govtypesv1beta1.RegisterProposalType(ProposalTypeIncreaseSupply)
	govtypesv1beta1.RegisterProposalType(ProposalTypeDecreaseSupply)
	govtypesv1beta1.RegisterProposalType(ProposalTypeSetAdministrator)
	govtypesv1beta1.RegisterProposalType(ProposalTypeRemoveAdministrator)
	govtypesv1beta1.RegisterProposalType(ProposalTypeChangeStatus)
	govtypesv1beta1.RegisterProposalType(ProposalTypeWithdrawEscrow)
	govtypesv1beta1.RegisterProposalType(ProposalTypeSetDenomMetadata)
}

// NewSupplyIncreaseProposal creates a new proposal
func NewSupplyIncreaseProposal(title, description string, amount sdk.Coin, destination string) *SupplyIncreaseProposal {
	return &SupplyIncreaseProposal{title, description, amount, destination}
}

// Implements Proposal Interface

func (sip SupplyIncreaseProposal) ProposalRoute() string { return RouterKey }
func (sip SupplyIncreaseProposal) ProposalType() string  { return ProposalTypeIncreaseSupply }
func (sip SupplyIncreaseProposal) ValidateBasic() error {
	if sip.Amount.IsNegative() {
		return fmt.Errorf("amount to increase must be greater than zero")
	}
	return govtypesv1beta1.ValidateAbstract(&sip)
}

func (sip SupplyIncreaseProposal) String() string {
	return fmt.Sprintf(`MarkerAccount Token Supply Increase Proposal:
  Marker:      %s
  Title:       %s
  Description: %s
  Amount to Increase: %s
`, sip.Amount.Denom, sip.Title, sip.Description, sip.Amount.Amount.String())
}

// NewSupplyDecreaseProposal creates a new proposal
func NewSupplyDecreaseProposal(title, description string, amount sdk.Coin) *SupplyDecreaseProposal {
	return &SupplyDecreaseProposal{title, description, amount}
}

// Implements Proposal Interface

func (sdp SupplyDecreaseProposal) ProposalRoute() string { return RouterKey }
func (sdp SupplyDecreaseProposal) ProposalType() string  { return ProposalTypeDecreaseSupply }
func (sdp SupplyDecreaseProposal) ValidateBasic() error {
	if sdp.Amount.IsNegative() {
		return fmt.Errorf("amount to decrease must be greater than zero")
	}
	return govtypesv1beta1.ValidateAbstract(&sdp)
}

func (sdp SupplyDecreaseProposal) String() string {
	return fmt.Sprintf(`MarkerAccount Token Supply Decrease Proposal:
  Marker:      %s
  Title:       %s
  Description: %s
  Amount to Decrease: %s
`, sdp.Amount.Denom, sdp.Title, sdp.Description, sdp.Amount.Amount.String())
}

func NewSetAdministratorProposal(
	title, description, denom string, accessGrants []AccessGrant,
) *SetAdministratorProposal {
	return &SetAdministratorProposal{title, description, denom, accessGrants}
}

// Implements Proposal Interface

func (sap SetAdministratorProposal) ProposalRoute() string { return RouterKey }
func (sap SetAdministratorProposal) ProposalType() string  { return ProposalTypeSetAdministrator }
func (sap SetAdministratorProposal) ValidateBasic() error {
	for _, a := range sap.Access {
		if err := a.Validate(); err != nil {
			return fmt.Errorf("invalid access grant for administrator: %w", err)
		}
	}
	return govtypesv1beta1.ValidateAbstract(&sap)
}

func (sap SetAdministratorProposal) String() string {
	return fmt.Sprintf(`MarkerAccount Set Administrator Proposal:
  Marker:      %s
  Title:       %s
  Description: %s
  Administrator Access Grant: %v
`, sap.Denom, sap.Title, sap.Description, sap.Access)
}

func NewRemoveAdministratorProposal(
	title, description, denom string, administrators []string,
) *RemoveAdministratorProposal {
	return &RemoveAdministratorProposal{title, description, denom, administrators}
}

// Implements Proposal Interface

func (rap RemoveAdministratorProposal) ProposalRoute() string { return RouterKey }
func (rap RemoveAdministratorProposal) ProposalType() string  { return ProposalTypeRemoveAdministrator }
func (rap RemoveAdministratorProposal) ValidateBasic() error {
	for _, ra := range rap.RemovedAddress {
		if _, err := sdk.AccAddressFromBech32(ra); err != nil {
			return fmt.Errorf("administrator account address is invalid: %w", err)
		}
	}

	return govtypesv1beta1.ValidateAbstract(&rap)
}

func (rap RemoveAdministratorProposal) String() string {
	return fmt.Sprintf(`MarkerAccount Remove Administrator Proposal:
  Marker:      %s
  Title:       %s
  Description: %s
  Administrators To Remove: %v
`, rap.Denom, rap.Title, rap.Description, rap.RemovedAddress)
}

func NewChangeStatusProposal(title, description, denom string, status MarkerStatus) *ChangeStatusProposal {
	return &ChangeStatusProposal{title, description, denom, status}
}

// Implements Proposal Interface

func (csp ChangeStatusProposal) ProposalRoute() string { return RouterKey }
func (csp ChangeStatusProposal) ProposalType() string  { return ProposalTypeChangeStatus }
func (csp ChangeStatusProposal) ValidateBasic() error {
	return govtypesv1beta1.ValidateAbstract(&csp)
}

func (csp ChangeStatusProposal) String() string {
	return fmt.Sprintf(`MarkerAccount Change Status Proposal:
  Marker:      %s
  Title:       %s
  Description: %s
  Change Status To: %s
`, csp.Denom, csp.Title, csp.Description, csp.NewStatus)
}

func NewWithdrawEscrowProposal(title, description, denom string, amount sdk.Coins, target string) *WithdrawEscrowProposal {
	return &WithdrawEscrowProposal{title, description, denom, amount, target}
}

// Implements Proposal Interface

func (wep WithdrawEscrowProposal) ProposalRoute() string { return RouterKey }
func (wep WithdrawEscrowProposal) ProposalType() string  { return ProposalTypeWithdrawEscrow }
func (wep WithdrawEscrowProposal) ValidateBasic() error {
	return govtypesv1beta1.ValidateAbstract(&wep)
}

func (wep WithdrawEscrowProposal) String() string {
	return fmt.Sprintf(`MarkerAccount Withdraw Escrow Proposal:
  Marker:      %s
  Title:       %s
  Description: %s
  Withdraw %s and transfer to %s
`, wep.Denom, wep.Title, wep.Description, wep.Amount, wep.TargetAddress)
}

func NewSetDenomMetadataProposal(title, description string, metadata banktypes.Metadata) *SetDenomMetadataProposal {
	return &SetDenomMetadataProposal{
		Title:       title,
		Description: description,
		Metadata:    metadata,
	}
}

// Implements Proposal Interface

func (sdmdp SetDenomMetadataProposal) ProposalRoute() string { return RouterKey }
func (sdmdp SetDenomMetadataProposal) ProposalType() string  { return ProposalTypeSetDenomMetadata }
func (sdmdp SetDenomMetadataProposal) ValidateBasic() error {
	if err := sdmdp.Metadata.Validate(); err != nil {
		return govtypes.ErrInvalidProposalContent.Wrap("invalid metadata: " + err.Error())
	}
	return govtypesv1beta1.ValidateAbstract(&sdmdp)
}

func (sdmdp SetDenomMetadataProposal) String() string {
	return fmt.Sprintf(`Set Denom Metadata Proposal:
  Marker:      %s
  Title:       %s
  Description: %s
  Metadata:    %s
`, sdmdp.Metadata.Base, sdmdp.Title, sdmdp.Description, sdmdp.Metadata.String())
}
