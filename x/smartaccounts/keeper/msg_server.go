package keeper

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/go-webauthn/webauthn/protocol"

	errorsmod "cosmossdk.io/errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/smartaccounts/types"
)

// compile-time check
var _ types.MsgServer = MsgServer{}

type MsgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &MsgServer{keeper: keeper}
}

func (m MsgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if m.keeper.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.keeper.authority, msg.Authority)
	}
	return nil, m.keeper.SmartAccountParams.Set(ctx, msg.Params)
}

func (m MsgServer) RegisterFido2Credential(ctx context.Context, msg *types.MsgRegisterFido2Credential) (*types.MsgRegisterFido2CredentialResponse, error) {
	if enabled, err := m.keeper.IsSmartAccountsEnabled(ctx); err != nil || !enabled {
		return nil, types.ErrSmartAccountsNotEnabled
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k := m.keeper
	ak := k.AccountKeeper
	creator, err := k.addressCodec.StringToBytes(msg.Sender)
	if err != nil {
		return nil, err
	}
	baseAcc := ak.GetAccount(sdkCtx, creator)
	if baseAcc == nil {
		// this should never happen, since the signer is sender at least for the first time
		return nil, fmt.Errorf("base account does not exist")
	}
	// Step 2: Parse the attestation object
	// see // https://w3c.github.io/webauthn/#iface-pkcredential
	response, err := ParseAndValidateAttestation(msg)
	if err != nil {
		return nil, err
	}
	resp := &types.MsgRegisterFido2CredentialResponse{}
	attestationPubKey := response.Response.AttestationObject.AuthData.AttData.CredentialPublicKey

	anyPubKey, err := types.ParsePublicKey(attestationPubKey)
	if err != nil {
		return nil, fmt.Errorf("parse public key failed %w", err)
	}
	// check rp id is same
	// check and set the rpid
	rpid, err := url.Parse(response.Response.CollectedClientData.Origin)
	if err != nil {
		return resp, fmt.Errorf("failed to parse rp id: %w", err)
	}
	rpIDHash := sha256.Sum256([]byte(rpid.Hostname()))
	if !bytes.Equal(rpIDHash[:], response.Response.AttestationObject.AuthData.RPIDHash) {
		return resp, fmt.Errorf("rp id hash mismatch")
	}

	// Check User Present flag first
	if (response.Response.AttestationObject.AuthData.Flags & 0x01) != 0x01 {
		return nil, fmt.Errorf("user presence flag not set")
	}

	var credentialType types.CredentialType
	// In MakeNewFidoCredential or RegisterWebAuthnCredential
	if (response.Response.AttestationObject.AuthData.Flags & 0x04) == 0x04 {
		// UV flag is set
		credentialType = types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN_UV
	} else {
		// UV flag is not set
		credentialType = types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN
	}
	// Create base credential with only the base fields
	baseCredential := &types.BaseCredential{
		PublicKey:  anyPubKey,
		Variant:    credentialType,
		CreateTime: sdkCtx.BlockTime().Unix(),
	}

	// Then create a FidoCredential
	fido2Authenticator := &types.Fido2Authenticator{
		Id: response.ID,
		// this is the user id used by the authenticator to prompt the user for a signature.
		// Where Is the User ID?
		// The user ID is provided during registration inside the PublicKeyCredentialCreationOptions.User.ID.
		// However, it is NOT returned in the CredentialCreationResponse after registration.
		// You must store the user ID separately when initiating registration and retrieve it later.
		Username:                   msg.UserIdentifier,
		Aaguid:                     response.Response.AttestationObject.AuthData.AttData.AAGUID,
		CredentialCreationResponse: msg.EncodedAttestation,
		RpId:                       rpid.Hostname(),
		RpOrigin:                   response.Response.CollectedClientData.Origin,
	}
	credential := &types.Credential{
		BaseCredential: baseCredential,
		Authenticator: &types.Credential_Fido2Authenticator{
			Fido2Authenticator: fido2Authenticator,
		},
	}
	existingAcc, lookupErr := k.LookupAccountByAddress(sdkCtx, creator)
	// smart account doesn't exists
	if lookupErr != nil && errors.Is(lookupErr, types.ErrSmartAccountDoesNotExist) {
		hasDuplicate := HasDuplicateCredentialID(existingAcc.Credentials, fido2Authenticator.Id)
		if hasDuplicate {
			return nil, fmt.Errorf("credential already exists")
		}
		smartAccount, initErr := k.Init(sdkCtx, &types.MsgInit{
			Sender:      msg.Sender,
			Credentials: []*types.Credential{credential},
		})
		if initErr != nil {
			return nil, initErr
		}

		resp = &types.MsgRegisterFido2CredentialResponse{
			CredentialNumber:  GetAddedCredentialNumber(smartAccount.Credentials, fido2Authenticator.Id),
			ProvenanceAccount: smartAccount,
		}

		// After successful initialization, emit event
		numCreds := len(smartAccount.Credentials)
		errFromEventManager := sdkCtx.EventManager().EmitTypedEvent(
			types.NewEventSmartAccountInit(smartAccount.Address, uint64(numCreds)),
		)
		if errFromEventManager != nil {
			return nil, errFromEventManager
		}
		// After successful registration, emit typed event
		errFromEventManager = sdkCtx.EventManager().EmitTypedEvent(
			&types.EventFido2CredentialAdd{
				Address:          msg.Sender,
				CredentialNumber: 0,
				CredentialId:     fido2Authenticator.Id,
			},
		)
		if errFromEventManager != nil {
			return nil, errFromEventManager
		}
		return resp, nil
	}

	// some other error
	if lookupErr != nil {
		return nil, lookupErr
	}

	// account exists
	// check for duplicate credential
	hasDuplicate := HasDuplicateCredentialID(existingAcc.Credentials, fido2Authenticator.Id)
	if hasDuplicate {
		return nil, errorsmod.Wrap(types.ErrDuplicateCredential, "credential already exists")
	}
	smartAccount, credentialNumberAdded, errorAddingWebauthn := k.AddSmartAccountCredential(sdkCtx, credential, &existingAcc)
	if errorAddingWebauthn != nil {
		return nil, errorAddingWebauthn
	}
	resp = &types.MsgRegisterFido2CredentialResponse{
		CredentialNumber:  credentialNumberAdded,
		ProvenanceAccount: smartAccount,
	}

	// After successful registration, emit typed event
	err = sdkCtx.EventManager().EmitTypedEvent(
		types.NewEventFido2CredentialAdd(msg.Sender, credentialNumberAdded, fido2Authenticator.Id),
	)
	if err != nil {
		return nil, err
	}
	return resp, err
}

// HasDuplicateCredentialID checks if a credential ID already exists in the provided list of credentials
func HasDuplicateCredentialID(credentials []*types.Credential, credentialID string) bool {
	for _, cred := range credentials {
		if cred.GetVariant() == types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN || cred.GetVariant() == types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN_UV {
			fidoCredentialWrapper, ok := cred.GetAuthenticator().(*types.Credential_Fido2Authenticator)
			if !ok {
				continue
			}
			if fidoCredentialWrapper.Fido2Authenticator.Id == credentialID {
				return true
			}
		}
	}
	return false
}

func GetAddedCredentialNumber(credentials []*types.Credential, credentialID string) uint64 {
	for _, cred := range credentials {
		if cred.GetVariant() == types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN || cred.GetVariant() == types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN_UV {
			fidoCredentialWrapper, ok := cred.GetAuthenticator().(*types.Credential_Fido2Authenticator)
			if !ok {
				continue
			}
			if fidoCredentialWrapper.Fido2Authenticator.Id == credentialID {
				return cred.CredentialNumber
			}
		}
	}
	return 0
}

func ParseAndValidateAttestation(msg *types.MsgRegisterFido2Credential) (*protocol.ParsedCredentialCreationData, error) {
	rawAttestationObject, err := base64.RawURLEncoding.DecodeString(msg.EncodedAttestation)
	if err != nil {
		return nil, types.ErrParseCredential.Wrapf("failed to decode base64 client data JSON: %v", err)
	}
	body := io.NopCloser(bytes.NewReader(rawAttestationObject))
	// Generating an Attestation Object
	// attestationFormat
	// An attestation statement format.
	//	authData
	// A byte array containing authenticator data.
	//	hash
	// The hash of the serialized client data.
	//	the authenticator MUST:
	// Let attStmt be the result of running attestationFormat ’s signing procedure given authData and hash .
	//	Let fmt be attestationFormat ’s attestation statement format identifier
	// Return the attestation object as a CBOR map with the following syntax, filled in with variables initialized by this algorithm:
	// 1
	// attObj = { authData: bytes, $$attStmtType } attStmtTemplate = ( fmt: text, attStmt: { * tstr => any } ; Map is filled in by each concrete attStmtType ) ; Every attestation statement format must have the above fields attStmtTemplate .within $$attStmtType
	credentialCreationResp, err := protocol.ParseCredentialCreationResponseBody(body)
	if err != nil {
		return nil, types.ErrParseCredential.Wrapf("failed to parse credential creation response: %v", err)
	}

	if len(credentialCreationResp.Raw.ID) == 0 {
		return nil, types.ErrParseCredential.Wrapf("Credential ID is empty")
	}
	if len(credentialCreationResp.Raw.AttestationResponse.AttestationObject) == 0 {
		return nil, types.ErrParseCredential.Wrapf("Attestation object is empty")
	}

	if len(credentialCreationResp.Response.AttestationObject.Format) == 0 {
		return nil, types.ErrParseCredential.Wrapf("Attestation format is empty")
	}

	if len(credentialCreationResp.ParsedPublicKeyCredential.AuthenticatorAttachment) == 0 {
		// has to be `platform` or `cross-platform`
		return nil, types.ErrParseCredential.Wrapf("Authenticator attachment is empty")
	}
	return credentialCreationResp, nil
}

func (m MsgServer) RegisterCosmosCredential(ctx context.Context, msg *types.MsgRegisterCosmosCredential) (*types.MsgRegisterCosmosCredentialResponse, error) {
	if enabled, err := m.keeper.IsSmartAccountsEnabled(ctx); err != nil || !enabled {
		return nil, types.ErrSmartAccountsNotEnabled
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	// Convert the sender address string to AccAddress
	addr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	// Check if the smart account exists
	smartAccount, err := m.keeper.LookupAccountByAddress(ctx, addr)

	if err != nil {
		if errors.Is(err, types.ErrSmartAccountDoesNotExist) {
			// If account doesn't exist, initialize a new smart account with this credential
			baseCredential := &types.BaseCredential{
				PublicKey:  msg.Pubkey,
				Variant:    types.CredentialType_CREDENTIAL_TYPE_K256,
				CreateTime: sdkCtx.BlockTime().Unix(),
			}

			credential := &types.Credential{
				BaseCredential: baseCredential,
			}

			initMsg := &types.MsgInit{
				Sender:      msg.Sender,
				Credentials: []*types.Credential{credential},
			}

			smartAccountResponse, errInit := m.keeper.Init(ctx, initMsg)

			if errInit != nil {
				return nil, errInit
			}

			// After successful initialization, emit event
			numCreds := len(smartAccountResponse.Credentials)
			err = sdkCtx.EventManager().EmitTypedEvent(
				types.NewEventSmartAccountInit(smartAccountResponse.Address, uint64(numCreds)),
			)
			if err != nil {
				return nil, err
			}

			// After successful registration, emit typed event
			err = sdkCtx.EventManager().EmitTypedEvent(
				types.NewEventCosmosCredentialAdd(msg.Sender, 0),
			)
			if err != nil {
				return nil, err
			}

			// First credential will have number 0
			return &types.MsgRegisterCosmosCredentialResponse{
				CredentialNumber: 0,
			}, nil
		}
		return nil, err
	}
	// For existing account, create a new credential
	baseCredential := &types.BaseCredential{
		PublicKey: msg.Pubkey,
		Variant:   types.CredentialType_CREDENTIAL_TYPE_K256,
	}

	credential := &types.Credential{
		BaseCredential: baseCredential,
	}

	// check for duplicate credential
	hasDuplicate := HasDuplicateCredentialIDK256(smartAccount.Credentials, credential.BaseCredential.PublicKey)
	if hasDuplicate {
		return nil, errorsmod.Wrap(types.ErrDuplicateCredential, "credential already exists")
	}

	// Add credential to the existing account
	_, credentialNumberAdded, err := m.keeper.AddSmartAccountCredential(ctx, credential, &smartAccount)
	if err != nil {
		return nil, err
	}

	// After successful registration, emit typed event
	err = sdkCtx.EventManager().EmitTypedEvent(
		types.NewEventCosmosCredentialAdd(msg.Sender, credentialNumberAdded),
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterCosmosCredentialResponse{
		CredentialNumber: credentialNumberAdded,
	}, nil
}

func (m MsgServer) DeleteCredential(ctx context.Context, deleteCredential *types.MsgDeleteCredential) (*types.MsgDeleteCredentialResponse, error) {
	if enabled, err := m.keeper.IsSmartAccountsEnabled(ctx); err != nil || !enabled {
		return nil, types.ErrSmartAccountsNotEnabled
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	invoker, err := m.keeper.addressCodec.StringToBytes(deleteCredential.Sender)
	if err != nil {
		return nil, err
	}

	smartAccount, lookupErr := m.keeper.LookupAccountByAddress(ctx, invoker)
	if lookupErr != nil && !errors.Is(lookupErr, types.ErrSmartAccountDoesNotExist) {
		return nil, errorsmod.Wrap(lookupErr, "failed to lookup smart account")
	}
	credentialDeleted, err := m.keeper.DeleteCredential(ctx, &smartAccount, deleteCredential.CredentialNumber)
	if err != nil {
		return nil, err
	}

	// After successful deletion, emit event
	if err := sdkCtx.EventManager().EmitTypedEvent(
		types.NewEventCredentialDelete(
			smartAccount.Address,
			credentialDeleted.CredentialNumber,
		),
	); err != nil {
		return nil, err
	}

	return &types.MsgDeleteCredentialResponse{
		CredentialNumber: credentialDeleted.CredentialNumber,
	}, nil
}

// HasDuplicateCredentialIDK256 checks if the same secp256k1 public key already exists
// in the provided list of credentials
func HasDuplicateCredentialIDK256(credentials []*types.Credential, pubKey *codectypes.Any) bool {
	if pubKey == nil {
		return false
	}

	// Get incoming pubkey value bytes
	incomingKeyBytes := pubKey.Value

	for _, cred := range credentials {
		if cred.BaseCredential == nil || cred.BaseCredential.PublicKey == nil {
			continue
		}

		// Only compare with other secp256k1 keys
		if cred.BaseCredential.PublicKey.TypeUrl == pubKey.TypeUrl {
			// Compare the raw bytes of the public keys
			if bytes.Equal(cred.BaseCredential.PublicKey.Value, incomingKeyBytes) {
				return true
			}
		}
	}
	return false
}
