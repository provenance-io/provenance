package testutil

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/protoadapt"

	"cosmossdk.io/x/tx/signing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil/assertions"
)

// MsgMaker is a function that returns a new message that has the provided signer.
// If your Msg can have multiple signers, use a MsgMakerMulti.
type MsgMaker func(signer string) sdk.Msg

// MsgMakerMulti is a function that returns a new message that has the provided signers.
// If your Msg can only have one signer, use a MsgMaker.
type MsgMakerMulti func(signers []string) sdk.Msg

// RunGetSignersTests make sure that we can get signers from all msgs.
// Also ensures that all entries in allRequestMsgs have one maker provided, and that all msg types returned by the makers are in allRequestMsgs.
func RunGetSignersTests[R []M, M sdk.Msg](t *testing.T, allRequestMsgs R, msgMakers []MsgMaker, msgMakersMulti []MsgMakerMulti) {
	t.Helper()

	encCfg := app.MakeTestEncodingConfig(t)
	sigCtx := encCfg.InterfaceRegistry.SigningContext()

	var msgTypes []string
	msgTests := make(map[string][]*msgTestCases)
	inMsgs := make(map[string]bool)
	for _, msg := range allRequestMsgs {
		typeName := getTypeName(msg)
		inMsgs[typeName] = true
		msgTypes = appendIfNew(msgTypes, typeName)
	}

	for _, msgMaker := range msgMakers {
		tests := newMsgTestCasesSingle(msgMaker)
		msgTypes = appendIfNew(msgTypes, tests.TypeName)
		msgTests[tests.TypeName] = append(msgTests[tests.TypeName], tests)
	}

	for _, msgMakerMulti := range msgMakersMulti {
		tests := newMsgTestCasesMulti(msgMakerMulti)
		msgTypes = appendIfNew(msgTypes, tests.TypeName)
		msgTests[tests.TypeName] = append(msgTests[tests.TypeName], tests)
	}

	for _, msgType := range msgTypes {
		t.Run(msgType, func(t *testing.T) {
			t.Run("has AllRequestMsgs entry", func(t *testing.T) {
				assert.True(t, inMsgs[msgType], "AllRequestMsgs does not have an entry for %s", msgType)
			})
			testGroups, hasMaker := msgTests[msgType]
			t.Run("one maker provided", func(t *testing.T) {
				require.True(t, hasMaker, "Missing maker entry that returns a %s", msgType)
				require.Len(t, testGroups, 1, "Duplicate maker entries provided that return a %s", msgType)
			})
			if len(testGroups) == 0 {
				return
			}
			tests := testGroups[0]

			for _, tc := range tests.TestCases {
				genericRunner := tc.GetGenericTestRunner(sigCtx)
				if tests.HasLegacy || tests.HasStrs {
					t.Run(tc.Name+" generic", genericRunner)
				} else {
					t.Run(tc.Name, genericRunner)
				}

				if tests.HasLegacy {
					t.Run(tc.Name+" legacy", tc.GetLegacyTestRunner())
				}

				if tests.HasStrs {
					t.Run(tc.Name+" strings", tc.GetStringsTestRunner())
				}
			}
		})
	}
}

// getTypeName gets just the type name of the provided thing, e.g. "MsgGovCreateMarketRequest".
func getTypeName(thing interface{}) string {
	rv := fmt.Sprintf("%T", thing) // e.g. "*types.MsgGovCreateMarketRequest"
	lastDot := strings.LastIndex(rv, ".")
	if lastDot < 0 || lastDot+1 >= len(rv) {
		return rv
	}
	return rv[lastDot+1:]
}

// hasGetSigners is an interface satisfied when a thing has a legacy GetSigners() method.
type hasGetSigners interface {
	GetSigners() []sdk.AccAddress
}

// hasGetSignersStrs is an interface satisfied when a thing has a GetSignerStrs() method.
type hasGetSignersStrs interface {
	GetSignerStrs() []string
}

// sigTestCase is a struct used to define a GetSigners test case.
type sigTestCase struct {
	Name           string
	Msg            sdk.Msg
	ExpInErr       []string
	ExpSigners     []sdk.AccAddress
	ExpSignersBz   [][]byte
	ExpSignersStrs []string
}

