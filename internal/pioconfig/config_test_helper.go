package pioconfig

// ChangeMsgFeeFloorDenom changes FloorGasPrice, even though practically the denom for Floor checking should be the same
// as fee denom, however it seems tests like to play around with it being different, so for now keeping it, in a test helper
// for the sole reason for the awareness of someone using it of it's current purpose.
func ChangeMsgFeeFloorDenom(amount int64, denom string) {
	provConfig.MsgFloorDenom = denom
	provConfig.MsgFeeFloorGasPrice = amount
}
