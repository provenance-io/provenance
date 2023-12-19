package metadata

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

// NewHandler returns a handler for metadata messages.
// TODO[1760]: metadata: Delete the metadata NewHandler.
func NewHandler(k keeper.Keeper) func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgWriteScopeRequest:
			res, err := msgServer.WriteScope(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteScopeRequest:
			res, err := msgServer.DeleteScope(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgAddScopeDataAccessRequest:
			res, err := msgServer.AddScopeDataAccess(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteScopeDataAccessRequest:
			res, err := msgServer.DeleteScopeDataAccess(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgAddScopeOwnerRequest:
			res, err := msgServer.AddScopeOwner(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteScopeOwnerRequest:
			res, err := msgServer.DeleteScopeOwner(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgUpdateValueOwnersRequest:
			res, err := msgServer.UpdateValueOwners(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgMigrateValueOwnerRequest:
			res, err := msgServer.MigrateValueOwner(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgWriteRecordRequest:
			res, err := msgServer.WriteRecord(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteRecordRequest:
			res, err := msgServer.DeleteRecord(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgWriteSessionRequest:
			res, err := msgServer.WriteSession(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgWriteScopeSpecificationRequest:
			res, err := msgServer.WriteScopeSpecification(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteScopeSpecificationRequest:
			res, err := msgServer.DeleteScopeSpecification(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgWriteContractSpecificationRequest:
			res, err := msgServer.WriteContractSpecification(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteContractSpecificationRequest:
			res, err := msgServer.DeleteContractSpecification(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgAddContractSpecToScopeSpecRequest:
			res, err := msgServer.AddContractSpecToScopeSpec(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteContractSpecFromScopeSpecRequest:
			res, err := msgServer.DeleteContractSpecFromScopeSpec(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgWriteRecordSpecificationRequest:
			res, err := msgServer.WriteRecordSpecification(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteRecordSpecificationRequest:
			res, err := msgServer.DeleteRecordSpecification(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgBindOSLocatorRequest:
			res, err := msgServer.BindOSLocator(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteOSLocatorRequest:
			res, err := msgServer.DeleteOSLocator(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgModifyOSLocatorRequest:
			res, err := msgServer.ModifyOSLocator(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgSetAccountDataRequest:
			res, err := msgServer.SetAccountData(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf("unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
