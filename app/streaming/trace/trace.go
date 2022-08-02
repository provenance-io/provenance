package trace

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/provenance-io/provenance/app/streaming/trace/service"
	"github.com/provenance-io/provenance/internal/streaming"
	"github.com/spf13/cast"
)

const (
	// TomlKey is the top-level TOML key for TraceStreamingService configuration
	TomlKey = "trace"

	// PrintDataToStdoutParam is the TraceStreamingService flag that logs full streamed data at debug level
	PrintDataToStdoutParam = "print-data-to-stdout"
)

var StreamServiceInitializer = &streamServiceInitializer{}

type streamServiceInitializer struct{}

var _ streaming.StreamServiceInitializer = (*streamServiceInitializer)(nil)

func (a streamServiceInitializer) Init(appOpts servertypes.AppOptions, marshaller codec.BinaryCodec) streaming.StreamService {
	// read in configuration properties
	tomlKeyPrefix := fmt.Sprintf("%s.%s", streaming.TomlKey, TomlKey)
	printDataToStdout := cast.ToBool(appOpts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, PrintDataToStdoutParam)))

	return service.NewTraceStreamingService(printDataToStdout, marshaller)
}
