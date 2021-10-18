package antewrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/msgfees/types"
)

// MsgBasedFeeKeeper for additional msg fees.
type MsgBasedFeeKeeper interface {
  GetMsgBasedFee(ctx sdk.Context, msgType string) (*types.MsgBasedFee, error)
  GetFeeCollectorName() string
}

