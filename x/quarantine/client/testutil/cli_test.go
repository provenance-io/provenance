//go:build norace
// +build norace

package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/network"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig()
	cfg.NumValidators = 2
	cfg.TimeoutCommit = 500 * time.Millisecond
	suite.Run(t, NewIntegrationTestSuite(cfg))
}
