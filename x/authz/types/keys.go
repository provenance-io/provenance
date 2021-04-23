package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "authz"

	// StoreKey is the store key string for authz
	StoreKey = ModuleName

	// RouterKey is the message route for authz
	RouterKey = ModuleName

	// QuerierRoute is the querier route for authz
	QuerierRoute = ModuleName
)

// Keys for authz store
// Items are stored with the following key: values
//
// - 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>: Grant

var (
	// Keys for store prefixes
	GrantKey = []byte{0x01} // prefix for each key
)

// GetAuthorizationStoreKey - return authorization store key
func GetAuthorizationStoreKey(grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) []byte {
	return append(append(append(
		GrantKey,
		MustLengthPrefix(granter)...),
		MustLengthPrefix(grantee)...),
		[]byte(msgType)...,
	)
}

// ExtractAddressesFromGrantKey - split granter & grantee address from the authorization key
func ExtractAddressesFromGrantKey(key []byte) (granterAddr, granteeAddr sdk.AccAddress) {
	// key if of format:
	// 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>
	granterAddrLen := key[1] // remove prefix key
	granterAddr = sdk.AccAddress(key[2 : 2+granterAddrLen])
	granteeAddrLen := int(key[2+granterAddrLen])
	granteeAddr = sdk.AccAddress(key[3+granterAddrLen : 3+granterAddrLen+byte(granteeAddrLen)])

	return granterAddr, granteeAddr
}

// TODO Delete this when we goto 0.43
// MaxAddrLen is the maximum allowed length (in bytes) for an address.
const MaxAddrLen = 255

// LengthPrefix prefixes the address bytes with its length, this is used
// for example for variable-length components in store keys.
func LengthPrefix(bz []byte) ([]byte, error) {
	bzLen := len(bz)
	if bzLen == 0 {
		return bz, nil
	}

	if bzLen > MaxAddrLen {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "address length should be max %d bytes, got %d", MaxAddrLen, bzLen)
	}

	return append([]byte{byte(bzLen)}, bz...), nil
}

// MustLengthPrefix is LengthPrefix with panic on error.
func MustLengthPrefix(bz []byte) []byte {
	res, err := LengthPrefix(bz)
	if err != nil {
		panic(err)
	}

	return res
}
