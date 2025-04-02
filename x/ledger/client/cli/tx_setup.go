package cli

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/provenance-io/provenance/x/ledger"
	"github.com/spf13/pflag"
)

func MakeMsgCreateAppend(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*ledger.MsgAppendRequest, error) {
	msg := &ledger.MsgAppendRequest{}

	errs := make([]error, 8)
	// msg.AskOrder.Seller, errs[0] = ReadAddrFlagOrFrom(clientCtx, flagSet, FlagSeller)
	// msg.AskOrder.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	// msg.AskOrder.Assets, errs[2] = ReadReqCoinFlag(flagSet, FlagAssets)
	// msg.AskOrder.Price, errs[3] = ReadReqCoinFlag(flagSet, FlagPrice)
	// msg.AskOrder.SellerSettlementFlatFee, errs[4] = ReadCoinFlag(flagSet, FlagSettlementFee)
	// msg.AskOrder.AllowPartial, errs[5] = flagSet.GetBool(FlagPartial)
	// msg.AskOrder.ExternalId, errs[6] = flagSet.GetString(FlagExternalID)
	// msg.OrderCreationFee, errs[7] = ReadCoinFlag(flagSet, FlagCreationFee)

	return msg, errors.Join(errs...)
}
