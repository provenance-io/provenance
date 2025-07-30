package keeper

import (
	"context"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	accountkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/x/smartaccounts/provenanceaccount"
	"github.com/provenance-io/provenance/x/smartaccounts/types"
)

var (
	// CredentialNumberKey is the key for the account number.
	CredentialNumberKey = collections.NewPrefix(1)
	// SmartAccountNumberKey global smart account number key
	SmartAccountNumberKey = collections.NewPrefix(2)
	// SmartAccountStatePrefix is the prefix for the smart account state.
	SmartAccountStatePrefix = collections.NewPrefix(3)
	// ParamsKey is the key for the params of the module.
	ParamsKey = collections.NewPrefix(4)
)

func NewSmartAccountIndexes(sb *collections.SchemaBuilder) SmartAccountsIndexes {
	return SmartAccountsIndexes{
		Address: indexes.NewUnique(
			sb, types.SmartAccountNumberStoreKeyPrefix, "smart_account_by_number", sdk.AccAddressKey, sdk.AccAddressKey,
			func(addr sdk.AccAddress, v types.ProvenanceAccount) (sdk.AccAddress, error) {
				return addr, nil
			},
		),
	}
}

func NewKeeper(
	cdc codec.Codec,
	addressCodec address.Codec,
	storeService store.KVStoreService,
	handlerMap *signing.HandlerMap,
	accountKeeper accountkeeper.AccountKeeper,
	authority string,
	logger log.Logger,
) Keeper {
	if addressCodec == nil {
		panic(fmt.Errorf("addressCodec cannot be nil passed to New SmartAccountKeeper"))
	}
	sb := collections.NewSchemaBuilder(storeService)
	keeper := Keeper{
		Codec:              cdc,
		addressCodec:       addressCodec,
		Schema:             collections.Schema{},
		CredentialNumber:   collections.NewSequence(sb, CredentialNumberKey, "credential_number"),
		SmartAccountNumber: collections.NewSequence(sb, SmartAccountNumberKey, "smart_account_number"),
		SmartAccounts:      collections.NewIndexedMap(sb, SmartAccountStatePrefix, "smart_accounts", sdk.AccAddressKey, codec.CollValue[types.ProvenanceAccount](cdc), NewSmartAccountIndexes(sb)),
		AccountKeeper:      accountKeeper,
		authority:          authority,
		SmartAccountParams: collections.NewItem(sb, ParamsKey, "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(fmt.Errorf("cannot build schema for SmartAccountKeeper: %w", err))
	}
	keeper.Schema = schema
	keeper.provenanceAccount, err = provenanceaccount.NewProvenanceAccountHandler(provenanceaccount.AccountCreatorDependencies{
		AddressCodec: addressCodec,
		StoreService: storeService,
		Cdc:          cdc,
	}, handlerMap)

	if err != nil {
		panic(fmt.Errorf("error in SmartAccountKeeper: %w", err))
	}

	return keeper
}

type SmartAccountsIndexes struct {
	// Address is a unique index that indexes accounts by their address.
	Address *indexes.Unique[sdk.AccAddress, sdk.AccAddress, types.ProvenanceAccount]
}

func (a SmartAccountsIndexes) IndexesList() []collections.Index[sdk.AccAddress, types.ProvenanceAccount] {
	return []collections.Index[sdk.AccAddress, types.ProvenanceAccount]{
		a.Address,
	}
}

type Keeper struct {
	addressCodec address.Codec
	Codec        codec.Codec

	// Schema is the schema for the module.
	Schema collections.Schema
	// SmartAccountNumber is the last global account number.
	SmartAccountNumber collections.Sequence
	// CredentialNumber is the last global credential number.
	CredentialNumber collections.Sequence

	StoreService store.KVStoreService
	HandlerMap   *signing.HandlerMap

	provenanceAccount *provenanceaccount.ProvenanceSmartAccountHandler // for now there is only one implementation, keeping it simple for now
	SmartAccounts     *collections.IndexedMap[sdk.AccAddress, types.ProvenanceAccount, SmartAccountsIndexes]
	AccountKeeper     accountkeeper.AccountKeeper

	authority          string
	SmartAccountParams collections.Item[types.Params]
}

func (k Keeper) NextCredentialNumber(
	ctx context.Context,
) (credentialNumber uint64, err error) {
	credentialNumber, err = k.CredentialNumber.Next(ctx)
	if err != nil {
		return 0, err
	}

	return credentialNumber, nil
}

func (k Keeper) NextSmartAccountNumber(
	ctx context.Context,
) (accNum uint64, err error) {
	accNum, err = k.SmartAccountNumber.Next(ctx)
	if err != nil {
		return 0, err
	}

	return accNum, nil
}

// Init called when initializing the provenance smart account for the first time.
func (k Keeper) Init(ctx context.Context, msg *types.MsgInit) (*types.ProvenanceAccount, error) {
	// Get module params to check max credentials allowed
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	// Check if the number of credentials exceeds the maximum allowed
	if len(msg.Credentials) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "at least one credential must be provided")
	}
	if len(msg.Credentials) > int(params.MaxCredentialAllowed) {
		return nil, fmt.Errorf("maximum number of credentials (%d) exceeded", params.MaxCredentialAllowed)
	}

	// Convert the string address to sdk.AccAddress
	senderAddr, err := k.addressCodec.StringToBytes(msg.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	// Fetch the existing base account
	baseAcc := k.AccountKeeper.GetAccount(ctx, senderAddr)
	if baseAcc == nil {
		return nil, fmt.Errorf("base account for %s not found", msg.Sender)
	}

	// Type assert to BaseAccount
	baseAccount, ok := baseAcc.(*authtypes.BaseAccount)
	if !ok {
		return nil, fmt.Errorf("account %s is not a base account", msg.Sender)
	}

	// Get the next smart account number
	nextNumber, err := k.NextSmartAccountNumber(ctx)
	if err != nil {
		return nil, err
	}

	// Process and validate credentials
	var validCredentials []*types.Credential
	for _, cred := range msg.Credentials {
		credentialNumber, err := k.NextCredentialNumber(ctx)
		if err != nil {
			return nil, err
		}
		cred.CredentialNumber = credentialNumber

		// Verify public key
		err = k.verifyPubKey(ctx, cred.PublicKey)
		if err != nil {
			return nil, err
		}
		validCredentials = append(validCredentials, cred)
	}

	// Create account using the NewProvenanceAccount helper
	acc := types.NewProvenanceAccount(
		baseAccount,
		nextNumber,
		validCredentials,
		false, // Default to false for IsSmartAccountOnlyAuthentication
	)

	// Save account details - note we need to pass a pointer to match the function signature
	account, err := k.SaveAccountDetails(ctx, &acc, senderAddr)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (k Keeper) verifyPubKey(ctx context.Context, anyPk *codectypes.Any) error {
	if k.provenanceAccount == nil {
		return fmt.Errorf("provenanceAccount handler not initialized")
	}
	// Check for nil pointer
	if anyPk == nil {
		return fmt.Errorf("pubkey cannot be nil")
	}
	// check if known
	name := nameFromTypeURL(anyPk.TypeUrl)
	impl, exists := k.provenanceAccount.SupportedPubKeys[name]
	if !exists {
		return fmt.Errorf("unknown pubkey type %s", name)
	}
	pk, err := impl.Decode(anyPk.Value)
	if err != nil {
		return fmt.Errorf("unable to decode pubkey: %w", err)
	}
	err = impl.Validate(pk)
	if err != nil {
		return fmt.Errorf("unable to validate pubkey: %w", err)
	}
	return nil
}

func (k Keeper) SaveAccountDetails(ctx context.Context, acc *types.ProvenanceAccount, address sdk.AccAddress) (*types.ProvenanceAccount, error) {
	// Set the account in the store
	err := k.SmartAccounts.Set(ctx, address, *acc)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

// LookupAccountByAddress looks up an account by its address.
func (k Keeper) LookupAccountByAddress(ctx context.Context, address sdk.AccAddress) (types.ProvenanceAccount, error) {
	// Check if the account exists
	exists, err := k.SmartAccounts.Has(ctx, address)
	if err != nil {
		return types.ProvenanceAccount{}, err
	}
	if !exists {
		return types.ProvenanceAccount{}, types.ErrSmartAccountDoesNotExist
	}
	// Retrieve the account details directly using the address
	account, err := k.SmartAccounts.Get(ctx, address)
	if err != nil {
		return types.ProvenanceAccount{}, err
	}
	return account, nil
}

// IterateSmartAccounts iterates over all the stored smart accounts and performs a callback function.
// Stops iteration when callback returns true.
func (k Keeper) IterateSmartAccounts(ctx context.Context, cb func(smartAccounts types.ProvenanceAccount) (stop bool)) error {
	return k.SmartAccounts.Walk(ctx, nil, func(_ sdk.AccAddress, value types.ProvenanceAccount) (bool, error) {
		return cb(value), nil
	})
}

// GetAllSmartAccounts returns all accounts in the smartaccountKeeper.
func (k Keeper) GetAllSmartAccounts(ctx context.Context) ([]types.ProvenanceAccount, error) {
	var accounts []types.ProvenanceAccount
	err := k.IterateSmartAccounts(ctx, func(acc types.ProvenanceAccount) (stop bool) {
		accounts = append(accounts, acc)
		return false
	})
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (keeper Keeper) GetParams(c context.Context) (*types.Params, error) {
	p, err := keeper.SmartAccountParams.Get(c)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// SetParams sets the account parameters to the param store.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	return k.SmartAccountParams.Set(ctx, params)
}

func (k Keeper) AddSmartAccountCredential(ctx context.Context, credential *types.Credential, smartAccount *types.ProvenanceAccount) (*types.ProvenanceAccount, uint64, error) {
	// Get module params to check max credentials allowed
	params, errGetParams := k.GetParams(ctx)
	if errGetParams != nil {
		return nil, 0, errGetParams
	}

	// Check if adding a new credential exceeds the maximum allowed
	if len(smartAccount.Credentials)+1 > int(params.MaxCredentialAllowed) {
		return nil, 0, fmt.Errorf("maximum number of credentials (%d) exceeded", params.MaxCredentialAllowed)
	}

	var credentialNumber uint64
	// get the next credential number
	credentialNumber, err := k.NextCredentialNumber(ctx)
	if err != nil {
		return nil, 0, err
	}
	credential.CredentialNumber = credentialNumber
	// save the public key, and make sure it is valid.
	err = k.verifyPubKey(ctx, credential.PublicKey)
	if err != nil {
		return nil, 0, err
	}

	// Add the new credential to existing ones
	smartAccount.Credentials = append(smartAccount.Credentials, credential)
	addressOfSender, err := k.addressCodec.StringToBytes(smartAccount.Address)
	if err != nil {
		return nil, 0, err
	}
	// Save the updated account
	err = k.SmartAccounts.Set(ctx, addressOfSender, *smartAccount)
	if err != nil {
		return nil, 0, err
	}

	return smartAccount, credentialNumber, nil
}

// DeleteCredential deletes a credential from the smart account.
// It returns the updated smart account and an error if any.
func (k Keeper) DeleteCredential(ctx context.Context, smartAccount *types.ProvenanceAccount, credentialNumber uint64) (*types.Credential, error) {
	var deletedCredential *types.Credential

	modifiedCredentials, deletedCredential := removeCredentialByNumber(smartAccount.Credentials, credentialNumber)
	// If credential not found, return error
	if deletedCredential == nil {
		return nil, fmt.Errorf("credential with number %d not found", credentialNumber)
	}

	// Save the updated smart account
	addr, err := sdk.AccAddressFromBech32(smartAccount.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %w", err)
	}

	smartAccount.Credentials = modifiedCredentials

	if _, err := k.SaveAccountDetails(ctx, smartAccount, addr); err != nil {
		return nil, fmt.Errorf("failed to save account details: %w", err)
	}

	return deletedCredential, nil
}

func removeCredentialByNumber(credentials []*types.Credential, credNumber uint64) ([]*types.Credential, *types.Credential) {
	result := make([]*types.Credential, 0, len(credentials))
	var deletedCredential *types.Credential
	for _, cred := range credentials {
		if cred.BaseCredential.CredentialNumber != credNumber {
			result = append(result, cred)
		} else {
			deletedCredential = cred
		}
	}
	return result, deletedCredential
}

func (k Keeper) IsSmartAccountsEnabled(ctx context.Context) (bool, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return false, err
	}
	return params.Enabled, nil
}

func nameFromTypeURL(url string) string {
	name := url
	if i := strings.LastIndexByte(url, '/'); i >= 0 {
		name = name[i+len("/"):]
	}
	return name
}
