package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func init() {

}

func accAddressFromBech32(t *testing.T, addrStr string) sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(addrStr)
	require.NoError(t, err)
	return addr
}

func creator(t *testing.T) *authtypes.BaseAccount {
	addr := accAddressFromBech32(t, "cosmos184kae0avnncs5vnfzfj4tertwppnp0pyn0yy03")
	return authtypes.NewBaseAccount(addr, nil, 0, 0)
}

func TestNewEmptyMarkerValidate(t *testing.T) {
	creatorAddr := accAddressFromBech32(t, creator(t).Address)
	m := NewEmptyMarkerAccount(
		"test",
		creatorAddr.String(),
		[]AccessGrant{{Address: creatorAddr.String(), Permissions: []Access{Access_Mint, Access_Admin}}},
	)

	err := m.Validate()
	require.NoError(t, err)

	require.EqualValues(t, "test", m.GetDenom())
	require.EqualValues(t, "proposed", m.GetStatus().String())
	require.True(t, m.HasGovernanceEnabled())
	require.True(t, m.HasFixedSupply())
	require.True(t, m.AddressHasAccess(creatorAddr, Access_Mint), "creator was assigned mint permission")
	require.False(t, m.AddressHasAccess(creatorAddr, Access_Burn), "creator was not assigned burn permission")
	require.ElementsMatch(t, m.AddressListForPermission(Access_Mint), []sdk.AccAddress{creatorAddr})

	require.NoError(t, m.GrantAccess(NewAccessGrant(creatorAddr, []Access{Access_Burn})))
	require.True(t, m.AddressHasAccess(creatorAddr, Access_Mint), "creator still has mint permission")
	require.True(t, m.AddressHasAccess(creatorAddr, Access_Burn), "creator also has burn permission")

	require.Error(t, m.RevokeAccess(sdk.AccAddress([]byte{})), "can't revoke for an empty/invalid address")
	require.NoError(t, m.RevokeAccess(creatorAddr))
	require.False(t, m.AddressHasAccess(creatorAddr, Access_Burn), "creator permissions were revoked")
	require.NoError(t,
		m.GrantAccess(NewAccessGrant(creatorAddr, []Access{Access_Mint, Access_Admin})), "permissions restored")

	require.Equal(t, m.GetMarkerType(), MarkerType_Coin, "default marker type should be of MarkerType_Coin")
	require.Equal(t, m.GetPubKey(), nil, "set public key should not be supported.")
	require.Error(t, m.SetPubKey(nil), "set public key should not be supported.")
	require.Error(t, m.SetSequence(100), "set sequence should not be supported.")

	require.Error(t, m.SetStatus(StatusUndefined), "unknown status should fault")
	require.NoError(t, m.SetManager(creatorAddr), "should be able to set manager for proposed status event")
	require.NoError(t, m.SetStatus(StatusFinalized), "no error expected from setting a valid status")

	require.Error(t, m.SetManager(creatorAddr), "should not be able to set manager for active status event")

	require.EqualValues(t, m.GetManager(), creatorAddr, "creator address should match manager")
	require.NoError(t, m.SetStatus(StatusActive), "no error expected from setting a valid status")
	require.EqualValues(t, m.GetManager(), sdk.AccAddress([]byte{}), "manager should be empty on active status")

	require.EqualValues(t, m.GetSupply(), sdk.NewCoin("test", sdk.ZeroInt()), "initial supply will be zero")
	require.NoError(t, m.SetSupply(sdk.NewCoin("test", sdk.OneInt())))
	require.EqualValues(t, m.GetSupply(), sdk.NewCoin("test", sdk.OneInt()), "supply should be be one")
	require.Error(t,
		m.SetSupply(sdk.NewCoin("other", sdk.NewInt(1000))), "expected failure setting supply to invalid denom of coin")
	require.EqualValues(t, m.GetSupply(), sdk.NewCoin("test", sdk.OneInt()), "supply should be still be one")

	yaml, merr := m.MarshalYAML()
	require.NoError(t, merr, "marshall of yaml should succeed without error")
	require.Equal(t, yaml, m.String(), "should use yaml for string() view")

	if err != nil {
		t.Fatalf("expect no errors from marker validation: %s", err)
	}
	bz, err := m.Marshal()
	require.NoError(t, err)
	restored := &MarkerAccount{}
	//	fmt.Printf("%s", bz)
	require.NoError(t, restored.Unmarshal(bz))
	//require.True(t, restored.Equals(*m), "restored version should match serialized one")
}

