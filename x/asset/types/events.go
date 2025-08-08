package types

const (
	// EventTypeAssetClassCreated emitted when an asset class is created
	EventTypeAssetClassCreated string = "asset_class_created"
	// EventTypeAssetCreated emitted when an asset is created
	EventTypeAssetCreated string = "asset_created"
	// EventTypePoolCreated emitted when a pool is created
	EventTypePoolCreated string = "pool_created"
	// EventTypeTokenizationCreated emitted when a tokenization marker is created
	EventTypeTokenizationCreated string = "tokenization_created"
	// EventTypeSecuritizationCreated emitted when a securitization is created
	EventTypeSecuritizationCreated string = "securitization_created"

	// Attribute keys
	AttributeKeyAssetClassId      string = "asset_class_id"
	AttributeKeyAssetCount        string = "asset_count"
	AttributeKeyAssetId           string = "asset_id"
	AttributeKeyAssetName         string = "asset_name"
	AttributeKeyAssetSymbol       string = "asset_symbol"
	AttributeKeyOwner             string = "owner"
	AttributeKeyPoolDenom         string = "pool_denom"
	AttributeKeyPoolAmount        string = "pool_amount"
	AttributeKeyTokenizationDenom string = "tokenization_denom"
	AttributeKeySecuritizationId  string = "securitization_id"
	AttributeKeyTrancheCount      string = "tranche_count"
	AttributeKeyPoolCount         string = "pool_count"
)
