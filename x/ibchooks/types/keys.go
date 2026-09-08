package types

import "cosmossdk.io/collections"

const (
	ModuleName = "ibchooks"
	StoreKey   = "hooks-for-ibc" // not using the module name because of collisions with key "ibc"

	IBCCallbackKey = "ibc_callback"
	IBCAsyncAckKey = "ibc_async_ack"

	MsgEmitAckKey           = "emit_ack"
	AttributeSender         = "sender"
	AttributeChannel        = "channel"
	AttributePacketSequence = "sequence"

	SenderPrefix = "ibc-wasm-hook-intermediary"
)

var (

	// Params keeps 0x01 (byte-identical to the legacy layout).
	ParamsKeyBz         = []byte{0x01}
	PacketCallbackKeyBz = []byte{0x02}
	PacketAckKeyBz      = []byte{0x03}

	// IbcHooksParamStoreKey key for ibchooks module's params
	IbcHooksParamStoreKey = ParamsKeyBz

	// Collection prefixes.
	ParamsKeyPrefix         = collections.NewPrefix(ParamsKeyBz)
	PacketCallbackKeyPrefix = collections.NewPrefix(PacketCallbackKeyBz)
	PacketAckKeyPrefix      = collections.NewPrefix(PacketAckKeyBz)
)
