package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil"

	. "github.com/provenance-io/provenance/x/asset/types"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgCreateAsset{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgCreateAssetClass{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgCreatePool{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgCreateTokenization{Signer: signer} },
		func(signer string) sdk.Msg { return &MsgCreateSecuritization{Signer: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}

func TestMsgCreateAsset_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     MsgCreateAsset
		wantErr bool
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
			wantErr: false,
		},
		{
			name: "nil asset",
			msg: MsgCreateAsset{
				Asset:  nil,
				Owner:  "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: true,
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
			wantErr: true,
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
			wantErr: true,
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
			wantErr: true,
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
			wantErr: true,
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
			wantErr: true,
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
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreateAssetClass_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     MsgCreateAssetClass
		wantErr bool
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
			wantErr: false,
		},
		{
			name: "nil asset class",
			msg: MsgCreateAssetClass{
				AssetClass: nil,
				Signer:     "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: true,
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
			wantErr: true,
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
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreatePool_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     MsgCreatePool
		wantErr bool
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
			wantErr: false,
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
			wantErr: true,
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
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreateTokenization_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     MsgCreateTokenization
		wantErr bool
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
			wantErr: false,
		},
		{
			name: "invalid denom",
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
			wantErr: true,
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
			wantErr: true,
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
			wantErr: true,
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
			wantErr: true,
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
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreateSecuritization_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     MsgCreateSecuritization
		wantErr bool
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
			wantErr: false,
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
			wantErr: true,
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
			wantErr: true,
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
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
