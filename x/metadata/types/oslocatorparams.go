package types

const (
	// DefaultMaxURILength defines the default maximum length for a URI.
	DefaultMaxURILength = 2048
)

// NewOSLocatorParams creates a new parameter object
func NewOSLocatorParams(maxURILength uint32) OSLocatorParams {
	return OSLocatorParams{MaxUriLength: maxURILength}
}

// DefaultOSLocatorParams defines the parameters for this module
func DefaultOSLocatorParams() OSLocatorParams {
	return NewOSLocatorParams(DefaultMaxURILength)
}
