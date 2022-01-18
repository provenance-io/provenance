package keeper

// This is a _test file that's part of the keeper package (instead of keeper_test like the rest of the unit test files).
// It exists as a way to expose package-private stuff to the unit tests.
// The unit tests are in package keeper_test because of some automated wiring that gets messed up due to the unit test
// structs and extra stuff.

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// IndexContractSpecBad indexes a contract spec the wrong way.
// If the 15th byte of the uuid is even, it's also indexed correctly.
func IndexContractSpecBad(store *sdk.KVStore, spec *types.ContractSpecification) {
	if spec == nil {
		return
	}
	for _, indexKey := range getContractSpecIndexValues(spec).IndexKeys() {
		if indexKey[0] == types.AddressContractSpecCacheKeyPrefix[0] {
			badKey := make([]byte, len(indexKey)-1)
			badKey[0] = indexKey[0]
			copy(badKey[1:], indexKey[2:])
			(*store).Set(badKey, []byte{0x01})
			// 15 chosen based on analysis of input files.
			// main: 19 more odd than even (0.39%).
			// test: 1 more even than odd (0.009%).
			// small: 10 more odd than even (1.0%).
			if spec.SpecificationId[15]%2 == 0 {
				(*store).Set(indexKey, []byte{0x01})
			}
		} else {
			(*store).Set(indexKey, []byte{0x01})
		}
	}
}

// IndexScopeSpecBad indexes a scope spec the wrong way.
// If the 1st byte of the uuid is even, it's also indexed correctly.
func IndexScopeSpecBad(store *sdk.KVStore, spec *types.ScopeSpecification) {
	if spec == nil {
		return
	}
	for _, indexKey := range getScopeSpecIndexValues(spec).IndexKeys() {
		if indexKey[0] == types.AddressScopeSpecCacheKeyPrefix[0] {
			badKey := make([]byte, len(indexKey)-1)
			badKey[0] = indexKey[0]
			copy(badKey[1:], indexKey[2:])
			(*store).Set(badKey, []byte{0x01})
			// 1 chosen based on analysis of input files.
			// main: 0 more even than odd (0.0%).
			// test: 1 more even than odd (6.7%).
			// small: 0 more even than odd (0.0%).
			if spec.SpecificationId[1]%2 == 0 {
				(*store).Set(indexKey, []byte{0x01})
			}
		} else {
			(*store).Set(indexKey, []byte{0x01})
		}
	}
}

// IndexScopeBad indexes a scope the wrong way.
// If the 6th byte of the uuid is even, it's also indexed correctly.
func IndexScopeBad(store *sdk.KVStore, scope *types.Scope) {
	if scope == nil {
		return
	}
	for _, indexKey := range getScopeIndexValues(scope).IndexKeys() {
		if indexKey[0] == types.AddressScopeCacheKeyPrefix[0] || indexKey[0] == types.ValueOwnerScopeCacheKeyPrefix[0] {
			badKey := make([]byte, len(indexKey)-1)
			badKey[0] = indexKey[0]
			copy(badKey[1:], indexKey[2:])
			(*store).Set(badKey, []byte{0x01})
			// 6 chosen based on analysis of input files.
			// main: 30 more even than odd (0.035%).
			// test: 32 more odd than even (0.011%).
			// small: 38 more even than odd (3.8%). One of the worst, but the small set is just a superficial test.
			if scope.ScopeId[6]%2 == 0 {
				(*store).Set(indexKey, []byte{0x01})
			}
		} else {
			(*store).Set(indexKey, []byte{0x01})
		}
	}
}
