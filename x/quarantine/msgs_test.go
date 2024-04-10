package quarantine_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/quarantine"
	"github.com/provenance-io/provenance/x/quarantine/testutil"
)

func TestNewMsgOptIn(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("nmoi", 0)
	testAddr1 := testutil.MakeTestAddr("nmoi", 1)

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected *quarantine.MsgOptIn
	}{
		{
			name:     "addr 0",
			toAddr:   testAddr0,
			expected: &quarantine.MsgOptIn{ToAddress: testAddr0.String()},
		},
		{
			name:     "addr 1",
			toAddr:   testAddr1,
			expected: &quarantine.MsgOptIn{ToAddress: testAddr1.String()},
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: &quarantine.MsgOptIn{ToAddress: ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := quarantine.NewMsgOptIn(tc.toAddr)
			assert.Equal(t, tc.expected, actual, "NewMsgOptIn")
		})
	}
}

func TestMsgOptIn_ValidateBasic(t *testing.T) {
	addr := testutil.MakeTestAddr("moivb", 0).String()

	tests := []struct {
		name          string
		addr          string
		expectedInErr []string
	}{
		{
			name:          "addr",
			addr:          addr,
			expectedInErr: nil,
		},
		{
			name:          "bad",
			addr:          "not an actual address",
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "empty",
			addr:          "",
			expectedInErr: []string{"invalid to address"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := quarantine.MsgOptIn{ToAddress: tc.addr}
			msg := quarantine.MsgOptIn{ToAddress: tc.addr}
			err := msg.ValidateBasic()
			assertions.AssertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, msgOrig, msg, "MsgOptIn before and after")
		})
	}
}

func TestMsgOptIn_GetSigners(t *testing.T) {
	addr := testutil.MakeTestAddr("moigs", 0)

	tests := []struct {
		name     string
		addr     string
		expected []sdk.AccAddress
	}{
		{
			name:     "addr",
			addr:     addr.String(),
			expected: []sdk.AccAddress{addr},
		},
		{
			name:     "bad",
			addr:     "not an actual address",
			expected: []sdk.AccAddress{nil},
		},
		{
			name:     "empty",
			addr:     "",
			expected: []sdk.AccAddress{{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := quarantine.MsgOptIn{ToAddress: tc.addr}
			msg := quarantine.MsgOptIn{ToAddress: tc.addr}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, msgOrig, msg, "MsgOptIn before and after")
		})
	}
}

func TestNewMsgOptOut(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("nmoo", 0)
	testAddr1 := testutil.MakeTestAddr("nmoo", 1)

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected *quarantine.MsgOptOut
	}{
		{
			name:     "addr 0",
			toAddr:   testAddr0,
			expected: &quarantine.MsgOptOut{ToAddress: testAddr0.String()},
		},
		{
			name:     "addr 1",
			toAddr:   testAddr1,
			expected: &quarantine.MsgOptOut{ToAddress: testAddr1.String()},
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: &quarantine.MsgOptOut{ToAddress: ""},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := quarantine.NewMsgOptOut(tc.toAddr)
			assert.Equal(t, tc.expected, actual, "NewMsgOptOut")
		})
	}
}

func TestMsgOptOut_ValidateBasic(t *testing.T) {
	addr := testutil.MakeTestAddr("moovb", 0).String()

	tests := []struct {
		name          string
		addr          string
		expectedInErr []string
	}{
		{
			name:          "addr",
			addr:          addr,
			expectedInErr: nil,
		},
		{
			name:          "bad",
			addr:          "not an actual address",
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "empty",
			addr:          "",
			expectedInErr: []string{"invalid to address"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := quarantine.MsgOptOut{ToAddress: tc.addr}
			msg := quarantine.MsgOptOut{ToAddress: tc.addr}
			err := msg.ValidateBasic()
			assertions.AssertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, msgOrig, msg, "MsgOptOut before and after")
		})
	}
}

func TestMsgOptOut_GetSigners(t *testing.T) {
	addr := testutil.MakeTestAddr("moogs", 0)

	tests := []struct {
		name     string
		addr     string
		expected []sdk.AccAddress
	}{
		{
			name:     "addr",
			addr:     addr.String(),
			expected: []sdk.AccAddress{addr},
		},
		{
			name:     "bad",
			addr:     "not an actual address",
			expected: []sdk.AccAddress{nil},
		},
		{
			name:     "empty",
			addr:     "",
			expected: []sdk.AccAddress{{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := quarantine.MsgOptOut{ToAddress: tc.addr}
			msg := quarantine.MsgOptOut{ToAddress: tc.addr}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, msgOrig, msg, "MsgOptOut before and after")
		})
	}
}

