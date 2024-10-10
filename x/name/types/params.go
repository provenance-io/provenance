package types

const (
	DefaultMinSegmentLength       = uint32(2)
	DefaultMaxSegmentLength       = uint32(32)
	DefaultMaxNameLevels          = uint32(16)
	DefaultAllowUnrestrictedNames = true
)

// NewParams creates a new parameter object
func NewParams(
	maxSegmentLength uint32,
	minSegmentLength uint32,
	maxNameLevels uint32,
	allowUnrestrictedNames bool,
) Params {
	return Params{
		MaxSegmentLength:       maxSegmentLength,
		MinSegmentLength:       minSegmentLength,
		MaxNameLevels:          maxNameLevels,
		AllowUnrestrictedNames: allowUnrestrictedNames,
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(
		DefaultMaxSegmentLength,
		DefaultMinSegmentLength,
		DefaultMaxNameLevels,
		DefaultAllowUnrestrictedNames,
	)
}

// Equal returns true if the given value is equivalent to the current instance of params
func (p *Params) Equal(that interface{}) bool {
	if that == nil {
		return p == nil
	}

	that1, ok := that.(*Params)
	if !ok {
		that2, ok := that.(Params)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return p == nil
	} else if p == nil {
		return false
	}
	if p.AllowUnrestrictedNames != that1.AllowUnrestrictedNames {
		return false
	}
	if p.MaxNameLevels != that1.MaxNameLevels {
		return false
	}
	if p.MaxSegmentLength != that1.MaxSegmentLength {
		return false
	}
	if p.MinSegmentLength != that1.MinSegmentLength {
		return false
	}

	return true
}
