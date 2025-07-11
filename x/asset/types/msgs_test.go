package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

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
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: false,
		},
		{
			name: "nil asset",
			msg: MsgCreateAsset{
				Asset:       nil,
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
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
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
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
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: true,
		},
		{
			name: "empty from_address",
			msg: MsgCreateAsset{
				Asset: &Asset{
					ClassId: "test-class",
					Id:      "test-id",
				},
				FromAddress: "",
			},
			wantErr: true,
		},
		{
			name: "invalid from_address",
			msg: MsgCreateAsset{
				Asset: &Asset{
					ClassId: "test-class",
					Id:      "test-id",
				},
				FromAddress: "invalid-address",
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
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: false,
		},
		{
			name: "nil asset class",
			msg: MsgCreateAssetClass{
				AssetClass:  nil,
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
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
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
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
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
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
				Pool: &sdk.Coin{
					Denom:  "pool",
					Amount: sdkmath.NewInt(1000),
				},
				Nfts: []*Nft{
					{
						ClassId: "test-class",
						Id:      "test-id",
					},
				},
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: false,
		},
		{
			name: "nil pool",
			msg: MsgCreatePool{
				Pool: nil,
				Nfts: []*Nft{
					{
						ClassId: "test-class",
						Id:      "test-id",
					},
				},
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: true,
		},
		{
			name: "empty nfts",
			msg: MsgCreatePool{
				Pool: &sdk.Coin{
					Denom:  "pool",
					Amount: sdkmath.NewInt(1000),
				},
				Nfts:        []*Nft{},
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: true,
		},
		{
			name: "nil nft",
			msg: MsgCreatePool{
				Pool: &sdk.Coin{
					Denom:  "pool",
					Amount: sdkmath.NewInt(1000),
				},
				Nfts: []*Nft{
					nil,
				},
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
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
				Denom: sdk.Coin{
					Denom:  "tokenization",
					Amount: sdkmath.NewInt(1000),
				},
				Nft: &Nft{
					ClassId: "test-class",
					Id:      "test-id",
				},
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: false,
		},
		{
			name: "invalid denom",
			msg: MsgCreateTokenization{
				Denom: sdk.Coin{
					Denom:  "",
					Amount: sdkmath.NewInt(1000),
				},
				Nft: &Nft{
					ClassId: "test-class",
					Id:      "test-id",
				},
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: true,
		},
		{
			name: "nil nft",
			msg: MsgCreateTokenization{
				Denom: sdk.Coin{
					Denom:  "tokenization",
					Amount: sdkmath.NewInt(1000),
				},
				Nft:         nil,
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: true,
		},
		{
			name: "empty nft class_id",
			msg: MsgCreateTokenization{
				Denom: sdk.Coin{
					Denom:  "tokenization",
					Amount: sdkmath.NewInt(1000),
				},
				Nft: &Nft{
					ClassId: "",
					Id:      "test-id",
				},
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: true,
		},
		{
			name: "empty nft id",
			msg: MsgCreateTokenization{
				Denom: sdk.Coin{
					Denom:  "tokenization",
					Amount: sdkmath.NewInt(1000),
				},
				Nft: &Nft{
					ClassId: "test-class",
					Id:      "",
				},
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: true,
		},
		{
			name: "empty from_address",
			msg: MsgCreateTokenization{
				Denom: sdk.Coin{
					Denom:  "tokenization",
					Amount: sdkmath.NewInt(1000),
				},
				Nft: &Nft{
					ClassId: "test-class",
					Id:      "test-id",
				},
				FromAddress: "",
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
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
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
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
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
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
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
				Tranches:    []*sdk.Coin{},
				FromAddress: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
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