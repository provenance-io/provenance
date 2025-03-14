package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/testutil/assertions"
)

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

	require.EqualValues(t, m.GetSupply(), sdk.NewInt64Coin("test", 0), "initial supply will be zero")
	require.NoError(t, m.SetSupply(sdk.NewInt64Coin("test", 1)))
	require.EqualValues(t, m.GetSupply(), sdk.NewInt64Coin("test", 1), "supply should be be one")
	require.Error(t,
		m.SetSupply(sdk.NewInt64Coin("other", 1000)), "expected failure setting supply to invalid denom of coin")
	require.EqualValues(t, m.GetSupply(), sdk.NewInt64Coin("test", 1), "supply should be still be one")

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
			"empty marker is invalid",
			NewEmptyMarkerAccount("test", "", nil),
			fmt.Errorf("a manager is required if there are no accounts with ACCESS_ADMIN and marker is not ACTIVE"),
		},
		{
			"insufficient supply",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("test", 0), manager, nil, StatusFinalized, MarkerType_Coin, true, true, false, []string{}),
			fmt.Errorf("cannot create a marker with zero total supply and no authorization for minting more"),
		},
		{
			"invalid status",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("test", 0), manager, nil, StatusUndefined, MarkerType_Coin, true, false, false, []string{}),
			fmt.Errorf("invalid marker status"),
		},
		{
			"invalid name and address pair",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("nottest", 1), manager, nil, StatusProposed, MarkerType_Coin, true, false, false, []string{}),
			fmt.Errorf("address %s cannot be derived from the marker denom 'nottest'", baseAcc.GetAddress()),
		},
		{
			"invalid marker account permissions",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("test", 1), manager,
				[]AccessGrant{{Address: MustGetMarkerAddress("foo").String(),
					Permissions: []Access{Access_Unknown}}}, StatusProposed, MarkerType_Coin, true, false, false, []string{}),
			fmt.Errorf("invalid access privileges granted: ACCESS_UNSPECIFIED is not supported for marker type MARKER_TYPE_COIN"),
		},
		{
			"invalid restricted marker account permissions",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("test", 1), manager,
				[]AccessGrant{{Address: MustGetMarkerAddress("foo").String(),
					Permissions: []Access{Access_Unknown}}}, StatusProposed, MarkerType_RestrictedCoin, true, false, false, []string{}),
			fmt.Errorf("invalid access privileges granted: ACCESS_UNSPECIFIED is not supported for marker type MARKER_TYPE_RESTRICTED"),
		},
		{
			"marker account permissions assigned to self",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("test", 1), manager, []AccessGrant{{Address: baseAcc.Address,
				Permissions: []Access{Access_Mint, Access_Admin}}}, StatusProposed, MarkerType_Coin, true, false, false, []string{}),
			fmt.Errorf("permissions cannot be granted to 'test' marker account: [ACCESS_MINT ACCESS_ADMIN]"),
		},
		{
			"invalid marker account permissions for type",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("test", 1), manager,
				[]AccessGrant{{Address: MustGetMarkerAddress("foo").String(),
					Permissions: []Access{Access_Mint, Access_Admin, Access_Transfer}}}, StatusActive, MarkerType_Coin, true, false, false, []string{}),
			fmt.Errorf("invalid access privileges granted: ACCESS_TRANSFER is not supported for marker type MARKER_TYPE_COIN"),
		},
		{
			"invalid marker ibc type fixed supply",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("ibc/test", 1), manager,
				[]AccessGrant{{Address: MustGetMarkerAddress("foo").String(),
					Permissions: []Access{Access_Admin, Access_Withdraw}}}, StatusActive, MarkerType_Coin, true, false, false, []string{}),
			fmt.Errorf("invalid ibc denom configuration: fixed supply is not supported for ibc marker"),
		},
		{
			"invalid marker ibc type has mint",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("ibc/test", 1), manager,
				[]AccessGrant{{Address: MustGetMarkerAddress("foo").String(),
					Permissions: []Access{Access_Mint, Access_Admin, Access_Withdraw}}}, StatusActive, MarkerType_Coin, false, false, false, []string{}),
			fmt.Errorf("invalid ibc denom configuration: ACCESS_MINT is not supported for ibc marker"),
		},
		{
			"invalid marker ibc type has burn",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("ibc/test", 1), manager,
				[]AccessGrant{{Address: MustGetMarkerAddress("foo").String(),
					Permissions: []Access{Access_Burn, Access_Admin, Access_Withdraw}}}, StatusActive, MarkerType_Coin, false, false, false, []string{}),
			fmt.Errorf("invalid ibc denom configuration: ACCESS_BURN is not supported for ibc marker"),
		},
		{
			"valid marker account",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("test", 1), manager, nil, StatusProposed, MarkerType_Coin, true, false, false, []string{}),
			nil,
		},
		{
			"valid marker account",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("test", 1), manager, nil, StatusActive, MarkerType_Coin, true, false, false, []string{}),
			nil,
		},
		{
			"coin type with forced transfer is invalid",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("test", 1), manager, nil, StatusActive, MarkerType_Coin, true, true, true, []string{}),
			fmt.Errorf("forced transfers can only be allowed on restricted markers"),
		},
		{
			"coin type without forced transfer is ok",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("test", 1), manager, nil, StatusActive, MarkerType_Coin, true, true, false, []string{}),
			nil,
		},
		{
			"restricted type with froced transfer is ok",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("test", 1), manager, nil, StatusActive, MarkerType_RestrictedCoin, true, true, true, []string{}),
			nil,
		},
		{
			"restricted type without forced transfer is ok",
			NewMarkerAccount(baseAcc, sdk.NewInt64Coin("test", 1), manager, nil, StatusActive, MarkerType_RestrictedCoin, true, true, false, []string{}),
			nil,
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
	newMsgMarker := NewMsgAddMarkerRequest("testcoin", sdkmath.OneInt(), base.GetAddress(), base.GetAddress(), MarkerType_Coin, false, false, false, []string{}, 0, 0)

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
			"standard coin",
			"coin",
			MarkerType_Coin,
			nil,
		},
		{
			"upper coin",
			"COIN",
			MarkerType_Coin,
			nil,
		},
		{
			"enum coin value",
			"MARKER_TYPE_COIN",
			MarkerType_Coin,
			nil,
		},
		{
			"plain restricted",
			"restricted",
			MarkerType_RestrictedCoin,
			nil,
		},
		{
			"restrictedcoin style",
			"restrictedcoin",
			MarkerType_RestrictedCoin,
			nil,
		},
		{
			"enum string restricted",
			"MARKER_TYPE_RESTRICTED",
			MarkerType_RestrictedCoin,
			nil,
		},
		{
			"invalid",
			"invalid",
			MarkerType_Unknown,
			fmt.Errorf("'invalid' is not a valid marker status"),
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

func TestAddToRequiredAttributes(t *testing.T) {
	tests := []struct {
		name          string
		addList       []string
		reqAttrs      []string
		expectedAttrs []string
		expectedError string
	}{
		{
			name:          "should fail, duplicate value",
			addList:       []string{"foo", "bar"},
			reqAttrs:      []string{"foo", "baz"},
			expectedError: `attribute "foo" is already required`,
		},
		{
			name:          "should succeed, add elements to none empty list",
			addList:       []string{"qux", "fix"},
			reqAttrs:      []string{"foo", "bar", "baz"},
			expectedAttrs: []string{"foo", "bar", "baz", "qux", "fix"},
		},
		{
			name:          "should succeed, add elements to empty list",
			addList:       []string{"qux", "fix"},
			reqAttrs:      []string{},
			expectedAttrs: []string{"qux", "fix"},
		},
		{
			name:          "should succeed, nothing added",
			addList:       []string{},
			reqAttrs:      []string{"foo", "bar", "baz"},
			expectedAttrs: []string{"foo", "bar", "baz"},
		},
		{
			name:          "should succeed, two empty lists",
			addList:       []string{},
			reqAttrs:      []string{},
			expectedAttrs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualAttrs, err := AddToRequiredAttributes(tt.addList, tt.reqAttrs)
			if len(tt.expectedError) == 0 {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedAttrs, actualAttrs)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Nil(t, tt.expectedAttrs)
			}
		})
	}
}

