package types

const (
	// EventTypeAssetClassCreated emitted when an asset class is created
	EventTypeAssetClassCreated string = "asset_class_created"
	// EventTypeAssetCreated emitted when an asset is created
	EventTypeAssetCreated string = "asset_created"
	// EventTypePoolCreated emitted when a pool is created
	EventTypePoolCreated string = "pool_created"
	// EventTypeTokenizationCreated emitted when a tokenization marker is created
	EventTypeTokenizationCreated string = "tokenization_created" //nolint:gosec // This is an event type, not credentials
	// EventTypeSecuritizationCreated emitted when a securitization is created
	EventTypeSecuritizationCreated string = "securitization_created"

	// Attribute keys
	AttributeKeyAssetClassID      string = "asset_class_id"
	AttributeKeyAssetCount        string = "asset_count"
	AttributeKeyAssetID           string = "asset_id"
	AttributeKeyAssetName         string = "asset_name"
	AttributeKeyAssetSymbol       string = "asset_symbol"
	AttributeKeyOwner             string = "owner"
	AttributeKeyPool              string = "pool"
	AttributeKeyTokenization      string = "tokenization"
	AttributeKeySecuritizationID  string = "securitization_id"
	AttributeKeyTrancheCount      string = "tranche_count"
	AttributeKeyPoolCount         string = "pool_count"
)