func TestNewMsgAccept(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("nma", 0)
	testAddr1 := testutil.MakeTestAddr("nma", 1)

	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []string
		permanent bool
		expected  *quarantine.MsgAccept
	}{
		{
			name:      "control",
			toAddr:    testAddr0,
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected: &quarantine.MsgAccept{
				ToAddress:     testAddr0.String(),
				FromAddresses: []string{testAddr1.String()},
				Permanent:     false,
			},
		},
		{
			name:      "nil toAddr",
			toAddr:    nil,
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected: &quarantine.MsgAccept{
				ToAddress:     "",
				FromAddresses: []string{testAddr1.String()},
				Permanent:     false,
			},
		},
		{
			name:      "nil fromAddrsStrs",
			toAddr:    testAddr1,
			fromAddrs: nil,
			permanent: false,
			expected: &quarantine.MsgAccept{
				ToAddress:     testAddr1.String(),
				FromAddresses: nil,
				Permanent:     false,
			},
		},
		{
			name:      "empty fromAddrsStrs",
			toAddr:    testAddr1,
			fromAddrs: []string{},
			permanent: false,
			expected: &quarantine.MsgAccept{
				ToAddress:     testAddr1.String(),
				FromAddresses: []string{},
				Permanent:     false,
			},
		},
		{
			name:      "three bad fromAddrsStrs",
			toAddr:    testAddr1,
			fromAddrs: []string{"one", "two", "three"},
			permanent: false,
			expected: &quarantine.MsgAccept{
				ToAddress:     testAddr1.String(),
				FromAddresses: []string{"one", "two", "three"},
				Permanent:     false,
			},
		},
		{
			name:      "permanent",
			toAddr:    testAddr1,
			fromAddrs: []string{testAddr0.String()},
			permanent: true,
			expected: &quarantine.MsgAccept{
				ToAddress:     testAddr1.String(),
				FromAddresses: []string{testAddr0.String()},
				Permanent:     true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := quarantine.NewMsgAccept(tc.toAddr, tc.fromAddrs, tc.permanent)
			assert.Equal(t, tc.expected, actual, "NewMsgAccept")
		})
	}
}

func TestMsgAccept_ValidateBasic(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("mavb", 0).String()
	testAddr1 := testutil.MakeTestAddr("mavb", 1).String()
	testAddr2 := testutil.MakeTestAddr("mavb", 2).String()

	tests := []struct {
		name          string
		toAddr        string
		fromAddrs     []string
		permanent     bool
		expectedInErr []string
	}{
		{
			name:          "control",
			toAddr:        testAddr0,
			fromAddrs:     []string{testAddr1},
			permanent:     false,
			expectedInErr: nil,
		},
		{
			name:          "permanent",
			toAddr:        testAddr0,
			fromAddrs:     []string{testAddr1},
			permanent:     true,
			expectedInErr: nil,
		},
		{
			name:          "empty to address",
			toAddr:        "",
			fromAddrs:     []string{testAddr1},
			permanent:     false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "bad to address",
			toAddr:        "this address isn't",
			fromAddrs:     []string{testAddr0},
			permanent:     false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "nil from addresses",
			toAddr:        testAddr1,
			fromAddrs:     nil,
			permanent:     false,
			expectedInErr: []string{"at least one from address is required", "unknown address"},
		},
		{
			name:          "empty from addresses",
			toAddr:        testAddr1,
			fromAddrs:     []string{},
			permanent:     false,
			expectedInErr: []string{"at least one from address is required", "unknown address"},
		},
		{
			name:          "bad from address",
			toAddr:        testAddr0,
			fromAddrs:     []string{"this one is a tunic"},
			permanent:     false,
			expectedInErr: []string{"invalid from address[0]"},
		},
		{
			name:          "bad third from address",
			toAddr:        testAddr0,
			fromAddrs:     []string{testAddr1, testAddr2, "Michael Jackson (he's bad)"},
			permanent:     false,
			expectedInErr: []string{"invalid from address[2]"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := quarantine.MsgAccept{
				ToAddress:     tc.toAddr,
				FromAddresses: testutil.MakeCopyOfStringSlice(tc.fromAddrs),
				Permanent:     tc.permanent,
			}
			msg := quarantine.MsgAccept{
				ToAddress:     tc.toAddr,
				FromAddresses: tc.fromAddrs,
				Permanent:     tc.permanent,
			}
			err := msg.ValidateBasic()
			assertions.AssertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, msgOrig, msg, "MsgAccept before and after")
		})
	}
}

