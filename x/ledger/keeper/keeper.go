package keeper

import (
	"fmt"
	"strings"

	"cosmossdk.io/core/store"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/provenance-io/provenance/x/ledger"
	"github.com/provenance-io/provenance/x/registry"
)

var _ Keeper = (*BaseKeeper)(nil)

type Keeper interface {
}

// Keeper defines the mymodule keeper.
type BaseKeeper struct {
	BaseViewKeeper
	BaseConfigKeeper
	BaseEntriesKeeper
	BaseFundTransferKeeper
}

const (
	ledgerPrefix                 = "ledgers"
	entriesPrefix                = "ledger_entries"
	fundTransfersPrefix          = "fund_transfers"
	ledgerClassesPrefix          = "ledger_classes"
	ledgerClassEntryTypesPrefix  = "ledger_class_entry_types"
	ledgerClassStatusTypesPrefix = "ledger_class_status_types"
	ledgerClassBucketTypesPrefix = "ledger_class_bucket_types"

	ledgerKeyHrp   = "ledger"
	ledgerClassHrp = "ledgerc"
	ledgerEntryHrp = "ledgere"
)

// NewKeeper returns a new mymodule Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, storeService store.KVStoreService, bankKeeper BankKeeper, registryKeeper RegistryKeeper) BaseKeeper {
	viewKeeper := NewBaseViewKeeper(cdc, storeKey, storeService, registryKeeper)

	return BaseKeeper{
		BaseViewKeeper: viewKeeper,
		BaseConfigKeeper: BaseConfigKeeper{
			BaseViewKeeper: viewKeeper,
			BankKeeper:     bankKeeper,
		},
		BaseEntriesKeeper: BaseEntriesKeeper{
			BaseViewKeeper: viewKeeper,
		},
		BaseFundTransferKeeper: BaseFundTransferKeeper{
			BankKeeper: bankKeeper,
		},
	}
}

// Combine the asset class id and nft id into a bech32 string.
// Using bech32 here just allows us a readable identifier for the ledger.
func LedgerKeyToString(key *ledger.LedgerKey) (*string, error) {
	joined := strings.Join([]string{key.AssetClassId, key.NftId}, ":")

	b32, err := bech32.ConvertAndEncode(ledgerKeyHrp, []byte(joined))
	if err != nil {
		return nil, err
	}

	return &b32, nil
}

func StringToLedgerKey(s string) (*ledger.LedgerKey, error) {
	hrp, b, err := bech32.DecodeAndConvert(s)
	if err != nil {
		return nil, err
	}

	if hrp != ledgerKeyHrp {
		return nil, fmt.Errorf("invalid hrp: %s", hrp)
	}

	parts := strings.Split(string(b), ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid key: %s", s)
	}

	return &ledger.LedgerKey{
		AssetClassId: parts[0],
		NftId:        parts[1],
	}, nil
}

func RequireAuthority(ctx sdk.Context, rk RegistryKeeper, addr string, key *registry.RegistryKey) error {
	has, err := assertAuthority(ctx, rk, addr, key)
	if err != nil {
		return err
	}
	if !has {
		return NewLedgerCodedError(ErrCodeUnauthorized, "authority is not the owner or servicer")
	}
	return nil
}

func assertOwner(ctx sdk.Context, k RegistryKeeper, authorityAddr string, ledgerKey *ledger.LedgerKey) error {
	// Check if the authority has ownership of the NFT
	nftOwner := k.GetNFTOwner(ctx, &ledgerKey.AssetClassId, &ledgerKey.NftId)
	if nftOwner == nil || nftOwner.String() != authorityAddr {
		return NewLedgerCodedError(ErrCodeUnauthorized, "authority is not the nft owner")
	}

	return nil
}

// Assert that the authority address is either the registered servicer, or the owner of the NFT if there is no registered servicer.
func assertAuthority(ctx sdk.Context, k RegistryKeeper, authorityAddr string, rk *registry.RegistryKey) (bool, error) {
	// Get the registry entry for the NFT to determine if the authority has the servicer role.
	registryEntry, err := k.GetRegistry(ctx, rk)
	if err != nil {
		return false, err
	}

	lk := &ledger.LedgerKey{
		AssetClassId: rk.AssetClassId,
		NftId:        rk.NftId,
	}

	if registryEntry == nil {
		err = assertOwner(ctx, k, authorityAddr, lk)
		if err != nil {
			return false, err
		}

		return true, nil
	} else {
		// Since the authority doesn't have the servicer role, let's see if there is any servicer set. If there is, we'll return an error
		// so that only the assigned servicer can append entries.
		var servicerRegistered bool = false
		for _, role := range registryEntry.Roles {
			if role.Role == registry.RegistryRole_REGISTRY_ROLE_SERVICER {
				// Note that there is a registered servicer since we allow the owner to be the servicer if there is a registry without one.
				servicerRegistered = true
				for _, address := range role.Addresses {
					// Check if the authority is the servicer
					if address == authorityAddr {
						return true, nil
					}
				}

				// Since there isn't a registered servicer, we'll check if the authority is the owner
				if !servicerRegistered {
					err = assertOwner(ctx, k, authorityAddr, lk)
					if err != nil {
						return false, err
					}
				} else {
					return false, NewLedgerCodedError(ErrCodeUnauthorized, "registered servicer")
				}
			}
		}
	}

	// Default to false if the authority is not the owner or servicer
	return false, nil
}
