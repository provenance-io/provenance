package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateParams(t *testing.T) {
	osParam := NewOSLocatorParams(3000)
	require.Equal(t, int(3000), int(osParam.MaxUriLength))
}

func TestDefault(t *testing.T) {
	metadataData := DefaultOSLocatorParams()
	require.Equal(t, 2048, int(metadataData.MaxUriLength))
}