func TestMsgAccept_GetSigners(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("mags", 0)
	testAddr1 := testutil.MakeTestAddr("mags", 1)
	testAddr2 := testutil.MakeTestAddr("mags", 2)

	tests := []struct {
		name      string
		toAddr    string
		fromAddrs []string
		permanent bool
		expected  []sdk.AccAddress
	}{
		{
			name:      "control",
			toAddr:    testAddr0.String(),
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr0},
		},
		{
			name:      "permanent",
			toAddr:    testAddr0.String(),
			fromAddrs: []string{testAddr1.String()},
			permanent: true,
			expected:  []sdk.AccAddress{testAddr0},
		},
		{
			name:      "empty to address",
			toAddr:    "",
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected:  []sdk.AccAddress{{}},
		},
		{
			name:      "bad to address",
			toAddr:    "this address isn't",
			fromAddrs: []string{testAddr0.String()},
			permanent: false,
			expected:  []sdk.AccAddress{nil},
		},
		{
			name:      "empty from addresses",
			toAddr:    testAddr1.String(),
			fromAddrs: []string{},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr1},
		},
		{
			name:      "two from addresses",
			toAddr:    testAddr2.String(),
			fromAddrs: []string{testAddr0.String(), testAddr1.String()},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr2},
		},
		{
			name:      "bad from address",
			toAddr:    testAddr0.String(),
			fromAddrs: []string{"this one is a tunic"},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := quarantine.MsgAccept{
				ToAddress:     tc.toAddr,
				FromAddresses: testutil.MakeCopyOfStringSlice(tc.fromAddrs),
				Permanent:     tc.permanent,
			}
			msg := quarantine.MsgAccept{
				ToAddress:     tc.toAddr,
				FromAddresses: tc.fromAddrs,
				Permanent:     tc.permanent,
			}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, msgOrig, msg, "MsgAccept before and after")
		})
	}
}

func TestNewMsgDecline(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("nmd", 0)
	testAddr1 := testutil.MakeTestAddr("nmd", 1)

	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []string
		permanent bool
		expected  *quarantine.MsgDecline
	}{
		{
			name:      "control",
			toAddr:    testAddr0,
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected: &quarantine.MsgDecline{
				ToAddress:     testAddr0.String(),
				FromAddresses: []string{testAddr1.String()},
				Permanent:     false,
			},
		},
		{
			name:      "nil toAddr",
			toAddr:    nil,
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected: &quarantine.MsgDecline{
				ToAddress:     "",
				FromAddresses: []string{testAddr1.String()},
				Permanent:     false,
			},
		},
		{
			name:      "nil fromAddrsStrs",
			toAddr:    testAddr1,
			fromAddrs: nil,
			permanent: false,
			expected: &quarantine.MsgDecline{
				ToAddress:     testAddr1.String(),
				FromAddresses: nil,
				Permanent:     false,
			},
		},
		{
			name:      "empty fromAddrsStrs",
			toAddr:    testAddr1,
			fromAddrs: []string{},
			permanent: false,
			expected: &quarantine.MsgDecline{
				ToAddress:     testAddr1.String(),
				FromAddresses: []string{},
				Permanent:     false,
			},
		},
		{
			name:      "three bad fromAddrsStrs",
			toAddr:    testAddr1,
			fromAddrs: []string{"one", "two", "three"},
			permanent: false,
			expected: &quarantine.MsgDecline{
				ToAddress:     testAddr1.String(),
				FromAddresses: []string{"one", "two", "three"},
				Permanent:     false,
			},
		},
		{
			name:      "permanent",
			toAddr:    testAddr1,
			fromAddrs: []string{testAddr0.String()},
			permanent: true,
			expected: &quarantine.MsgDecline{
				ToAddress:     testAddr1.String(),
				FromAddresses: []string{testAddr0.String()},
				Permanent:     true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := quarantine.NewMsgDecline(tc.toAddr, tc.fromAddrs, tc.permanent)
			assert.Equal(t, tc.expected, actual, "NewMsgDecline")
		})
	}
}

