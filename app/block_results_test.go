package app

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtstate "github.com/cometbft/cometbft/proto/tendermint/state"
	cmttypes "github.com/cometbft/cometbft/proto/tendermint/types"
)

func TestBlockResults(t *testing.T) {
	event := func(t string, attrs ...abci.EventAttribute) abci.Event {
		return abci.Event{Type: t, Attributes: attrs}
	}
	attr := func(key, value string) abci.EventAttribute {
		return abci.EventAttribute{Key: key, Value: value, Index: true}
	}

	var expBz, actBz []byte

	tests := []struct {
		name       string
		oldEntry   *cmtstate.LegacyABCIResponses
		oldEntryBz []byte
		saveActBz  bool
		saveExpBz  bool
	}{
		{
			name: "one of each, no updates",
			oldEntry: &cmtstate.LegacyABCIResponses{
				DeliverTxs: []*abci.ExecTxResult{
					{
						Code:      1,
						Data:      []byte("somedata"),
						Log:       "the log",
						Info:      "the info",
						GasWanted: 5,
						GasUsed:   4,
						Events: []abci.Event{
							event("txevent1", attr("tx1attr1", "tx1val1"), attr("tx1attr2", "tx1val2")),
						},
						Codespace: "thecodespace",
					},
				},
				EndBlock: &cmtstate.ResponseEndBlock{
					ValidatorUpdates:      nil,
					ConsensusParamUpdates: nil,
					Events: []abci.Event{
						event("endevent1", attr("end1attr1", "end1val1"), attr("end1attr2", "end1val2")),
					},
				},
				BeginBlock: &cmtstate.ResponseBeginBlock{
					Events: []abci.Event{
						event("beginevent1", attr("begin1attr1", "begin1val1"), attr("begin1attr2", "begin1val2")),
					},
				},
			},
		},
		{
			name: "no tx, 1 end, 1 begin, no updates",
			oldEntry: &cmtstate.LegacyABCIResponses{
				DeliverTxs: nil,
				EndBlock: &cmtstate.ResponseEndBlock{
					ValidatorUpdates:      nil,
					ConsensusParamUpdates: nil,
					Events: []abci.Event{
						event("endevent1", attr("end1attr1", "end1val1"), attr("end1attr2", "end1val2")),
					},
				},
				BeginBlock: &cmtstate.ResponseBeginBlock{
					Events: []abci.Event{
						event("beginevent1", attr("begin1attr1", "begin1val1"), attr("begin1attr2", "begin1val2")),
					},
				},
			},
		},
		{
			name:       "raw bz from real example",
			oldEntryBz: []byte("\n\xdc\x0e\x12\"\x12 \n\x1e/cosmos.gov.v1.MsgVoteResponse\x1a\xf7\x02[{\"msg_index\":0,\"events\":[{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"/cosmos.gov.v1.MsgVote\"},{\"key\":\"module\",\"value\":\"governance\"},{\"key\":\"sender\",\"value\":\"tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25\"}]},{\"type\":\"proposal_vote\",\"attributes\":[{\"key\":\"option\",\"value\":\"option:VOTE_OPTION_YES weight:\\\"1.000000000000000000\\\"\"},{\"key\":\"proposal_id\",\"value\":\"1\"}]}]}](\xc0\x9a\f0\x99\xe7\x04:`\n\ncoin_spent\x126\n\aspender\x12)tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25\x18\x01\x12\x1a\n\x06amount\x12\x0e381000000nhash\x18\x01:d\n\rcoin_received\x127\n\breceiver\x12)tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt\x18\x01\x12\x1a\n\x06amount\x12\x0e381000000nhash\x18\x01:\x97\x01\n\btransfer\x128\n\trecipient\x12)tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt\x18\x01\x125\n\x06sender\x12)tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25\x18\x01\x12\x1a\n\x06amount\x12\x0e381000000nhash\x18\x01:@\n\amessage\x125\n\x06sender\x12)tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25\x18\x01:\\\n\x02tx\x12\x1c\n\x03fee\x12\x1310000000000000nhash\x18\x01\x128\n\tfee_payer\x12)tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25\x18\x01:c\n\x02tx\x12#\n\x0fmin_fee_charged\x12\x0e381000000nhash\x18\x01\x128\n\tfee_payer\x12)tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25\x18\x01:>\n\x02tx\x128\n\aacc_seq\x12+tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25/2\x18\x01:m\n\x02tx\x12g\n\tsignature\x12X8J3dGzkMy8bz3ldIS/80nNgtLV2+A/0NU7RNbWpYy8RqNTJT/NLkXFe2OqFugwYl7QQVjT3iG1vzW8jhswSV1A==\x18\x01:-\n\amessage\x12\"\n\x06action\x12\x16/cosmos.gov.v1.MsgVote\x18\x01:e\n\rproposal_vote\x12@\n\x06option\x124option:VOTE_OPTION_YES weight:\"1.000000000000000000\"\x18\x01\x12\x12\n\vproposal_id\x12\x011\x18\x01:X\n\amessage\x12\x16\n\x06module\x12\ngovernance\x18\x01\x125\n\x06sender\x12)tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25\x18\x01:d\n\ncoin_spent\x126\n\aspender\x12)tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25\x18\x01\x12\x1e\n\x06amount\x12\x129999619000000nhash\x18\x01:h\n\rcoin_received\x127\n\breceiver\x12)tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt\x18\x01\x12\x1e\n\x06amount\x12\x129999619000000nhash\x18\x01:\x9b\x01\n\btransfer\x128\n\trecipient\x12)tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt\x18\x01\x125\n\x06sender\x12)tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25\x18\x01\x12\x1e\n\x06amount\x12\x129999619000000nhash\x18\x01:@\n\amessage\x125\n\x06sender\x12)tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25\x18\x01\x12)\x12'\n\n\b\x80\x80\xc0\n\x10\x80\x8e\xce\x1c\x12\x0e\b\xa0\x8d\x06\x12\x04\b\x80\xc6\n\x18\x80\x80@\x1a\t\n\aed25519\x1a\x8c\n\n@\n\amessage\x125\n\x06sender\x12)tp1m3h30wlvsf8llruxtpukdvsy0km2kum8j5f3gk\x18\x01\n\x8f\x01\n\x04mint\x12&\n\fbonded_ratio\x12\x140.000010000000000000\x18\x01\x12#\n\tinflation\x12\x140.000000000000000000\x18\x01\x12+\n\x11annual_provisions\x12\x140.000000000000000000\x18\x01\x12\r\n\x06amount\x12\x010\x18\x01\ne\n\ncoin_spent\x126\n\aspender\x12)tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt\x18\x01\x12\x1f\n\x06amount\x12\x1310000000000000nhash\x18\x01\ni\n\rcoin_received\x127\n\breceiver\x12)tp1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8a624xf\x18\x01\x12\x1f\n\x06amount\x12\x1310000000000000nhash\x18\x01\n\x9c\x01\n\btransfer\x128\n\trecipient\x12)tp1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8a624xf\x18\x01\x125\n\x06sender\x12)tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt\x18\x01\x12\x1f\n\x06amount\x12\x1310000000000000nhash\x18\x01\n@\n\amessage\x125\n\x06sender\x12)tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt\x18\x01\n\x84\x01\n\x0fproposer_reward\x120\n\x06amount\x12$500000000000.000000000000000000nhash\x18\x01\x12?\n\tvalidator\x120tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3\x18\x01\n~\n\ncommission\x12/\n\x06amount\x12#50000000000.000000000000000000nhash\x18\x01\x12?\n\tvalidator\x120tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3\x18\x01\n|\n\arewards\x120\n\x06amount\x12$500000000000.000000000000000000nhash\x18\x01\x12?\n\tvalidator\x120tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3\x18\x01\n\x7f\n\ncommission\x120\n\x06amount\x12$930000000000.000000000000000000nhash\x18\x01\x12?\n\tvalidator\x120tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3\x18\x01\n}\n\arewards\x121\n\x06amount\x12%9300000000000.000000000000000000nhash\x18\x01\x12?\n\tvalidator\x120tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3\x18\x01"),
			saveExpBz:  true,
		},
		{
			name: "real example starting from struct",
			oldEntry: &cmtstate.LegacyABCIResponses{
				DeliverTxs: []*abci.ExecTxResult{
					{
						Code:      0,
						Data:      []byte("\x12 \n\x1e/cosmos.gov.v1.MsgVoteResponse"),
						Log:       "[{\"msg_index\":0,\"events\":[{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"/cosmos.gov.v1.MsgVote\"},{\"key\":\"module\",\"value\":\"governance\"},{\"key\":\"sender\",\"value\":\"tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25\"}]},{\"type\":\"proposal_vote\",\"attributes\":[{\"key\":\"option\",\"value\":\"option:VOTE_OPTION_YES weight:\\\"1.000000000000000000\\\"\"},{\"key\":\"proposal_id\",\"value\":\"1\"}]}]}]",
						Info:      "",
						GasWanted: 200000,
						GasUsed:   78745,
						Codespace: "",
						Events: []abci.Event{
							event("coin_spent",
								attr("spender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25"),
								attr("amount", "381000000nhash")),
							event("coin_received",
								attr("receiver", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"),
								attr("amount", "381000000nhash")),
							event("transfer",
								attr("recipient", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"),
								attr("sender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25"),
								attr("amount", "381000000nhash")),
							event("message", attr("sender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25")),
							event("tx",
								attr("fee", "10000000000000nhash"),
								attr("fee_payer", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25")),
							event("tx",
								attr("min_fee_charged", "381000000nhash"),
								attr("fee_payer", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25")),
							event("tx", attr("acc_seq", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25/2")),
							event("tx", attr("signature", "8J3dGzkMy8bz3ldIS/80nNgtLV2+A/0NU7RNbWpYy8RqNTJT/NLkXFe2OqFugwYl7QQVjT3iG1vzW8jhswSV1A==")),
							event("message", attr("action", "/cosmos.gov.v1.MsgVote")),
							event("proposal_vote",
								attr("option", "option:VOTE_OPTION_YES weight:\"1.000000000000000000\""),
								attr("proposal_id", "1")),
							event("message",
								attr("module", "governance"),
								attr("sender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25")),
							event("coin_spent",
								attr("spender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25"),
								attr("amount", "9999619000000nhash")),
							event("coin_received",
								attr("receiver", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"),
								attr("amount", "9999619000000nhash")),
							event("transfer",
								attr("recipient", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"),
								attr("sender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25"),
								attr("amount", "9999619000000nhash")),
							event("message", attr("sender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25")),
						},
					},
				},
				EndBlock: &cmtstate.ResponseEndBlock{
					ValidatorUpdates: nil,
					ConsensusParamUpdates: &cmttypes.ConsensusParams{
						Block:     &cmttypes.BlockParams{MaxBytes: 22020096, MaxGas: 60000000},
						Evidence:  &cmttypes.EvidenceParams{MaxAgeNumBlocks: 100000, MaxAgeDuration: 172800000000000, MaxBytes: 1048576},
						Validator: &cmttypes.ValidatorParams{PubKeyTypes: []string{"ed25519"}},
						Version:   nil,
						Abci:      nil,
					},
					Events: nil,
				},
				BeginBlock: &cmtstate.ResponseBeginBlock{
					Events: []abci.Event{
						event("message", attr("sender", "tp1m3h30wlvsf8llruxtpukdvsy0km2kum8j5f3gk")),
						event("mint",
							attr("bonded_ratio", "0.000010000000000000"),
							attr("inflation", "0.000000000000000000"),
							attr("annual_provisions", "0.000000000000000000"),
							attr("amount", "0")),
						event("coin_spent",
							attr("spender", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"),
							attr("amount", "10000000000000nhash")),
						event("coin_received",
							attr("receiver", "tp1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8a624xf"),
							attr("amount", "10000000000000nhash")),
						event("transfer",
							attr("recipient", "tp1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8a624xf"),
							attr("sender", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"),
							attr("amount", "10000000000000nhash")),
						event("message", attr("sender", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt")),
						event("proposer_reward",
							attr("amount", "500000000000.000000000000000000nhash"),
							attr("validator", "tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3")),
						event("commission",
							attr("amount", "50000000000.000000000000000000nhash"),
							attr("validator", "tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3")),
						event("rewards",
							attr("amount", "500000000000.000000000000000000nhash"),
							attr("validator", "tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3")),
						event("commission",
							attr("amount", "930000000000.000000000000000000nhash"),
							attr("validator", "tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3")),
						event("rewards",
							attr("amount", "9300000000000.000000000000000000nhash"),
							attr("validator", "tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3")),
					},
				},
			},
			saveActBz: true,
		},
		{
			name: "real example with the Data field cleaned up",
			oldEntry: &cmtstate.LegacyABCIResponses{
				DeliverTxs: []*abci.ExecTxResult{
					{
						Code:      0,
						Data:      []byte("/cosmos.gov.v1.MsgVoteResponse"),
						Log:       "[{\"msg_index\":0,\"events\":[{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"/cosmos.gov.v1.MsgVote\"},{\"key\":\"module\",\"value\":\"governance\"},{\"key\":\"sender\",\"value\":\"tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25\"}]},{\"type\":\"proposal_vote\",\"attributes\":[{\"key\":\"option\",\"value\":\"option:VOTE_OPTION_YES weight:\\\"1.000000000000000000\\\"\"},{\"key\":\"proposal_id\",\"value\":\"1\"}]}]}]",
						Info:      "",
						GasWanted: 200000,
						GasUsed:   78745,
						Codespace: "",
						Events: []abci.Event{
							event("coin_spent",
								attr("spender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25"),
								attr("amount", "381000000nhash")),
							event("coin_received",
								attr("receiver", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"),
								attr("amount", "381000000nhash")),
							event("transfer",
								attr("recipient", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"),
								attr("sender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25"),
								attr("amount", "381000000nhash")),
							event("message", attr("sender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25")),
							event("tx",
								attr("fee", "10000000000000nhash"),
								attr("fee_payer", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25")),
							event("tx",
								attr("min_fee_charged", "381000000nhash"),
								attr("fee_payer", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25")),
							event("tx", attr("acc_seq", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25/2")),
							event("tx", attr("signature", "8J3dGzkMy8bz3ldIS/80nNgtLV2+A/0NU7RNbWpYy8RqNTJT/NLkXFe2OqFugwYl7QQVjT3iG1vzW8jhswSV1A==")),
							event("message", attr("action", "/cosmos.gov.v1.MsgVote")),
							event("proposal_vote",
								attr("option", "option:VOTE_OPTION_YES weight:\"1.000000000000000000\""),
								attr("proposal_id", "1")),
							event("message",
								attr("module", "governance"),
								attr("sender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25")),
							event("coin_spent",
								attr("spender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25"),
								attr("amount", "9999619000000nhash")),
							event("coin_received",
								attr("receiver", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"),
								attr("amount", "9999619000000nhash")),
							event("transfer",
								attr("recipient", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"),
								attr("sender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25"),
								attr("amount", "9999619000000nhash")),
							event("message", attr("sender", "tp1q8qctuwmvudpvlvl5csu3gwa5zkg2hz5cw3l25")),
						},
					},
				},
				EndBlock: &cmtstate.ResponseEndBlock{
					ValidatorUpdates: nil,
					ConsensusParamUpdates: &cmttypes.ConsensusParams{
						Block:     &cmttypes.BlockParams{MaxBytes: 22020096, MaxGas: 60000000},
						Evidence:  &cmttypes.EvidenceParams{MaxAgeNumBlocks: 100000, MaxAgeDuration: 172800000000000, MaxBytes: 1048576},
						Validator: &cmttypes.ValidatorParams{PubKeyTypes: []string{"ed25519"}},
						Version:   nil,
						Abci:      nil,
					},
					Events: nil,
				},
				BeginBlock: &cmtstate.ResponseBeginBlock{
					Events: []abci.Event{
						event("message", attr("sender", "tp1m3h30wlvsf8llruxtpukdvsy0km2kum8j5f3gk")),
						event("mint",
							attr("bonded_ratio", "0.000010000000000000"),
							attr("inflation", "0.000000000000000000"),
							attr("annual_provisions", "0.000000000000000000"),
							attr("amount", "0")),
						event("coin_spent",
							attr("spender", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"),
							attr("amount", "10000000000000nhash")),
						event("coin_received",
							attr("receiver", "tp1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8a624xf"),
							attr("amount", "10000000000000nhash")),
						event("transfer",
							attr("recipient", "tp1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8a624xf"),
							attr("sender", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt"),
							attr("amount", "10000000000000nhash")),
						event("message", attr("sender", "tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt")),
						event("proposer_reward",
							attr("amount", "500000000000.000000000000000000nhash"),
							attr("validator", "tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3")),
						event("commission",
							attr("amount", "50000000000.000000000000000000nhash"),
							attr("validator", "tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3")),
						event("rewards",
							attr("amount", "500000000000.000000000000000000nhash"),
							attr("validator", "tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3")),
						event("commission",
							attr("amount", "930000000000.000000000000000000nhash"),
							attr("validator", "tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3")),
						event("rewards",
							attr("amount", "9300000000000.000000000000000000nhash"),
							attr("validator", "tpvaloper1q8qctuwmvudpvlvl5csu3gwa5zkg2hz588vtl3")),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.False(t, tc.oldEntry != nil && len(tc.oldEntryBz) > 0, "cannot provide both an oldEntry and oldEntryBz")

			var oldEntryBz []byte
			var oldEntry *cmtstate.LegacyABCIResponses
			if tc.oldEntry != nil {
				var err error
				oldEntryBz, err = tc.oldEntry.Marshal()
				require.NoError(t, err, "oldEntry.Marshal()")
				oldEntry = tc.oldEntry
			} else {
				oldEntry = new(cmtstate.LegacyABCIResponses)
				err := oldEntry.Unmarshal(tc.oldEntryBz)
				if err != nil {
					t.Logf("oldEntry.Unmarshal(tc.oldEntryBz) error: %q", err.Error())
				}
				oldEntryBz = tc.oldEntryBz
			}

			t.Logf("oldEntryBz:\n%q", string(oldEntryBz))
			t.Logf("oldEntry:\n%s", LegacyABCIResponsesString(oldEntry))

			if tc.saveActBz {
				require.Nil(t, actBz, "Only one test can have saveActBz: true")
				actBz = oldEntryBz
			}

			if tc.saveExpBz {
				require.Nil(t, expBz, "Only one test can have saveExpBz: true")
				expBz = oldEntryBz
			}

			newEntry := new(abci.ResponseFinalizeBlock)
			err := newEntry.Unmarshal(oldEntryBz)
			if assert.Error(t, err, "newEntry.Unmarshal(oldEntryBz)") {
				t.Logf("newEntry.Unmarshal(oldEntryBz) error: %q", err.Error())
			} else {
				t.Logf("newEntry:\n%s", ResponseFinalizeBlockString(newEntry))
			}
		})
	}

	t.Run("compare results", func(t *testing.T) {
		assert.Equal(t, expBz, actBz, "state entry bytes")
	})
}

func ResponseFinalizeBlockString(val *abci.ResponseFinalizeBlock) string {
	entries := []string{
		fmt.Sprintf("Events: %s", EventsString(val.Events)),
		fmt.Sprintf("TxResults: %s", TxResultsString(val.TxResults)),
		fmt.Sprintf("ValidatorUpdates: %s", ValidatorUpdatesString(val.ValidatorUpdates)),
		fmt.Sprintf("ConsensusParamUpdates: %s", ConsensusParamsString(val.ConsensusParamUpdates)),
		fmt.Sprintf("AppHash: %s", bzStr(val.AppHash)),
	}
	return strings.Join(entries, "\n")
}

func indent(str, prefix string) string {
	lines := strings.Split(str, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}

func bzStr(bz []byte) string {
	if bz == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%q", string(bz))
}

func EventsString(vals []abci.Event) string {
	if vals == nil {
		return "<nil>"
	}
	if len(vals) == 0 {
		return "(0) []"
	}
	entries := make([]string, len(vals))
	for i, v := range vals {
		entries[i] = fmt.Sprintf("[%d]: %s", i, EventString(v))
	}
	return fmt.Sprintf("(%d) [\n%s\n]", len(vals), indent(strings.Join(entries, "\n"), "  "))
}

func EventString(val abci.Event) string {
	return fmt.Sprintf("{%q:%s}", val.Type, AttrsString(val.Attributes))
}

func AttrsString(vals []abci.EventAttribute) string {
	if vals == nil {
		return "<nil>"
	}
	entries := make([]string, len(vals))
	for i, v := range vals {
		entries[i] = AttrString(v)
	}
	return fmt.Sprintf("[%s]", strings.Join(entries, ","))
}

func AttrString(val abci.EventAttribute) string {
	ind := ""
	if val.Index {
		ind = "(i)"
	}
	return fmt.Sprintf("%q=%q%s", val.Key, val.Value, ind)
}

func TxResultsString(vals []*abci.ExecTxResult) string {
	if vals == nil {
		return "<nil>"
	}
	if len(vals) == 0 {
		return "(0) []"
	}
	entries := make([]string, len(vals))
	for i, v := range vals {
		entries[i] = fmt.Sprintf("[%d]: %s", i, ExecTxResultString(v))
	}
	return fmt.Sprintf("(%d) [\n%s\n]", len(vals), indent(strings.Join(entries, "\n"), "  "))
}

func ExecTxResultString(val *abci.ExecTxResult) string {
	if val == nil {
		return "<nil>"
	}
	fields := []string{
		fmt.Sprintf("     Code: %d", val.Code),
		fmt.Sprintf("     Data: %q", bzStr(val.Data)),
		fmt.Sprintf("      Log: %q", val.Log),
		fmt.Sprintf("     Info: %q", val.Info),
		fmt.Sprintf("GasWanted: %d", val.GasWanted),
		fmt.Sprintf("  GasUsed: %d", val.GasUsed),
		fmt.Sprintf("Codespace: %q", val.Codespace),
		fmt.Sprintf("   Events: %s", EventsString(val.Events)),
	}
	return fmt.Sprintf("{\n%s\n}", indent(strings.Join(fields, "\n"), "  "))
}

func ValidatorUpdatesString(vals []abci.ValidatorUpdate) string {
	if vals == nil {
		return "<nil>"
	}
	if len(vals) == 0 {
		return "(0) []"
	}
	entries := make([]string, len(vals))
	for i, v := range vals {
		entries[i] = fmt.Sprintf("[%d]: %s", i, ValidatorUpdateString(v))
	}
	return fmt.Sprintf("(%d) [\n%s\n]", len(vals), indent(strings.Join(entries, "\n"), "  "))
}

func ValidatorUpdateString(val abci.ValidatorUpdate) string {
	return fmt.Sprintf("{PubKey: %q, Power: %d}", val.PubKey.String(), val.Power)
}

func ConsensusParamsString(val *cmttypes.ConsensusParams) string {
	if val == nil {
		return "<nil>"
	}
	entries := []string{
		fmt.Sprintf("Block: %s", BlockParamsString(val.Block)),
		fmt.Sprintf("Evidence: %s", EvidenceParamsString(val.Evidence)),
		fmt.Sprintf("Validator: %s", ValidatorParamsString(val.Validator)),
		fmt.Sprintf("Version: %s", VersionParamsString(val.Version)),
		fmt.Sprintf("Abci: %s", ABCIParamsString(val.Abci)),
	}
	return fmt.Sprintf("{\n%s\n}", indent(strings.Join(entries, "\n"), "  "))
}

func BlockParamsString(val *cmttypes.BlockParams) string {
	if val == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{MaxBytes: %d, MaxGas: %d}", val.MaxBytes, val.MaxGas)
}

func EvidenceParamsString(val *cmttypes.EvidenceParams) string {
	if val == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{MaxAgeNumBlocks: %d, MaxAgeDuration: %d, MaxBytes: %d}",
		val.MaxAgeNumBlocks, val.MaxAgeDuration, val.MaxBytes)
}

func ValidatorParamsString(val *cmttypes.ValidatorParams) string {
	if val == nil {
		return "<nil>"
	}
	pkt := ""
	switch {
	case val.PubKeyTypes == nil:
		pkt = "<nil>"
	case len(val.PubKeyTypes) == 0:
		pkt = "(0) []"
	default:
		pkts := make([]string, len(val.PubKeyTypes))
		for i, t := range val.PubKeyTypes {
			pkts[i] = fmt.Sprintf("%q", t)
		}
		pkt = fmt.Sprintf("(%d) [%s]", len(val.PubKeyTypes), strings.Join(pkts, ","))
	}
	return fmt.Sprintf("{PubKeyTypes: %s}", pkt)
}

func VersionParamsString(val *cmttypes.VersionParams) string {
	if val == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{App: %d}", val.App)
}

func ABCIParamsString(val *cmttypes.ABCIParams) string {
	if val == nil {
		return "<nil>"
	}
	return fmt.Sprintf("{VoteExtensionsEnableHeight: %d}", val.VoteExtensionsEnableHeight)
}

func LegacyABCIResponsesString(val *cmtstate.LegacyABCIResponses) string {
	if val == nil {
		return "<nil>"
	}
	entries := []string{
		fmt.Sprintf("DeliverTxs: %s", TxResultsString(val.DeliverTxs)),
		fmt.Sprintf("EndBlock: %s", ResponseEndBlockString(val.EndBlock)),
		fmt.Sprintf("BeginBlock: %s", ResponseBeginBlockString(val.BeginBlock)),
	}
	return fmt.Sprintf("{\n%s\n}", indent(strings.Join(entries, "\n"), "  "))
}

func ResponseEndBlockString(val *cmtstate.ResponseEndBlock) string {
	if val == nil {
		return "<nil>"
	}
	entries := []string{
		fmt.Sprintf("ValidatorUpdates: %s", ValidatorUpdatesString(val.ValidatorUpdates)),
		fmt.Sprintf("ConsensusParamUpdates: %s", ConsensusParamsString(val.ConsensusParamUpdates)),
		fmt.Sprintf("Events: %s", EventsString(val.Events)),
	}
	return fmt.Sprintf("{\n%s\n}", indent(strings.Join(entries, "\n"), "  "))
}

func ResponseBeginBlockString(val *cmtstate.ResponseBeginBlock) string {
	if val == nil {
		return "<nil>"
	}
	entries := []string{
		fmt.Sprintf("Events: %s", EventsString(val.Events)),
	}
	return fmt.Sprintf("{\n%s\n}", indent(strings.Join(entries, "\n"), "  "))
}
