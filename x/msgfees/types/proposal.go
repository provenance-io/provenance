package types

import (
	"errors"

	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeAddMsgFee to add a new msg based fee
	ProposalTypeAddMsgFee string = "AddMsgFee"
	// ProposalTypeUpdateMsgFee to update an existing msg based fee
	ProposalTypeUpdateMsgFee string = "UpdateMsgFee"
	// ProposalTypeRemoveMsgFee to remove an existing msg based fee
	ProposalTypeRemoveMsgFee string = "RemoveMsgFee"
	// ProposalTypeUpdateUsdConversionRate to update the usd conversion rate param
	ProposalTypeUpdateUsdConversionRate string = "UpdateUsdConversionRate"
	// ProposalTypeUpdateConversionFeeDenom to update the conversion rate denom
	ProposalTypeUpdateConversionFeeDenom string = "UpdateConversionFeeDenom"
)

var (
	_ govtypesv1beta1.Content = &AddMsgFeeProposal{}
	_ govtypesv1beta1.Content = &UpdateMsgFeeProposal{}
	_ govtypesv1beta1.Content = &RemoveMsgFeeProposal{}
	_ govtypesv1beta1.Content = &UpdateNhashPerUsdMilProposal{}
	_ govtypesv1beta1.Content = &UpdateConversionFeeDenomProposal{}
)

func (p AddMsgFeeProposal) ProposalRoute() string { return RouterKey }
func (p AddMsgFeeProposal) ProposalType() string  { return ProposalTypeAddMsgFee }
func (p AddMsgFeeProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

func (p UpdateMsgFeeProposal) ProposalRoute() string { return RouterKey }

func (p UpdateMsgFeeProposal) ProposalType() string { return ProposalTypeUpdateMsgFee }

func (p UpdateMsgFeeProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

func (p RemoveMsgFeeProposal) ProposalRoute() string { return RouterKey }

func (p RemoveMsgFeeProposal) ProposalType() string { return ProposalTypeRemoveMsgFee }

func (p RemoveMsgFeeProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

func (p UpdateNhashPerUsdMilProposal) ProposalRoute() string { return RouterKey }

func (p UpdateNhashPerUsdMilProposal) ProposalType() string {
	return ProposalTypeUpdateUsdConversionRate
}

func (p UpdateNhashPerUsdMilProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}

func (p UpdateConversionFeeDenomProposal) ProposalRoute() string { return RouterKey }

func (p UpdateConversionFeeDenomProposal) ProposalType() string {
	return ProposalTypeUpdateConversionFeeDenom
}

func (p UpdateConversionFeeDenomProposal) ValidateBasic() error {
	return errors.New("deprecated and unusable")
}
