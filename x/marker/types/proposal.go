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

func (p AddMarkerProposal) ProposalRoute() string { return RouterKey }
func (p AddMarkerProposal) ProposalType() string  { return "AddMarker" }
func (p AddMarkerProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

func (sip SupplyIncreaseProposal) ProposalRoute() string { return RouterKey }
func (sip SupplyIncreaseProposal) ProposalType() string  { return ProposalTypeIncreaseSupply }
func (sip SupplyIncreaseProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

func (sdp SupplyDecreaseProposal) ProposalRoute() string { return RouterKey }
func (sdp SupplyDecreaseProposal) ProposalType() string  { return ProposalTypeDecreaseSupply }
func (sdp SupplyDecreaseProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

func (sap SetAdministratorProposal) ProposalRoute() string { return RouterKey }
func (sap SetAdministratorProposal) ProposalType() string  { return ProposalTypeSetAdministrator }
func (sap SetAdministratorProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

func (rap RemoveAdministratorProposal) ProposalRoute() string { return RouterKey }
func (rap RemoveAdministratorProposal) ProposalType() string  { return ProposalTypeRemoveAdministrator }
func (rap RemoveAdministratorProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

func (csp ChangeStatusProposal) ProposalRoute() string { return RouterKey }
func (csp ChangeStatusProposal) ProposalType() string  { return ProposalTypeChangeStatus }
func (csp ChangeStatusProposal) ValidateBasic() error {
	return govtypesv1beta1.ValidateAbstract(&csp)
}

func (wep WithdrawEscrowProposal) ProposalRoute() string { return RouterKey }
func (wep WithdrawEscrowProposal) ProposalType() string  { return ProposalTypeWithdrawEscrow }
func (wep WithdrawEscrowProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

func (sdmdp SetDenomMetadataProposal) ProposalRoute() string { return RouterKey }
func (sdmdp SetDenomMetadataProposal) ProposalType() string  { return ProposalTypeSetDenomMetadata }
func (sdmdp SetDenomMetadataProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}
