package types

import (
	"errors"
	"fmt"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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
	// ProposalTypeUpdateDenomMetadata to update denom metadata
	ProposalTypeUpdateDenomMetadata string = "UpdateDenomMetadata"
)

var (
	_ govtypes.Content = &AddMsgFeeProposal{}
	_ govtypes.Content = &UpdateMsgFeeProposal{}
	_ govtypes.Content = &RemoveMsgFeeProposal{}
	_ govtypes.Content = &UpdateNhashPerUsdMilProposal{}
	_ govtypes.Content = &UpdateDenomMetadataProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeAddMsgFee)
	govtypes.RegisterProposalTypeCodec(AddMsgFeeProposal{}, "provenance/msgfees/AddMsgFeeProposal")

	govtypes.RegisterProposalType(ProposalTypeUpdateMsgFee)
	govtypes.RegisterProposalTypeCodec(UpdateMsgFeeProposal{}, "provenance/msgfees/UpdateMsgFeeProposal")

	govtypes.RegisterProposalType(ProposalTypeRemoveMsgFee)
	govtypes.RegisterProposalTypeCodec(RemoveMsgFeeProposal{}, "provenance/msgfees/RemoveMsgFeeProposal")
	govtypes.RegisterProposalType(ProposalTypeUpdateUsdConversionRate)
	govtypes.RegisterProposalTypeCodec(UpdateNhashPerUsdMilProposal{}, "provenance/msgfees/UpdateNhashPerUsdMilProposal")

	govtypes.RegisterProposalType(ProposalTypeUpdateDenomMetadata)
	govtypes.RegisterProposalTypeCodec(UpdateDenomMetadataProposal{}, "provenance/msgfees/UpdateDenomMetadataProposal")
}

func NewAddMsgFeeProposal(
	title string,
	description string,
	msg string,
	additionalFee sdk.Coin,
	recipient string,
	recipientBasisPoints string,
) *AddMsgFeeProposal {
	return &AddMsgFeeProposal{
		Title:                title,
		Description:          description,
		MsgTypeUrl:           msg,
		AdditionalFee:        additionalFee,
		Recipient:            recipient,
		RecipientBasisPoints: recipientBasisPoints,
	}
}

func (p AddMsgFeeProposal) ProposalRoute() string { return RouterKey }
func (p AddMsgFeeProposal) ProposalType() string  { return ProposalTypeAddMsgFee }
func (p AddMsgFeeProposal) ValidateBasic() error {
	if len(p.MsgTypeUrl) == 0 {
		return ErrEmptyMsgType
	}

	if !p.AdditionalFee.IsPositive() {
		return ErrInvalidFee
	}

	if err := p.AdditionalFee.Validate(); err != nil {
		return err
	}

	if len(p.Recipient) != 0 {
		_, err := sdk.AccAddressFromBech32(p.Recipient)
		if err != nil {
			return err
		}
	}

	if len(p.RecipientBasisPoints) > 0 && len(p.Recipient) > 0 {
		bips, err := strconv.ParseUint(p.RecipientBasisPoints, 0, 64)
		if err != nil {
			return err
		}
		if bips > 10_000 {
			return fmt.Errorf("recipient basis points can only be between 0 and 10,000 : %v", p.RecipientBasisPoints)
		}
	} else if len(p.RecipientBasisPoints) > 0 && len(p.Recipient) == 0 {
		return fmt.Errorf("")
	}

	return govtypes.ValidateAbstract(&p)
}

func NewUpdateMsgFeeProposal(
	title string,
	description string,
	msg string,
	additionalFee sdk.Coin,
	recipient string,
	recipientBasisPoints string,
) *UpdateMsgFeeProposal {
	return &UpdateMsgFeeProposal{
		Title:                title,
		Description:          description,
		MsgTypeUrl:           msg,
		AdditionalFee:        additionalFee,
		Recipient:            recipient,
		RecipientBasisPoints: recipientBasisPoints,
	}
}

func (p UpdateMsgFeeProposal) ProposalRoute() string { return RouterKey }

func (p UpdateMsgFeeProposal) ProposalType() string { return ProposalTypeUpdateMsgFee }

func (p UpdateMsgFeeProposal) ValidateBasic() error {
	if len(p.MsgTypeUrl) == 0 {
		return ErrEmptyMsgType
	}

	if !p.AdditionalFee.IsPositive() {
		return ErrInvalidFee
	}

	if len(p.Recipient) != 0 {
		_, err := sdk.AccAddressFromBech32(p.Recipient)
		if err != nil {
			return err
		}
	}
	if len(p.RecipientBasisPoints) > 0 {
		bips, err := strconv.ParseUint(p.RecipientBasisPoints, 0, 64)
		if err != nil {
			return err
		}
		if bips > 10_000 {
			return fmt.Errorf("recipient basis points can only be between 0 and 10,000 : %v", p.RecipientBasisPoints)
		}
	}

	return govtypes.ValidateAbstract(&p)
}

func NewRemoveMsgFeeProposal(
	title string,
	description string,
	msgTypeURL string,
) *RemoveMsgFeeProposal {
	return &RemoveMsgFeeProposal{
		Title:       title,
		Description: description,
		MsgTypeUrl:  msgTypeURL,
	}
}

func (p RemoveMsgFeeProposal) ProposalRoute() string { return RouterKey }

func (p RemoveMsgFeeProposal) ProposalType() string { return ProposalTypeRemoveMsgFee }

func (p RemoveMsgFeeProposal) ValidateBasic() error {
	if len(p.MsgTypeUrl) == 0 {
		return ErrEmptyMsgType
	}
	return govtypes.ValidateAbstract(&p)
}

func NewUpdateNhashPerUsdMilProposal(
	title string,
	description string,
	nhashPerUsdMil uint64,
) *UpdateNhashPerUsdMilProposal {
	return &UpdateNhashPerUsdMilProposal{
		Title:          title,
		Description:    description,
		NhashPerUsdMil: nhashPerUsdMil,
	}
}

func (p UpdateNhashPerUsdMilProposal) ProposalRoute() string { return RouterKey }

func (p UpdateNhashPerUsdMilProposal) ProposalType() string {
	return ProposalTypeUpdateUsdConversionRate
}

func (p UpdateNhashPerUsdMilProposal) ValidateBasic() error {
	if p.NhashPerUsdMil < 1 {
		return errors.New("nhash per usd mil must be greater than 0")
	}
	return govtypes.ValidateAbstract(&p)
}

func NewUpdateDenomMetadataProposal(
	title string,
	description string,
	metadata banktypes.Metadata,
) *UpdateDenomMetadataProposal {
	return &UpdateDenomMetadataProposal{
		Title: title,
		Description: description,
		Metadata: metadata,
	}
}

func (p UpdateDenomMetadataProposal) ProposalRoute() string { return RouterKey }

func (p UpdateDenomMetadataProposal) ProposalType() string {
	return ProposalTypeUpdateDenomMetadata
}

func (p UpdateDenomMetadataProposal) ValidateBasic() error {
	return govtypes.ValidateAbstract(&p)
}