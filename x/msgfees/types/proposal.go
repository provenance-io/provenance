package types

import (
	"errors"
	"fmt"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	// ProposalTypeUpdateDenomMetadata to update denom metadata
	ProposalTypeUpdateDenomMetadata string = "UpdateDenomMetadata"
)

var (
	_ govtypesv1beta1.Content = &AddMsgFeeProposal{}
	_ govtypesv1beta1.Content = &UpdateMsgFeeProposal{}
	_ govtypesv1beta1.Content = &RemoveMsgFeeProposal{}
	_ govtypesv1beta1.Content = &UpdateNhashPerUsdMilProposal{}
	_ govtypesv1beta1.Content = &UpdateDenomMetadataProposal{}
)

func init() {
	govtypesv1beta1.RegisterProposalType(ProposalTypeAddMsgFee)
	govtypesv1beta1.RegisterProposalType(ProposalTypeUpdateMsgFee)
	govtypesv1beta1.RegisterProposalType(ProposalTypeRemoveMsgFee)
	govtypesv1beta1.RegisterProposalType(ProposalTypeUpdateUsdConversionRate)
	govtypesv1beta1.RegisterProposalType(ProposalTypeUpdateDenomMetadata)
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

	return govtypesv1beta1.ValidateAbstract(&p)
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

	return govtypesv1beta1.ValidateAbstract(&p)
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
	return govtypesv1beta1.ValidateAbstract(&p)
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
	return govtypesv1beta1.ValidateAbstract(&p)
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
	return govtypesv1beta1.ValidateAbstract(&p)
}