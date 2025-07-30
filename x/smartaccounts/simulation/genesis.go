package simulation

import (
	"encoding/base64"
	"math/rand"
	"strconv"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/x/smartaccounts/types"
)

// Simulation parameter constants
var (
	MaxCredentialAllowed = "max_credential_allowed" // #nosec G101
	Enabled              = "enabled"
)

// GenMaxCredentialAllowed randomized MaxCredentialAllowed
func GenMaxCredentialAllowed(r *rand.Rand) uint32 {
	// The result of r.Intn(20) + 1 is always between 1 and 20, which safely fits in a uint32.
	return uint32(r.Int31n(20) + 1)
}

// GenEnabled returns a randomized Enabled parameter
func GenEnabled(r *rand.Rand) bool {
	return r.Intn(100) < 90 // 90% chance of being enabled
}

// RandomizedGenState generates a random GenesisState for the smartaccounts module
func RandomizedGenState(simState *module.SimulationState) {
	var maxCredentialAllowed uint32
	simState.AppParams.GetOrGenerate(
		MaxCredentialAllowed,
		&maxCredentialAllowed,
		simState.Rand,
		func(r *rand.Rand) { maxCredentialAllowed = GenMaxCredentialAllowed(r) },
	)

	var enabled bool
	simState.AppParams.GetOrGenerate(
		Enabled,
		&enabled,
		simState.Rand,
		func(r *rand.Rand) { enabled = GenEnabled(r) },
	)

	// Create a ProvenanceAccount struct directly instead of using JSON
	// Generate random public key for account
	pubKey, _ := codectypes.NewAnyWithValue(GenRandomSecp256k1PubKey())
	randAccount := simState.Accounts[simState.Rand.Intn(len(simState.Accounts))]
	baseAccount := authtypes.BaseAccount{
		Address:       randAccount.Address.String(),
		PubKey:        pubKey,
		AccountNumber: GetUniqueAccountNumber(simState.Rand),
		Sequence:      simState.Rand.Uint64(),
	}

	// Create multiple credentials (1-3)
	numCredentials := simState.Rand.Intn(3) + 1
	credentials := make([]*types.Credential, numCredentials)

	for i := 0; i < numCredentials; i++ {
		// Randomly choose credential type
		variant := simState.Rand.Intn(2)

		if variant == 0 {
			// Create K256 credential
			credPubKey, _ := codectypes.NewAnyWithValue(GenRandomSecp256k1PubKey())
			credentials[i] = &types.Credential{
				BaseCredential: &types.BaseCredential{
					CredentialNumber: uint64(i),
					PublicKey:        credPubKey,
					Variant:          types.CredentialType_CREDENTIAL_TYPE_K256,
					CreateTime:       simState.Rand.Int63n(time.Now().Unix()),
				},
			}
		} else {
			// Create WebAuthn credential with EC2 public key
			ec2PubKey := GenRandomEC2PublicKey()
			ec2PubKeyAny, _ := codectypes.NewAnyWithValue(ec2PubKey)

			credentials[i] = &types.Credential{
				BaseCredential: &types.BaseCredential{
					CredentialNumber: uint64(i),
					PublicKey:        ec2PubKeyAny,
					Variant:          types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN_UV,
					CreateTime:       simtypes.RandTimestamp(simState.Rand).Unix(),
				},
				Authenticator: &types.Credential_Fido2Authenticator{
					Fido2Authenticator: GenRandomFido2Authenticator(simState.Rand),
				},
			}
		}
	}

	// Create a simple K256 credential
	credPubKey, _ := codectypes.NewAnyWithValue(GenRandomSecp256k1PubKey())

	credential := types.Credential{
		BaseCredential: &types.BaseCredential{
			CredentialNumber: 1,
			PublicKey:        credPubKey,
			Variant:          types.CredentialType_CREDENTIAL_TYPE_K256,
			CreateTime:       simState.Rand.Int63n(time.Now().Unix()-172800) + 172800, // Random time in last 2 days
		},
	}

	testAccount := types.ProvenanceAccount{
		BaseAccount:                      &baseAccount,
		SmartAccountNumber:               simState.Rand.Uint64(),
		Credentials:                      []*types.Credential{&credential},
		IsSmartAccountOnlyAuthentication: false,
	}

	smartAccountsGenesis := types.GenesisState{
		Params: types.Params{
			Enabled:              enabled,
			MaxCredentialAllowed: maxCredentialAllowed,
		},
		Accounts: []types.ProvenanceAccount{testAccount}, // Start with no accounts in genesis
	}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&smartAccountsGenesis)
}

