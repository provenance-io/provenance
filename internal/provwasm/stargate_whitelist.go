package provwasm

import (
	"fmt"
	"sync"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/hold"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

// stargateWhitelist keeps whitelist and its deterministic
// response binding for stargate queries.
//
// The query can be multi-thread, so we have to use
// thread safe sync.Map.
var stargateWhitelist sync.Map

// Note: When adding a migration here, we should also add it to the Async ICQ params in the upgrade.
// In the future we may want to find a better way to keep these in sync

func init() {
	// ibc queries
	setWhitelistedQuery("/ibc.applications.transfer.v1.Query/DenomTrace", &ibctransfertypes.QueryDenomTraceResponse{})

	// ==========================================================
	// cosmos-sdk queries
	// ==========================================================

	// auth
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/Account", &authtypes.QueryAccountResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/Params", &authtypes.QueryParamsResponse{})

	// bank
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/Balance", &banktypes.QueryBalanceResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/DenomMetadata", &banktypes.QueryDenomsMetadataResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/Params", &banktypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/SupplyOf", &banktypes.QuerySupplyOfResponse{})

	// distribution
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/Params", &distributiontypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress", &distributiontypes.QueryDelegatorWithdrawAddressResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/ValidatorCommission", &distributiontypes.QueryValidatorCommissionResponse{})

	// gov
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Deposit", &govtypes.QueryDepositResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Params", &govtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Vote", &govtypes.QueryVoteResponse{})

	// slashing
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/Params", &slashingtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/SigningInfo", &slashingtypes.QuerySigningInfoResponse{})

	// staking
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Delegation", &stakingtypes.QueryDelegationResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Params", &stakingtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Validator", &stakingtypes.QueryValidatorResponse{})

	// ==========================================================
	// provenance queries
	// ==========================================================

	// attribute
	setWhitelistedQuery("/provenance.attribute.v1.Query/Params", &attributetypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.attribute.v1.Query/Attribute", &attributetypes.QueryAttributeResponse{})
	setWhitelistedQuery("/provenance.attribute.v1.Query/Attributes", &attributetypes.QueryAttributesResponse{})
	setWhitelistedQuery("/provenance.attribute.v1.Query/Scan", &attributetypes.QueryScanResponse{})

	// exchange
	setWhitelistedQuery("/provenance.exchange.v1.Query/OrderFeeCalc", &exchange.QueryOrderFeeCalcResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetOrder", &exchange.QueryGetOrderResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetOrderByExternalID", &exchange.QueryGetOrderByExternalIDResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetMarketOrders", &exchange.QueryGetMarketOrdersResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetOwnerOrders", &exchange.QueryGetOwnerOrdersResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetAssetOrders", &exchange.QueryGetAssetOrdersResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetAllOrders", &exchange.QueryGetAllOrdersResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetMarket", &exchange.QueryGetMarketResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetAllMarkets", &exchange.QueryGetAllMarketsResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/Params", &exchange.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/ValidateCreateMarket", &exchange.QueryValidateCreateMarketResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/ValidateMarket", &exchange.QueryValidateMarketResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/ValidateManageFees", &exchange.QueryValidateManageFeesResponse{})

	// hold
	setWhitelistedQuery("/provenance.hold.v1.Query/GetHolds", &hold.GetHoldsResponse{})
	setWhitelistedQuery("/provenance.hold.v1.Query/GetAllHolds", &hold.GetAllHoldsResponse{})

	// marker
	setWhitelistedQuery("/provenance.marker.v1.Query/Params", &markertypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/Marker", &markertypes.QueryMarkerResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/Holding", &markertypes.QueryHoldingResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/Supply", &markertypes.QuerySupplyResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/Escrow", &markertypes.QueryEscrowResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/Access", &markertypes.QueryAccessResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/DenomMetadata", &markertypes.QueryDenomMetadataResponse{})

	// metadata
	setWhitelistedQuery("/provenance.metadata.v1.Query/Params", &metadatatypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/Scope", &metadatatypes.ScopeResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/Sessions", &metadatatypes.SessionsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/Records", &metadatatypes.RecordsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/Ownership", &metadatatypes.OwnershipResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/ValueOwnership", &metadatatypes.ValueOwnershipResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/ScopeSpecification", &metadatatypes.ScopeSpecificationResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/ContractSpecification", &metadatatypes.ContractSpecificationResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/RecordSpecificationsForContractSpecification", &metadatatypes.RecordSpecificationsForContractSpecificationResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/RecordSpecification", &metadatatypes.RecordSpecificationResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/OSLocatorParams", &metadatatypes.OSLocatorParamsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/OSLocator", &metadatatypes.OSLocatorResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/OSLocatorsByURI", &metadatatypes.OSLocatorsByURIResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/OSLocatorsByScope", &metadatatypes.OSLocatorsByScopeResponse{})

	// msg fee
	setWhitelistedQuery("/provenance.msgfees.v1.Query/Params", &msgfeestypes.QueryParamsResponse{})

	// name
	setWhitelistedQuery("/provenance.name.v1.Query/Params", &nametypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.name.v1.Query/Resolve", &nametypes.QueryResolveResponse{})
	setWhitelistedQuery("/provenance.name.v1.Query/ReverseLookup", &nametypes.QueryReverseLookupResponse{})

	// trigger
	setWhitelistedQuery("/provenance.trigger.v1.Query/TriggerByID", &triggertypes.QueryTriggerByIDResponse{})
	setWhitelistedQuery("/provenance.trigger.v1.Query/Triggers", &triggertypes.QueryTriggersResponse{})
}

// GetWhitelistedQuery returns the whitelisted query at the provided path.
// If the query does not exist, or it was setup wrong by the chain, this returns an error.
func GetWhitelistedQuery(queryPath string) (codec.ProtoMarshaler, error) {
	protoResponseAny, isWhitelisted := stargateWhitelist.Load(queryPath)
	if !isWhitelisted {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", queryPath)}
	}
	protoResponseType, ok := protoResponseAny.(codec.ProtoMarshaler)
	if !ok {
		return nil, wasmvmtypes.Unknown{}
	}
	return protoResponseType, nil
}

func setWhitelistedQuery(queryPath string, protoType codec.ProtoMarshaler) {
	stargateWhitelist.Store(queryPath, protoType)
}

func GetStargateWhitelistedPaths() (keys []string) {
	// Iterate over the map and collect the keys
	stargateWhitelist.Range(func(key, _ interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			panic("key is not a string")
		}
		keys = append(keys, keyStr)
		return true
	})

	return keys
}
