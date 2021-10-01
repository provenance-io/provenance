package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	// ProposalTypeAddMsgBasedFees to add a new msg based fees
	ProposalTypeAddMsgBasedFees string = "AddMsgBasedFees"
	// ProposalTypeUpdateMsgBasedFees to update an existing msg based fees
	ProposalTypeUpdateMsgBasedFees string = "UpdateMsgBasedFees"
	// ProposalTypeRemoveMsgBasedFees to remove an existing msg based fees
	ProposalTypeRemoveMsgBasedFees string = "RemoveMsgBasedFees"
)

var (
	_ govtypes.Content = &AddMsgBasedFeesProposal{}
	_ govtypes.Content = &UpdateMsgBasedFeesProposal{}
	_ govtypes.Content = &RemoveMsgBasedFeesProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddMsgBasedFees)
	govtypes.RegisterProposalTypeCodec(AddMsgBasedFeesProposal{}, "provenance/msgfees/AddMsgBasedFeesProposal")

	govtypes.RegisterProposalType(ProposalTypeUpdateMsgBasedFees)
	govtypes.RegisterProposalTypeCodec(UpdateMsgBasedFeesProposal{}, "provenance/msgfees/UpdateMsgBasedFeesProposal")

	govtypes.RegisterProposalType(ProposalTypeRemoveMsgBasedFees)
	govtypes.RegisterProposalTypeCodec(RemoveMsgBasedFeesProposal{}, "provenance/msgfees/RemoveMsgBasedFeesProposal")
}

func (ambfp AddMsgBasedFeesProposal) ProposalRoute() string { return RouterKey }
func (ambfp AddMsgBasedFeesProposal) ProposalType() string  { return ProposalTypeAddMsgBasedFees }
func (ambfp AddMsgBasedFeesProposal) ValidateBasic() error  { return nil }
func (ambfp AddMsgBasedFeesProposal) String() string        { return "" }

func (umbfp UpdateMsgBasedFeesProposal) ProposalRoute() string { return RouterKey }
func (umbfp UpdateMsgBasedFeesProposal) ProposalType() string  { return ProposalTypeUpdateMsgBasedFees }
func (umbfp UpdateMsgBasedFeesProposal) ValidateBasic() error  { return nil }
func (umbfp UpdateMsgBasedFeesProposal) String() string        { return "" }

func (rmbfp RemoveMsgBasedFeesProposal) ProposalRoute() string { return RouterKey }
func (rmbfp RemoveMsgBasedFeesProposal) ProposalType() string  { return ProposalTypeRemoveMsgBasedFees }
func (rmbfp RemoveMsgBasedFeesProposal) ValidateBasic() error  { return nil }
func (rmbfp RemoveMsgBasedFeesProposal) String() string        { return "" }
