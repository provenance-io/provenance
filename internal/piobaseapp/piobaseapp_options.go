package piobaseapp

import (
	msgbasedfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type pioBaseAppKeeperOptions struct {
	AccountKeeper     msgbasedfeetypes.AccountKeeper
	BankKeeper        msgbasedfeetypes.BankKeeper
	FeegrantKeeper    msgbasedfeetypes.FeegrantKeeper
	MsgBasedFeeKeeper msgbasedfeetypes.MsgBasedFeeKeeper
}
