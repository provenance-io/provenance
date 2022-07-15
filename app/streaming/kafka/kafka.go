package kafka

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/provenance-io/provenance/app/streaming/kafka/service"
	"github.com/provenance-io/provenance/internal/streaming"
	"github.com/spf13/cast"
	"strings"
)

// TOML configuration parameter keys
const (
	// TomlKey is the top-level TOML key for KafkaStreamingService configuration
	TomlKey = "kafka"

	// TopicPrefixParam is the KafkaStreamingService topic where events will be streamed to
	TopicPrefixParam = "topic_prefix"

	// ProducerTomlKey is the top-level key for the Kafka Producer configuration properties
	ProducerTomlKey = "producer"
)

var StreamServiceInitializer = &streamServiceInitializer{}

type streamServiceInitializer struct{}

var _ streaming.StreamServiceInitializer = (*streamServiceInitializer)(nil)

func (a streamServiceInitializer) Init(appOpts servertypes.AppOptions, marshaller codec.BinaryCodec) streaming.StreamService {
	// load all the params required from the provided AppOptions
	tomlKeyPrefix := fmt.Sprintf("%s.%s", streaming.TomlKey, TomlKey)
	topicPrefix := cast.ToString(appOpts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, TopicPrefixParam)))
	producerConfig := cast.ToStringMap(appOpts.Get(fmt.Sprintf("%s.%s", tomlKeyPrefix, ProducerTomlKey)))
	producerConfigKey := fmt.Sprintf("%s.%s", tomlKeyPrefix, ProducerTomlKey)

	// Validate minimum required producer config properties
	if len(producerConfig) == 0 {
		panic(fmt.Errorf("unable to connect to Kafka cluster: unset properties for '%s': ", producerConfigKey))
	} else {
		bootstrapServers := strings.TrimSpace(cast.ToString(producerConfig["bootstrap_servers"]))
		if len(bootstrapServers) == 0 {
			panic(fmt.Errorf("unable to connect to Kafka cluster: unset property \"%s.%s\" ", producerConfigKey, "bootstrap_servers"))
		}
	}

	// load producer config into a kafka.ConfigMap
	producerConfigMap := kafka.ConfigMap{}
	for key, element := range producerConfig {
		key = strings.ReplaceAll(key, "_", ".")
		if err := producerConfigMap.SetKey(key, element); err != nil {
			panic(err)
		}
	}

	return service.NewKafkaStreamingService(topicPrefix, &producerConfigMap, marshaller)
}
