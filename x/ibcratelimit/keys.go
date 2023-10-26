package ibcratelimit

const (
	ModuleName = "ratelimitedibc" // IBC at the end to avoid conflicts with the ibc prefix
	StoreKey   = ModuleName
	RouterKey  = ModuleName
)

var (
	// ParamsKey is the key to obtain the module's params.
	ParamsKey = []byte{0x01}
)
