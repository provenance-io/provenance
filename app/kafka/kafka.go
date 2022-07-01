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

	// Enabled is the KafkaStreamingService flag that enabled streaming to Kafka
	EnableKafkaStreamingParam = "enabled"

	// TopicPrefixParam is the KafkaStreamingService topic where events will be streamed to
	TopicPrefixParam = "topic_prefix"

	// FlushTimeoutMsParam is the timeout setting passed to the producer.Flush(timeoutMs)
	FlushTimeoutMsParam = "flush_timeout_ms"

	// KafkaProducerTomlKey is the top-level key for the KafkaStreamingService Producer configuration properties
	KafkaProducerTomlKey = "producer"

	// HaltAppOnDeliveryErrorParam whether or not to halt the application when plugin fails to deliver message(s)
	HaltAppOnDeliveryErrorParam = "halt_app_on_delivery_error"
)

// EventTypeValueTypeTopic Kafka topic name enum types
type EventTypeValueTypeTopic string
const (
	BeginBlockReqTopic EventTypeValueTypeTopic = "begin-block-req"
	BeginBlockResTopic                         = "begin-block-res"
	EndBlockReqTopic                           = "end-block-req"
	EndBlockResTopic                           = "end-block-res"
)

// KafkaStreamingService is a concrete implementation of streaming.Service that writes state changes out to KafkaStreamingService
type StreamingService struct {
	topicPrefix            string            // topicPrefix prefix name
	producer               *kafka.Producer   // the producer instance that will be used to send messages to KafkaStreamingService
	flushTimeoutMs         int               // the time to wait for outstanding messages and requests to complete delivery (milliseconds)
	codec                  codec.BinaryCodec // binary marshaller used for re-marshalling the ABCI messages to write them out to the destination files
	deliveryChan           chan kafka.Event  // KafkaStreamingService producer delivery report channel
	haltAppOnDeliveryError bool              // true if the app should be halted on streaming errors, false otherwise
}

func NewStreamingService(appOpts serverTypes.AppOptions, marshaller codec.BinaryCodec) *StreamingService {
	// load all the params required from the provided AppOptions
	topicPrefix := cast.ToString(appOpts.Get(fmt.Sprintf("%s.%s", TomlKey, TopicPrefixParam)))
	flushTimeoutMs := cast.ToInt(appOpts.Get(fmt.Sprintf("%s.%s", TomlKey, FlushTimeoutMsParam)))
	haltAppOnDeliveryError := cast.ToBool(appOpts.Get(fmt.Sprintf("%s.%s", TomlKey, HaltAppOnDeliveryErrorParam)))
	producerConfig := cast.ToStringMap(appOpts.Get(fmt.Sprintf("%s.%s", TomlKey, KafkaProducerTomlKey)))

	// Validate minimum producer config properties
	producerConfigKey := fmt.Sprintf("%s.%s", TomlKey, KafkaProducerTomlKey)

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
		if !haltAppOnDeliveryError {
			// disable delivery reports when operating in fire-and-forget fashion
			if err := producerConfigMap.SetKey("go.delivery.reports", false); err != nil {
				panic(err)
			}
		}
	}

	// Initialize the producer and connect to KafkaStreamingService cluster
	p, err := kafka.NewProducer(&producerConfigMap)
	if err != nil {
		if haltAppOnDeliveryError {
			panic(err)
		} else {
			// producing to Kafka will be ignored
			// TODO: should we log a warning here?
			return nil
		}
	}

	kss := &StreamingService{
		topicPrefix: topicPrefix,
		producer: p,
		flushTimeoutMs: flushTimeoutMs,
		codec: marshaller,
		haltAppOnDeliveryError: haltAppOnDeliveryError,
	}

	// setup private delivery channel to listen for delivery errors.
	if haltAppOnDeliveryError {
		kss.deliveryChan = make(chan kafka.Event)
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
		if kss.haltAppOnDeliveryError {
			panic(err)
		}
	}

	// write res
	if err := kss.writeToKafka(ctx, BeginBlockResTopic, &res); err != nil {
		ctx.Logger().Error(errMsg, "height", ctx.BlockHeight(), "err", err)
		if kss.haltAppOnDeliveryError {
			panic(err)
		}
	}
}

// ListenEndBlocker writes out the received BeginBlockEvent request and response to Kafka
func (kss *StreamingService) ListenEndBlocker(
	ctx sdk.Context,
	req abci.RequestEndBlock,
	res abci.ResponseEndBlock,
) {
	errMsg := "EndBlocker listening hook failed"

	// write req
	if err := kss.writeToKafka(ctx, EndBlockReqTopic, &req); err != nil {
		ctx.Logger().Error(errMsg, "height", ctx.BlockHeight(), "err", err)
		if kss.haltAppOnDeliveryError {
			panic(err)
		}
	}

	// write res
	if err := kss.writeToKafka(ctx, EndBlockResTopic, &res); err != nil {
		ctx.Logger().Error(errMsg, "height", ctx.BlockHeight(), "err", err)
		if kss.haltAppOnDeliveryError {
			panic(err)
		}
	}
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
		Key: 			[]byte(strconv.FormatInt(ctx.BlockHeight(), 10)),
		Value:          value,
	}, kss.deliveryChan); err != nil {
		return err
	}

	return kss.checkDeliveryReport(ctx)
}

// checkDeliveryReport checks kafka.Producer delivery report for successful or failed messages
func (kss *StreamingService) checkDeliveryReport(ctx sdk.Context) error {
	if kss.deliveryChan == nil {
		return nil
	}

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
