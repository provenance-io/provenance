package keeper_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	simapp "github.com/provenance-io/provenance/app"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

var (
	encoding   = simapp.MakeEncodingConfig()
	signerData = authsigning.SignerData{
		ChainID:       "test-chain",
		AccountNumber: 50,
		Sequence:      23,
	}
)

func CreateTxFactory(t *testing.T) tx.Factory {
	requireT := require.New(t)
	path := hd.CreateHDPath(118, 0, 0).String()
	kr, err := keyring.New(t.Name(), "test", t.TempDir(), nil)
	requireT.NoError(err)

	var from1 = "test_key1"
	var from2 = "test_key2"
	var from3 = "test_key3"

	acc1, _, err := kr.NewMnemonic(from1, keyring.English, path, "", hd.Secp256k1)
	requireT.NoError(err)

	acc2, _, err := kr.NewMnemonic(from2, keyring.English, path, "", hd.Secp256k1)
	requireT.NoError(err)

	acc3, _, err := kr.NewMnemonic(from3, keyring.English, path, "", hd.Secp256k1)
	requireT.NoError(err)

	pubKey1 := acc1.GetPubKey()
	pubKey2 := acc2.GetPubKey()
	pubKey3 := acc3.GetPubKey()

	multi := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{pubKey1, pubKey2})
	kr.SaveMultisig("test_multi1", multi)

	requireT.NotEqual(pubKey1.Bytes(), pubKey2.Bytes())
	requireT.NotEqual(pubKey1.Bytes(), pubKey3.Bytes())
	t.Log("Pub keys:", pubKey1, pubKey2, pubKey3)

	return tx.Factory{}.
		WithTxConfig(encoding.TxConfig).
		WithAccountNumber(signerData.AccountNumber).
		WithSequence(signerData.Sequence).
		WithFees("50stake").
		WithMemo("memo").
		WithChainID(signerData.ChainID).
		WithKeybase(kr)

	// Use TxFactory with a signmode WithSignMode(signingtypes.SignMode_SIGN_MODE_DIRECT)
}

func TestValidateRawSingleSignature(t *testing.T) {
	app := simapp.Setup(false)

	txf := CreateTxFactory(t).WithSignMode(signingtypes.SignMode_SIGN_MODE_DIRECT)
	testkey1, err := txf.Keybase().Key("test_key1")
	require.NoError(t, err)

	testkey2, err := txf.Keybase().Key("test_key2")
	require.NoError(t, err)

	s := *types.NewScope(types.ScopeMetadataAddress(uuid.New()), nil, ownerPartyList(testkey1.GetAddress().String()), []string{}, "")
	txb, err := tx.BuildUnsignedTx(txf, types.NewMsgWriteScopeRequest(s, []string{testkey1.GetAddress().String()}))
	require.NoError(t, err)
	require.NotNil(t, txb)

	bytesToSign, err := encoding.TxConfig.SignModeHandler().GetSignBytes(signingtypes.SignMode_SIGN_MODE_DIRECT, signerData, txb.GetTx())

	err = app.MetadataKeeper.CreateRawSignature(txf, "test_key1", txb, bytesToSign, true)
	require.NoError(t, err)

	err = app.MetadataKeeper.CreateRawSignature(txf, "test_key2", txb, bytesToSign, true)
	require.NoError(t, err)

	signedTx := txb.GetTx()
	require.NotNil(t, signedTx)

	sigs, err := signedTx.GetSignaturesV2()
	require.NoError(t, err)
	require.NotEmpty(t, sigs)

	descriptors, err := sigV2ToDescriptors(sigs)
	require.NoError(t, err)
	require.NotEmpty(t, descriptors)
	require.Equal(t, 2, len(sigs))

	addr, err := app.MetadataKeeper.ValidateRawSignature(*descriptors[0], bytesToSign)
	require.NoError(t, err)
	require.EqualValues(t, testkey1.GetAddress(), addr)

	addr, err = app.MetadataKeeper.ValidateRawSignature(*descriptors[1], bytesToSign)
	require.NoError(t, err)
	require.EqualValues(t, testkey2.GetAddress(), addr)
}

