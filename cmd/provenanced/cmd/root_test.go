package cmd

import (
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"

	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdksim "github.com/cosmos/cosmos-sdk/simapp"
)

func TestIAVLConfig(t *testing.T) {
	require.Equal(t, getIAVLCacheSize(sdksim.EmptyAppOptions{}), cast.ToInt(serverconfig.DefaultConfig().IAVLCacheSize))
}
