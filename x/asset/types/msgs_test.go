package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/testutil/assertions"

	. "github.com/provenance-io/provenance/x/asset/types"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgBurnAsset{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgCreateAsset{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgCreateAssetClass{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgCreatePool{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgCreateTokenization{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgCreateSecuritization{Signer: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}

func TestMsgBurnAsset_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgBurnAsset
		expErr string
	}{
		{
			name: "valid message",
			msg: MsgBurnAsset{
				Asset: &AssetKey{
					ClassId: "class-id",
					Id:      "asset-id",
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "",
		},
		{
			name: "nil asset",
			msg: MsgBurnAsset{
				Asset:  nil,
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset: asset_key cannot be nil: invalid field",
		},
		{
			name: "empty class id",
			msg: MsgBurnAsset{
				Asset: &AssetKey{
					ClassId: "",
					Id:      "asset-id",
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset: invalid class_id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "empty id",
			msg: MsgBurnAsset{
				Asset: &AssetKey{
					ClassId: "class-id",
					Id:      "",
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset: invalid id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "asset validation failure",
			msg: MsgBurnAsset{
				Asset: &AssetKey{
					ClassId: "class-id",
					Id:      "",
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset: invalid id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "empty signer",
			msg: MsgBurnAsset{
				Asset: &AssetKey{
					ClassId: "class-id",
					Id:      "asset-id",
				},
				Signer: "",
			},
			expErr: "invalid signer: empty address string is not allowed: invalid field",
		},
		{
			name: "invalid signer",
			msg: MsgBurnAsset{
				Asset: &AssetKey{
					ClassId: "class-id",
					Id:      "asset-id",
				},
				Signer: "invalid-address",
			},
			expErr: "invalid signer: decoding bech32 failed: invalid separator index -1: invalid field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			assertions.AssertErrorValue(t, err, tt.expErr, "MsgBurnAsset.ValidateBasic()")
		})
	}
}

func TestMsgCreateAsset_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCreateAsset
		expErr string
	}{
		{
			name: "valid message",
			msg: MsgCreateAsset{
				Asset: &Asset{
					ClassId: "test-class",
					Id:      "test-id",
				},
				Owner:  "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "",
		},
		{
			name: "nil asset",
			msg: MsgCreateAsset{
				Asset:  nil,
				Owner:  "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset: asset cannot be nil: invalid field",
		},
		{
			name: "empty class_id",
			msg: MsgCreateAsset{
				Asset: &Asset{
					ClassId: "",
					Id:      "test-id",
				},
				Owner:  "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset: invalid class_id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "empty id",
			msg: MsgCreateAsset{
				Asset: &Asset{
					ClassId: "test-class",
					Id:      "",
				},
				Owner:  "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset: invalid id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "empty owner",
			msg: MsgCreateAsset{
				Asset: &Asset{
					ClassId: "test-class",
					Id:      "test-id",
				},
				Owner:  "",
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid owner: empty address string is not allowed: invalid field",
		},
		{
			name: "invalid owner",
			msg: MsgCreateAsset{
				Asset: &Asset{
					ClassId: "test-class",
					Id:      "test-id",
				},
				Owner:  "invalid-address",
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid owner: decoding bech32 failed: invalid separator index -1: invalid field",
		},
		{
			name: "empty signer",
			msg: MsgCreateAsset{
				Asset: &Asset{
					ClassId: "test-class",
					Id:      "test-id",
				},
				Owner:  "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
				Signer: "",
			},
			expErr: "invalid signer: empty address string is not allowed: invalid field",
		},
		{
			name: "invalid signer",
			msg: MsgCreateAsset{
				Asset: &Asset{
					ClassId: "test-class",
					Id:      "test-id",
				},
				Owner:  "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
				Signer: "invalid-address",
			},
			expErr: "invalid signer: decoding bech32 failed: invalid separator index -1: invalid field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			assertions.AssertErrorValue(t, err, tt.expErr, "MsgCreateAsset.ValidateBasic()")
		})
	}
}

func TestMsgCreateAssetClass_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCreateAssetClass
		expErr string
	}{
		{
			name: "valid message",
			msg: MsgCreateAssetClass{
				AssetClass: &AssetClass{
					Id:   "test-class",
					Name: "Test Class",
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "",
		},
		{
			name: "nil asset class",
			msg: MsgCreateAssetClass{
				AssetClass: nil,
				Signer:     "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset_class: asset_class cannot be nil: invalid field",
		},
		{
			name: "empty id",
			msg: MsgCreateAssetClass{
				AssetClass: &AssetClass{
					Id:   "",
					Name: "Test Class",
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset_class: invalid id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "empty name",
			msg: MsgCreateAssetClass{
				AssetClass: &AssetClass{
					Id:   "test-class",
					Name: "",
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset_class: invalid name: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "empty signer",
			msg: MsgCreateAssetClass{
				AssetClass: &AssetClass{
					Id:   "test-class",
					Name: "Test Class",
				},
				Signer: "",
			},
			expErr: "invalid signer: empty address string is not allowed: invalid field",
		},
		{
			name: "invalid signer",
			msg: MsgCreateAssetClass{
				AssetClass: &AssetClass{
					Id:   "test-class",
					Name: "Test Class",
				},
				Signer: "invalid-address",
			},
			expErr: "invalid signer: decoding bech32 failed: invalid separator index -1: invalid field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			assertions.AssertErrorValue(t, err, tt.expErr, "MsgCreateAssetClass.ValidateBasic()")
		})
	}
}

func TestMsgCreatePool_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCreatePool
		expErr string
	}{
		{
			name: "valid message",
			msg: MsgCreatePool{
				Pool: sdk.Coin{
					Denom:  "pool",
					Amount: sdkmath.NewInt(1000),
				},
				Assets: []*AssetKey{
					{
						ClassId: "test-class",
						Id:      "test-id",
					},
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "",
		},
		{
			name: "invalid pool coin",
			msg: MsgCreatePool{
				Pool: sdk.Coin{
					Denom:  "",
					Amount: sdkmath.NewInt(1000),
				},
				Assets: []*AssetKey{
					{
						ClassId: "test-class",
						Id:      "test-id",
					},
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid pool: invalid denom: : invalid field",
		},
		{
			name: "empty assets",
			msg: MsgCreatePool{
				Pool: sdk.Coin{
					Denom:  "pool",
					Amount: sdkmath.NewInt(1000),
				},
				Assets: []*AssetKey{},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid assets: cannot be empty: invalid field",
		},
		{
			name: "nil asset",
			msg: MsgCreatePool{
				Pool: sdk.Coin{
					Denom:  "pool",
					Amount: sdkmath.NewInt(1000),
				},
				Assets: []*AssetKey{
					nil,
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid assets[0]: asset_key cannot be nil: invalid field",
		},
		{
			name: "asset with empty class id",
			msg: MsgCreatePool{
				Pool: sdk.Coin{
					Denom:  "pool",
					Amount: sdkmath.NewInt(1000),
				},
				Assets: []*AssetKey{
					{
						ClassId: "",
						Id:      "test-id",
					},
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid assets[0]: invalid class_id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "asset with empty id",
			msg: MsgCreatePool{
				Pool: sdk.Coin{
					Denom:  "pool",
					Amount: sdkmath.NewInt(1000),
				},
				Assets: []*AssetKey{
					{
						ClassId: "test-class",
						Id:      "",
					},
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid assets[0]: invalid id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "duplicate assets",
			msg: MsgCreatePool{
				Pool: sdk.Coin{
					Denom:  "pool",
					Amount: sdkmath.NewInt(1000),
				},
				Assets: []*AssetKey{
					{
						ClassId: "test-class",
						Id:      "test-id",
					},
					{
						ClassId: "test-class",
						Id:      "test-id",
					},
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid assets: duplicate asset at index 0 and 1: invalid field",
		},
		{
			name: "empty signer",
			msg: MsgCreatePool{
				Pool: sdk.Coin{
					Denom:  "pool",
					Amount: sdkmath.NewInt(1000),
				},
				Assets: []*AssetKey{
					{
						ClassId: "test-class",
						Id:      "test-id",
					},
				},
				Signer: "",
			},
			expErr: "invalid signer: empty address string is not allowed: invalid field",
		},
		{
			name: "invalid signer",
			msg: MsgCreatePool{
				Pool: sdk.Coin{
					Denom:  "pool",
					Amount: sdkmath.NewInt(1000),
				},
				Assets: []*AssetKey{
					{
						ClassId: "test-class",
						Id:      "test-id",
					},
				},
				Signer: "invalid-address",
			},
			expErr: "invalid signer: decoding bech32 failed: invalid separator index -1: invalid field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			assertions.AssertErrorValue(t, err, tt.expErr, "MsgCreatePool.ValidateBasic()")
		})
	}
}

func TestMsgCreateTokenization_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCreateTokenization
		expErr string
	}{
		{
			name: "valid message",
			msg: MsgCreateTokenization{
				Token: sdk.Coin{
					Denom:  "tokenization",
					Amount: sdkmath.NewInt(1000),
				},
				Asset: &AssetKey{
					ClassId: "test-class",
					Id:      "test-id",
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "",
		},
		{
			name: "invalid token denom",
			msg: MsgCreateTokenization{
				Token: sdk.Coin{
					Denom:  "",
					Amount: sdkmath.NewInt(1000),
				},
				Asset: &AssetKey{
					ClassId: "test-class",
					Id:      "test-id",
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid token: invalid denom: : invalid field",
		},
		{
			name: "negative token amount",
			msg: MsgCreateTokenization{
				Token: sdk.Coin{
					Denom:  "tokenization",
					Amount: sdkmath.NewInt(-1000),
				},
				Asset: &AssetKey{
					ClassId: "test-class",
					Id:      "test-id",
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid token: negative coin amount: -1000: invalid field",
		},
		{
			name: "nil asset",
			msg: MsgCreateTokenization{
				Token: sdk.Coin{
					Denom:  "tokenization",
					Amount: sdkmath.NewInt(1000),
				},
				Asset:  nil,
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset: asset_key cannot be nil: invalid field",
		},
		{
			name: "empty asset class_id",
			msg: MsgCreateTokenization{
				Token: sdk.Coin{
					Denom:  "tokenization",
					Amount: sdkmath.NewInt(1000),
				},
				Asset: &AssetKey{
					ClassId: "",
					Id:      "test-id",
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset: invalid class_id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "empty asset id",
			msg: MsgCreateTokenization{
				Token: sdk.Coin{
					Denom:  "tokenization",
					Amount: sdkmath.NewInt(1000),
				},
				Asset: &AssetKey{
					ClassId: "test-class",
					Id:      "",
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid asset: invalid id: must be between 1 and 128 characters: invalid field",
		},
		{
			name: "empty signer",
			msg: MsgCreateTokenization{
				Token: sdk.Coin{
					Denom:  "tokenization",
					Amount: sdkmath.NewInt(1000),
				},
				Asset: &AssetKey{
					ClassId: "test-class",
					Id:      "test-id",
				},
				Signer: "",
			},
			expErr: "invalid signer: empty address string is not allowed: invalid field",
		},
		{
			name: "invalid signer",
			msg: MsgCreateTokenization{
				Token: sdk.Coin{
					Denom:  "tokenization",
					Amount: sdkmath.NewInt(1000),
				},
				Asset: &AssetKey{
					ClassId: "test-class",
					Id:      "test-id",
				},
				Signer: "invalid-address",
			},
			expErr: "invalid signer: decoding bech32 failed: invalid separator index -1: invalid field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			assertions.AssertErrorValue(t, err, tt.expErr, "MsgCreateTokenization.ValidateBasic()")
		})
	}
}

func TestMsgCreateSecuritization_ValidateBasic(t *testing.T) {
	tests := []struct {
		name   string
		msg    MsgCreateSecuritization
		expErr string
	}{
		{
			name: "valid message",
			msg: MsgCreateSecuritization{
				Id: "test-sec",
				Pools: []string{
					"pool1",
					"pool2",
				},
				Tranches: []*sdk.Coin{
					{
						Denom:  "tranche1",
						Amount: sdkmath.NewInt(1000),
					},
					{
						Denom:  "tranche2",
						Amount: sdkmath.NewInt(2000),
					},
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "",
		},
		{
			name: "empty id",
			msg: MsgCreateSecuritization{
				Id: "",
				Pools: []string{
					"pool1",
				},
				Tranches: []*sdk.Coin{
					{
						Denom:  "tranche1",
						Amount: sdkmath.NewInt(1000),
					},
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid id: cannot be empty: invalid field",
		},
		{
			name: "empty pools",
			msg: MsgCreateSecuritization{
				Id:    "test-sec",
				Pools: []string{},
				Tranches: []*sdk.Coin{
					{
						Denom:  "tranche1",
						Amount: sdkmath.NewInt(1000),
					},
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid pools: cannot be empty: invalid field",
		},
		{
			name: "empty pool string",
			msg: MsgCreateSecuritization{
				Id: "test-sec",
				Pools: []string{
					"pool1",
					"",
				},
				Tranches: []*sdk.Coin{
					{
						Denom:  "tranche1",
						Amount: sdkmath.NewInt(1000),
					},
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid pools[1]: cannot be empty: invalid field",
		},
		{
			name: "empty tranches",
			msg: MsgCreateSecuritization{
				Id: "test-sec",
				Pools: []string{
					"pool1",
				},
				Tranches: []*sdk.Coin{},
				Signer:   "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid tranches: cannot be empty: invalid field",
		},
		{
			name: "nil tranche",
			msg: MsgCreateSecuritization{
				Id: "test-sec",
				Pools: []string{
					"pool1",
				},
				Tranches: []*sdk.Coin{
					nil,
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid tranches[0]: cannot be nil: invalid field",
		},
		{
			name: "invalid tranche coin",
			msg: MsgCreateSecuritization{
				Id: "test-sec",
				Pools: []string{
					"pool1",
				},
				Tranches: []*sdk.Coin{
					{
						Denom:  "",
						Amount: sdkmath.NewInt(1000),
					},
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			expErr: "invalid tranches[0]: invalid denom: : invalid field",
		},
		{
			name: "empty signer",
			msg: MsgCreateSecuritization{
				Id: "test-sec",
				Pools: []string{
					"pool1",
				},
				Tranches: []*sdk.Coin{
					{
						Denom:  "tranche1",
						Amount: sdkmath.NewInt(1000),
					},
				},
				Signer: "",
			},
			expErr: "invalid signer: empty address string is not allowed: invalid field",
		},
		{
			name: "invalid signer",
			msg: MsgCreateSecuritization{
				Id: "test-sec",
				Pools: []string{
					"pool1",
				},
				Tranches: []*sdk.Coin{
					{
						Denom:  "tranche1",
						Amount: sdkmath.NewInt(1000),
					},
				},
				Signer: "invalid-address",
			},
			expErr: "invalid signer: decoding bech32 failed: invalid separator index -1: invalid field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			assertions.AssertErrorValue(t, err, tt.expErr, "MsgCreateSecuritization.ValidateBasic()")
		})
	}
}
