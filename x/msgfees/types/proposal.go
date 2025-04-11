package types

import (
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	PropTypeAddMsgFee                string = "AddMsgFee"
	PropTypeUpdateMsgFee             string = "UpdateMsgFee"
	PropTypeRemoveMsgFee             string = "RemoveMsgFee"
	PropTypeUpdateUsdConversionRate  string = "UpdateUsdConversionRate"
	PropTypeUpdateConversionFeeDenom string = "UpdateConversionFeeDenom"
)

var routerKey = ModuleName

var (
	_ govtypesv1beta1.Content = &AddMsgFeeProposal{}
	_ govtypesv1beta1.Content = &UpdateMsgFeeProposal{}
	_ govtypesv1beta1.Content = &RemoveMsgFeeProposal{}
	_ govtypesv1beta1.Content = &UpdateNhashPerUsdMilProposal{}
	_ govtypesv1beta1.Content = &UpdateConversionFeeDenomProposal{}
)

func (p AddMsgFeeProposal) ProposalRoute() string { return routerKey }
func (p AddMsgFeeProposal) ProposalType() string  { return "AddMsgFee" }
func (p AddMsgFeeProposal) ValidateBasic() error  { return errDep }

func (p UpdateMsgFeeProposal) ProposalRoute() string { return routerKey }
func (p UpdateMsgFeeProposal) ProposalType() string  { return "UpdateMsgFee" }
func (p UpdateMsgFeeProposal) ValidateBasic() error  { return errDep }

func (p RemoveMsgFeeProposal) ProposalRoute() string { return routerKey }
func (p RemoveMsgFeeProposal) ProposalType() string  { return "RemoveMsgFee" }
func (p RemoveMsgFeeProposal) ValidateBasic() error  { return errDep }

func (p UpdateNhashPerUsdMilProposal) ProposalRoute() string { return routerKey }
func (p UpdateNhashPerUsdMilProposal) ProposalType() string  { return "UpdateNhashPerUsdMil" }
func (p UpdateNhashPerUsdMilProposal) ValidateBasic() error  { return errDep }

func (p UpdateConversionFeeDenomProposal) ProposalRoute() string { return routerKey }
func (p UpdateConversionFeeDenomProposal) ProposalType() string  { return "UpdateConversionFeeDenom" }
func (p UpdateConversionFeeDenomProposal) ValidateBasic() error  { return errDep }
