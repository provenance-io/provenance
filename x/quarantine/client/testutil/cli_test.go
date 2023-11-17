//go:build norace
// +build norace

package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
)

func TestIntegrationTestSuite(t *testing.T) {
	pioconfig.SetProvenanceConfig("stake", 0)
	cfg := testutil.DefaultTestNetworkConfig()
	cfg.NumValidators = 1
	cfg.TimeoutCommit = 500 * time.Millisecond
	suite.Run(t, NewIntegrationTestSuite(cfg))
}
