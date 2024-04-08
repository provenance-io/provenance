package queries

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

func GetRequest[T proto.Message](val *network.Validator, url string, emptyResp T) (T, error) {
	respBz, err := testutil.GetRequestWithHeaders(url, nil)
	if err != nil {
		return emptyResp, fmt.Errorf("failed to execute GET %q: %w", url, err)
	}

	err = val.ClientCtx.Codec.UnmarshalJSON(respBz, emptyResp)
	if err != nil {
		return emptyResp, fmt.Errorf("failed to unmarshal GET %q response as %T: %w\nResponse: %s", url, emptyResp, err, string(respBz))
	}

	return emptyResp, nil
}
