package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/sharding/types"
)

// GetPet returns the stored pet
func (k Keeper) GetPet(ctx sdk.Context, petID uint64) *types.Pet {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPetKey(petID)
	bz := store.Get(key)
	pet := types.Pet{}
	k.cdc.Unmarshal(bz, &pet)
	return &pet
}

// SetPet sets the pet
func (k Keeper) SetPet(ctx sdk.Context, pet *types.Pet) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(pet)
	store.Set(types.GetPetKey(pet.GetId()), bz)
}

// GetPetInfo returns the stored pet info
func (k Keeper) GetPetInfo(ctx sdk.Context, petID uint64) *types.PetInfo {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPetInfoKey(petID)
	bz := store.Get(key)
	petinfo := types.PetInfo{}
	k.cdc.Unmarshal(bz, &petinfo)
	return &petinfo
}

// SetPetInfo sets the pet info
func (k Keeper) SetPetInfo(ctx sdk.Context, petID uint64, petinfo *types.PetInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(petinfo)
	store.Set(types.GetPetInfoKey(petID), bz)
}

// GetPetOwner returns the stored pet info
func (k Keeper) GetPetOwner(ctx sdk.Context, petID uint64) (owner sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPetOwnerKey(petID)
	owner = store.Get(key)
	return owner
}

// SetPetOwner sets the pet info
func (k Keeper) SetPetOwner(ctx sdk.Context, petID uint64, owner sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPetOwnerKey(petID), owner)
}

// GetPetName returns the stored pet name
func (k Keeper) GetPetName(ctx sdk.Context, petID uint64) (name string) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPetNameKey(petID)
	name = string(store.Get(key))
	return name
}

// SetPetName sets the pet name
func (k Keeper) SetPetName(ctx sdk.Context, petID uint64, name string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPetNameKey(petID), []byte(name))
}

// GetPetColor returns the stored pet color
func (k Keeper) GetPetColor(ctx sdk.Context, petID uint64) (color string) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPetColorKey(petID)
	color = string(store.Get(key))
	return color
}

// SetPetColor sets the pet color
func (k Keeper) SetPetColor(ctx sdk.Context, petID uint64, color string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetPetColorKey(petID), []byte(color))
}

// GetPetSpots returns the stored pet color
func (k Keeper) GetPetSpots(ctx sdk.Context, petID uint64) (spots uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPetSpotsKey(petID)
	spots = binary.BigEndian.Uint64(store.Get(key))
	return spots
}

// SetPetSpots sets the pet color
func (k Keeper) SetPetSpots(ctx sdk.Context, petID uint64, spots uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, spots)
	store.Set(types.GetPetSpotsKey(petID), bz)
}

func (k Keeper) DefaultPet() *types.Pet {
	return &types.Pet{
		Id:    1,
		Owner: "Matt",
		PetInfo: types.PetInfo{
			Name:  "Bubbbles",
			Color: "Gray",
			Spots: 0,
		},
	}
}
