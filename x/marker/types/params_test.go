package types

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestDefaultParams(t *testing.T) {
	p := DefaultParams()

	require.NotNil(t, ParamKeyTable())

	require.NotNil(t, p)
	require.Equal(t, DefaultUnrestrictedDenomRegex, p.UnrestrictedDenomRegex)
	require.Equal(t, DefaultEnableGovernance, p.EnableGovernance)
	require.Equal(t, uint64(DefaultMaxTotalSupply), p.MaxTotalSupply)
	require.Equal(t, DefaultMaxSupply, p.MaxSupply.String())

	require.True(t, p.Equal(NewParams(DefaultMaxTotalSupply, DefaultEnableGovernance, DefaultUnrestrictedDenomRegex, StringToBigInt(DefaultMaxSupply))))
	require.False(t, p.Equal(NewParams(1000, DefaultEnableGovernance, DefaultUnrestrictedDenomRegex, StringToBigInt(DefaultMaxSupply))))
	require.False(t, p.Equal(NewParams(DefaultMaxTotalSupply, false, DefaultUnrestrictedDenomRegex, StringToBigInt(DefaultMaxSupply))))
	require.False(t, p.Equal(NewParams(DefaultMaxTotalSupply, DefaultEnableGovernance, "a-z", StringToBigInt(DefaultMaxSupply))))
	require.False(t, p.Equal(NewParams(DefaultMaxTotalSupply, DefaultEnableGovernance, DefaultUnrestrictedDenomRegex, StringToBigInt("1000"))))
	require.False(t, p.Equal(nil))

	var p2 *Params
	require.True(t, p2.Equal(nil))
	require.False(t, p2.Equal(p))

	r := p.GetUnrestrictedDenomRegex()
	require.NotEmpty(t, r)

	_, err := regexp.Compile(r)
	require.NoError(t, err)
}

func TestParamString(t *testing.T) {
	expected := `max_total_supply:100000000000 ` +
		`enable_governance:true ` +
		`unrestricted_denom_regex:"[a-zA-Z][a-zA-Z0-9\\-\\.]{2,83}" ` +
		`max_supply:"100000000000000000000" `
	p := DefaultParams()
	actual := p.String()
	require.Equal(t, expected, actual)
}

func TestParamSetPairs(t *testing.T) {
	p := DefaultParams()
	pairs := p.ParamSetPairs()
	require.Equal(t, 4, len(pairs))

	for i := range pairs {
		switch string(pairs[i].Key) {
		case string(ParamStoreKeyEnableGovernance):
			require.Error(t, pairs[i].ValidatorFn("foo"))
			require.NoError(t, pairs[i].ValidatorFn(true))
		case string(ParamStoreKeyMaxTotalSupply):
			require.Error(t, pairs[i].ValidatorFn("foo"))
			require.Error(t, pairs[i].ValidatorFn(-1000))
			require.NoError(t, pairs[i].ValidatorFn(uint64(1000)))
		case string(ParamStoreKeyMaxSupply):
			require.Error(t, pairs[i].ValidatorFn("foo"))
			require.Error(t, pairs[i].ValidatorFn(-1000))
			require.Error(t, pairs[i].ValidatorFn(1000))
			bigint, _ := sdkmath.NewIntFromString("1944674407370955516150")
			require.NoError(t, pairs[i].ValidatorFn(bigint))
			require.NoError(t, pairs[i].ValidatorFn(sdkmath.NewInt(1000)))
		case string(ParamStoreKeyUnrestrictedDenomRegex):
			require.Error(t, pairs[i].ValidatorFn(1))
			require.Error(t, pairs[i].ValidatorFn("\\!(")) // invalid regex
			require.NoError(t, pairs[i].ValidatorFn("[a-z].*"))

			// Prohibit use of anchors (these are always enforced and will be added to every expression)
			require.Error(t, pairs[i].ValidatorFn("^[a-z].*"))
			require.Error(t, pairs[i].ValidatorFn("^[a-z].*$"))
			require.Error(t, pairs[i].ValidatorFn("[a-z].*$"))

			// If the expression contains the anchors but they are not at the end of the expression that is allowed (however unrealistic)
			require.NoError(t, pairs[i].ValidatorFn("[a-z].*$."))
			require.NoError(t, pairs[i].ValidatorFn(".^[a-z].*$."))

		default:
			require.Fail(t, "unexpected param set pair")
		}
	}
}

func TestStringToBigInt(t *testing.T) {
	require.Equal(t, sdkmath.NewIntFromUint64(1), StringToBigInt("1"), "should handle uint64")
	require.Equal(t, sdkmath.NewIntFromUint64(0), StringToBigInt("0"), "should handle zero")
	require.Equal(t, "-1", StringToBigInt("-1").String(), "should handle negative")
	assertions.AssertPanicEquals(t, func() {
		StringToBigInt("abc")
	}, "unable to create sdkmath.Int from string: abc", "should panic on invalid input")
	bigNum, _ := sdkmath.NewIntFromString("100000000000000000000")
	require.Equal(t, bigNum, StringToBigInt("100000000000000000000"), "should handle large number that exceeds uint64")
}
