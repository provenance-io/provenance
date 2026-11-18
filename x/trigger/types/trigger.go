package types

import (
	proto "github.com/cosmos/gogoproto/proto"
)

// TriggerEventI is the interface for the events that a trigger could wait for.
// It's only kept so that historical trigger messages can still be decoded.
type TriggerEventI interface {
	proto.Message
}

var (
	_ TriggerEventI = (*TransactionEvent)(nil)
	_ TriggerEventI = (*BlockHeightEvent)(nil)
	_ TriggerEventI = (*BlockTimeEvent)(nil)
)
