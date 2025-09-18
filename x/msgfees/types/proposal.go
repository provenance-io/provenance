package types

import (
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var routerKey = ModuleName

var (
	_ govtypesv1beta1.Content = (*AddMsgFeeProposal)(nil)
	_ govtypesv1beta1.Content = (*UpdateMsgFeeProposal)(nil)
	_ govtypesv1beta1.Content = (*RemoveMsgFeeProposal)(nil)
	_ govtypesv1beta1.Content = (*UpdateNhashPerUsdMilProposal)(nil)
	_ govtypesv1beta1.Content = (*UpdateConversionFeeDenomProposal)(nil)
)

// ProposalRoute returns the routing key for AddMsgFeeProposal.
func (p AddMsgFeeProposal) ProposalRoute() string { return routerKey }

// ProposalType returns the type of the AddMsgFeeProposal.
func (p AddMsgFeeProposal) ProposalType() string { return "AddMsgFee" }

// ValidateBasic performs basic validation of AddMsgFeeProposal fields.
func (p AddMsgFeeProposal) ValidateBasic() error { return errDep }

// ProposalRoute returns the routing key for UpdateMsgFeeProposal.
func (p UpdateMsgFeeProposal) ProposalRoute() string { return routerKey }

// ProposalType returns the type of the UpdateMsgFeeProposal.
func (p UpdateMsgFeeProposal) ProposalType() string { return "UpdateMsgFee" }

// ValidateBasic performs basic validation of UpdateMsgFeeProposal fields.
func (p UpdateMsgFeeProposal) ValidateBasic() error { return errDep }

// ProposalRoute returns the routing key for RemoveMsgFeeProposal.
func (p RemoveMsgFeeProposal) ProposalRoute() string { return routerKey }

// ProposalType returns the type of the RemoveMsgFeeProposal.
func (p RemoveMsgFeeProposal) ProposalType() string { return "RemoveMsgFee" }

// ValidateBasic performs basic validation of RemoveMsgFeeProposal fields.
func (p RemoveMsgFeeProposal) ValidateBasic() error { return errDep }

// ProposalRoute returns the routing key for UpdateNhashPerUsdMilProposal.
func (p UpdateNhashPerUsdMilProposal) ProposalRoute() string { return routerKey }

// ProposalType returns the type of the UpdateNhashPerUsdMilProposal.
func (p UpdateNhashPerUsdMilProposal) ProposalType() string { return "UpdateNhashPerUsdMil" }

// ValidateBasic performs basic validation of UpdateNhashPerUsdMilProposal fields
func (p UpdateNhashPerUsdMilProposal) ValidateBasic() error { return errDep }

// ProposalRoute returns the routing key for UpdateConversionFeeDenomProposal.
func (p UpdateConversionFeeDenomProposal) ProposalRoute() string { return routerKey }

// ProposalType returns the type of the UpdateConversionFeeDenomProposal.
func (p UpdateConversionFeeDenomProposal) ProposalType() string { return "UpdateConversionFeeDenom" }

// ValidateBasic performs basic validation of UpdateConversionFeeDenomProposal fields.
func (p UpdateConversionFeeDenomProposal) ValidateBasic() error { return errDep }
