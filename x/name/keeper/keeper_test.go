package keeper_test

import (
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/name/keeper"
	"github.com/provenance-io/provenance/x/name/types"
)

func TestSetWithdrawAddr(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	nameKeeper := keeper.NewKeeper(app.AppCodec(), nil, app.GetSubspace(types.ModuleName), app.AccountKeeper)

	params := nameKeeper.GetParams(ctx)
	params.MaxNameLevels = 16
	params.MinSegmentLength = 2
	params.MaxSegmentLength = 16
	nameKeeper.SetParams(ctx, params)

	normalizeName(t, ctx, nameKeeper)
}

func normalizeName(t *testing.T, ctx sdk.Context, keeper keeper.Keeper) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// Valid names
		{"normalize upper case", args{name: "TEST.NORMALIZE.PIO"}, "test.normalize.pio", false},
		{"trim comp spaces", args{name: "test . normalize. pio "}, "test.normalize.pio", false},
		{"allow single dash per comp", args{name: "test-field.my-service.pio"}, "test-field.my-service.pio", false},
		{"allow digits", args{name: "test.normalize.v1.pio"}, "test.normalize.v1.pio", false},
		{"allow unicode chars", args{name: "tœst.nørmålize.v1.pio"}, "tœst.nørmålize.v1.pio", false},
		// TODO -- this uuid is rejected due to name length constraints, need to resolve.
		// {"allow uuid as comp", args{name: "6443a1e8-ec9b-4ff1-b200-d639424bcba4.service.pb"},
		// 	"6443a1e8-ec9b-4ff1-b200-d639424bcba4.service.pb", false},
		// Invalid names / components
		{"fail on empty name", args{name: ""}, "", true},
		{"fail when too short", args{name: "z"}, "", true},
		{"fail when too long", args{name: "too.looooooooooooooooooooooooooooooooooooooong.pio"}, "", true},
		{"fail on multiple dashes in comp", args{name: "fail-test-field.my-app.pio"}, "", true},
		{"fail on non-printable chars", args{name: "test.normalize" + string([]byte{0x01}) + ".pio"}, "", true},
		{"fail on too many components", args{name: "ab.bc.cd.de.ef.fg.gh.hi.ij.jk.kl.lm.mn.no.op.pq.qr"}, "", true},
		{"fail on unsupported chars", args{name: "fail_normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail!normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail|normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail,normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail~normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail*normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail&normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail^normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail@normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail#normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail=normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail+normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail`normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail%normalize.pio"}, "", true},
		{"fail on invalid uuid", args{name: "6443a1e8-ec9b-4ff1-b200-d639424bcba4-deadbeef.service.pb"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := keeper.Normalize(ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("normalizeName() = %v, want %v", got, tt.want)
			}
		})
	}
}
