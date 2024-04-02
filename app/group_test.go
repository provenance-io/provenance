package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	cerrs "cosmossdk.io/errors"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

func TestNewGroupCheckerFunc(t *testing.T) {
	querier := NewMockGroupPolicyQuerier(true)
	checker := NewGroupCheckerFunc(querier)
	assert.NotNil(t, checker, "should return a group checker function")
}

func TestIsGroupAddress(t *testing.T) {
	tests := []struct {
		name         string
		querySuccess bool
		address      sdk.AccAddress
	}{
		{
			name:         "should be true with group address",
			querySuccess: true,
			address:      sdk.AccAddress("test"),
		},
		{
			name:         "should return false with non group address",
			querySuccess: false,
			address:      sdk.AccAddress("test"),
		},
		{
			name:         "should return false with nil address",
			querySuccess: false,
			address:      nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			querier := NewMockGroupPolicyQuerier(tc.querySuccess)
			checker := NewGroupCheckerFunc(querier)
			ctx := sdk.NewContext(nil, cmtproto.Header{}, true, nil)
			success := checker.IsGroupAddress(ctx, tc.address)
			assert.Equal(t, tc.querySuccess, success, "should correctly detect if the supplied address is a group address")
		})
	}
}

// MockGroupPolicyQuerier mocks the querier so a GroupKeeper isn't needed.
type MockGroupPolicyQuerier struct {
	isGroupAddress bool
}

// NewMockGroupPolicyQuerier creates a new MockGroupPolicyQuerier.
func NewMockGroupPolicyQuerier(isGroupAddress bool) *MockGroupPolicyQuerier {
	return &MockGroupPolicyQuerier{
		isGroupAddress: isGroupAddress,
	}
}

// GroupPolicyInfo provides a stubbed implementation of the GroupPolicyInfo method.
func (t MockGroupPolicyQuerier) GroupPolicyInfo(goCtx context.Context, request *group.QueryGroupPolicyInfoRequest) (*group.QueryGroupPolicyInfoResponse, error) {
	var err error
	if !t.isGroupAddress {
		err = cerrs.New("", 1, "")
	}
	return nil, err
}
