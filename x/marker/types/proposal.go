package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeAddMarker to add a new marker
	ProposalTypeAddMarker string = "AddMarker"
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
	_ govtypes.Content = &AddMarkerProposal{}
	_ govtypes.Content = &SupplyIncreaseProposal{}
	_ govtypes.Content = &SupplyDecreaseProposal{}
	_ govtypes.Content = &SetAdministratorProposal{}
	_ govtypes.Content = &RemoveAdministratorProposal{}
	_ govtypes.Content = &ChangeStatusProposal{}
	_ govtypes.Content = &WithdrawEscrowProposal{}
	_ govtypes.Content = &SetDenomMetadataProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddMarker)
	govtypes.RegisterProposalTypeCodec(AddMarkerProposal{}, "provenance/marker/AddMarkerProposal")

	govtypes.RegisterProposalType(ProposalTypeIncreaseSupply)
	govtypes.RegisterProposalTypeCodec(SupplyIncreaseProposal{}, "provenance/marker/SupplyIncreaseProposal")
	govtypes.RegisterProposalType(ProposalTypeDecreaseSupply)
	govtypes.RegisterProposalTypeCodec(SupplyDecreaseProposal{}, "provenance/marker/SupplyDecreaseProposal")

	govtypes.RegisterProposalType(ProposalTypeSetAdministrator)
	govtypes.RegisterProposalTypeCodec(SetAdministratorProposal{}, "provenance/marker/SetAdministratorProposal")
	govtypes.RegisterProposalType(ProposalTypeRemoveAdministrator)
	govtypes.RegisterProposalTypeCodec(RemoveAdministratorProposal{}, "provenance/marker/RemoveAdministratorProposal")

	govtypes.RegisterProposalType(ProposalTypeChangeStatus)
	govtypes.RegisterProposalTypeCodec(ChangeStatusProposal{}, "provenance/marker/ChangeStatusProposal")

	govtypes.RegisterProposalType(ProposalTypeWithdrawEscrow)
	govtypes.RegisterProposalTypeCodec(WithdrawEscrowProposal{}, "provenance/marker/WithdrawEscrowProposal")

	govtypes.RegisterProposalType(ProposalTypeSetDenomMetadata)
	govtypes.RegisterProposalTypeCodec(SetDenomMetadataProposal{}, "provenance/marker/SetDenomMetadataProposal")
}

// NewAddMarkerProposal creates a new proposal
func NewAddMarkerProposal(
	title,
	description string,
	denom string,
	totalSupply sdk.Int,
	manager sdk.AccAddress,
	status MarkerStatus,
	markerType MarkerType,
	access []AccessGrant,
	fixed bool,
	allowGov bool, // nolint:interfacer
) *AddMarkerProposal {
	return &AddMarkerProposal{
		Title:                  title,
		Description:            description,
		Amount:                 sdk.NewCoin(denom, totalSupply),
		Manager:                manager.String(),
		Status:                 status,
		MarkerType:             markerType,
		AccessList:             access,
		SupplyFixed:            fixed,
		AllowGovernanceControl: allowGov,
	}
}

// Implements Proposal Interface

func (amp AddMarkerProposal) ProposalRoute() string { return RouterKey }
func (amp AddMarkerProposal) ProposalType() string  { return ProposalTypeAddMarker }
func (amp AddMarkerProposal) ValidateBasic() error {
	if amp.Status == StatusUndefined {
		return ErrInvalidMarkerStatus
	}
	// A proposed marker must have a manager assigned to allow updates to be made by the caller.
	if len(amp.Manager) == 0 && amp.Status == StatusProposed {
		return fmt.Errorf("marker manager cannot be empty when creating a proposed marker")
	}
	testCoin := sdk.Coin{
		Denom:  amp.Amount.Denom,
		Amount: amp.Amount.Amount,
	}
	if !testCoin.IsValid() {
		return fmt.Errorf("invalid marker denom/total supply: %w", sdkerrors.ErrInvalidCoins)
	}
	return govtypes.ValidateAbstract(&amp)
}

func (amp AddMarkerProposal) String() string {
	return fmt.Sprintf(`Add Marker Proposal:
  Marker:      %s
  Title:       %s
  Description: %s
  Supply:      %s
  Status:      %s
  Type:        %s
`, amp.Amount.Denom, amp.Title, amp.Description, amp.Amount.Amount, amp.Status, amp.MarkerType)
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
	return govtypes.ValidateAbstract(&sip)
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
	return govtypes.ValidateAbstract(&sdp)
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
	title, description, denom string, accessGrants []AccessGrant, // nolint:interfacer
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
	return govtypes.ValidateAbstract(&sap)
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

	return govtypes.ValidateAbstract(&rap)
}

func (rap RemoveAdministratorProposal) String() string {
	return fmt.Sprintf(`MarkerAccount Remove Administrator Proposal:
  Marker:      %s
  Title:       %s
  Description: %s
  Administrators To Remove: %v
`, rap.Denom, rap.Title, rap.Description, rap.RemovedAddress)
}

func NewChangeStatusProposal(title, description, denom string, status MarkerStatus) *ChangeStatusProposal { // nolint:interfacer
	return &ChangeStatusProposal{title, description, denom, status}
}

// Implements Proposal Interface

func (csp ChangeStatusProposal) ProposalRoute() string { return RouterKey }
func (csp ChangeStatusProposal) ProposalType() string  { return ProposalTypeChangeStatus }
func (csp ChangeStatusProposal) ValidateBasic() error {
	return govtypes.ValidateAbstract(&csp)
}

func (csp ChangeStatusProposal) String() string {
	return fmt.Sprintf(`MarkerAccount Change Status Proposal:
  Marker:      %s
  Title:       %s
  Description: %s
  Change Status To: %s
`, csp.Denom, csp.Title, csp.Description, csp.NewStatus)
}

func NewWithdrawEscrowProposal(title, description, denom string, amount sdk.Coins, target string) *WithdrawEscrowProposal { // nolint:interfacer
	return &WithdrawEscrowProposal{title, description, denom, amount, target}
}

// Implements Proposal Interface

func (wep WithdrawEscrowProposal) ProposalRoute() string { return RouterKey }
func (wep WithdrawEscrowProposal) ProposalType() string  { return ProposalTypeWithdrawEscrow }
func (wep WithdrawEscrowProposal) ValidateBasic() error {
	return govtypes.ValidateAbstract(&wep)
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
		return sdkerrors.Wrap(govtypes.ErrInvalidProposalContent, "invalid metadata: "+err.Error())
	}
	return govtypes.ValidateAbstract(&sdmdp)
}

func (sdmdp SetDenomMetadataProposal) String() string {
	return fmt.Sprintf(`Set Denom Metadata Proposal:
  Marker:      %s
  Title:       %s
  Description: %s
  Metadata:    %s
`, sdmdp.Metadata.Base, sdmdp.Title, sdmdp.Description, sdmdp.Metadata.String())
}
