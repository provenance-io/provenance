package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
	p := DefaultParams()

	require.NotNil(t, p)
	require.Equal(t, DefaultMinSegmentLength, p.MinSegmentLength)
	require.Equal(t, DefaultMaxSegmentLength, p.MaxSegmentLength)
	require.Equal(t, DefaultMaxNameLevels, p.MaxNameLevels)
	require.Equal(t, DefaultAllowUnrestrictedNames, p.AllowUnrestrictedNames)

	require.True(t, p.Equal(NewParams(DefaultMaxSegmentLength, DefaultMinSegmentLength, DefaultMaxNameLevels, DefaultAllowUnrestrictedNames)))
	require.False(t, p.Equal(NewParams(1, DefaultMinSegmentLength, DefaultMaxNameLevels, DefaultAllowUnrestrictedNames)))
	require.False(t, p.Equal(NewParams(DefaultMaxSegmentLength, 1, DefaultMaxNameLevels, DefaultAllowUnrestrictedNames)))
	require.False(t, p.Equal(NewParams(DefaultMaxSegmentLength, DefaultMinSegmentLength, 1, DefaultAllowUnrestrictedNames)))
	require.False(t, p.Equal(NewParams(DefaultMaxSegmentLength, DefaultMinSegmentLength, DefaultMaxNameLevels, false)))

	var p2 *Params
	require.True(t, p2.Equal(nil))
	require.False(t, p2.Equal(p))
}

func TestParamString(t *testing.T) {
	p := DefaultParams()
	require.Equal(t, `max_segment_length:32 min_segment_length:2 max_name_levels:16 allow_unrestricted_names:true `, p.String())
}
