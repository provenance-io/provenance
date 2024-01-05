package app

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
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
			db := dbm.GoLevelDB{}
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

func TestPruningWrapper(t *testing.T) {

}

func createMockStoreLoader() baseapp.StoreLoader {
	return func(ms sdk.CommitMultiStore) error {
		return nil
	}
}

func createMockFlipWrapper(flag *bool) StoreLoaderWrapper {
	return func(cms sdk.CommitMultiStore, sl baseapp.StoreLoader) error {
		*flag = !(*flag)
		return nil
	}
}

func createMockStoreWrapper(flag *bool) StoreLoaderWrapper {
	return func(cms sdk.CommitMultiStore, sl baseapp.StoreLoader) error {
		*flag = sl != nil
		return nil
	}
}
