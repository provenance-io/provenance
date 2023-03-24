package keeper_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	simapp "github.com/provenance-io/provenance/app"
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
	kr, err := keyring.New(t.Name(), "test", t.TempDir(), nil, encoding.Marshaler)
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

	pubKey1, err := acc1.GetPubKey()
	requireT.NotEqual(err, "getting acc1 pub key")
	pubKey2, err := acc2.GetPubKey()
	requireT.NotEqual(err, "getting acc2 pub key")
	pubKey3, err := acc3.GetPubKey()
	requireT.NotEqual(err, "getting acc3 pub key")

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
	app := simapp.Setup(t)

	txf := CreateTxFactory(t).WithSignMode(signingtypes.SignMode_SIGN_MODE_DIRECT)
	testkey1, err := txf.Keybase().Key("test_key1")
	require.NoError(t, err)
	testkey1addr, err := testkey1.GetAddress()
	require.NoError(t, err, "getting test_key1 address")

	testkey2, err := txf.Keybase().Key("test_key2")
	require.NoError(t, err)
	testkey2addr, err := testkey2.GetAddress()
	require.NoError(t, err, "getting test_key2 address")

	s := *types.NewScope(types.ScopeMetadataAddress(uuid.New()), nil, ownerPartyList(testkey1addr.String()), []string{}, "")
	txb, err := txf.BuildUnsignedTx(types.NewMsgWriteScopeRequest(s, []string{testkey1addr.String()}))
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
	require.EqualValues(t, testkey1addr, addr)

	addr, err = app.MetadataKeeper.ValidateRawSignature(*descriptors[1], bytesToSign)
	require.NoError(t, err)
	require.EqualValues(t, testkey2addr, addr)
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
