package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/provenance-io/provenance/x/authz/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

func TestGenericAuthorization(t *testing.T) {
	t.Log("verify ValidateBasic returns error for non-service msg")
	authorization := types.NewGenericAuthorization(banktypes.TypeMsgSend)
	require.Error(t, authorization.ValidateBasic())

	t.Log("verify ValidateBasic returns nil for service msg")
	authorization = types.NewGenericAuthorization(markertypes.MarkerSendAuthorization{}.MethodName())
	require.NoError(t, authorization.ValidateBasic())
	require.Equal(t, markertypes.MarkerSendAuthorization{}.MethodName(), authorization.MessageName)
}