func TestNewMarkerValidate(t *testing.T) {
	manager := MustGetMarkerAddress("manager")
	mAddr := MustGetMarkerAddress("test")
	fmt.Printf("Marker address: %s", mAddr)
	baseAcc := authtypes.NewBaseAccount(mAddr, nil, 0, 0)
	tests := []struct {
		name   string
		acc    authtypes.GenesisAccount
		expErr error
	}{
		{
			name: "empty marker is invalid",
			acc: NewEmptyMarkerAccount("test", "", nil),
			expErr: fmt.Errorf("a manager is required if there are no accounts with ACCESS_ADMIN and marker is not ACTIVE"),
		},
		{
			name: "insufficient supply",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("test", sdk.ZeroInt()), manager, nil, StatusFinalized, MarkerType_Coin, true, true, false, []string{}),
			expErr: fmt.Errorf("cannot create a marker with zero total supply and no authorization for minting more"),
		},
		{
			name: "invalid status",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("test", sdk.ZeroInt()), manager, nil, StatusUndefined, MarkerType_Coin, true, false, false, []string{}),
			expErr: fmt.Errorf("invalid marker status"),
		},
		{
			name: "invalid name and address pair",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("nottest", sdk.OneInt()), manager, nil, StatusProposed, MarkerType_Coin, true, false, false, []string{}),
			expErr: fmt.Errorf("address %s cannot be derived from the marker denom 'nottest'", baseAcc.GetAddress()),
		},
		{
			name: "invalid marker account permissions",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("test", sdk.OneInt()), manager,
				[]AccessGrant{{Address: MustGetMarkerAddress("foo").String(),
					Permissions: []Access{Access_Unknown}}}, StatusProposed, MarkerType_Coin, true, false, false, []string{}),
			expErr: fmt.Errorf("invalid access privileges granted: ACCESS_UNSPECIFIED is not supported for marker type MARKER_TYPE_COIN"),
		},
		{
			name: "invalid restricted marker account permissions",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("test", sdk.OneInt()), manager,
				[]AccessGrant{{Address: MustGetMarkerAddress("foo").String(),
					Permissions: []Access{Access_Unknown}}}, StatusProposed, MarkerType_RestrictedCoin, true, false, false, []string{}),
			expErr: fmt.Errorf("invalid access privileges granted: ACCESS_UNSPECIFIED is not supported for marker type MARKER_TYPE_RESTRICTED"),
		},
		{
			name: "marker account permissions assigned to self",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("test", sdk.OneInt()), manager, []AccessGrant{{Address: baseAcc.Address,
				Permissions: []Access{Access_Mint, Access_Admin}}}, StatusProposed, MarkerType_Coin, true, false, false, []string{}),
			expErr: fmt.Errorf("permissions cannot be granted to 'test' marker account: [ACCESS_MINT ACCESS_ADMIN]"),
		},
		{
			name: "invalid marker account permissions for type",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("test", sdk.OneInt()), manager,
				[]AccessGrant{{Address: MustGetMarkerAddress("foo").String(),
					Permissions: []Access{Access_Mint, Access_Admin, Access_Transfer}}}, StatusActive, MarkerType_Coin, true, false, false, []string{}),
			expErr: fmt.Errorf("invalid access privileges granted: ACCESS_TRANSFER is not supported for marker type MARKER_TYPE_COIN"),
		},
		{
			name: "invalid marker ibc type fixed supply",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("ibc/test", sdk.OneInt()), manager,
				[]AccessGrant{{Address: MustGetMarkerAddress("foo").String(),
					Permissions: []Access{Access_Admin, Access_Withdraw}}}, StatusActive, MarkerType_Coin, true, false, false, []string{}),
			expErr: fmt.Errorf("invalid ibc denom configuration: fixed supply is not supported for ibc marker"),
		},
		{
			name: "invalid marker ibc type has mint",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("ibc/test", sdk.OneInt()), manager,
				[]AccessGrant{{Address: MustGetMarkerAddress("foo").String(),
					Permissions: []Access{Access_Mint, Access_Admin, Access_Withdraw}}}, StatusActive, MarkerType_Coin, false, false, false, []string{}),
			expErr: fmt.Errorf("invalid ibc denom configuration: ACCESS_MINT is not supported for ibc marker"),
		},
		{
			name: "invalid marker ibc type has burn",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("ibc/test", sdk.OneInt()), manager,
				[]AccessGrant{{Address: MustGetMarkerAddress("foo").String(),
					Permissions: []Access{Access_Burn, Access_Admin, Access_Withdraw}}}, StatusActive, MarkerType_Coin, false, false, false, []string{}),
			expErr: fmt.Errorf("invalid ibc denom configuration: ACCESS_BURN is not supported for ibc marker"),
		},
		{
			name: "valid marker account",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("test", sdk.OneInt()), manager, nil, StatusProposed, MarkerType_Coin, true, false, false, []string{}),
			expErr: nil,
		},
		{
			name: "valid marker account",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("test", sdk.OneInt()), manager, nil, StatusActive, MarkerType_Coin, true, false, false, []string{}),
			expErr: nil,
		},
		{
			name: "coin type with forced transfer is invalid",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("test", sdk.OneInt()), manager, nil, StatusActive, MarkerType_Coin, true, true, true, []string{}),
			expErr: fmt.Errorf("forced transfers can only be allowed on restricted markers"),
		},
		{
			name: "coin type without forced transfer is ok",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("test", sdk.OneInt()), manager, nil, StatusActive, MarkerType_Coin, true, true, false, []string{}),
			expErr: nil,
		},
		{
			name: "restricted type with froced transfer is ok",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("test", sdk.OneInt()), manager, nil, StatusActive, MarkerType_RestrictedCoin, true, true, true, []string{}),
			expErr: nil,
		},
		{
			name: "restricted type without forced transfer is ok",
			acc: NewMarkerAccount(baseAcc, sdk.NewCoin("test", sdk.OneInt()), manager, nil, StatusActive, MarkerType_RestrictedCoin, true, true, false, []string{}),
			expErr: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.acc.Validate()
			if err == nil {
				require.Equal(t, tt.expErr, err)
			} else {
				errStr := fmt.Sprintf("%v", err)
				chkStr := fmt.Sprintf("%v", tt.expErr)
				require.Equal(t, chkStr, errStr)
			}
		})
	}
}

