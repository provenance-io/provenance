package keeper

var (
	ledgerPrefix                      = []byte{0x01}
	entriesPrefix                     = []byte{0x02}
	ledgerClassesPrefix               = []byte{0x03}
	ledgerClassEntryTypesPrefix       = []byte{0x04}
	ledgerClassStatusTypesPrefix      = []byte{0x05}
	ledgerClassBucketTypesPrefix      = []byte{0x06}
	fundTransfersPrefix               = []byte{0x07} // Reserved for future use...
	fundTransfersWithSettlementPrefix = []byte{0x08}
)
