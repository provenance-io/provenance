package provwasm_test

import (
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/proto"
	provapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/stretchr/testify/require"
)

func Test_StargateWhitelistedURLs(t *testing.T) {
	app := provapp.Setup(t)
	router := app.GRPCQueryRouter()

	paths := provwasm.GetStargateWhitelistedPaths()
	require.NotEmpty(t, paths, "expected stargate whitelist to contain at least one entry")

	sort.Strings(paths)

	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			parts := strings.Split(p, "/")
			require.Len(t, parts, 3, "path %q should follow /package.Service/Method format", p)
			require.NotEmpty(t, parts[1], "path %q missing service name", p)
			require.NotEmpty(t, parts[2], "path %q missing method name", p)

			_, err := provwasm.GetWhitelistedQuery(p)
			require.NoErrorf(t, err, "whitelist entry for path %q should return a valid proto type", p)

			route := router.Route(p)
			require.NotNilf(t, route, "expected router to have a route for path %q", p)
		})
	}
}
func Test_GetWhitelistedQueryReturnsFreshInstance(t *testing.T) {
	const path = "/cosmos.auth.v1beta1.Query/Account"

	first, err := provwasm.GetWhitelistedQuery(path)
	require.NoError(t, err, "first GetWhitelistedQuery call")
	require.NotNil(t, first)

	second, err := provwasm.GetWhitelistedQuery(path)
	require.NoError(t, err, "second GetWhitelistedQuery call")
	require.NotNil(t, second)

	require.NotSamef(t, first, second,
		"GetWhitelistedQuery must return a fresh instance per call; got the same pointer twice")

	require.Equal(t,
		reflect.TypeOf(first),
		reflect.TypeOf(second),
		"both calls should return the same concrete type")

	if acct, ok := first.(*authtypes.QueryAccountResponse); ok {
		acct.Reset()
	}
	require.NotSamef(t, first, second, "still distinct instances after mutation")
}

func Test_GetWhitelistedQueryConcurrentSafe(t *testing.T) {
	const path = "/cosmos.bank.v1beta1.Query/Balance"
	const goroutines = 64

	results := make(chan proto.Message, goroutines)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			m, err := provwasm.GetWhitelistedQuery(path)
			require.NoError(t, err)
			results <- m
		}()
	}
	wg.Wait()
	close(results)

	seen := make(map[proto.Message]struct{}, goroutines)
	for m := range results {
		_, dup := seen[m]
		require.Falsef(t, dup, "GetWhitelistedQuery returned a duplicate pointer under concurrency")
		seen[m] = struct{}{}
	}
	require.Lenf(t, seen, goroutines, "expected %d distinct pointers", goroutines)
}