// GetGenericTestRunner returns a new test runner that ensures the sigCtx.GetSigners(...) method behaves as expected.
func (tc *sigTestCase) GetGenericTestRunner(sigCtx *signing.Context) func(t *testing.T) {
	return func(t *testing.T) {
		var actualBZ [][]byte
		var err error
		testFunc := func() {
			msgV2 := protoadapt.MessageV2Of(tc.Msg)
			actualBZ, err = sigCtx.GetSigners(msgV2)
		}
		require.NotPanics(t, testFunc, "sigCtx.GetSigners(msgV2)")
		assertions.AssertErrorContents(t, err, tc.ExpInErr, "sigCtx.GetSigners(msgV2) error")
		assert.Equal(t, tc.ExpSignersBz, actualBZ, "sigCtx.GetSigners(msgV2) result")
	}
}

// GetLegacyTestRunner returns a new test runner that ensures the msg.GetSigner() method behaves as expected.
func (tc *sigTestCase) GetLegacyTestRunner() func(t *testing.T) {
	return func(t *testing.T) {
		smsg, ok := tc.Msg.(hasGetSigners)
		require.True(t, ok, "%T does not have a .GetSigners method.", tc.Msg)

		var signers []sdk.AccAddress
		testFunc := func() {
			signers = smsg.GetSigners()
		}

		assertions.RequirePanicContents(t, testFunc, tc.ExpInErr, "GetSigners")
		assert.Equal(t, tc.ExpSigners, signers, "GetSigners")
	}
}

// GetStringsTestRunner returns a new test runner that ensures the msg.GetSignerStrs() method behaves as expected.
func (tc *sigTestCase) GetStringsTestRunner() func(t *testing.T) {
	return func(t *testing.T) {
		smsg, ok := tc.Msg.(hasGetSignersStrs)
		require.True(t, ok, "%T does not have a .GetSignerStrs method.", tc.Msg)

		var signers []string
		testFunc := func() {
			signers = smsg.GetSignerStrs()
		}
		require.NotPanics(t, testFunc, "GetSignerStrs")
		assert.Equal(t, tc.ExpSignersStrs, signers, "GetSignerStrs")
	}
}

// msgTestCases groups several sigTestCase entries together for a msg type.
type msgTestCases struct {
	TypeName  string
	HasLegacy bool
	HasStrs   bool
	TestCases []*sigTestCase
}

// newMsgTestCasesSingle creates a new msgTestCases for a single-signer message.
func newMsgTestCasesSingle(msgMaker MsgMaker) *msgTestCases {
	msg := msgMaker("")
	rv := &msgTestCases{TypeName: getTypeName(msg)}
	_, rv.HasLegacy = msg.(hasGetSigners)
	_, rv.HasStrs = msg.(hasGetSignersStrs)

	for _, tc := range singleSignerCases {
		var expSigners []sdk.AccAddress
		if len(tc.ExpSigner) > 0 {
			expSigners = []sdk.AccAddress{copyAccAddr(tc.ExpSigner)}
		}
		var expInErr []string
		if len(tc.ExpErr) > 0 {
			expInErr = []string{tc.ExpErr}
		}
		rv.TestCases = append(rv.TestCases, &sigTestCase{
			Name:           tc.Name,
			Msg:            msgMaker(tc.MsgSigner),
			ExpInErr:       expInErr,
			ExpSigners:     expSigners,
			ExpSignersBz:   addrsToBZs(expSigners),
			ExpSignersStrs: []string{tc.MsgSigner},
		})
	}

	return rv
}

// newMsgTestCasesMulti creates a new msgTestCases (empty) for a multi-signer message.
func newMsgTestCasesMulti(msgMakerMulti MsgMakerMulti) *msgTestCases {
	msg := msgMakerMulti(nil)
	rv := &msgTestCases{TypeName: getTypeName(msg)}
	_, rv.HasLegacy = msg.(hasGetSigners)
	_, rv.HasStrs = msg.(hasGetSignersStrs)

	for _, tc := range multiSignerCases {
		var expInErr []string
		if len(tc.ExpErr) > 0 {
			expInErr = []string{tc.ExpErr}
		}
		rv.TestCases = append(rv.TestCases, &sigTestCase{
			Name:           tc.Name,
			Msg:            msgMakerMulti(copyStrs(tc.MsgSigners)),
			ExpInErr:       expInErr,
			ExpSigners:     copyAccAddrs(tc.ExpSigners),
			ExpSignersBz:   addrsToBZs(tc.ExpSigners),
			ExpSignersStrs: copyStrs(tc.MsgSigners),
		})
	}

	return rv
}

