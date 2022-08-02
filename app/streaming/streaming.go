package streaming

import (
	"sort"
	"strings"

	"github.com/provenance-io/provenance/app/streaming/trace"
	"github.com/provenance-io/provenance/internal/streaming"
)

// StreamServiceInitializers contains a map of supported StreamServiceInitializer implementations
var StreamServiceInitializers = map[string]streaming.StreamServiceInitializer{
	"trace": trace.StreamServiceInitializer,
}

func ConfigOptions() string {
	keys := make([]string, 0, len(StreamServiceInitializers))
	for k := range StreamServiceInitializers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}