// GenRandomSecp256k1PubKey generates a random secp256k1 public key for simulation purposes
func GenRandomSecp256k1PubKey() *secp256k1.PubKey {
	privKey := secp256k1.GenPrivKey()
	return privKey.PubKey().(*secp256k1.PubKey)
}

// Used to track generated account numbers to avoid collisions
var generatedAccountNumbers = make(map[uint64]bool)

// GetUniqueAccountNumber generates a guaranteed unique account number
func GetUniqueAccountNumber(r *rand.Rand) uint64 {
	var accountNum uint64

	// Keep generating numbers until we get a unique one
	for {
		accountNum = r.Uint64()
		if !generatedAccountNumbers[accountNum] {
			// Found a unique account number
			generatedAccountNumbers[accountNum] = true
			break
		}
	}

	return accountNum
}

// GenRandomFido2Authenticator generates random FIDO2 authenticator data
func GenRandomFido2Authenticator(r *rand.Rand) *types.Fido2Authenticator {
	// Generate random ID (16 bytes)
	idBytes := make([]byte, 16)
	r.Read(idBytes)
	id := base64.RawURLEncoding.EncodeToString(idBytes)

	// Generate random AAGUID (16 bytes)
	aaguidBytes := make([]byte, 16)
	r.Read(aaguidBytes)
	aaguid := base64.RawURLEncoding.EncodeToString(aaguidBytes)

	// Generate fake response (would normally be a complex structure)
	responseBytes := make([]byte, 512)
	r.Read(responseBytes)
	responseB64 := base64.RawURLEncoding.EncodeToString(responseBytes)

	return &types.Fido2Authenticator{
		Id:                         id,
		Username:                   "user" + strconv.Itoa(r.Intn(1000)),
		Aaguid:                     []byte(aaguid),
		CredentialCreationResponse: responseB64,
		RpId:                       "localhost",
		RpOrigin:                   "http://localhost:18080",
	}
}

// GenRandomEC2PublicKey generates random EC2 public key data for WebAuthn
func GenRandomEC2PublicKey() *types.EC2PublicKeyData {
	// Use cosmos/crypto/secp256r1 to generate a proper P-256 key
	privKey, err := secp256r1.GenPrivKey()
	if err != nil {
		panic("failed to generate secp256r1 private key: " + err.Error())
	}
	pubKey := privKey.PubKey().(*secp256r1.PubKey)

	// Get the ECDSA public key
	ecdsaPub := pubKey.Key

	// Extract the X and Y coordinates
	xCoord := make([]byte, 32)
	yCoord := make([]byte, 32)

	// Copy X coordinate with proper padding
	xBytes := ecdsaPub.X.Bytes()
	copy(xCoord[32-len(xBytes):], xBytes)

	// Copy Y coordinate with proper padding
	yBytes := ecdsaPub.Y.Bytes()
	copy(yCoord[32-len(yBytes):], yBytes)

	// Create the raw public key in uncompressed format (0x04 + X + Y)
	// This is only for the test, never do such calculation in production, without using a trusted library.
	rawPubKey := make([]byte, 65)
	rawPubKey[0] = 0x04 // Uncompressed point format
	copy(rawPubKey[1:33], xCoord)
	copy(rawPubKey[33:], yCoord)

	return &types.EC2PublicKeyData{
		PublicKeyData: &types.PublicKeyData{
			PublicKey: rawPubKey,
			KeyType:   2,  // EC2 key type
			Algorithm: -7, // ES256 algorithm
		},
		Curve:  1, // P-256
		XCoord: xCoord,
		YCoord: yCoord,
	}
}