func TestRemovesFromRequiredAttributes(t *testing.T) {
	tests := []struct {
		name          string
		currentAttrs  []string
		removeAttrs   []string
		expectedAttrs []string
		expectedError string
	}{
		{
			name:          "should succeed, removing a single element",
			currentAttrs:  []string{"foo", "bar", "baz"},
			removeAttrs:   []string{"bar"},
			expectedAttrs: []string{"foo", "baz"},
		},
		{
			name:          "should fail, element doesn't exist",
			currentAttrs:  []string{"foo", "bar", "baz"},
			removeAttrs:   []string{"qux"},
			expectedAttrs: nil,
			expectedError: `attribute "qux" is already not required`,
		},
		{
			name:          "should succeed, removing multiple elements",
			currentAttrs:  []string{"foo", "bar", "baz"},
			removeAttrs:   []string{"foo", "baz"},
			expectedAttrs: []string{"bar"},
		},
		{
			name:          "should succeed, removing no elements",
			currentAttrs:  []string{"foo", "bar", "baz"},
			removeAttrs:   []string{},
			expectedAttrs: []string{"foo", "bar", "baz"},
		},
		{
			name:          "should succeed, remove all elements",
			currentAttrs:  []string{"foo", "bar", "baz"},
			removeAttrs:   []string{"baz", "foo", "bar"},
			expectedAttrs: []string{},
		},
		{
			name:          "should succeed, both empty lists",
			currentAttrs:  []string{},
			removeAttrs:   []string{},
			expectedAttrs: []string{},
		},
		{
			name:          "should fail, trying to remove elements from empty list",
			currentAttrs:  []string{},
			removeAttrs:   []string{"blah"},
			expectedAttrs: []string{},
			expectedError: `attribute "blah" is already not required`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualAttrs, err := RemoveFromRequiredAttributes(tt.currentAttrs, tt.removeAttrs)
			if len(tt.expectedError) == 0 {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expectedAttrs, actualAttrs)
			} else {
				assert.NotNil(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			}
		})
	}
}

