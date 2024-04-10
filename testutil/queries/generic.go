package queries

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

// AssertGetRequest does an HTTP get on the provided url and unmarshalls the response into the provided emptyResp.
// The returned bool will be true on success, or false if something goes wrong.
func AssertGetRequest[T proto.Message](t *testing.T, val *network.Validator, url string, emptyResp T) (T, bool) {
	t.Helper()
	respBz, err := testutil.GetRequestWithHeaders(url, nil)
	if !assert.NoError(t, err, "failed to execute GET %q", url) {
		return emptyResp, false
	}
	t.Logf("GET %q\nResponse: %s", url, string(respBz))

	err = val.ClientCtx.Codec.UnmarshalJSON(respBz, emptyResp)
	if !assert.NoError(t, err, "failed to unmarshal response as %T", emptyResp) {
		return emptyResp, false
	}

	return emptyResp, true
}
