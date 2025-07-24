package antewrapper

import (
	"cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	errorsmod "cosmossdk.io/errors"
	txsigning "cosmossdk.io/x/tx/signing"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	smartaccountkeeper "github.com/provenance-io/provenance/x/smartaccounts/keeper"
	smartaccounttypes "github.com/provenance-io/provenance/x/smartaccounts/types"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	// simulation signature values used to estimate gas consumption
	key                = make([]byte, secp256k1.PubKeySize)
	simSecp256k1Pubkey = &secp256k1.PubKey{Key: key}
)

func init() {
	// This decodes a valid hex string into a sepc256k1Pubkey for use in transaction simulation
	bz, _ := hex.DecodeString("035AD6810A47F073553FF30D2FCC7E0D3B1C0B74B61A1AAA2582344037151E143A")
	copy(key, bz)
	simSecp256k1Pubkey.Key = key
}

// ProvenanceSigVerificationDecorator verifies all signatures for a tx and returns an error if any are invalid. Note,
// the ProvenanceSigVerificationDecorator will not check signatures on ReCheck.
//
// CONTRACT: Pubkeys are set in context for all signers before this decorator runs
// CONTRACT: Tx must implement SigVerifiableTx interface
type ProvenanceSigVerificationDecorator struct {
	ak                 ante.AccountKeeper
	signModeHandler    *txsigning.HandlerMap
	smartaccountkeeper smartaccountkeeper.Keeper
}

func NewProvenanceSigVerificationDecorator(ak ante.AccountKeeper, signModeHandler *txsigning.HandlerMap, smartaccountkeeper smartaccountkeeper.Keeper) ProvenanceSigVerificationDecorator {
	return ProvenanceSigVerificationDecorator{
		ak:                 ak,
		signModeHandler:    signModeHandler,
		smartaccountkeeper: smartaccountkeeper,
	}
}

