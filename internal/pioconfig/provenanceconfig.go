package pioconfig

const (
	// DefaultBondDenom is the denomination of coin to use for bond/staking
	DefaultBondDenom = "nhash" // nano-hash
	// DefaultFeeDenom is the denomination of coin to use for fees
	DefaultFeeDenom = "nhash" // nano-hash
	// DefaultMinGasPrices is the minimum gas prices coin value.
	DefaultMinGasPrices = "1905" + DefaultFeeDenom
	// DefaultReDnmString is the allowed denom regex expression
	DefaultReDnmString = `[a-zA-Z][a-zA-Z0-9/\-\.]{2,127}`
)
