package keeper_test

import (
	"testing"

	"github.com/google/uuid"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func FreshCtx(app *simapp.App) sdk.Context {
	return keeper.AddAuthzCacheToContext(app.BaseApp.NewContext(false, tmproto.Header{}))
}

func TestValidateSetAccountData(t *testing.T) {
	app := simapp.Setup(t)

	uuid1 := uuid.MustParse("F55CB94D-02C2-40A5-BB19-ED36DF322A33")
	uuid2 := uuid.MustParse("304BF4FB-0C3F-4B0B-BB01-4EAA615123D6")
	recordName := "arecord"

	newMsg := func(addr types.MetadataAddress) *types.MsgSetAccountDataRequest {
		return &types.MsgSetAccountDataRequest{
			MetadataAddr: addr,
			Value:        "Some value.",
			Signers:      []string{sdk.AccAddress("addr________________").String()},
		}
	}

	errUnsupported := func(pre string) string {
		return "unsupported metadata address type: " + pre
	}

	tests := []struct {
		name string
		msg  *types.MsgSetAccountDataRequest
		exp  string
	}{
		{
			name: "no metadata address",
			msg:  newMsg(nil),
			exp:  errUnsupported(""),
		},
		{
			name: "scope id",
			msg:  newMsg(types.ScopeMetadataAddress(uuid1)),
			exp:  "scope not found with id " + types.ScopeMetadataAddress(uuid1).String(),
		},
		{
			name: "session id",
			msg:  newMsg(types.SessionMetadataAddress(uuid1, uuid2)),
			exp:  errUnsupported(types.PrefixSession),
		},
		{
			name: "record id",
			msg:  newMsg(types.RecordMetadataAddress(uuid1, recordName)),
			exp:  errUnsupported(types.PrefixRecord),
		},
		{
			name: "scope spec id",
			msg:  newMsg(types.ScopeSpecMetadataAddress(uuid1)),
			exp:  errUnsupported(types.PrefixScopeSpecification),
		},
		{
			name: "contract spec id",
			msg:  newMsg(types.ContractSpecMetadataAddress(uuid1)),
			exp:  errUnsupported(types.PrefixContractSpecification),
		},
		{
			name: "record spec id",
			msg:  newMsg(types.RecordSpecMetadataAddress(uuid1, recordName)),
			exp:  errUnsupported(types.PrefixRecordSpecification),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := app.MetadataKeeper.ValidateSetAccountData(FreshCtx(app), tc.msg)
			AssertErrorValue(t, err, tc.exp, "ValidateSetAccountData")
		})
	}
}
