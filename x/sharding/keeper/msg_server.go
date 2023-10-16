package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/sharding/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the account MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Reads from the store
func (s msgServer) Read(goCtx context.Context, msg *types.MsgReadRequest) (*types.MsgReadResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	id := uint64(1)

	for i := 0; i < int(msg.Iterations); i++ {
		if msg.FullRead {
			s.GetPet(ctx, id)
		}
		if msg.GroupRead {
			s.GetPetInfo(ctx, id)
		}
		if msg.OwnerRead {
			s.GetPetOwner(ctx, id)
		}
		if msg.NameRead {
			s.GetPetName(ctx, id)
		}
		if msg.ColorRead {
			s.GetPetColor(ctx, id)
		}
		if msg.SpotsRead {
			s.GetPetSpots(ctx, id)
		}
	}

	return &types.MsgReadResponse{}, nil
}

// Writes to the store
func (s msgServer) Write(goCtx context.Context, msg *types.MsgWriteRequest) (*types.MsgWriteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	id := uint64(1)
	pet := s.DefaultPet()

	for i := 0; i < int(msg.Iterations); i++ {
		if msg.FullWrite {
			s.SetPet(ctx, pet)
		}
		if msg.GroupWrite {
			s.SetPetInfo(ctx, id, &pet.PetInfo)
		}
		if msg.OwnerWrite {
			s.SetPetOwner(ctx, id, sdk.AccAddress(pet.Owner))
		}
		if msg.NameWrite {
			s.SetPetName(ctx, id, pet.PetInfo.Name)
		}
		if msg.ColorWrite {
			s.SetPetColor(ctx, id, pet.PetInfo.Color)
		}
		if msg.SpotsWrite {
			s.SetPetSpots(ctx, id, pet.PetInfo.Spots)
		}
	}

	return &types.MsgWriteResponse{}, nil
}

// Read and writes to the store
func (s msgServer) Update(goCtx context.Context, msg *types.MsgUpdateRequest) (*types.MsgUpdateResponse, error) {
	_ = sdk.UnwrapSDKContext(goCtx)

	return &types.MsgUpdateResponse{}, nil
}
