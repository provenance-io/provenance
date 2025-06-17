package types

const (
	// EventTypeAssetClassCreated emitted when an asset class is created
	EventTypeAssetClassCreated string = "asset_class_created"
	// EventTypeAssetCreated emitted when an asset is created
	EventTypeAssetCreated string = "asset_created"
	// EventTypePoolCreated emitted when a pool is created
	EventTypePoolCreated string = "pool_created"
	// EventTypeParticipationCreated emitted when a participation marker is created
	EventTypeParticipationCreated string = "participation_created"
	// EventTypeSecuritizationCreated emitted when a securitization is created
	EventTypeSecuritizationCreated string = "securitization_created"

	// Attribute keys
	AttributeKeyAssetClassId   string = "asset_class_id"
	AttributeKeyAssetId        string = "asset_id"
	AttributeKeyAssetName      string = "asset_name"
	AttributeKeyAssetSymbol    string = "asset_symbol"
	AttributeKeyLedgerClass    string = "ledger_class"
	AttributeKeyOwner          string = "owner"
	AttributeKeyPoolDenom      string = "pool_denom"
	AttributeKeyPoolAmount     string = "pool_amount"
	AttributeKeyNftCount       string = "nft_count"
	AttributeKeyParticipationDenom string = "participation_denom"
	AttributeKeySecuritizationId string = "securitization_id"
	AttributeKeyTrancheCount   string = "tranche_count"
	AttributeKeyPoolCount      string = "pool_count"
) 