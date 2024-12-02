package nav

const (
	// ModuleName is the name of the NAV module.
	ModuleName = "nav"

	// StoreKey is the store key string for the NAV module.
	StoreKey = ModuleName

	// NotYetImplemented is a string to give a panic for stuff that's not yet implemented.
	NotYetImplemented = "not yet implemented" // TODO: Delete this

	// SourceMaxLen is the maximum number of bytes the NetAssetValueRecord.Source field can have.
	SourceMaxLen = 100

	StorePrefixNAVs = byte(0x01)
)
