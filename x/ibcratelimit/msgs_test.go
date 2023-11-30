package ibcratelimit_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/ibcratelimit"
	"github.com/stretchr/testify/assert"
)

func TestNewMsgGovUpdateParamsRequest(t *testing.T) {
	expected := &ibcratelimit.MsgGovUpdateParamsRequest{
		Authority: "authority",
		Params:    ibcratelimit.NewParams("contract"),
	}
	event := ibcratelimit.NewMsgGovUpdateParamsRequest(expected.Authority, expected.Params.ContractAddress)
	assert.Equal(t, expected, event, "should create the correct with correct content")
}

func TestNewMsgGovUpdateParamsValidateBasic(t *testing.T) {
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
			msg := ibcratelimit.NewMsgGovUpdateParamsRequest(tc.authority, tc.contract)
			err := msg.ValidateBasic()

			if len(tc.err) > 0 {
				assert.EqualError(t, err, tc.err, "should return correct error")
			} else {
				assert.NoError(t, err, "should not throw an error")
			}
		})
	}
}

func TestMsgGovUpdateParamsRequestGetSigners(t *testing.T) {
	tests := []struct {
		name      string
		authority string
		err       string
	}{
		{
			name:      "success - valid signer",
			authority: "cosmos1qm0hhug8kszhcp9f3ryuecz5yw8s3e5v0n2ckd",
		},
		{
			name:      "failure - missing signer",
			authority: "",
			err:       "empty address string is not allowed",
		},
		{
			name:      "failure - invalid signer",
			authority: "authority",
			err:       "decoding bech32 failed: invalid separator index -1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := ibcratelimit.NewMsgGovUpdateParamsRequest(tc.authority, "contract")

			if len(tc.err) > 0 {
				assertions.RequirePanicEquals(t, func() {
					msg.GetSigners()
				}, tc.err, "should panic with correct message")
			} else {
				signers := make([]sdk.AccAddress, 1)
				assert.NotPanics(t, func() {
					signers = msg.GetSigners()
				}, "should not panic")
				assert.Equal(t, signers[0].String(), tc.authority)
			}
		})
	}
}
