package provwasm

import (
	"fmt"
	"sync"

	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	circuittypes "cosmossdk.io/x/circuit/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/hold"
	ibchookstypes "github.com/provenance-io/provenance/x/ibchooks/types"
	"github.com/provenance-io/provenance/x/ibcratelimit"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
	oracletypes "github.com/provenance-io/provenance/x/oracle/types"
	"github.com/provenance-io/provenance/x/quarantine"
	"github.com/provenance-io/provenance/x/sanction"
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
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/Accounts", &authtypes.QueryAccountsResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/Account", &authtypes.QueryAccountResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/AccountAddressByID", &authtypes.QueryAccountAddressByIDResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/Params", &authtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/ModuleAccounts", &authtypes.QueryModuleAccountsResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/ModuleAccountByName", &authtypes.QueryModuleAccountByNameResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/Bech32Prefix", &authtypes.Bech32PrefixResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/AddressBytesToString", &authtypes.AddressBytesToStringResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/AddressStringToBytes", &authtypes.AddressStringToBytesResponse{})
	setWhitelistedQuery("/cosmos.auth.v1beta1.Query/AccountInfo", &authtypes.QueryAccountInfoResponse{})

	// authz
	setWhitelistedQuery("/cosmos.authz.v1beta1.Query/Grants", &authztypes.QueryGrantsResponse{})
	setWhitelistedQuery("/cosmos.authz.v1beta1.Query/GranterGrants", &authztypes.QueryGranterGrantsResponse{})
	setWhitelistedQuery("/cosmos.authz.v1beta1.Query/GranteeGrants", &authztypes.QueryGranteeGrantsResponse{})

	// bank
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/Balance", &banktypes.QueryBalanceResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/AllBalances", &banktypes.QueryAllBalancesResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/SpendableBalances", &banktypes.QuerySpendableBalancesResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/SpendableBalanceByDenom", &banktypes.QuerySpendableBalanceByDenomResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/TotalSupply", &banktypes.QueryTotalSupplyResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/SupplyOf", &banktypes.QuerySupplyOfResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/Params", &banktypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/DenomMetadata", &banktypes.QueryDenomMetadataResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/DenomMetadataByQueryString", &banktypes.QueryDenomMetadataByQueryStringResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/DenomsMetadata", &banktypes.QueryDenomsMetadataResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/DenomOwners", &banktypes.QueryDenomOwnersResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/DenomOwnersByQuery", &banktypes.QueryDenomOwnersByQueryResponse{})
	setWhitelistedQuery("/cosmos.bank.v1beta1.Query/SendEnabled", &banktypes.QuerySendEnabledResponse{})

	// circuit
	setWhitelistedQuery("/cosmos.circuit.v1.Query/Account", &circuittypes.AccountResponse{})
	setWhitelistedQuery("/cosmos.circuit.v1.Query/Accounts", &circuittypes.AccountsResponse{})
	setWhitelistedQuery("/cosmos.circuit.v1.Query/DisabledList", &circuittypes.DisabledListResponse{})

	// consensus
	setWhitelistedQuery("/cosmos.consensus.v1.Query/Params", &consensustypes.QueryParamsResponse{})

	// distribution
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/Params", &distributiontypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/ValidatorDistributionInfo", &distributiontypes.QueryValidatorDistributionInfoResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards", &distributiontypes.QueryValidatorOutstandingRewardsResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/ValidatorCommission", &distributiontypes.QueryValidatorCommissionResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/ValidatorSlashes", &distributiontypes.QueryValidatorSlashesResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/DelegationRewards", &distributiontypes.QueryDelegationRewardsResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/DelegationTotalRewards", &distributiontypes.QueryDelegationTotalRewardsResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/DelegatorValidators", &distributiontypes.QueryDelegatorValidatorsResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress", &distributiontypes.QueryDelegatorWithdrawAddressResponse{})
	setWhitelistedQuery("/cosmos.distribution.v1beta1.Query/CommunityPool", &distributiontypes.QueryCommunityPoolResponse{})

	// evidence
	setWhitelistedQuery("/cosmos.evidence.v1beta1.Query/Evidence", &evidencetypes.QueryEvidenceResponse{})
	setWhitelistedQuery("/cosmos.evidence.v1beta1.Query/AllEvidence", &evidencetypes.QueryAllEvidenceResponse{})

	// feegrant
	setWhitelistedQuery("/cosmos.feegrant.v1beta1.Query/Allowance", &feegrant.QueryAllowanceResponse{})
	setWhitelistedQuery("/cosmos.feegrant.v1beta1.Query/Allowances", &feegrant.QueryAllowancesResponse{})
	setWhitelistedQuery("/cosmos.feegrant.v1beta1.Query/AllowancesByGranter", &feegrant.QueryAllowancesByGranterResponse{})

	// gov
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Proposal", &govtypesv1beta1.QueryProposalResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Proposals", &govtypesv1beta1.QueryProposalsResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Vote", &govtypesv1beta1.QueryVoteResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Votes", &govtypesv1beta1.QueryVotesResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Params", &govtypesv1beta1.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Deposit", &govtypesv1beta1.QueryDepositResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/Deposits", &govtypesv1beta1.QueryDepositsResponse{})
	setWhitelistedQuery("/cosmos.gov.v1beta1.Query/TallyResult", &govtypesv1beta1.QueryTallyResultResponse{})
	setWhitelistedQuery("/cosmos.gov.v1.Query/Constitution", &govtypes.QueryConstitutionResponse{})
	setWhitelistedQuery("/cosmos.gov.v1.Query/Proposal", &govtypes.QueryProposalResponse{})
	setWhitelistedQuery("/cosmos.gov.v1.Query/Proposals", &govtypes.QueryProposalsResponse{})
	setWhitelistedQuery("/cosmos.gov.v1.Query/Vote", &govtypes.QueryVoteResponse{})
	setWhitelistedQuery("/cosmos.gov.v1.Query/Votes", &govtypes.QueryVotesResponse{})
	setWhitelistedQuery("/cosmos.gov.v1.Query/Params", &govtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.gov.v1.Query/Deposit", &govtypes.QueryDepositResponse{})
	setWhitelistedQuery("/cosmos.gov.v1.Query/Deposits", &govtypes.QueryDepositsResponse{})
	setWhitelistedQuery("/cosmos.gov.v1.Query/TallyResult", &govtypes.QueryTallyResultResponse{})

	// group
	setWhitelistedQuery("/cosmos.group.v1.Query/GroupInfo", &group.QueryGroupInfoResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/GroupPolicyInfo", &group.QueryGroupPolicyInfoResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/GroupMembers", &group.QueryGroupMembersResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/GroupsByAdmin", &group.QueryGroupsByAdminResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/GroupPoliciesByGroup", &group.QueryGroupPoliciesByGroupResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/GroupPoliciesByAdmin", &group.QueryGroupPoliciesByAdminResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/Proposal", &group.QueryProposalResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/ProposalsByGroupPolicy", &group.QueryProposalsByGroupPolicyResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/VoteByProposalVoter", &group.QueryVoteByProposalVoterResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/VotesByProposal", &group.QueryVotesByProposalResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/VotesByVoter", &group.QueryVotesByVoterResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/GroupsByMember", &group.QueryGroupsByMemberResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/TallyResult", &group.QueryTallyResultResponse{})
	setWhitelistedQuery("/cosmos.group.v1.Query/Groups", &group.QueryGroupsResponse{})

	// mint
	setWhitelistedQuery("/cosmos.mint.v1beta1.Query/Params", &minttypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.mint.v1beta1.Query/Inflation", &minttypes.QueryInflationResponse{})
	setWhitelistedQuery("/cosmos.mint.v1beta1.Query/AnnualProvisions", &minttypes.QueryAnnualProvisionsResponse{})

	// slashing
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/Params", &slashingtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/SigningInfo", &slashingtypes.QuerySigningInfoResponse{})
	setWhitelistedQuery("/cosmos.slashing.v1beta1.Query/SigningInfos", &slashingtypes.QuerySigningInfosResponse{})

	// staking
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Validators", &stakingtypes.QueryValidatorsResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Validator", &stakingtypes.QueryValidatorResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/ValidatorDelegations", &stakingtypes.QueryValidatorDelegationsResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/ValidatorUnbondingDelegations", &stakingtypes.QueryValidatorUnbondingDelegationsResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Delegation", &stakingtypes.QueryDelegationResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/UnbondingDelegation", &stakingtypes.QueryUnbondingDelegationResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/DelegatorDelegations", &stakingtypes.QueryDelegatorDelegationsResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations", &stakingtypes.QueryDelegatorUnbondingDelegationsResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Redelegations", &stakingtypes.QueryRedelegationsResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/DelegatorValidators", &stakingtypes.QueryDelegatorValidatorsResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/DelegatorValidator", &stakingtypes.QueryDelegatorValidatorResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/HistoricalInfo", &stakingtypes.QueryHistoricalInfoResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Pool", &stakingtypes.QueryPoolResponse{})
	setWhitelistedQuery("/cosmos.staking.v1beta1.Query/Params", &stakingtypes.QueryParamsResponse{})

	// upgrade
	setWhitelistedQuery("/cosmos.upgrade.v1beta1.Query/CurrentPlan", &upgradetypes.QueryCurrentPlanResponse{})
	setWhitelistedQuery("/cosmos.upgrade.v1beta1.Query/AppliedPlan", &upgradetypes.QueryAppliedPlanResponse{})
	setWhitelistedQuery("/cosmos.upgrade.v1beta1.Query/ModuleVersions", &upgradetypes.QueryModuleVersionsResponse{})
	setWhitelistedQuery("/cosmos.upgrade.v1beta1.Query/Authority", &upgradetypes.QueryAuthorityResponse{})

	// wasm
	setWhitelistedQuery("/cosmwasm.wasm.v1.Query/ContractHistory", &wasmtypes.QueryContractInfoResponse{})
	setWhitelistedQuery("/cosmwasm.wasm.v1.Query/ContractsByCode", &wasmtypes.QueryContractsByCodeResponse{})
	setWhitelistedQuery("/cosmwasm.wasm.v1.Query/SmartContractState", &wasmtypes.QuerySmartContractStateResponse{})
	setWhitelistedQuery("/cosmwasm.wasm.v1.Query/Code", &wasmtypes.QueryCodeResponse{})
	setWhitelistedQuery("/cosmwasm.wasm.v1.Query/Codes", &wasmtypes.QueryCodesResponse{})
	setWhitelistedQuery("/cosmwasm.wasm.v1.Query/PinnedCodes", &wasmtypes.QueryPinnedCodesResponse{})
	setWhitelistedQuery("/cosmwasm.wasm.v1.Query/Params", &wasmtypes.QueryParamsResponse{})
	setWhitelistedQuery("/cosmwasm.wasm.v1.Query/ContractsByCreator", &wasmtypes.QueryContractsByCreatorResponse{})

	// ==========================================================
	// provenance queries
	// ==========================================================

	// attribute
	setWhitelistedQuery("/provenance.attribute.v1.Query/Params", &attributetypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.attribute.v1.Query/Attribute", &attributetypes.QueryAttributeResponse{})
	setWhitelistedQuery("/provenance.attribute.v1.Query/Attributes", &attributetypes.QueryAttributesResponse{})
	setWhitelistedQuery("/provenance.attribute.v1.Query/Scan", &attributetypes.QueryScanResponse{})
	setWhitelistedQuery("/provenance.attribute.v1.Query/AttributeAccounts", &attributetypes.QueryAttributeAccountsResponse{})
	setWhitelistedQuery("/provenance.attribute.v1.Query/AccountData", &attributetypes.QueryAccountDataResponse{})

	// exchange
	setWhitelistedQuery("/provenance.exchange.v1.Query/OrderFeeCalc", &exchange.QueryOrderFeeCalcResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetOrder", &exchange.QueryGetOrderResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetOrderByExternalID", &exchange.QueryGetOrderByExternalIDResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetMarketOrders", &exchange.QueryGetMarketOrdersResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetOwnerOrders", &exchange.QueryGetOwnerOrdersResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetAssetOrders", &exchange.QueryGetAssetOrdersResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetAllOrders", &exchange.QueryGetAllOrdersResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetCommitment", &exchange.QueryGetCommitmentResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetAccountCommitments", &exchange.QueryGetAccountCommitmentsResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetMarketCommitments", &exchange.QueryGetMarketCommitmentsResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetAllCommitments", &exchange.QueryGetAllCommitmentsResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetMarket", &exchange.QueryGetMarketResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetAllMarkets", &exchange.QueryGetAllMarketsResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/Params", &exchange.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/CommitmentSettlementFeeCalc", &exchange.QueryCommitmentSettlementFeeCalcResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/ValidateCreateMarket", &exchange.QueryValidateCreateMarketResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/ValidateMarket", &exchange.QueryValidateMarketResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/ValidateManageFees", &exchange.QueryValidateManageFeesResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetPayment", &exchange.QueryGetPaymentResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetPaymentsWithSource", &exchange.QueryGetPaymentsWithSourceResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetPaymentsWithTarget", &exchange.QueryGetPaymentsWithTargetResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/GetAllPayments", &exchange.QueryGetAllPaymentsResponse{})
	setWhitelistedQuery("/provenance.exchange.v1.Query/PaymentFeeCalc", &exchange.QueryPaymentFeeCalcResponse{})

	// hold
	setWhitelistedQuery("/provenance.hold.v1.Query/GetHolds", &hold.GetHoldsResponse{})
	setWhitelistedQuery("/provenance.hold.v1.Query/GetAllHolds", &hold.GetAllHoldsResponse{})

	// ibcratelimit
	setWhitelistedQuery("/provenance.ibcratelimit.v1.Query/Params", &ibcratelimit.ParamsResponse{})

	// ibchooks
	setWhitelistedQuery("/provenance.ibchooks.v1.Query/Params", &ibchookstypes.QueryParamsResponse{})

	// marker
	setWhitelistedQuery("/provenance.marker.v1.Query/Params", &markertypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/AllMarkers", &markertypes.QueryAllMarkersResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/Marker", &markertypes.QueryMarkerResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/Holding", &markertypes.QueryHoldingResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/Supply", &markertypes.QuerySupplyResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/Escrow", &markertypes.QueryEscrowResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/Access", &markertypes.QueryAccessResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/DenomMetadata", &markertypes.QueryDenomMetadataResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/AccountData", &markertypes.QueryAccountDataResponse{})
	setWhitelistedQuery("/provenance.marker.v1.Query/NetAssetValues", &markertypes.QueryNetAssetValuesResponse{})

	// metadata
	setWhitelistedQuery("/provenance.metadata.v1.Query/Params", &metadatatypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/Scope", &metadatatypes.ScopeResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/ScopesAll", &metadatatypes.ScopesAllResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/Sessions", &metadatatypes.SessionsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/SessionsAll", &metadatatypes.SessionsAllResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/Records", &metadatatypes.RecordsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/RecordsAll", &metadatatypes.RecordsAllResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/Ownership", &metadatatypes.OwnershipResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/ValueOwnership", &metadatatypes.ValueOwnershipResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/ScopeSpecification", &metadatatypes.ScopeSpecificationResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/ScopeSpecificationsAll", &metadatatypes.ScopeSpecificationsAllResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/ContractSpecification", &metadatatypes.ContractSpecificationResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/ContractSpecificationsAll", &metadatatypes.ContractSpecificationsAllResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/RecordSpecificationsForContractSpecification", &metadatatypes.RecordSpecificationsForContractSpecificationResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/RecordSpecification", &metadatatypes.RecordSpecificationResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/RecordSpecificationsAll", &metadatatypes.RecordSpecificationsAllResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/GetByAddr", &metadatatypes.GetByAddrResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/OSLocatorParams", &metadatatypes.OSLocatorParamsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/OSLocator", &metadatatypes.OSLocatorResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/OSLocatorsByURI", &metadatatypes.OSLocatorsByURIResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/OSLocatorsByScope", &metadatatypes.OSLocatorsByScopeResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/OSAllLocators", &metadatatypes.OSAllLocatorsResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/AccountData", &metadatatypes.AccountDataResponse{})
	setWhitelistedQuery("/provenance.metadata.v1.Query/ScopeNetAssetValues", &metadatatypes.QueryScopeNetAssetValuesResponse{})

	// msg fee
	setWhitelistedQuery("/provenance.msgfees.v1.Query/Params", &msgfeestypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.msgfees.v1.Query/QueryAllMsgFees", &msgfeestypes.QueryAllMsgFeesResponse{})
	setWhitelistedQuery("/provenance.msgfees.v1.Query/CalculateTxFees", &msgfeestypes.CalculateTxFeesResponse{})

	// name
	setWhitelistedQuery("/provenance.name.v1.Query/Params", &nametypes.QueryParamsResponse{})
	setWhitelistedQuery("/provenance.name.v1.Query/Resolve", &nametypes.QueryResolveResponse{})
	setWhitelistedQuery("/provenance.name.v1.Query/ReverseLookup", &nametypes.QueryReverseLookupResponse{})

	// oracle
	setWhitelistedQuery("/provenance.oracle.v1.Query/OracleAddress", &oracletypes.QueryOracleAddressResponse{})
	setWhitelistedQuery("/provenance.oracle.v1.Query/Oracle", &oracletypes.QueryOracleResponse{})

	// quarantine
	setWhitelistedQuery("/cosmos.quarantine.v1beta1.Query/IsQuarantined", &quarantine.QueryIsQuarantinedResponse{})
	setWhitelistedQuery("/cosmos.quarantine.v1beta1.Query/QuarantinedFunds", &quarantine.QueryQuarantinedFundsResponse{})
	setWhitelistedQuery("/cosmos.quarantine.v1beta1.Query/AutoResponses", &quarantine.QueryAutoResponsesResponse{})

	// sanction
	setWhitelistedQuery("/cosmos.sanction.v1beta1.Query/IsSanctioned", &sanction.QueryIsSanctionedResponse{})
	setWhitelistedQuery("/cosmos.sanction.v1beta1.Query/SanctionedAddresses", &sanction.QuerySanctionedAddressesResponse{})
	setWhitelistedQuery("/cosmos.sanction.v1beta1.Query/TemporaryEntries", &sanction.QueryTemporaryEntriesResponse{})
	setWhitelistedQuery("/cosmos.sanction.v1beta1.Query/Params", &sanction.QueryParamsResponse{})

	// trigger
	setWhitelistedQuery("/provenance.trigger.v1.Query/TriggerByID", &triggertypes.QueryTriggerByIDResponse{})
	setWhitelistedQuery("/provenance.trigger.v1.Query/Triggers", &triggertypes.QueryTriggersResponse{})
}

// GetWhitelistedQuery returns the whitelisted query at the provided path.
// If the query does not exist, or it was setup wrong by the chain, this returns an error.
func GetWhitelistedQuery(queryPath string) (proto.Message, error) {
	protoResponseAny, isWhitelisted := stargateWhitelist.Load(queryPath)
	if !isWhitelisted {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", queryPath)}
	}
	protoResponseType, ok := protoResponseAny.(proto.Message)
	if !ok {
		return nil, wasmvmtypes.Unknown{}
	}
	return protoResponseType, nil
}

func setWhitelistedQuery(queryPath string, protoType proto.Message) {
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