// OnlyLegacyAminoSigners checks SignatureData to see if all
// signers are using SIGN_MODE_LEGACY_AMINO_JSON. If this is the case
// then the corresponding SignatureV2 struct will not have account sequence
// explicitly set, and we should skip the explicit verification of sig.Sequence
// in the ProvenanceSigVerificationDecorator's AnteHandler function.
func OnlyLegacyAminoSigners(sigData signing.SignatureData) bool {
	switch v := sigData.(type) {
	case *signing.SingleSignatureData:
		return v.SignMode == signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON
	case *signing.MultiSignatureData:
		for _, s := range v.Signatures {
			if !OnlyLegacyAminoSigners(s) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func (svd ProvenanceSigVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	sigTx, ok := tx.(authsigning.Tx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	// stdSigs contains the sequence number, account number, and signatures.
	// When simulating, this would just be a 0-length slice.
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return ctx, err
	}

	signers, err := sigTx.GetSigners()
	if err != nil {
		return ctx, err
	}

	// check that signer length and signature length are the same
	if len(sigs) != len(signers) {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "invalid number of signer;  expected: %d, got %d", len(signers), len(sigs))
	}

	for i, sig := range sigs {

		acc, err := GetSignerAcc(ctx, svd.ak, signers[i])
		if err != nil {
			return ctx, err
		}

		// retrieve pubkey
		// assumption is that should always be set even for a smart account
		pubKey := acc.GetPubKey()
		if !simulate && pubKey == nil {
			return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidPubKey, "pubkey on account is not set")
		}

		// Check account sequence number.
		// Should still be same for smart accounts
		if sig.Sequence != acc.GetSequence() {
			return ctx, errorsmod.Wrapf(
				sdkerrors.ErrWrongSequence,
				"account sequence mismatch, expected %d, got %d", acc.GetSequence(), sig.Sequence,
			)
		}

		// retrieve signer data
		genesis := ctx.BlockHeight() == 0
		chainID := ctx.ChainID()
		var accNum uint64
		if !genesis {
			accNum = acc.GetAccountNumber()
		}

		// no need to verify signatures on recheck tx
		if !simulate && !ctx.IsReCheckTx() && ctx.IsSigverifyTx() {
			anyPk, _ := codectypes.NewAnyWithValue(pubKey)

			signerData := txsigning.SignerData{
				Address:       acc.GetAddress().String(),
				ChainID:       chainID,
				AccountNumber: accNum,
				Sequence:      acc.GetSequence(),
				PubKey: &anypb.Any{
					TypeUrl: anyPk.TypeUrl,
					Value:   anyPk.Value,
				},
			}
			adaptableTx, validAdaptableTx := tx.(authsigning.V2AdaptableTx)
			if !validAdaptableTx {
				return ctx, fmt.Errorf("expected tx to implement V2AdaptableTx, got %T", tx)
			}
			txData := adaptableTx.GetSigningTxData()

			// First, try the default cosmos signature verification.
			err := authsigning.VerifySignature(ctx, pubKey, signerData, sig.Data, svd.signModeHandler, txData)
			if err == nil {
				// Standard signature verification was successful.
				continue
			}

			// Standard verification failed. Check if smart account verification should be attempted.
			shouldAttemptSmartAccountVerification := ShouldSkipSmartAccountVerification(tx.GetMsgs())

			if shouldAttemptSmartAccountVerification {
				// verifySmartAccountSignature will return the original error if it's not a smart account,
				// nil on success, or a new specific error if smart account verification fails.
				err = svd.verifySmartAccountSignature(ctx, signers[i], sig, signerData, svd.signModeHandler, txData, err)
				if err == nil {
					// Smart account verification was successful.
					continue
				}
			}

			// If we reach here, all verification methods have failed for this signer.
			var errMsg string
			if OnlyLegacyAminoSigners(sig.Data) {
				// If all signers are using SIGN_MODE_LEGACY_AMINO, we rely on VerifySignature to check account sequence number,
				// and therefore communicate sequence number as a potential cause of error.
				errMsg = fmt.Sprintf("signature verification failed; please verify account number (%d), sequence (%d) and chain-id (%s)", accNum, acc.GetSequence(), chainID)
			} else {
				errMsg = fmt.Sprintf("signature verification failed; please verify account number (%d) and chain-id (%s): %v", accNum, chainID, err)
			}
			return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, errMsg)
		}
	}

	return next(ctx, tx, simulate)
}

// ShouldSkipSmartAccountVerification checks if the transaction contains messages which should skip smart account verification.
// returns true if all messages are allowed to be authenticated by a smart account.
// returns false if any message is not allowed to be authenticated by a smart account.
func ShouldSkipSmartAccountVerification(msgs []sdk.Msg) (shouldSkipSmartAccountVerification bool) {
	for _, msg := range msgs {
		switch msg.(type) {
		case
			*smartaccounttypes.MsgDeleteCredential:
			return false
		default:
			continue
		}
	}
	// All messages are allowed to be authenticated by a smart account.
	return true
}

// GetSignerAcc returns an account for a given address that is expected to sign
// a transaction.
func GetSignerAcc(ctx sdk.Context, ak ante.AccountKeeper, addr sdk.AccAddress) (sdk.AccountI, error) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, nil
	}

	return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
}

