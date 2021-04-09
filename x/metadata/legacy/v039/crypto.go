package v039

import (
	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcrypt "github.com/tendermint/tendermint/crypto"
	tmcurve "github.com/tendermint/tendermint/crypto/secp256k1"
)

// RecoverPublicKey recovers a tendermint secp256k1 public key from a signtaure and message hash.
func RecoverPublicKey(sig, hash []byte) (tmcrypt.PubKey, sdk.AccAddress, error) {
	// Recover public key
	pubKey, _, err := btcec.RecoverCompact(btcec.S256(), sig, hash)
	if err != nil {
		return nil, nil, err
	}
	// Create tendermint public key type and return with address.
	tmKey := tmcurve.PubKey{} // .PubKeySecp256k1{}
	copy(tmKey[:], pubKey.SerializeCompressed())
	return tmKey, tmKey.Address().Bytes(), nil
}

// ParsePublicKey parses a secp256k1 public key, calculates the account address, and returns both.
func ParsePublicKey(data []byte) (tmcrypt.PubKey, sdk.AccAddress, error) {
	// Parse the secp256k1 public key.
	pk, err := btcec.ParsePubKey(data, btcec.S256())
	if err != nil {
		return nil, nil, err
	}
	// Create tendermint public key type and return with address.
	tmKey := tmcurve.PubKey(pk.SerializeCompressed()) // PubKeySecp256k1{}

	return tmKey, tmKey.Address().Bytes(), nil
}
