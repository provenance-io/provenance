package provwasm_test

import (
	"sort"
	"strings"
	"testing"

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
