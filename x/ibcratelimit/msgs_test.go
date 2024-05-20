package ibcratelimit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil"

	. "github.com/provenance-io/provenance/x/ibcratelimit"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgGovUpdateParamsRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgUpdateParamsRequest{Authority: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}

func TestNewMsgGovUpdateParamsRequest(t *testing.T) {
	expected := &MsgUpdateParamsRequest{
		Authority: "authority",
		Params:    NewParams("contract"),
	}
	event := NewUpdateParamsRequest(expected.Authority, expected.Params.ContractAddress)
	assert.Equal(t, expected, event, "should create the correct with correct content")
}

func TestNewMsgUpdateParamsValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		authority string
		contract  string
		err       string
	}{
		{
			name:      "success - valid message",
			authority: "cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd",
			contract:  "cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd",
		},
		{
			name:      "success - empty contract",
			authority: "cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd",
			contract:  "",
		},
		{
			name:      "failure - invalid authority",
			authority: "authority",
			contract:  "cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd",
			err:       "invalid authority: decoding bech32 failed: invalid separator index -1",
		},
		{
			name:      "failure - invalid contract",
			authority: "cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd",
			contract:  "contract",
			err:       "decoding bech32 failed: invalid separator index -1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := NewUpdateParamsRequest(tc.authority, tc.contract)
			err := msg.ValidateBasic()

			if len(tc.err) > 0 {
				assert.EqualError(t, err, tc.err, "should return correct error")
			} else {
				assert.NoError(t, err, "should not throw an error")
			}
		})
	}
}
