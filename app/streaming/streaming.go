package streaming

// StreamServiceInitializers contains a map of predefined StreamServiceInitializer implementations
import (
	"github.com/provenance-io/provenance/app/streaming/kafka"
	"github.com/provenance-io/provenance/internal/streaming"
)

// StreamServiceInitializers contains a map of supported StreamServiceInitializer implementations
var StreamServiceInitializers = map[string]streaming.StreamServiceInitializer{
	"kafka": kafka.StreamServiceInitializer,
}