func TestNetAssetValueConstructor(t *testing.T) {
	price := sdk.NewInt64Coin("jackthecat", 406)
	volume := uint64(100)
	actual := NewNetAssetValue(price, volume)
	assert.Equal(t, price, actual.Price)
	assert.Equal(t, volume, actual.Volume)
	assert.Equal(t, uint64(0), actual.UpdatedBlockHeight, "update time should not be set")
}

func TestNetAssetValueValidate(t *testing.T) {
	tests := []struct {
		name   string
		nav    NetAssetValue
		expErr string
	}{
		{
			name: "invalid denom",
			nav: NetAssetValue{
				Volume: 406,
			},
			expErr: "invalid denom: ",
		},
		{
			name: "volume is not positive",
			nav: NetAssetValue{
				Price:  sdk.NewInt64Coin("jackthecat", 420),
				Volume: 0,
			},
			expErr: "marker net asset value volume must be positive value",
		},
		{
			name: "volume must be positive if value is greater than 1",
			nav: NetAssetValue{
				Price:  sdk.NewInt64Coin("usdmills", 1),
				Volume: 0,
			},
			expErr: "marker net asset value volume must be positive value",
		},
		{
			name: "successful with 0 volume and coin",
			nav: NetAssetValue{
				Price:  sdk.NewInt64Coin("usdmills", 0),
				Volume: 0,
			},
		},
		{
			name: "successful",
			nav: NetAssetValue{
				Price:  sdk.NewInt64Coin("jackthecat", 420),
				Volume: 406,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.nav.Validate()
			if len(tt.expErr) > 0 {
				assert.EqualErrorf(t, err, tt.expErr, "NetAssetValue validate expected error")
			} else {
				assert.NoError(t, err, "NetAssetValue validate should have passed")
			}
		})
	}
}