func TestMsgDecline_ValidateBasic(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("mdvb", 0).String()
	testAddr1 := testutil.MakeTestAddr("mdvb", 1).String()
	testAddr2 := testutil.MakeTestAddr("mdvb", 2).String()

	tests := []struct {
		name          string
		toAddr        string
		fromAddrs     []string
		permanent     bool
		expectedInErr []string
	}{
		{
			name:          "control",
			toAddr:        testAddr0,
			fromAddrs:     []string{testAddr1},
			permanent:     false,
			expectedInErr: nil,
		},
		{
			name:          "permanent",
			toAddr:        testAddr0,
			fromAddrs:     []string{testAddr1},
			permanent:     true,
			expectedInErr: nil,
		},
		{
			name:          "empty to address",
			toAddr:        "",
			fromAddrs:     []string{testAddr1},
			permanent:     false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "bad to address",
			toAddr:        "this address isn't",
			fromAddrs:     []string{testAddr0},
			permanent:     false,
			expectedInErr: []string{"invalid to address"},
		},
		{
			name:          "nil from addresses",
			toAddr:        testAddr1,
			fromAddrs:     nil,
			permanent:     false,
			expectedInErr: []string{"at least one from address is required", "unknown address"},
		},
		{
			name:          "empty from addresses",
			toAddr:        testAddr1,
			fromAddrs:     []string{},
			permanent:     false,
			expectedInErr: []string{"at least one from address is required", "unknown address"},
		},
		{
			name:          "bad from address",
			toAddr:        testAddr0,
			fromAddrs:     []string{"this one is a tunic"},
			permanent:     false,
			expectedInErr: []string{"invalid from address[0]"},
		},
		{
			name:          "bad third from address",
			toAddr:        testAddr0,
			fromAddrs:     []string{testAddr1, testAddr2, "Michael Jackson (he's bad)"},
			permanent:     false,
			expectedInErr: []string{"invalid from address[2]"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := quarantine.MsgDecline{
				ToAddress:     tc.toAddr,
				FromAddresses: testutil.MakeCopyOfStringSlice(tc.fromAddrs),
				Permanent:     tc.permanent,
			}
			msg := quarantine.MsgDecline{
				ToAddress:     tc.toAddr,
				FromAddresses: tc.fromAddrs,
				Permanent:     tc.permanent,
			}
			err := msg.ValidateBasic()
			assertions.AssertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, msgOrig, msg, "MsgDecline before and after")
		})
	}
}

func TestMsgDecline_GetSigners(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("mdgs", 0)
	testAddr1 := testutil.MakeTestAddr("mdgs", 1)
	testAddr2 := testutil.MakeTestAddr("mdgs", 2)

	tests := []struct {
		name      string
		toAddr    string
		fromAddrs []string
		permanent bool
		expected  []sdk.AccAddress
	}{
		{
			name:      "control",
			toAddr:    testAddr0.String(),
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr0},
		},
		{
			name:      "permanent",
			toAddr:    testAddr0.String(),
			fromAddrs: []string{testAddr1.String()},
			permanent: true,
			expected:  []sdk.AccAddress{testAddr0},
		},
		{
			name:      "empty to address",
			toAddr:    "",
			fromAddrs: []string{testAddr1.String()},
			permanent: false,
			expected:  []sdk.AccAddress{{}},
		},
		{
			name:      "bad to address",
			toAddr:    "this address isn't",
			fromAddrs: []string{testAddr0.String()},
			permanent: false,
			expected:  []sdk.AccAddress{nil},
		},
		{
			name:      "empty from addresses",
			toAddr:    testAddr1.String(),
			fromAddrs: []string{},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr1},
		},
		{
			name:      "two from addresses",
			toAddr:    testAddr2.String(),
			fromAddrs: []string{testAddr0.String(), testAddr1.String()},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr2},
		},
		{
			name:      "bad from address",
			toAddr:    testAddr0.String(),
			fromAddrs: []string{"this one is a tunic"},
			permanent: false,
			expected:  []sdk.AccAddress{testAddr0},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msgOrig := quarantine.MsgDecline{
				ToAddress:     tc.toAddr,
				FromAddresses: testutil.MakeCopyOfStringSlice(tc.fromAddrs),
				Permanent:     tc.permanent,
			}
			msg := quarantine.MsgDecline{
				ToAddress:     tc.toAddr,
				FromAddresses: tc.fromAddrs,
				Permanent:     tc.permanent,
			}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, msgOrig, msg, "MsgDecline before and after")
		})
	}
}

