package streaming

import (
	"github.com/provenance-io/provenance/app/streaming/kafka"
	"github.com/provenance-io/provenance/app/streaming/trace"
	"github.com/provenance-io/provenance/internal/streaming"
)

// StreamServiceInitializers contains a map of supported StreamServiceInitializer implementations
var StreamServiceInitializers = map[string]streaming.StreamServiceInitializer{
	"kafka": kafka.StreamServiceInitializer,
	"trace": trace.StreamServiceInitializer,
}