func TestHasAccess(t *testing.T) {
	addrAll := sdk.AccAddress("addrAll_____________")
	addrAllButWithdraw := sdk.AccAddress("addrAllButWithdraw__")
	addrOnlyTransfer := sdk.AccAddress("addrOnlyTransfer____")
	addrMintBurn := sdk.AccAddress("addrMintBurn________")
	addrDup := sdk.AccAddress("addrDup_____________")
	marker := MarkerAccount{
		BaseAccount: authtypes.NewBaseAccountWithAddress(MustGetMarkerAddress("mycooltestcoin")),
		Denom:       "mycooltestcoin",
		AccessControl: []AccessGrant{
			{
				Address: addrAll.String(),
				Permissions: []Access{
					Access_Mint, Access_Burn, Access_Deposit, Access_Withdraw,
					Access_Delete, Access_Admin, Access_Transfer, Access_ForceTransfer,
				},
			},
			{Address: addrDup.String(), Permissions: []Access{Access_Admin}},
			{
				Address: addrAllButWithdraw.String(),
				Permissions: []Access{
					Access_Mint, Access_Burn, Access_Deposit,
					Access_Delete, Access_Admin, Access_Transfer, Access_ForceTransfer,
				},
			},
			{Address: addrOnlyTransfer.String(), Permissions: []Access{Access_Transfer}},
			{Address: addrMintBurn.String(), Permissions: []Access{Access_Mint, Access_Burn}},
			{Address: addrDup.String(), Permissions: []Access{Access_Delete}},
		},
	}

	tests := []struct {
		name   string
		addr   sdk.AccAddress
		role   Access
		expHas bool
	}{
		{name: "address not known", addr: sdk.AccAddress("addrUnknown_________"), role: Access_Admin, expHas: false},
		{name: "address has all other roles", addr: addrAllButWithdraw, role: Access_Withdraw, expHas: false},
		{name: "address only has that role", addr: addrOnlyTransfer, role: Access_Transfer, expHas: true},
		{name: "address only has one role but not that one", addr: addrOnlyTransfer, role: Access_Delete, expHas: false},
		{name: "address has other roles too", addr: addrAllButWithdraw, role: Access_Deposit, expHas: true},
		{name: "address has two roles: first", addr: addrMintBurn, role: Access_Mint, expHas: true},
		{name: "address has two roles: second", addr: addrMintBurn, role: Access_Burn, expHas: true},
		{name: "address has two roles: neither", addr: addrMintBurn, role: Access_Deposit, expHas: false},
		{name: "address in list twice: first has role", addr: addrDup, role: Access_Admin, expHas: true},
		{name: "address in list twice: second has role", addr: addrDup, role: Access_Delete, expHas: true},
		{name: "address in list twice: neither has role", addr: addrDup, role: Access_Mint, expHas: false},
		{name: "address has all: unknown", addr: addrAll, role: Access_Unknown, expHas: false},
		{name: "address has all: mint", addr: addrAll, role: Access_Mint, expHas: true},
		{name: "address has all: burn", addr: addrAll, role: Access_Burn, expHas: true},
		{name: "address has all: deposit", addr: addrAll, role: Access_Deposit, expHas: true},
		{name: "address has all: withdraw", addr: addrAll, role: Access_Withdraw, expHas: true},
		{name: "address has all: delete", addr: addrAll, role: Access_Delete, expHas: true},
		{name: "address has all: admin", addr: addrAll, role: Access_Admin, expHas: true},
		{name: "address has all: transfer", addr: addrAll, role: Access_Transfer, expHas: true},
		{name: "address has all: force transfer", addr: addrAll, role: Access_ForceTransfer, expHas: true},
	}

	for _, tc := range tests {
		var expErr string
		if !tc.expHas {
			expErr = fmt.Sprintf("%s does not have %s on %s marker (%s)", tc.addr, tc.role, marker.Denom, marker.Address)
		}

		t.Run(tc.name+": HasAccess", func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = marker.HasAccess(tc.addr.String(), tc.role)
			}
			require.NotPanics(t, testFunc, "HasAccess(%s, %s)", string(tc.addr), tc.role)
			assert.Equal(t, tc.expHas, actual, "HasAccess(%s, %s) result", string(tc.addr), tc.role)
		})

		t.Run(tc.name+": ValidateHasAccess", func(t *testing.T) {
			var err error
			testFunc := func() {
				err = marker.ValidateHasAccess(tc.addr.String(), tc.role)
			}
			require.NotPanics(t, testFunc, "ValidateHasAccess(%s, %s)", string(tc.addr), tc.role)
			assertions.AssertErrorValue(t, err, expErr, "ValidateHasAccess(%s, %s) error", string(tc.addr), tc.role)
		})

		t.Run(tc.name+": AddressHasAccess", func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = marker.AddressHasAccess(tc.addr, tc.role)
			}
			require.NotPanics(t, testFunc, "AddressHasAccess(%s, %s)", string(tc.addr), tc.role)
			assert.Equal(t, tc.expHas, actual, "AddressHasAccess(%s, %s) result", string(tc.addr), tc.role)
		})

		t.Run(tc.name+": ValidateAddressHasAccess", func(t *testing.T) {
			var err error
			testFunc := func() {
				err = marker.ValidateAddressHasAccess(tc.addr, tc.role)
			}
			require.NotPanics(t, testFunc, "ValidateAddressHasAccess(%s, %s)", string(tc.addr), tc.role)
			assertions.AssertErrorValue(t, err, expErr, "ValidateAddressHasAccess(%s, %s) error", string(tc.addr), tc.role)
		})
	}
}

