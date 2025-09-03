package types

const (
	DefaultMaxValueLength = 10000 //nolint:revive
)

// NewParams create a new Params object
func NewParams(
	maxValueLength uint32,
) Params {
	return Params{
		MaxValueLength: maxValueLength,
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(
		DefaultMaxValueLength,
	)
}
