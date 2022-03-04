package cmd

import (
	"testing"

	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/app"
)

func TestIAVLConfig(t *testing.T) {
	require.Equal(t, getIAVLCacheSize(app.EmptyAppOptions{}), cast.ToInt(serverconfig.DefaultConfig().IAVLCacheSize))
}
