package kafka

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"strconv"
	"strings"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/codec"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
)

// TOML configuration parameter keys
const (
	// TomlKey is the top-level TOML key for KafkaStreamingService configuration
	TomlKey = "kafka"

	// EnableKafkaStreamingParam is the KafkaStreamingService flag that enabled streaming to Kafka
	EnableKafkaStreamingParam = "enabled"

	// TopicPrefixParam is the KafkaStreamingService topic where events will be streamed to
	TopicPrefixParam = "topic_prefix"

	// ProducerTomlKey is the top-level key for the KafkaStreamingService Producer configuration properties
	ProducerTomlKey = "producer"
)

// EventTypeTopic Kafka topic name enum types
type EventTypeTopic string

const (
	BeginBlockReqTopic EventTypeTopic = "begin-block-req"
	BeginBlockResTopic                = "begin-block-res"
	EndBlockReqTopic                  = "end-block-req"
	EndBlockResTopic                  = "end-block-res"
)

// StreamingService wraps a high-level [*kafka.Producer] instance
type StreamingService struct {
	topicPrefix  string            // topicPrefix prefix name
	producer     *kafka.Producer   // the producer instance that will be used to send messages to KafkaStreamingService
	codec        codec.BinaryCodec // binary marshaller used for re-marshalling the ABCI messages to write them out to the destination files
	deliveryChan chan kafka.Event  // KafkaStreamingService producer delivery report channel
}

func NewStreamingService(appOpts serverTypes.AppOptions, marshaller codec.BinaryCodec) *StreamingService {
	// load all the params required from the provided AppOptions
	topicPrefix := cast.ToString(appOpts.Get(fmt.Sprintf("%s.%s", TomlKey, TopicPrefixParam)))
	producerConfig := cast.ToStringMap(appOpts.Get(fmt.Sprintf("%s.%s", TomlKey, ProducerTomlKey)))

	// Validate minimum producer config properties
	producerConfigKey := fmt.Sprintf("%s.%s", TomlKey, ProducerTomlKey)

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

	// Initialize the producer and connect to KafkaStreamingService cluster
	p, err := kafka.NewProducer(&producerConfigMap)
	if err != nil {
		panic(err)
	}

	kss := &StreamingService{
		topicPrefix:  topicPrefix,
		producer:     p,
		codec:        marshaller,
		deliveryChan: make(chan kafka.Event),
	}

	return kss
}

// ListenBeginBlocker writes out the received BeginBlockEvent request and response to Kafka
func (kss *StreamingService) ListenBeginBlocker(
	ctx sdk.Context,
	req abci.RequestBeginBlock,
	res abci.ResponseBeginBlock,
) {
	errMsg := "BeginBlocker listening hook failed"

	// write req
	if err := kss.writeToKafka(ctx, string(BeginBlockReqTopic), &req); err != nil {
		ctx.Logger().Error(errMsg, "height", req.Header.Height, "err", err)
		panic(err)
	}

	// write res
	if err := kss.writeToKafka(ctx, BeginBlockResTopic, &res); err != nil {
		ctx.Logger().Error(errMsg, "height", ctx.BlockHeight(), "err", err)
		panic(err)
	}
}

// ListenEndBlocker writes out the received app.EndBlocker request and response to Kafka
func (kss *StreamingService) ListenEndBlocker(
	ctx sdk.Context,
	req abci.RequestEndBlock,
	res abci.ResponseEndBlock,
) {
	errMsg := "EndBlocker listening hook failed"

	// write req
	if err := kss.writeToKafka(ctx, EndBlockReqTopic, &req); err != nil {
		ctx.Logger().Error(errMsg, "height", ctx.BlockHeight(), "err", err)
		panic(err)
	}

	// write res
	if err := kss.writeToKafka(ctx, EndBlockResTopic, &res); err != nil {
		ctx.Logger().Error(errMsg, "height", ctx.BlockHeight(), "err", err)
		panic(err)
	}
}

// Close a Producer instance. The Producer object or its channels are no longer usable after this call.
//
// Flush and wait for outstanding messages and requests to complete delivery.
// Includes messages on ProduceChannel. Runs until value reaches zero or on timeoutMs.
// Returns the number of outstanding events still un-flushed
//
// NOTE: This is solely to be used for testing purposes.
func (kss *StreamingService) Close(timeoutMs int) {
	kss.producer.Flush(timeoutMs)
	close(kss.deliveryChan)
	kss.producer.Close()
}

func (kss *StreamingService) writeToKafka(
	ctx sdk.Context,
	topic string,
	msgValue codec.ProtoMarshaler,
) error {
	value, err := kss.codec.Marshal(msgValue)
	if err != nil {
		return err
	}

	if len(kss.topicPrefix) > 0 {
		topic = fmt.Sprintf("%s-%s", kss.topicPrefix, topic)
	}

	// produce message
	// when `halt_app_on_delivery_error = false`, kss.deliveryChan is `nil`
	// and the producer is configured with `go.delivery.reports: false`
	// this means that the producer operates in a fire-and-forget mode
	if err := kss.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(strconv.FormatInt(ctx.BlockHeight(), 10)),
		Value:          value,
	}, kss.deliveryChan); err != nil {
		return err
	}

	return kss.checkDeliveryReport(ctx)
}

// checkDeliveryReport checks [*kafka.Producer] delivery report for successful or failed messages
func (kss *StreamingService) checkDeliveryReport(ctx sdk.Context) error {
	e := <-kss.deliveryChan
	m := e.(*kafka.Message)
	topic := *m.TopicPartition.Topic
	partition := m.TopicPartition.Partition
	offset := m.TopicPartition.Offset
	key := string(m.Key)
	topicErr := m.TopicPartition.Error
	logger := ctx.Logger()

	if topicErr != nil {
		logger.Error("Delivery failed: ", "topic", topic, "partition", partition, "key", key, "err", topicErr)
	} else {
		logger.Debug("Delivered message:", "topic", topic, "partition", partition, "offset", offset, "key", key)
	}

	return topicErr
}