func TestNewMsgUpdateAutoResponses(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("nmuar", 0)
	testAddr1 := testutil.MakeTestAddr("nmuar", 1)
	testAddr2 := testutil.MakeTestAddr("nmuar", 2)
	testAddr3 := testutil.MakeTestAddr("nmuar", 3)
	testAddr4 := testutil.MakeTestAddr("nmuar", 4)
	testAddr5 := testutil.MakeTestAddr("nmuar", 5)

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		updates  []*quarantine.AutoResponseUpdate
		expected *quarantine.MsgUpdateAutoResponses
	}{
		{
			name:    "empty updates",
			toAddr:  testAddr0,
			updates: []*quarantine.AutoResponseUpdate{},
			expected: &quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0.String(),
				Updates:   []*quarantine.AutoResponseUpdate{},
			},
		},
		{
			name:    "one update no to addr",
			toAddr:  nil,
			updates: []*quarantine.AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT}},
			expected: &quarantine.MsgUpdateAutoResponses{
				ToAddress: "",
				Updates:   []*quarantine.AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT}},
			},
		},
		{
			name:    "one update accept",
			toAddr:  testAddr1,
			updates: []*quarantine.AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT}},
			expected: &quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr1.String(),
				Updates:   []*quarantine.AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT}},
			},
		},
		{
			name:    "one update decline",
			toAddr:  testAddr2,
			updates: []*quarantine.AutoResponseUpdate{{FromAddress: testAddr1.String(), Response: quarantine.AUTO_RESPONSE_DECLINE}},
			expected: &quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr2.String(),
				Updates:   []*quarantine.AutoResponseUpdate{{FromAddress: testAddr1.String(), Response: quarantine.AUTO_RESPONSE_DECLINE}},
			},
		},
		{
			name:    "one update unspecified",
			toAddr:  testAddr0,
			updates: []*quarantine.AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: quarantine.AUTO_RESPONSE_UNSPECIFIED}},
			expected: &quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0.String(),
				Updates:   []*quarantine.AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: quarantine.AUTO_RESPONSE_UNSPECIFIED}},
			},
		},
		{
			name:    "one update unspecified",
			toAddr:  testAddr0,
			updates: []*quarantine.AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: quarantine.AUTO_RESPONSE_UNSPECIFIED}},
			expected: &quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0.String(),
				Updates:   []*quarantine.AutoResponseUpdate{{FromAddress: testAddr2.String(), Response: quarantine.AUTO_RESPONSE_UNSPECIFIED}},
			},
		},
		{
			name:   "five updates",
			toAddr: testAddr0,
			updates: []*quarantine.AutoResponseUpdate{
				{FromAddress: testAddr1.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT},
				{FromAddress: testAddr2.String(), Response: quarantine.AUTO_RESPONSE_DECLINE},
				{FromAddress: testAddr3.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT},
				{FromAddress: testAddr4.String(), Response: quarantine.AUTO_RESPONSE_UNSPECIFIED},
				{FromAddress: testAddr5.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT},
			},
			expected: &quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0.String(),
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr1.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr2.String(), Response: quarantine.AUTO_RESPONSE_DECLINE},
					{FromAddress: testAddr3.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr4.String(), Response: quarantine.AUTO_RESPONSE_UNSPECIFIED},
					{FromAddress: testAddr5.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := quarantine.NewMsgUpdateAutoResponses(tc.toAddr, tc.updates)
			assert.Equal(t, tc.expected, actual, "NewMsgUpdateAutoResponses")
		})
	}
}

