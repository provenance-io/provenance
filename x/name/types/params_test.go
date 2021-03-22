package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	p := DefaultParams()

	require.NotNil(t, ParamKeyTable())

	require.NotNil(t, p)
	require.Equal(t, DefaultMinSegmentLength, p.MinSegmentLength)
	require.Equal(t, DefaultMaxSegmentLength, p.MaxSegmentLength)
	require.Equal(t, DefaultMaxSegments, p.MaxNameLevels)
	require.Equal(t, DefaultAllowUnrestrictedNames, p.AllowUnrestrictedNames)

	require.True(t, p.Equal(NewParams(DefaultMaxSegmentLength, DefaultMinSegmentLength, DefaultMaxSegments, DefaultAllowUnrestrictedNames)))
	require.False(t, p.Equal(NewParams(1, DefaultMinSegmentLength, DefaultMaxSegments, DefaultAllowUnrestrictedNames)))
	require.False(t, p.Equal(NewParams(DefaultMaxSegmentLength, 1, DefaultMaxSegments, DefaultAllowUnrestrictedNames)))
	require.False(t, p.Equal(NewParams(DefaultMaxSegmentLength, DefaultMinSegmentLength, 1, DefaultAllowUnrestrictedNames)))
	require.False(t, p.Equal(NewParams(DefaultMaxSegmentLength, DefaultMinSegmentLength, DefaultMaxSegments, false)))

	var p2 *Params
	require.True(t, p2.Equal(nil))
	require.False(t, p2.Equal(p))
}

func TestParamString(t *testing.T) {
	p := DefaultParams()
	require.Equal(t, `max_segment_length:32 min_segment_length:2 max_name_levels:16 allow_unrestricted_names:true `, p.String())
}

func TestParamSetPairs(t *testing.T) {
	p := DefaultParams()
	pairs := p.ParamSetPairs()
	require.Equal(t, 4, len(pairs))

	for i := range pairs {
		switch string(pairs[i].Key) {
		case string(ParamStoreKeyAllowUnrestrictedNames):
			require.Error(t, pairs[i].ValidatorFn("foo"))
			require.NoError(t, pairs[i].ValidatorFn(true))
		case string(ParamStoreKeyMaxSegmentLength):
			require.Error(t, pairs[i].ValidatorFn("foo"))
			require.Error(t, pairs[i].ValidatorFn(-1000))
			require.NoError(t, pairs[i].ValidatorFn(uint32(1000)))
		case string(ParamStoreKeyMinSegmentLength):
			require.Error(t, pairs[i].ValidatorFn("foo"))
			require.Error(t, pairs[i].ValidatorFn(-1000))
			require.NoError(t, pairs[i].ValidatorFn(uint32(1000)))
		case string(ParamStoreKeyMaxNameLevels):
			require.Error(t, pairs[i].ValidatorFn("foo"))
			require.Error(t, pairs[i].ValidatorFn(-1000))
			require.NoError(t, pairs[i].ValidatorFn(uint32(1000)))
		default:
			require.Fail(t, "unexpected param set pair")
		}
	}
}