func TestNewMarkerMsgEncoding(t *testing.T) {
	base := authtypes.NewBaseAccountWithAddress(MustGetMarkerAddress("testcoin"))
	newMsgMarker := NewMsgAddMarkerRequest("testcoin", sdk.OneInt(), base.GetAddress(), base.GetAddress(), MarkerType_Coin, false, false, false, []string{})

	require.NoError(t, newMsgMarker.ValidateBasic())
}

func TestMarkerTypeStrings(t *testing.T) {
	tests := []struct {
		name       string
		typeString string
		expType    MarkerType
		expErr     error
	}{
		{
			name: "standard coin",
			typeString: "coin",
			expType: MarkerType_Coin,
			expErr: nil,
		},
		{
			name: "upper coin",
			typeString: "COIN",
			expType: MarkerType_Coin,
			expErr: nil,
		},
		{
			name: "enum coin value",
			typeString: "MARKER_TYPE_COIN",
			expType: MarkerType_Coin,
			expErr: nil,
		},
		{
			name: "plain restricted",
			typeString: "restricted",
			expType: MarkerType_RestrictedCoin,
			expErr: nil,
		},
		{
			name: "restrictedcoin style",
			typeString: "restrictedcoin",
			expType: MarkerType_RestrictedCoin,
			expErr: nil,
		},
		{
			name: "enum string restricted",
			typeString: "MARKER_TYPE_RESTRICTED",
			expType: MarkerType_RestrictedCoin,
			expErr: nil,
		},
		{
			name: "invalid",
			typeString: "invalid",
			expType: MarkerType_Unknown,
			expErr: fmt.Errorf("'invalid' is not a valid marker status"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m, err := MarkerTypeFromString(tt.typeString)
			require.Equal(t, tt.expErr, err)
			require.Equal(t, tt.expType, m)
		})
	}
}
