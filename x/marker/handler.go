package marker

import (
	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// NewHandler returns a handler for marker messages.
func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgAddMarkerRequest:
			res, err := msgServer.AddMarker(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgAddAccessRequest:
			res, err := msgServer.AddAccess(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteAccessRequest:
			res, err := msgServer.DeleteAccess(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgFinalizeRequest:
			res, err := msgServer.Finalize(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgActivateRequest:
			res, err := msgServer.Activate(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgCancelRequest:
			res, err := msgServer.Cancel(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDeleteRequest:
			res, err := msgServer.Delete(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgMintRequest:
			res, err := msgServer.Mint(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgBurnRequest:
			res, err := msgServer.Burn(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgWithdrawRequest:
			res, err := msgServer.Withdraw(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgTransferRequest:
			res, err := msgServer.Transfer(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgSetDenomMetadataRequest:
			res, err := msgServer.SetDenomMetadata(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}

func NewProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.AddMarkerProposal:
			return keeper.HandleAddMarkerProposal(ctx, k, c)
		case *types.SupplyIncreaseProposal:
			return keeper.HandleSupplyIncreaseProposal(ctx, k, c)
		case *types.SupplyDecreaseProposal:
			return keeper.HandleSupplyDecreaseProposal(ctx, k, c)
		case *types.SetAdministratorProposal:
			return keeper.HandleSetAdministratorProposal(ctx, k, c)
		case *types.RemoveAdministratorProposal:
			return keeper.HandleRemoveAdministratorProposal(ctx, k, c)
		case *types.ChangeStatusProposal:
			return keeper.HandleChangeStatusProposal(ctx, k, c)
		case *types.WithdrawEscrowProposal:
			return keeper.HandleWithdrawEscrowProposal(ctx, k, c)
		case *types.SetDenomMetadataProposal:
			return keeper.HandleSetDenomMetadataProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized marker proposal content type: %T", c)
		}
	}
}
