package keeper_test

import (
	"encoding/binary"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/provenance-io/provenance/app"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	"github.com/provenance-io/provenance/x/name/keeper"
	"github.com/provenance-io/provenance/x/name/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestDeleteNameRemovingAttributeAccounts(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	msgServer := keeper.NewMsgServerImpl(app.NameKeeper)
	name := "test.io"
	owner1 := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"
	ownerAdd := sdk.MustAccAddressFromBech32(owner1)
	acct1 := app.AccountKeeper.NewAccountWithAddress(ctx, ownerAdd)
	app.AccountKeeper.SetAccount(ctx, acct1)
	require.NoError(t, app.NameKeeper.SetNameRecord(ctx, name, ownerAdd, false))
	attrAccounts := make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		attrAccounts[i] = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		require.NoError(t, app.AttributeKeeper.SetAttribute(ctx, attrtypes.NewAttribute(name, attrAccounts[i].String(), attrtypes.AttributeType_String, []byte(attrAccounts[i].String())), ownerAdd))
		attrStore := ctx.KVStore(app.GetKey(attrtypes.StoreKey))
		key := attrtypes.AttributeNameAddrKeyPrefix(name, attrAccounts[i])
		address, _ := attrtypes.GetAddressFromKey(key)
		bz := attrStore.Get(key)
		assert.Equal(t, attrAccounts[i], address)
		assert.Equal(t, uint64(1), binary.BigEndian.Uint64(bz))

	}
	attrAddresses, err := app.AttributeKeeper.AccountsByAttribute(ctx, name)
	assert.NoError(t, err)
	assert.ElementsMatch(t, attrAccounts, attrAddresses)

	result, err := msgServer.DeleteName(ctx, types.NewMsgDeleteNameRequest(types.NewNameRecord("test.io", ownerAdd, false)))
	assert.NotNil(t, result)
	assert.NoError(t, err)

	attrAddresses, err = app.AttributeKeeper.AccountsByAttribute(ctx, name)
	assert.NoError(t, err)
	assert.Len(t, attrAddresses, 0)

	for i := 0; i < 10; i++ {
		attrStore := ctx.KVStore(app.GetKey(attrtypes.StoreKey))
		key := attrtypes.AttributeNameAddrKeyPrefix(name, attrAccounts[i])
		bz := attrStore.Get(key)
		assert.Nil(t, bz)

	}
}
