package app

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
)

// StoreLoaderWrapper is a wrapper function that is called before the StoreLoader.
type StoreLoaderWrapper func(sdk.CommitMultiStore, baseapp.StoreLoader) error

// WrapStoreLoader creates a new StoreLoader by wrapping an existing one.
func WrapStoreLoader(wrapper StoreLoaderWrapper, storeLoader baseapp.StoreLoader) baseapp.StoreLoader {
	return func(ms sdk.CommitMultiStore) error {
		if storeLoader == nil {
			storeLoader = baseapp.DefaultStoreLoader
		}

		if wrapper == nil {
			return errors.New("wrapper must not be nil")
		}

		return wrapper(ms, storeLoader)
	}
}

// PruningWrapper creates a new StoreLoader that first validates the pruning settings before calling the provided StoreLoader.
func PruningWrapper(logger log.Logger, appOpts servertypes.AppOptions, storeLoader baseapp.StoreLoader) baseapp.StoreLoader {
	return WrapStoreLoader(func(ms sdk.CommitMultiStore, sl baseapp.StoreLoader) error {
		const MaxPruningInterval = 100

		// No error checking is needed because we leave that up to the sdk.
		pruningString, _ := appOpts.Get("pruning-interval").(string)
		interval, _ := strconv.ParseUint(pruningString, 10, 64)

		if interval > MaxPruningInterval {
			logger.Error(fmt.Sprintf("pruning-interval %s EXCEEDS %d AND IS NOT RECOMMENDED, AS IT CAN LEAD TO MISSED BLOCKS ON VALIDATORS", pruningString, MaxPruningInterval))
			time.Sleep(30 * time.Second)
		}
		return sl(ms)
	}, storeLoader)
}
