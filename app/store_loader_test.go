package app

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func TestWrapStoreLoader(t *testing.T) {
	var flag bool
	tests := []struct {
		name        string
		storeLoader baseapp.StoreLoader
		wrapper     StoreLoaderWrapper
		err         string
	}{
		{
			name:        "nil store loader is set with valid value",
			storeLoader: nil,
			wrapper:     createMockStoreWrapper(&flag),
		},
		{
			name:        "nil wrapper is handled",
			storeLoader: createMockStoreLoader(),
			wrapper:     nil,
			err:         "wrapper must not be nil",
		},
		{
			name:        "contents of wrapper are called",
			storeLoader: createMockStoreLoader(),
			wrapper:     createMockFlipWrapper(&flag),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storeLoader := WrapStoreLoader(tc.wrapper, tc.storeLoader)
			db := dbm.MemDB{}
			ms := rootmulti.NewStore(&db, nil)
			assert.NotNil(t, ms, "should create a new multistore for testing")
			flag = false

			err := storeLoader(ms)
			if len(tc.err) > 0 {
				assert.EqualError(t, err, tc.err, "should have correct error")
				assert.False(t, flag, "wrapper should not be executed")
			} else {
				assert.NoError(t, err, "should not return an error on success")
				assert.True(t, flag, "wrapper should execute and have correct logic")
			}

		})
	}
}

func TestValidatorWrapper(t *testing.T) {
	tests := []struct {
		name    string
		appOpts MockAppOptions
		delta   uint64
	}{
		{
			name: "recommended pruning, indexer, and db should not wait",
			appOpts: MockAppOptions{
				pruning: "13",
				db:      "goleveldb",
			},
			delta: 0,
		},
		{
			name: "non-recommended pruning should wait",
			appOpts: MockAppOptions{
				pruning: "1000",
			},
			delta: 30,
		},
		{
			name: "non-recommended indexer should wait",
			appOpts: MockAppOptions{
				pruning: "13",
				indexer: "kv",
			},
			delta: 30,
		},
		{
			name: "non-recommended db should wait",
			appOpts: MockAppOptions{
				pruning: "13",
				db:      "cleveldb",
			},
			delta: 30,
		},
		{
			name: "multiple non-recommended should wait",
			appOpts: MockAppOptions{
				pruning: "1000",
				indexer: "kv",
			},
			delta: 30,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := log.NewNopLogger()
			storeLoader := ValidatorWrapper(logger, tc.appOpts, createMockStoreLoader())
			db := dbm.MemDB{}
			ms := rootmulti.NewStore(&db, nil)
			assert.NotNil(t, ms, "should create a new multistore for testing")

			start := time.Now()
			err := storeLoader(ms)
			delta := uint64(time.Now().Sub(start).Seconds())
			assert.NoError(t, err, "should not throw error")
			assert.GreaterOrEqual(t, delta, tc.delta, "should wait correct amount of time")
		})
	}
}

// createMockStoreLoader creates an empty StoreLoader.
func createMockStoreLoader() baseapp.StoreLoader {
	return func(ms sdk.CommitMultiStore) error {
		return nil
	}
}

// createMockFlipWrapper creates a wrapper that has logic to flip a bit.
func createMockFlipWrapper(flag *bool) StoreLoaderWrapper {
	return func(cms sdk.CommitMultiStore, sl baseapp.StoreLoader) error {
		*flag = !(*flag)
		return nil
	}
}

// createMockStoreWrapper creates a wrapper that checks if the StoreLoader is nil and sets the flag accordingly.
func createMockStoreWrapper(flag *bool) StoreLoaderWrapper {
	return func(cms sdk.CommitMultiStore, sl baseapp.StoreLoader) error {
		*flag = sl != nil
		return nil
	}
}

// MockAppOptions is a mocked version of AppOpts that allows the developer to provide the pruning attribute.
type MockAppOptions struct {
	pruning string
	indexer string
	db      string
}

// Get returns the value for the provided option.
func (m MockAppOptions) Get(opt string) interface{} {
	switch opt {
	case "pruning-interval":
		return m.pruning
	case "tx_index":
		return map[string]interface{}{
			"indexer": m.indexer,
		}
	case "app-db-backend":
		return m.db
	case "db-backend":
		return m.db
	}

	return nil
}