func sigV2ToDescriptors(sigs []signingtypes.SignatureV2) ([]*signingtypes.SignatureDescriptor, error) {
	descs := make([]*signingtypes.SignatureDescriptor, len(sigs))
	for i, sig := range sigs {
		descData := signingtypes.SignatureDataToProto(sig.Data)
		any, err := codectypes.NewAnyWithValue(sig.PubKey)
		if err != nil {
			return nil, err
		}

		descs[i] = &signingtypes.SignatureDescriptor{
			PublicKey: any,
			Data:      descData,
			Sequence:  sig.Sequence,
		}
	}
	return descs, nil
}

func TestIsMarkerAndHasAuthority_IsMarker(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	pubkey := secp256k1.GenPrivKey().PubKey()
	userAddr := sdk.AccAddress(pubkey.Address().Bytes())
	user := userAddr.String()
	randomUser := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"

	markerAddr := markertypes.MustGetMarkerAddress("testcoin").String()
	err := app.MarkerKeeper.AddMarkerAccount(ctx, &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       markerAddr,
			AccountNumber: 23,
		},
		AccessControl: []markertypes.AccessGrant{
			{
				Address:     user,
				Permissions: markertypes.AccessListByNames("deposit,withdraw"),
			},
		},
		Denom:      "testcoin",
		Supply:     sdk.NewInt(1000),
		MarkerType: markertypes.MarkerType_Coin,
		Status:     markertypes.StatusActive,
	})
	require.NoError(t, err)
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, userAddr))

	isMarker, _ := app.MetadataKeeper.IsMarkerAndHasAuthority(ctx, markerAddr, []string{}, markertypes.Access_Unknown)
	assert.True(t, isMarker, "is a marker")
	isMarker, _ = app.MetadataKeeper.IsMarkerAndHasAuthority(ctx, user, []string{}, markertypes.Access_Unknown)
	assert.False(t, isMarker, "exists, is a user, not a marker")
	isMarker, _ = app.MetadataKeeper.IsMarkerAndHasAuthority(ctx, randomUser, []string{}, markertypes.Access_Unknown)
	assert.False(t, isMarker, "doesn't exist, not a marker")
	isMarker, _ = app.MetadataKeeper.IsMarkerAndHasAuthority(ctx, "invalid", []string{}, markertypes.Access_Unknown)
	assert.False(t, isMarker, "invalid address not a marker")
}

func TestIsMarkerAndHasAuthority_HasAuth(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	pubkey := secp256k1.GenPrivKey().PubKey()
	userAddr := sdk.AccAddress(pubkey.Address().Bytes())
	user := userAddr.String()
	randomUser := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"

	markerAddr := markertypes.MustGetMarkerAddress("testcoin").String()
	err := app.MarkerKeeper.AddMarkerAccount(ctx, &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       markerAddr,
			AccountNumber: 23,
		},
		AccessControl: []markertypes.AccessGrant{
			{
				Address:     user,
				Permissions: markertypes.AccessListByNames("deposit,withdraw"),
			},
		},
		Denom:      "testcoin",
		Supply:     sdk.NewInt(1000),
		MarkerType: markertypes.MarkerType_Coin,
		Status:     markertypes.StatusActive,
	})
	require.NoError(t, err)
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, userAddr))

	_, hasAuth := app.MetadataKeeper.IsMarkerAndHasAuthority(ctx, "invalid", []string{user}, markertypes.Access_Deposit)
	assert.False(t, hasAuth, "invalid value owner")

	_, hasAuth = app.MetadataKeeper.IsMarkerAndHasAuthority(ctx, randomUser, []string{user}, markertypes.Access_Deposit)
	assert.False(t, hasAuth, "value owner doesn't exist")

	_, hasAuth = app.MetadataKeeper.IsMarkerAndHasAuthority(ctx, user, []string{user}, markertypes.Access_Deposit)
	assert.False(t, hasAuth, "user isn't a marker")

	_, hasAuth = app.MetadataKeeper.IsMarkerAndHasAuthority(ctx, markerAddr, []string{user}, markertypes.Access_Deposit)
	assert.True(t, hasAuth, "user has access")

	_, hasAuth = app.MetadataKeeper.IsMarkerAndHasAuthority(ctx, markerAddr, []string{"invalidaddress", user}, markertypes.Access_Deposit)
	assert.True(t, hasAuth, "user has access even with invalid signer")

	_, hasAuth = app.MetadataKeeper.IsMarkerAndHasAuthority(ctx, markerAddr, []string{user}, markertypes.Access_Burn)
	assert.False(t, hasAuth, "user doesn't have this access")
}
