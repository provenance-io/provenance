package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/provenance-io/provenance/app"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/name/keeper"
	"github.com/provenance-io/provenance/x/name/types"
	"github.com/stretchr/testify/assert"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestDeleteName(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	msgServer := keeper.NewMsgServerImpl(app.NameKeeper)
	owner1 := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	ownerAdd := sdk.MustAccAddressFromBech32(owner1)
	app.NameKeeper.SetNameRecord(ctx, "test.io", ownerAdd, false)
	app.AttributeKeeper.SetAttribute(ctx, attrtypes.NewAttribute("test.io", owner1, attrtypes.AttributeType_Bytes, []byte("1")), ownerAdd)

	result, err := msgServer.DeleteName(ctx, types.NewMsgDeleteNameRequest(types.NewNameRecord("test.io", ownerAdd, false)))
	assert.Nil(t, result)
	assert.Error(t, err)
}