func VerifySignature(
	cdc codec.Codec,
	cred *smartaccounttypes.Credential,
	signerData txsigning.SignerData,
	signatureData signing.SignatureData,
	handler *txsigning.HandlerMap,
	txData txsigning.TxData,
	ctx sdk.Context,
) error {
	switch data := signatureData.(type) {
	case *signing.SingleSignatureData:
		signBytes, err := handler.GetSignBytes(ctx, signingv1beta1.SignMode(data.SignMode), signerData, txData)
		if err != nil {
			return err
		}
		// Check credential type and verify accordingly
		switch cred.BaseCredential.Variant {
		case smartaccounttypes.CredentialType_CREDENTIAL_TYPE_WEBAUTHN, smartaccounttypes.CredentialType_CREDENTIAL_TYPE_WEBAUTHN_UV:
			// Get the FidoAuthenticator from the oneof field
			fidoCredentialWrapper, ok := cred.GetAuthenticator().(*smartaccounttypes.Credential_Fido2Authenticator)
			if !ok || fidoCredentialWrapper == nil {
				return errorsmod.Wrapf(smartaccounttypes.ErrParseCredential, "Failed to get Fido2Authenticator from credential")
			}
			// Access the actual Fido2Authenticator object
			fidoCredential := fidoCredentialWrapper.Fido2Authenticator
			if fidoCredential.VerifySignature(ctx, *cred, signBytes, cdc, data.Signature) != nil {
				return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "unable to verify single signer signature")
			}
			return nil
		case smartaccounttypes.CredentialType_CREDENTIAL_TYPE_K256:
			// Get the K256Authenticator from the oneof field
			k256CredentialWrapper, ok := cred.GetAuthenticator().(*smartaccounttypes.Credential_K256Authenticator)
			if !ok || k256CredentialWrapper == nil {
				return errorsmod.Wrapf(smartaccounttypes.ErrParseCredential, "Failed to get K256 Authenticator from credential")
			}
			// Access the actual K256Authenticator object
			k256Authenticator := k256CredentialWrapper.K256Authenticator
			if k256Authenticator.VerifySignature(ctx, *cred, signBytes, cdc, data.Signature) != nil {
				return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "unable to verify secp256k1 signature")
			}
			return nil

		default:
			return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "unsupported credential type: %v", cred.BaseCredential.Variant)
		}
	default:
		return fmt.Errorf("multisig not supported yet.")
	}
}

// verifySmartAccountSignature attempts to verify a signature using credentials from a smart account, if it exists and has credentials.
// If the smart account does not exist or has no credentials or signature verification fails, it falls back to the original error.
func (svd ProvenanceSigVerificationDecorator) verifySmartAccountSignature(
	ctx sdk.Context,
	signerAddr sdk.AccAddress,
	sig signing.SignatureV2,
	signerData txsigning.SignerData,
	signModeHandler *txsigning.HandlerMap,
	txData txsigning.TxData,
	originalErr error,
) error {
	// Attempt to find a smart account for the given signer address.
	smartAccount, err := svd.smartaccountkeeper.LookupAccountByAddress(ctx, signerAddr)
	if err != nil {
		// If the error is specifically that a smart account does not exist for this address,
		// it's not a failure. It simply means this is a regular account, and we should
		// fall back to the original signature verification error.
		if errors.Is(err, smartaccounttypes.ErrSmartAccountDoesNotExist) {
			return originalErr
		}
		// For any other unexpected error during lookup (e.g., database issues), wrap and return it.
		return errorsmod.Wrap(err, "failed to lookup smart account during signature verification")
	}

	// If a smart account was found, check if it has credentials to verify the signature.
	if len(smartAccount.Credentials) > 0 {
		switch data := sig.Data.(type) {
		case *signing.SingleSignatureData:
			// FIDO2 signature verification only applies to SIGN_MODE_DIRECT
			if data.SignMode != signing.SignMode_SIGN_MODE_DIRECT {
				return originalErr
			}

			allErrors := true
			// if smart account is found, we need to verify the signature against the smart account
			for _, cred := range smartAccount.Credentials {
				if VerifySignature(svd.smartaccountkeeper.Codec, cred, signerData, sig.Data, signModeHandler, txData, ctx) == nil {
					allErrors = false
					break
				}
			}
			if allErrors {
				return errorsmod.Wrap(sdkerrors.ErrUnauthorized, "signature verification failed for all available smart account credentials")
			}
			// At least one credential successfully verified the signature.
			return nil
		case *signing.MultiSignatureData:
			// Fido2 signature verification is not currently supported for multisig
			return originalErr
		}
	}

	// Fallback for a smart account that exists but has no credentials,
	// or for any other unhandled cases.
	return originalErr
}
