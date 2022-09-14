package types

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewExpiration(
	moduleAssetID string,
	owner string,
	blockHeight int64,
	deposit sdk.Coin,
	message types.Any,
) *Expiration {
	return &Expiration{
		ModuleAssetId: moduleAssetID,
		Owner:         owner,
		BlockHeight:   blockHeight,
		Deposit:       deposit,
		Message:       message,
	}
}

// ValidateBasic basic format checking of the data
func (e *Expiration) ValidateBasic() error {
	if strings.TrimSpace(e.ModuleAssetId) == "" {
		return ErrEmptyModuleAssetID
	}
	if strings.TrimSpace(e.Owner) == "" {
		return ErrEmptyOwnerAddress
	}
	if e.BlockHeight <= 0 {
		return ErrBlockHeightLteZero
	}
	if !e.Deposit.IsValid() {
		return ErrInvalidDeposit
	}
	if err := e.validateMessage(); err != nil {
		return sdkerrors.New(
			ErrInvalidMessage.Codespace(),
			ErrInvalidMessage.ABCICode(),
			err.Error())
	}
	return nil
}

// validateMessage validates `expiration.Message` conforms to `sdk.Msg`
// and is registered and whitelisted with the InterfaceRegistry
func (e *Expiration) validateMessage() error {
	// this may occur in message decoding
	if e.Message.Value == nil {
		return fmt.Errorf("expecting non nil value to validate an Any message")
	}
	if e.Message.TypeUrl == "" {
		return fmt.Errorf("expecting message type URL to unpack Any")
	}
	// validate message is a whitelisted sdk.Msg during a BroadcastTx
	var msg sdk.Msg
	if err := ModuleCdc.UnpackAny(&e.Message, &msg); err != nil {
		return err
	}
	if msg == nil {
		return fmt.Errorf("failed to unpack Any message: %v", msg)
	}
	return msg.ValidateBasic()
}
