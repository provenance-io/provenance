package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil"

	. "github.com/provenance-io/provenance/x/oracle/types"
)

func TestAllMsgsGetSigners(t *testing.T) {
	msgMakers := []testutil.MsgMaker{
		func(signer string) sdk.Msg { return &MsgUpdateOracleRequest{Authority: signer} },
		func(signer string) sdk.Msg { return &MsgSendQueryOracleRequest{Authority: signer} },
	}

	testutil.RunGetSignersTests(t, AllRequestMsgs, msgMakers, nil)
}

func TestNewMsgQueryOracle(t *testing.T) {
	authority := "creator"
	channel := "channel"
	query := []byte{0x01, 0x02, 0x04}

	msg := NewMsgSendQueryOracle(authority, channel, query)
	assert.Equal(t, authority, msg.Authority, "must have the correct authority")
	assert.Equal(t, channel, msg.Channel, "must have the correct channel")
	assert.EqualValues(t, query, msg.Query, "must have the correct query")
}

func TestNewMsgUpdateOracle(t *testing.T) {
	authority := "creator"
	address := "address"

	msg := NewMsgUpdateOracle(authority, address)
	assert.Equal(t, authority, msg.Authority, "must have the correct authority")
	assert.Equal(t, address, msg.Address, "must have the correct address")
}

func TestMsgUpdateOracleRequestValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  *MsgUpdateOracleRequest
		err  string
	}{
		{
			name: "success - all fields are valid",
			msg:  NewMsgUpdateOracle("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
		},
		{
			name: "failure - invalid authority",
			msg:  NewMsgUpdateOracle("jackthecat", "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"),
			err:  "invalid authority address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "failure - invalid address",
			msg:  NewMsgUpdateOracle("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", "jackthecat"),
			err:  "invalid address for oracle: decoding bech32 failed: invalid separator index -1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.msg.ValidateBasic()
			if len(tc.err) > 0 {
				assert.EqualError(t, res, tc.err, "MsgUpdateOracleRequest.ValidateBasic")
			} else {
				assert.NoError(t, res, "MsgUpdateOracleRequest.ValidateBasic")
			}
		})
	}
}

func TestMsgSendQueryOracleRequestValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  *MsgSendQueryOracleRequest
		err  string
	}{
		{
			name: "success - all fields are valid",
			msg:  NewMsgSendQueryOracle("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", "channel-1", []byte("{}")),
		},
		{
			name: "failure - invalid authority",
			msg:  NewMsgSendQueryOracle("jackthecat", "channel-1", []byte("{}")),
			err:  "invalid authority address: decoding bech32 failed: invalid separator index -1",
		},
		{
			name: "failure - invalid channel",
			msg:  NewMsgSendQueryOracle("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", "bad", []byte("{}")),
			err:  "invalid channel id",
		},
		{
			name: "failure - invalid query",
			msg:  NewMsgSendQueryOracle("cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma", "channel-1", []byte{}),
			err:  "invalid query data: invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := tc.msg.ValidateBasic()
			if len(tc.err) > 0 {
				assert.EqualError(t, res, tc.err, "NewMsgSendQueryOracleRequest.ValidateBasic")
			} else {
				assert.NoError(t, res, "NewMsgSendQueryOracleRequest.ValidateBasic")
			}
		})
	}
}
