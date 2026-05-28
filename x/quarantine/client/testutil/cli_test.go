package testutil

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
)

func TestIntegrationTestSuite(t *testing.T) {
	pioconfig.SetProvConfig(sdk.DefaultBondDenom)
	govv1.DefaultMinDepositRatio = sdkmath.LegacyZeroDec()
	cfg := testutil.DefaultTestNetworkConfig()
	cfg.NumValidators = 2

	suite.Run(t, NewIntegrationTestSuite(cfg))
}
