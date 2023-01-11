package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewEventExpirationAdd(moduleAssetID string) *EventExpirationAdd {
	return &EventExpirationAdd{moduleAssetID}
}

func NewEventExpirationDeposit(moduleAssetID string, depositor string, deposit sdk.Coins) *EventExpirationDeposit {
	return &EventExpirationDeposit{
		ModuleAssetId: moduleAssetID,
		Depositor:     depositor,
		Deposit:       deposit,
	}
}

func NewEventExpirationExtend(moduleAssetID string) *EventExpirationExtend {
	return &EventExpirationExtend{moduleAssetID}
}

func NewEventExpirationInvoke(moduleAssetID string) *EventExpirationInvoke {
	return &EventExpirationInvoke{moduleAssetID}
}

func NewEventExpirationRemove(moduleAssetID string) *EventExpirationRemove {
	return &EventExpirationRemove{moduleAssetID}
}
