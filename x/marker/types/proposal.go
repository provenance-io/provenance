package types

import (
	"errors"

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
	_ govtypesv1beta1.Content = &AddMarkerProposal{}
	_ govtypesv1beta1.Content = &SupplyIncreaseProposal{}
	_ govtypesv1beta1.Content = &SupplyDecreaseProposal{}
	_ govtypesv1beta1.Content = &SetAdministratorProposal{}
	_ govtypesv1beta1.Content = &RemoveAdministratorProposal{}
	_ govtypesv1beta1.Content = &ChangeStatusProposal{}
	_ govtypesv1beta1.Content = &WithdrawEscrowProposal{}
	_ govtypesv1beta1.Content = &SetDenomMetadataProposal{}
)

// ProposalRoute returns the governance proposal route for AddMarkerProposal.
func (p AddMarkerProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the AddMarkerProposal.
func (p AddMarkerProposal) ProposalType() string { return "AddMarker" }

// ValidateBasic performs basic validation on AddMarkerProposal fields.
func (p AddMarkerProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

// ProposalRoute returns the governance proposal route for SupplyIncreaseProposal.
func (sip SupplyIncreaseProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the SupplyIncreaseProposal.
func (sip SupplyIncreaseProposal) ProposalType() string { return ProposalTypeIncreaseSupply }

// ValidateBasic performs basic validation on SupplyIncreaseProposal fields.
func (sip SupplyIncreaseProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

// ProposalRoute returns the governance proposal route for SupplyDecreaseProposal.
func (sdp SupplyDecreaseProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the SupplyDecreaseProposal.
func (sdp SupplyDecreaseProposal) ProposalType() string { return ProposalTypeDecreaseSupply }

// ValidateBasic performs basic validation on SupplyDecreaseProposal fields.
func (sdp SupplyDecreaseProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

// ProposalRoute returns the governance proposal route for SetAdministratorProposal.
func (sap SetAdministratorProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the SetAdministratorProposal.
func (sap SetAdministratorProposal) ProposalType() string { return ProposalTypeSetAdministrator }

// ValidateBasic performs basic validation on SetAdministratorProposal fields.
func (sap SetAdministratorProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

// ProposalRoute returns the governance proposal route for RemoveAdministratorProposal.
func (rap RemoveAdministratorProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the RemoveAdministratorProposal.
func (rap RemoveAdministratorProposal) ProposalType() string { return ProposalTypeRemoveAdministrator }

// ValidateBasic performs basic validation on RemoveAdministratorProposal fields.
func (rap RemoveAdministratorProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

// ProposalRoute returns the governance proposal route for ChangeStatusProposal.
func (csp ChangeStatusProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the ChangeStatusProposal.
func (csp ChangeStatusProposal) ProposalType() string { return ProposalTypeChangeStatus }

// ValidateBasic performs basic validation on ChangeStatusProposal fields.
func (csp ChangeStatusProposal) ValidateBasic() error {
	return govtypesv1beta1.ValidateAbstract(&csp)
}

// ProposalRoute returns the governance proposal route for WithdrawEscrowProposal.
func (wep WithdrawEscrowProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the WithdrawEscrowProposal.
func (wep WithdrawEscrowProposal) ProposalType() string { return ProposalTypeWithdrawEscrow }

// ValidateBasic performs basic validation on WithdrawEscrowProposal fields.
func (wep WithdrawEscrowProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

// ProposalRoute returns the governance proposal route for SetDenomMetadataProposal.
func (sdmdp SetDenomMetadataProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the SetDenomMetadataProposal.
func (sdmdp SetDenomMetadataProposal) ProposalType() string { return ProposalTypeSetDenomMetadata }

// ValidateBasic performs basic validation on SetDenomMetadataProposal fields.
func (sdmdp SetDenomMetadataProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}
