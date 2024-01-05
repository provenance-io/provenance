package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cast"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// ValidatorWrapper creates a new StoreLoader that first checks the validator settings before calling the provided StoreLoader.
func ValidatorWrapper(logger log.Logger, appOpts servertypes.AppOptions, storeLoader baseapp.StoreLoader) baseapp.StoreLoader {
	return WrapStoreLoader(func(ms sdk.CommitMultiStore, sl baseapp.StoreLoader) error {
		const MaxPruningInterval = 999
		const SleepSeconds = 30
		interval := cast.ToUint64(appOpts.Get("pruning-interval"))
		txIndexer := cast.ToStringMap(appOpts.Get("tx_index"))
		indexer := cast.ToString(txIndexer["indexer"])
		hasError := false

		if interval > MaxPruningInterval {
			logger.Error(fmt.Sprintf("pruning-interval %d EXCEEDS %d AND IS NOT RECOMMENDED, AS IT CAN LEAD TO MISSED BLOCKS ON VALIDATORS", interval, MaxPruningInterval))
			hasError = true
		}

		if indexer != "" {
			logger.Error(fmt.Sprintf("indexer \"%s\" IS NOT RECOMMENDED, AND IS RECOMMENDED TO USE %s", indexer, "\"\""))
			hasError = true
		}

		if hasError {
			logger.Error(fmt.Sprintf("NODE WILL CONTINUE AFTER %d SECONDS", SleepSeconds))
			time.Sleep(SleepSeconds * time.Second)
		}

		return sl(ms)
	}, storeLoader)
}