func TestValidateAtLeastOneAddrHasAccess(t *testing.T) {
	addr1 := sdk.AccAddress("1_addr______________")
	addr2 := sdk.AccAddress("2_addr______________")
	addr3 := sdk.AccAddress("3_addr______________")

	denom := "moomoo"
	ma := &MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{Address: MustGetMarkerAddress(denom).String()},
		Denom:       denom,
		AccessControl: []AccessGrant{
			{Address: addr1.String(), Permissions: AccessList{Access_Mint, Access_Burn}},
			{Address: addr2.String(), Permissions: AccessList{Access_Deposit, Access_Withdraw}},
			{Address: addr3.String(), Permissions: AccessList{Access_Transfer, Access_Admin, Access_Delete}},
		},
	}
	markerDesc := denom + " marker (" + ma.BaseAccount.Address + ")"

	addrOther1 := sdk.AccAddress("one_other_addr______")
	addrOther2 := sdk.AccAddress("two_other_addr______")
	addrOther3 := sdk.AccAddress("three_other_addr____")

	tests := []struct {
		name  string
		addrs []sdk.AccAddress
		role  Access
		exp   string
	}{
		{
			name:  "nil addrs",
			addrs: nil,
			role:  Access_Mint,
			exp:   "none of [] have permission ACCESS_MINT on " + markerDesc,
		},
		{
			name:  "empty addrs",
			addrs: []sdk.AccAddress{},
			role:  Access_Mint,
			exp:   "none of [] have permission ACCESS_MINT on " + markerDesc,
		},
		{
			name:  "one addr: no perms",
			addrs: []sdk.AccAddress{addrOther1},
			role:  Access_Burn,
			exp:   addrOther1.String() + " does not have ACCESS_BURN on " + markerDesc,
		},
		{
			name:  "one addr: unknown role",
			addrs: []sdk.AccAddress{addr1},
			role:  55,
			exp:   addr1.String() + " does not have 55 on " + markerDesc,
		},
		{
			name:  "one addr: other role",
			addrs: []sdk.AccAddress{addr1},
			role:  Access_ForceTransfer,
			exp:   addr1.String() + " does not have ACCESS_FORCE_TRANSFER on " + markerDesc,
		},
		{
			name:  "one addr: has role",
			addrs: []sdk.AccAddress{addr1},
			role:  Access_Mint,
			exp:   "",
		},
		{
			name:  "three addrs: no match",
			addrs: []sdk.AccAddress{addrOther1, addrOther2, addrOther3},
			role:  Access_Withdraw,
			exp: "none of [\"" + addrOther1.String() + "\" \"" + addrOther2.String() + "\" \"" +
				addrOther3.String() + "\"] have permission ACCESS_WITHDRAW on " + markerDesc,
		},
		{
			name:  "three addrs: match first",
			addrs: []sdk.AccAddress{addr1, addrOther2, addrOther3},
			role:  Access_Burn,
			exp:   "",
		},
		{
			name:  "three addrs: match second",
			addrs: []sdk.AccAddress{addrOther1, addr2, addrOther3},
			role:  Access_Deposit,
			exp:   "",
		},
		{
			name:  "three addrs: match third",
			addrs: []sdk.AccAddress{addrOther1, addrOther2, addr3},
			role:  Access_Admin,
			exp:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expIs := len(tc.exp) == 0
			var actIs bool
			testIsFunc := func() {
				actIs = AtLeastOneAddrHasAccess(ma, tc.addrs, tc.role)
			}
			if assert.NotPanics(t, testIsFunc, "AtLeastOneAddrHasAccess") {
				assert.Equal(t, expIs, actIs, "result from AtLeastOneAddrHasAccess")
			}

			var err error
			testValFunc := func() {
				err = ValidateAtLeastOneAddrHasAccess(ma, tc.addrs, tc.role)
			}
			if assert.NotPanics(t, testValFunc, "ValidateAtLeastOneAddrHasAccess") {
				assertions.AssertErrorValue(t, err, tc.exp, "result from ValidateAtLeastOneAddrHasAccess")
			}
		})
	}
}
