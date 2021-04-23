package simulation

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strings"

	"github.com/provenance-io/provenance/x/authz/msgservice"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/provenance-io/provenance/x/authz/keeper"
	"github.com/provenance-io/provenance/x/authz/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// authz message types
const (
	TypeMsgGrantAuthorization  = "/provenance.authz.v1.Msg/GrantAuthorization"
	TypeMsgRevokeAuthorization = "/provenance.authz.v1.Msg/RevokeAuthorization"
	TypeMsgExecDelegated       = "/provenance.authz.v1.Msg/ExecAuthorized"
)

// Simulation operation weights constants
const (
	OpWeightMsgGrantAuthorization = "op_weight_msg_grant_authorization"
	OpWeightRevokeAuthorization   = "op_weight_msg_revoke_authorization"
	OpWeightExecAuthorized        = "op_weight_msg_execute_authorized"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, appCdc cdctypes.AnyUnpacker, protoCdc *codec.ProtoCodec) simulation.WeightedOperations {
	var (
		weightMsgGrantAuthorization int
		weightRevokeAuthorization   int
		weightExecAuthorized        int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgGrantAuthorization, &weightMsgGrantAuthorization, nil,
		func(_ *rand.Rand) {
			weightMsgGrantAuthorization = simappparams.DefaultWeightMsgDelegate
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightRevokeAuthorization, &weightRevokeAuthorization, nil,
		func(_ *rand.Rand) {
			weightRevokeAuthorization = simappparams.DefaultWeightMsgUndelegate
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightExecAuthorized, &weightExecAuthorized, nil,
		func(_ *rand.Rand) {
			weightExecAuthorized = simappparams.DefaultWeightMsgSend
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgGrantAuthorization,
			SimulateMsgGrantAuthorization(ak, bk, k, protoCdc),
		),
		simulation.NewWeightedOperation(
			weightRevokeAuthorization,
			SimulateMsgRevokeAuthorization(ak, bk, k, protoCdc),
		),
		simulation.NewWeightedOperation(
			weightExecAuthorized,
			SimulateMsgExecuteAuthorized(ak, bk, k, appCdc, protoCdc),
		),
	}
}

// SimulateMsgGrantAuthorization generates a MsgGrantAuthorization with random values.
// nolint: funlen
func SimulateMsgGrantAuthorization(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
	protoCdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		granter := accs[0]
		grantee := accs[1]

		account := ak.GetAccount(ctx, granter.Address)

		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantAuthorization, err.Error()), nil, err
		}

		blockTime := ctx.BlockTime()
		spendLimit := spendableCoins.Sub(fees)
		if spendLimit == nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantAuthorization, "spend limit is nil"), nil, nil
		}
		msg, err := types.NewMsgGrantAuthorization(granter.Address, grantee.Address,
			markertypes.NewSendAuthorization(spendLimit), blockTime.AddDate(1, 0, 0))

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantAuthorization, err.Error()), nil, err
		}
		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		authzMsgClient := types.NewMsgClient(svcMsgClientConn)
		_, err = authzMsgClient.GrantAuthorization(context.Background(), msg)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantAuthorization, err.Error()), nil, err
		}
		tx, err := helpers.GenTx(
			txGen,
			svcMsgClientConn.GetMsgs(),
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			granter.PrivKey,
		)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantAuthorization, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, svcMsgClientConn.GetMsgs()[0].Type(), "unable to deliver tx"), nil, err
		}
		return NewOperationMsg(svcMsgClientConn.GetMsgs()[0], true, "", protoCdc), nil, err
	}
}

// SimulateMsgRevokeAuthorization generates a MsgRevokeAuthorization with random values.
// nolint: funlen
func SimulateMsgRevokeAuthorization(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, protoCdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		hasGrant := false
		var targetGrant types.AuthorizationGrant
		var granterAddr sdk.AccAddress
		var granteeAddr sdk.AccAddress
		k.IterateGrants(ctx, func(granter, grantee sdk.AccAddress, grant types.AuthorizationGrant) bool {
			targetGrant = grant
			granterAddr = granter
			granteeAddr = grantee
			hasGrant = true
			return true
		})

		if !hasGrant {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeAuthorization, "no grants"), nil, nil
		}

		granter, ok := simtypes.FindAccount(accs, granterAddr)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeAuthorization, "Account not found"), nil, nil
		}
		account := ak.GetAccount(ctx, granter.Address)

		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeAuthorization, "fee error"), nil, err
		}
		auth := targetGrant.GetAuthorizationGrant()
		msg := types.NewMsgRevokeAuthorization(granterAddr, granteeAddr, auth.MethodName())

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		authzMsgClient := types.NewMsgClient(svcMsgClientConn)
		_, err = authzMsgClient.RevokeAuthorization(context.Background(), &msg)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeAuthorization, err.Error()), nil, err
		}
		tx, err := helpers.GenTx(
			txGen,
			svcMsgClientConn.GetMsgs(),
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			granter.PrivKey,
		)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeAuthorization, err.Error()), nil, err
		}

		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		return NewOperationMsg(svcMsgClientConn.GetMsgs()[0], true, "", protoCdc), nil, err
	}
}

