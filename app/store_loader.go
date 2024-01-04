package app

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

// WrapStoreLoader creates a new StoreLoader by wrapping an existing one.
func WrapStoreLoader(wrapper func(sdk.CommitMultiStore, baseapp.StoreLoader) error, storeLoader baseapp.StoreLoader) baseapp.StoreLoader {
	return func(ms sdk.CommitMultiStore) error {
		if storeLoader == nil {
			storeLoader = baseapp.DefaultStoreLoader
		}
		return wrapper(ms, storeLoader)
	}
}

// DBBackendWrapper creates a new StoreLoader that first validates the DBBackend type before calling the provided StoreLoader.
func DBBackendWrapper(logger log.Logger, appOpts servertypes.AppOptions, storeLoader baseapp.StoreLoader) baseapp.StoreLoader {
	return WrapStoreLoader(func(ms sdk.CommitMultiStore, sl baseapp.StoreLoader) error {
		backend := server.GetAppDBBackend(appOpts)
		if backend != dbm.GoLevelDBBackend {
			logger.Error(fmt.Sprintf("%s IS NO LONGER SUPPORTED. MIGRATE TO %s", backend, dbm.GoLevelDBBackend))
			time.Sleep(30 * time.Second)
		}

		return sl(ms)
	}, storeLoader)
}

// PruningWrapper creates a new StoreLoader that first validates the pruning settings before calling the provided StoreLoader.
func PruningWrapper(logger log.Logger, appOpts servertypes.AppOptions, storeLoader baseapp.StoreLoader) baseapp.StoreLoader {
	return WrapStoreLoader(func(ms sdk.CommitMultiStore, sl baseapp.StoreLoader) error {
		const MAX_PRUNING_INTERVAL = 100

		// No error checking is needed because we leave that up to the sdk.
		pruningString, _ := appOpts.Get("pruning-interval").(string)
		interval, _ := strconv.ParseUint(pruningString, 10, 64)

		if interval > MAX_PRUNING_INTERVAL {
			logger.Error(fmt.Sprintf("pruning-interval %s IS INVALID AND CANNOT EXCEED %d", pruningString, MAX_PRUNING_INTERVAL))
			time.Sleep(30 * time.Second)
		}
		return sl(ms)
	}, storeLoader)
}
