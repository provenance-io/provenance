package types

const (
	QueryMarkers      = "all" // all instead of markers to prevent uri stuttering  in '/custom/marker/all'
	QueryMarker       = "detail"
	QueryHolders      = "holders"
	QueryMarkerSupply = "supply"
	QueryMarkerEscrow = "escrow"
	QueryMarkerAccess = "accesscontrol"
	QueryMarkerAssets = "assets"
)

// QueryMarkersParams defines the params for the following legacy queries:
// - 'custom/marker/all'
type QueryMarkersParams struct {
	Page, Limit   int
	Denom, Status string
}

// NewQueryMarkersParams object
func NewQueryMarkersParams(page, limit int, denom, status string) QueryMarkersParams {
	return QueryMarkersParams{page, limit, denom, status}
}
