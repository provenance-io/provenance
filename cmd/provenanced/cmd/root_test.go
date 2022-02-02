package cmd

import (
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/provenance-io/provenance/app"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIAVLConfig(t *testing.T) {
	require.Equal(t, getIAVLCacheSize(app.EmptyAppOptions{}), cast.ToInt(serverconfig.DefaultConfig().IAVLCacheSize))
}
