package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Ensure ProvenanceAccount implements the auth.AccountI interface
var _ sdk.AccountI = (*ProvenanceAccount)(nil)

func (pa ProvenanceAccount) ValidateBasic() error {
	if pa.Address == "" {
		return fmt.Errorf("base account has to be set")
	}
	return nil
}

// NewProvenanceAccount creates a new ProvenanceAccount with embedded BaseAccount
func NewProvenanceAccount(baseAccount *authtypes.BaseAccount, smartAccountNumber uint64, credentials []*Credential, isSmartAccountOnly bool) ProvenanceAccount {
	return ProvenanceAccount{
		BaseAccount:                      baseAccount,
		SmartAccountNumber:               smartAccountNumber,
		Credentials:                      credentials,
		IsSmartAccountOnlyAuthentication: isSmartAccountOnly,
	}
}