func TestMsgUpdateAutoResponses_ValidateBasic(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("muarvb", 0).String()
	testAddr1 := testutil.MakeTestAddr("muarvb", 1).String()
	testAddr2 := testutil.MakeTestAddr("muarvb", 2).String()
	testAddr3 := testutil.MakeTestAddr("muarvb", 3).String()
	testAddr4 := testutil.MakeTestAddr("muarvb", 4).String()
	testAddr5 := testutil.MakeTestAddr("muarvb", 5).String()

	tests := []struct {
		name          string
		orig          quarantine.MsgUpdateAutoResponses
		expectedInErr []string
	}{
		{
			name: "control accept",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr1, Response: quarantine.AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: nil,
		},
		{
			name: "control decline",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr2, Response: quarantine.AUTO_RESPONSE_DECLINE},
				},
			},
			expectedInErr: nil,
		},
		{
			name: "control unspecified",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr3, Response: quarantine.AUTO_RESPONSE_UNSPECIFIED},
				},
			},
			expectedInErr: nil,
		},
		{
			name: "bad to address",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: "not really that bad",
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr1, Response: quarantine.AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid to address"},
		},
		{
			name: "empty to address",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: "",
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr1, Response: quarantine.AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid to address"},
		},
		{
			name: "nil updates",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates:   nil,
			},
			expectedInErr: []string{"invalid value", "no updates"},
		},
		{
			name: "empty updates",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates:   []*quarantine.AutoResponseUpdate{},
			},
			expectedInErr: []string{"invalid value", "no updates"},
		},
		{
			name: "one update bad from address",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: "Okay, I'm bad again.", Response: quarantine.AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 1", "invalid from address"},
		},
		{
			name: "one update empty from address",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: "", Response: quarantine.AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 1", "invalid from address"},
		},
		{
			name: "one update negative resp",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr1, Response: -1},
				},
			},
			expectedInErr: []string{"invalid update 1", "unknown auto-response value: -1"},
		},
		{
			name: "one update resp too large",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr2, Response: 900},
				},
			},
			expectedInErr: []string{"invalid update 1", "unknown auto-response value: 900"},
		},
		{
			name: "five updates third bad from address",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr1, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr2, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: "still not good", Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr4, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr5, Response: quarantine.AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 3", "invalid from address"},
		},
		{
			name: "five updates fourth empty from address",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr1, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr2, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr3, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: "", Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr5, Response: quarantine.AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 4", "invalid from address"},
		},
		{
			name: "five updates first negative resp",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr1, Response: -88},
					{FromAddress: testAddr2, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr3, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr4, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr5, Response: quarantine.AUTO_RESPONSE_ACCEPT},
				},
			},
			expectedInErr: []string{"invalid update 1", "unknown auto-response value: -88"},
		},
		{
			name: "five update last resp too large",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0,
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr1, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr2, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr3, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr4, Response: quarantine.AUTO_RESPONSE_ACCEPT},
					{FromAddress: testAddr5, Response: 55},
				},
			},
			expectedInErr: []string{"invalid update 5", "unknown auto-response value: 55"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := quarantine.MsgUpdateAutoResponses{
				ToAddress: tc.orig.ToAddress,
				Updates:   nil,
			}
			if tc.orig.Updates != nil {
				msg.Updates = []*quarantine.AutoResponseUpdate{}
				for _, update := range tc.orig.Updates {
					msg.Updates = append(msg.Updates, &quarantine.AutoResponseUpdate{
						FromAddress: update.FromAddress,
						Response:    update.Response,
					})
				}
			}
			err := msg.ValidateBasic()
			assertions.AssertErrorContents(t, err, tc.expectedInErr, "ValidateBasic")
			assert.Equal(t, tc.orig, msg, "MsgUpdateAutoResponses before and after")
		})
	}
}

func TestMsgUpdateAutoResponses_GetSigners(t *testing.T) {
	testAddr0 := testutil.MakeTestAddr("muargs", 0)
	testAddr1 := testutil.MakeTestAddr("muargs", 1)
	testAddr2 := testutil.MakeTestAddr("muargs", 2)

	tests := []struct {
		name     string
		orig     quarantine.MsgUpdateAutoResponses
		expected []sdk.AccAddress
	}{
		{
			name: "control",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: testAddr0.String(),
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr1.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT},
				},
			},
			expected: []sdk.AccAddress{testAddr0},
		},
		{
			name: "bad addr",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: "bad bad bad",
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr2.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT},
				},
			},
			expected: []sdk.AccAddress{nil},
		},
		{
			name: "empty addr",
			orig: quarantine.MsgUpdateAutoResponses{
				ToAddress: "",
				Updates: []*quarantine.AutoResponseUpdate{
					{FromAddress: testAddr1.String(), Response: quarantine.AUTO_RESPONSE_ACCEPT},
				},
			},
			expected: []sdk.AccAddress{{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := quarantine.MsgUpdateAutoResponses{
				ToAddress: tc.orig.ToAddress,
				Updates:   nil,
			}
			if tc.orig.Updates != nil {
				msg.Updates = []*quarantine.AutoResponseUpdate{}
				for _, update := range tc.orig.Updates {
					msg.Updates = append(msg.Updates, &quarantine.AutoResponseUpdate{
						FromAddress: update.FromAddress,
						Response:    update.Response,
					})
				}
			}
			actual := msg.GetSigners()
			assert.Equal(t, tc.expected, actual, "GetSigners")
			assert.Equal(t, tc.orig, msg, "MsgUpdateAutoResponses before and after")
		})
	}
}