// singleSignerCase is the definition of a test case involving a single signer. It's used for all single-signer msg types.
type singleSignerCase struct {
	Name      string
	MsgSigner string
	ExpSigner sdk.AccAddress
	ExpErr    string
}

// multiSignerCase is the definition of a test case involving multiple signers. It's used for all multi-signer msg types.
type multiSignerCase struct {
	Name       string
	MsgSigners []string
	ExpSigners []sdk.AccAddress
	ExpErr     string
}

const (
	emptyAddrErr = "empty address string is not allowed"
	bech32Err    = "decoding bech32 failed: "
)

var (
	addr1      = sdk.AccAddress("addr1_______________")
	addr2      = sdk.AccAddress("addr2_______________")
	addr3      = sdk.AccAddress("addr3_______________")
	testAddr   = sdk.AccAddress("testAddr____________")
	badAddrStr = "badaddr"
	badAddrErr = bech32Err + "invalid bech32 string length 7"

	singleSignerCases = []singleSignerCase{
		{Name: "no signer", MsgSigner: "", ExpErr: emptyAddrErr},
		{Name: "good signer", MsgSigner: testAddr.String(), ExpSigner: testAddr},
		{Name: "bad signer", MsgSigner: badAddrStr, ExpErr: badAddrErr},
	}

	multiSignerCases = []multiSignerCase{
		{Name: "no signers", MsgSigners: []string{}, ExpSigners: []sdk.AccAddress{}},
		{Name: "one good signer", MsgSigners: []string{addr1.String()}, ExpSigners: []sdk.AccAddress{addr1}},
		{Name: "one bad signer", MsgSigners: []string{badAddrStr}, ExpErr: badAddrErr},
		{
			Name:       "three good signers",
			MsgSigners: []string{addr1.String(), addr2.String(), addr3.String()},
			ExpSigners: []sdk.AccAddress{addr1, addr2, addr3},
		},
		{
			Name:       "three signers 1st bad",
			MsgSigners: []string{badAddrStr, addr2.String(), addr3.String()},
			ExpErr:     badAddrErr,
		},
		{
			Name:       "three signers 2nd bad",
			MsgSigners: []string{addr1.String(), badAddrStr, addr3.String()},
			ExpErr:     badAddrErr,
		},
		{
			Name:       "three signers 3rd bad",
			MsgSigners: []string{addr1.String(), addr2.String(), badAddrStr},
			ExpErr:     badAddrErr,
		},
	}
)

// appendIfNew appends newEntry to known if the newEntry is not in known.
func appendIfNew(known []string, newEntry string) []string {
	for _, entry := range known {
		if entry == newEntry {
			return known
		}
	}
	return append(known, newEntry)
}

// copyAccAddr creates a copy of the provided AccAddress.
func copyAccAddr(orig sdk.AccAddress) sdk.AccAddress {
	if orig == nil {
		return nil
	}
	rv := make(sdk.AccAddress, len(orig))
	copy(rv, orig)
	return rv
}

// copyAccAddrs creates a copy of the provided AccAddresses.
func copyAccAddrs(origs []sdk.AccAddress) []sdk.AccAddress {
	if origs == nil {
		return nil
	}
	rv := make([]sdk.AccAddress, len(origs))
	for i, orig := range origs {
		rv[i] = copyAccAddr(orig)
	}
	return rv
}

// copyStrs creates a copy of the provided string slice.
func copyStrs(origs []string) []string {
	if origs == nil {
		return nil
	}
	rv := make([]string, len(origs))
	copy(rv, origs)
	return rv
}

// addrsToBZs creates a copy of the provided byte slices.
func addrsToBZs(addrs []sdk.AccAddress) [][]byte {
	if len(addrs) == 0 {
		return nil
	}
	rv := make([][]byte, len(addrs))
	for i, addr := range addrs {
		rv[i] = copyAccAddr(addr)
	}
	return rv
}
