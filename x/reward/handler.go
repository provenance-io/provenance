package reward

// func NewProposalHandler(k keeper.Keeper, registry cdctypes.InterfaceRegistry) govtypes.Handler {
// 	return func(ctx sdk.Context, content govtypes.Content) error {
// 		switch c := content.(type) {
// 		case *types.AddRewardProgramProposal:
// 			return keeper.HandleAddRewardProgramProposal(ctx, k, c, registry)
// 		default:
// 			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized reward proposal content type: %T", c)
// 		}
// 	}
// }
