package piobaseapp

import (
	msgbasedfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type PioBaseAppKeeperOptions struct {
	AccountKeeper     msgbasedfeetypes.AccountKeeper
	BankKeeper        msgbasedfeetypes.BankKeeper
	FeegrantKeeper    msgbasedfeetypes.FeegrantKeeper
	MsgBasedFeeKeeper msgbasedfeetypes.MsgBasedFeeKeeper
}
