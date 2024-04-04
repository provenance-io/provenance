package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
)

// ValidateWrapperSleeper is the sleeper that the ValidateWrapper will use.
// It primarily exists so it can be changed for unit tests on ValidateWrapper so they don't take so long.
var ValidateWrapperSleeper Sleeper = &DefaultSleeper{}

// ValidateWrapper creates a new StoreLoader that first checks the config settings before calling the provided StoreLoader.
func ValidateWrapper(logger log.Logger, appOpts servertypes.AppOptions, storeLoader baseapp.StoreLoader) baseapp.StoreLoader {
	return func(ms storetypes.CommitMultiStore) error {
		IssueConfigWarnings(logger, appOpts, ValidateWrapperSleeper)
		return storeLoader(ms)
	}
}

// Sleeper is an interface for something with a Sleep function.
type Sleeper interface {
	Sleep(d time.Duration)
}

// DefaultSleeper uses the time.Sleep function for sleeping.
type DefaultSleeper struct{}

// Sleep is a wrapper for time.Sleep(d).
func (s DefaultSleeper) Sleep(d time.Duration) {
	time.Sleep(d)
}

// IssueConfigWarnings checks a few values in the configs and issues warnings and sleeps if appropriate.
func IssueConfigWarnings(logger log.Logger, appOpts servertypes.AppOptions, sleeper Sleeper) {
	const MaxPruningInterval = 999
	const SleepSeconds = 30
	interval := cast.ToUint64(appOpts.Get("pruning-interval"))
	txIndexer := cast.ToStringMap(appOpts.Get("tx_index"))
	indexer := cast.ToString(txIndexer["indexer"])
	backend := server.GetAppDBBackend(appOpts)
	var errs []string

	if interval > MaxPruningInterval {
		errs = append(errs, fmt.Sprintf("pruning-interval %d EXCEEDS %d AND IS NOT RECOMMENDED, AS IT CAN LEAD TO MISSED BLOCKS ON VALIDATORS.", interval, MaxPruningInterval))
	}

	if indexer != "" && indexer != "null" {
		errs = append(errs, fmt.Sprintf("indexer \"%s\" IS NOT RECOMMENDED, AND IT IS RECOMMENDED TO USE \"%s\".", indexer, "null"))
	}

	if backend != dbm.GoLevelDBBackend {
		errs = append(errs, fmt.Sprintf("%s IS NO LONGER SUPPORTED. MIGRATE TO %s.", backend, dbm.GoLevelDBBackend))
	}

	if len(errs) > 0 {
		for _, err := range errs {
			logger.Error(err)
		}
		if !HaveAckWarn() {
			logger.Error(fmt.Sprintf("NODE WILL CONTINUE AFTER %d SECONDS.", SleepSeconds))
			logger.Error("This wait can be bypassed by fixing the above warnings or setting the PIO_ACKWARN environment variable to \"1\".")
			sleeper.Sleep(SleepSeconds * time.Second)
		}
	}
}

// HaveAckWarn returns true if the PIO_ACKWARN env var is set and isn't a false value (e.g. "0", "f" or "false").
func HaveAckWarn() bool {
	ackWarn := strings.TrimSpace(os.Getenv("PIO_ACKWARN"))
	if len(ackWarn) == 0 {
		return false
	}

	rv, err := strconv.ParseBool(ackWarn)
	// We return false only if it parsed successfully to a false value.
	// If parsing failed or it parsed to a true value, we return true.
	return err != nil || rv
}
