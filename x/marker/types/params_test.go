package types

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	p := DefaultParams()

	require.NotNil(t, ParamKeyTable())

	require.NotNil(t, p)
	require.Equal(t, DefaultUnrestrictedDenomRegex, p.UnrestrictedDenomRegex)
	require.Equal(t, DefaultEnableGovernance, p.EnableGovernance)
	require.Equal(t, uint64(DefaultMaxTotalSupply), p.MaxTotalSupply)

	require.True(t, p.Equal(NewParams(DefaultMaxTotalSupply, DefaultEnableGovernance, DefaultUnrestrictedDenomRegex)))
	require.False(t, p.Equal(NewParams(1000, DefaultEnableGovernance, DefaultUnrestrictedDenomRegex)))
	require.False(t, p.Equal(NewParams(DefaultMaxTotalSupply, false, DefaultUnrestrictedDenomRegex)))
	require.False(t, p.Equal(NewParams(DefaultMaxTotalSupply, DefaultEnableGovernance, "a-z")))
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
	p := DefaultParams()
	require.Equal(t, `maxtotalsupply: 100000000000
enablegovernance: true
unrestricteddenomregex: '[a-zA-Z][a-zA-Z0-9\-\.]{2,83}'
`, p.String())
}

func TestParamSetPairs(t *testing.T) {
	p := DefaultParams()
	pairs := p.ParamSetPairs()
	require.Equal(t, 3, len(pairs))

	for i := range pairs {
		switch string(pairs[i].Key) {
		case string(ParamStoreKeyEnableGovernance):
			require.Error(t, pairs[i].ValidatorFn("foo"))
			require.NoError(t, pairs[i].ValidatorFn(true))
		case string(ParamStoreKeyMaxTotalSupply):
			require.Error(t, pairs[i].ValidatorFn("foo"))
			require.Error(t, pairs[i].ValidatorFn(-1000))
			require.NoError(t, pairs[i].ValidatorFn(uint64(1000)))
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
