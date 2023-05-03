package types

import types "github.com/cosmos/cosmos-sdk/codec/types"

type TriggerID = uint64

func NewTrigger(id TriggerID, event Event, action *types.Any) Trigger {
	return Trigger{
		id,
		event,
		action,
	}
}
