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
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: false,
		},
		{
			name: "nil asset",
			msg: MsgCreateAsset{
				Asset:  nil,
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
				Pool: &sdk.Coin{
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
			name: "nil pool",
			msg: MsgCreatePool{
				Pool: nil,
				Assets: []*AssetKey{
					{
						ClassId: "test-class",
						Id:      "test-id",
					},
				},
				Signer: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma",
			},
			wantErr: true,
		},
		{
			name: "empty assets",
			msg: MsgCreatePool{
				Pool: &sdk.Coin{
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
				Pool: &sdk.Coin{
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
				Denom: sdk.Coin{
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
				Denom: sdk.Coin{
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
				Denom: sdk.Coin{
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
				Denom: sdk.Coin{
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
				Denom: sdk.Coin{
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
				Denom: sdk.Coin{
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