// SimulateMsgExecuteAuthorized generates a MsgExecuteAuthorized with random values.
// nolint: funlen
func SimulateMsgExecuteAuthorized(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, cdc cdctypes.AnyUnpacker, protoCdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		hasGrant := false
		var targetGrant types.AuthorizationGrant
		var granterAddr sdk.AccAddress
		var granteeAddr sdk.AccAddress
		k.IterateGrants(ctx, func(granter, grantee sdk.AccAddress, grant types.AuthorizationGrant) bool {
			targetGrant = grant
			granterAddr = granter
			granteeAddr = grantee
			hasGrant = true
			return true
		})

		if !hasGrant {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecDelegated, "Not found"), nil, nil
		}

		grantee, _ := simtypes.FindAccount(accs, granteeAddr)
		granterAccount := ak.GetAccount(ctx, granterAddr)
		granteeAccount := ak.GetAccount(ctx, granteeAddr)

		granterspendableCoins := bk.SpendableCoins(ctx, granterAccount.GetAddress())
		if granterspendableCoins.Empty() {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecDelegated, "no coins"), nil, nil
		}

		if targetGrant.Expiration.Before(ctx.BlockHeader().Time) {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecDelegated, "grant expired"), nil, nil
		}

		granteespendableCoins := bk.SpendableCoins(ctx, granteeAccount.GetAddress())
		fees, err := simtypes.RandomFees(r, ctx, granteespendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecDelegated, "fee error"), nil, err
		}
		sendCoins := sdk.NewCoin("foo", sdk.NewInt(10))

		execMsg := sdk.ServiceMsg{
			MethodName: markertypes.SendAuthorization{}.MethodName(),
			Request: markertypes.NewTransferRequest(
				granterAddr,
				granterAddr,
				granteeAddr,
				sendCoins,
			),
		}

		msg := types.NewMsgExecAuthorized(grantee.Address, []sdk.ServiceMsg{execMsg})
		sendGrant := targetGrant.Authorization.GetCachedValue().(*markertypes.SendAuthorization)
		_, _, err = sendGrant.Accept(ctx, execMsg)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecDelegated, err.Error()), nil, nil
		}

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		authzMsgClient := types.NewMsgClient(svcMsgClientConn)
		_, err = authzMsgClient.ExecAuthorized(context.Background(), &msg)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecDelegated, err.Error()), nil, err
		}
		tx, err := helpers.GenTx(
			txGen,
			svcMsgClientConn.GetMsgs(),
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{granteeAccount.GetAccountNumber()},
			[]uint64{granteeAccount.GetSequence()},
			grantee.PrivKey,
		)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecDelegated, err.Error()), nil, err
		}
		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		if err != nil {
			if strings.Contains(err.Error(), "insufficient fee") {
				return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecDelegated, "insufficient fee"), nil, nil
			}
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecDelegated, err.Error()), nil, err
		}
		msg.UnpackInterfaces(cdc)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgExecDelegated, "unmarshal error"), nil, err
		}
		return NewOperationMsg(svcMsgClientConn.GetMsgs()[0], true, "success", protoCdc), nil, nil
	}
}

// NewOperationMsg - create a new operation message from sdk.Msg
func NewOperationMsg(msg sdk.Msg, ok bool, comment string, cdc *codec.ProtoCodec) simtypes.OperationMsg {
	if reflect.TypeOf(msg) == reflect.TypeOf(sdk.ServiceMsg{}) {
		srvMsg, ok := msg.(sdk.ServiceMsg)
		if !ok {
			panic(fmt.Sprintf("Expecting %T to implement sdk.ServiceMsg", msg))
		}
		bz := cdc.MustMarshalJSON(srvMsg.Request)

		return simtypes.NewOperationMsgBasic(srvMsg.MethodName, srvMsg.MethodName, comment, ok, bz)
	}
	return simtypes.NewOperationMsgBasic(msg.Route(), msg.Type(), comment, ok, msg.GetSignBytes())
}
