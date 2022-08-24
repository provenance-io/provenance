package types

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewExpiration(
	moduleAssetID string,
	owner string,
	blockHeight int64,
	deposit sdk.Coin,
	messages []*types.Any,
) *Expiration {
	return &Expiration{
		ModuleAssetId: moduleAssetID,
		Owner:         owner,
		BlockHeight:   blockHeight,
		Deposit:       deposit,
		Messages:      messages,
	}
}

// ValidateBasic basic format checking of the data
func (e Expiration) ValidateBasic() error {
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
	return nil
}
