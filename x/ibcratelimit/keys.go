package ibcratelimit

const (
	// ModuleName defines the module name
	ModuleName = "ratelimitedibc"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

var (
	// ParamsKey is the key to obtain the module's params.
	ParamsKey = []byte{0x01}
)
