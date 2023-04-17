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
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"
)

// stargateWhitelist keeps whitelist and its deterministic
// response binding for stargate queries.
//
// The query can be multi-thread, so we have to use
// thread safe sync.Map.
var stargateWhitelist sync.Map

// Note: When adding a migration here, we should also add it to the Async ICQ params in the upgrade.
// In the future we may want to find a better way to keep these in sync

//nolint:staticcheck
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
	setWhitelistedQuery("/provenance.attribute.v1.QueryParamsRequest", &attributetypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.attribute.v1.QueryAttributeRequest", &attributetypes.QueryAttributeResponse{})
	setWhitelistedQuery("/provenance.attribute.v1.QueryAttributesRequest", &attributetypes.QueryAttributesResponse{})
	setWhitelistedQuery("/provenance.attribute.v1.QueryScanRequest", &attributetypes.QueryScanResponse{})
	setWhitelistedQuery("/provenance.attribute.v1.QueryAttributeAccountsRequest", &attributetypes.QueryAttributeAccountsResponse{})

	// marker
	setWhitelistedQuery("/provenance.marker.v1.QueryParamsRequest", &markertypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.marker.v1.QueryAllMarkersRequest", &markertypes.QueryAllMarkersResponse{})
	setWhitelistedQuery("/provenance.marker.v1.QueryMarkerRequest", &markertypes.QueryMarkerResponse{})
	setWhitelistedQuery("/provenance.marker.v1.QueryHoldingRequest", &markertypes.QueryHoldingResponse{})
	setWhitelistedQuery("/provenance.marker.v1.QuerySupplyRequest", &markertypes.QuerySupplyResponse{})
	setWhitelistedQuery("/provenance.marker.v1.QueryEscrowRequest", &markertypes.QueryEscrowResponse{})
	setWhitelistedQuery("/provenance.marker.v1.QueryAccessRequest", &markertypes.QueryAccessResponse{})
	setWhitelistedQuery("/provenance.marker.v1.QueryDenomMetadataRequest", &markertypes.QueryDenomMetadataResponse{})

	// metadata
	setWhitelistedQuery("/provenance.metadata.v1.QueryParamsRequest", &metadatatypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.ScopeRequest", &metadatatypes.ScopeResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.ScopesAllRequest", &metadatatypes.ScopesAllResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.SessionsRequest", &metadatatypes.SessionsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.SessionsAllRequest", &metadatatypes.SessionsAllResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.RecordsRequest", &metadatatypes.RecordsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.RecordsAllRequest", &metadatatypes.RecordsAllResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.OwnershipRequest", &metadatatypes.OwnershipResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.ValueOwnershipRequest", &metadatatypes.ValueOwnershipResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.ScopeSpecificationRequest", &metadatatypes.ScopeSpecificationResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.ScopeSpecificationsAllRequest", &metadatatypes.ScopeSpecificationsAllResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.ContractSpecificationRequest", &metadatatypes.ContractSpecificationResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.ContractSpecificationsAllRequest", &metadatatypes.ContractSpecificationsAllResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.RecordSpecificationsForContractSpecificationRequest", &metadatatypes.RecordSpecificationsForContractSpecificationResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.RecordSpecificationRequest", &metadatatypes.RecordSpecificationResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.RecordSpecificationsAllRequest", &metadatatypes.RecordSpecificationsAllResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.OSLocatorParamsRequest", &metadatatypes.OSLocatorParamsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.OSLocatorRequest", &metadatatypes.OSLocatorResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.OSLocatorsByURIRequest", &metadatatypes.OSLocatorsByURIResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.OSLocatorsByScopeRequest", &metadatatypes.OSLocatorsByScopeResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.OSAllLocatorsRequest", &metadatatypes.OSAllLocatorsResponse{})

	// msg fee
	setWhitelistedQuery("/provenance.msgfees.v1.QueryParamsRequest", &msgfeestypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.msgfees.v1.QueryAllMsgFeesRequest", &msgfeestypes.QueryAllMsgFeesResponse{})

	// name
	setWhitelistedQuery("/provenance.name.v1.QueryParamsRequest", &nametypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.name.v1.QueryResolveRequest", &nametypes.QueryResolveResponse{})
	setWhitelistedQuery("/provenance.name.v1.QueryReverseLookupRequest", &nametypes.QueryReverseLookupResponse{})

	// reward
	setWhitelistedQuery("/provenance.reward.v1.QueryRewardProgramByIDRequest", &rewardtypes.QueryRewardProgramByIDResponse{})
	setWhitelistedQuery("/provenance.reward.v1.QueryRewardProgramsRequest", &rewardtypes.QueryRewardProgramsResponse{})
	setWhitelistedQuery("/provenance.reward.v1.QueryClaimPeriodRewardDistributionsRequest", &rewardtypes.QueryClaimPeriodRewardDistributionsResponse{})
	setWhitelistedQuery("/provenance.reward.v1.QueryClaimPeriodRewardDistributionsByIDRequest", &rewardtypes.QueryClaimPeriodRewardDistributionsByIDResponse{})
	setWhitelistedQuery("/provenance.reward.v1.QueryRewardDistributionsByAddressRequest", &rewardtypes.QueryRewardDistributionsByAddressResponse{})
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
	stargateWhitelist.Range(func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			panic("key is not a string")
		}
		keys = append(keys, keyStr)
		return true
	})

	return keys
}
