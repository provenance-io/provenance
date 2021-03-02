package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeAddMarker to add a new marker
	ProposalTypeAddMarker string = "AddMarker"
	// ProposalTypeIncreaseSupply to mint coins
	ProposalTypeIncreaseSupply string = "IncreaseSupply"
	// ProposalTypeDecreaseSupply to burn coins
	ProposalTypeDecreaseSupply string = "DescreaseSupply"
	// ProposalTypeSetAdministrator to set permissions for an account address on marker account
	ProposalTypeSetAdministrator string = "SetAdministrator"
	// ProposalTypeRemoveAdministrator to remove an existing address and all permissions from marker account
	ProposalTypeRemoveAdministrator string = "RemoveAdministrator"
	// ProposalTypeChangeStatus to transition the status of a marker account.
	ProposalTypeChangeStatus string = "ChangeStatus"
)

var (
	_ govtypes.Content = &AddMarkerProposal{}
	_ govtypes.Content = &SupplyIncreaseProposal{}
	_ govtypes.Content = &SupplyDecreaseProposal{}
	_ govtypes.Content = &SetAdministratorProposal{}
	_ govtypes.Content = &RemoveAdministratorProposal{}
	_ govtypes.Content = &ChangeStatusProposal{}
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
}

// NewAddMarkerProposal creates a new proposal
func NewAddMarkerProposal(
	title,
	description string,
	denom string,
	totalSupply sdk.Int,
	fromAddress sdk.AccAddress,
	manager sdk.AccAddress,
	markerType MarkerType, // nolint:interfacer
) *AddMarkerProposal {
	return &AddMarkerProposal{
		Title:       title,
		Description: description,
		Amount:      sdk.NewCoin(denom, totalSupply),
		Manager:     manager.String(),
		Status:      StatusProposed,
		MarkerType:  markerType,
	}
}

// Implements Proposal Interface

func (amp AddMarkerProposal) ProposalRoute() string { return RouterKey }
func (amp AddMarkerProposal) ProposalType() string  { return ProposalTypeIncreaseSupply }
func (amp AddMarkerProposal) ValidateBasic() error {
	if amp.Status == StatusUndefined {
		return ErrInvalidMarkerStatus
	}
	// A proposed marker must have a manager assigned to allow updates to be made by the caller.
	if len(amp.Manager) == 0 && amp.Status == StatusProposed {
		return fmt.Errorf("marker manage cannot be empty when creating a proposed marker")
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
  Supply:      %d
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
  Amount to Increase: %d
`, sip.Amount.Denom, sip.Title, sip.Description, sip.Amount.Amount)
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
  Amount to Decrease: %d
`, sdp.Amount.Denom, sdp.Title, sdp.Description, sdp.Amount.Amount)
}

func NewSetAdministratorProposal(
	title, description string, marker sdk.AccAddress, accessGrants []AccessGrant, // nolint:interfacer
) *SetAdministratorProposal {
	return &SetAdministratorProposal{title, description, marker.String(), accessGrants}
}

// Implements Proposal Interface

func (sap SetAdministratorProposal) ProposalRoute() string { return RouterKey }
func (sap SetAdministratorProposal) ProposalType() string  { return ProposalTypeDecreaseSupply }
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
	title, description string, denom string, administrator sdk.AccAddress,
) *RemoveAdministratorProposal {
	return &RemoveAdministratorProposal{title, description, denom, []string{administrator.String()}}
}

// Implements Proposal Interface

func (rap RemoveAdministratorProposal) ProposalRoute() string { return RouterKey }
func (rap RemoveAdministratorProposal) ProposalType() string  { return ProposalTypeDecreaseSupply }
func (rap RemoveAdministratorProposal) ValidateBasic() error {
	for _, ra := range rap.RemovedAddress {
		if err := sdk.VerifyAddressFormat([]byte(ra)); err != nil {
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

func NewChangeStatusProposal(title, description string, marker sdk.AccAddress, status MarkerStatus) *ChangeStatusProposal { // nolint:interfacer
	return &ChangeStatusProposal{title, description, marker.String(), status}
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
  Change Status To : %s
`, csp.Denom, csp.Title, csp.Description, csp.NewStatus)
}
