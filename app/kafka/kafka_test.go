package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	types2 "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	types1 "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	interfaceRegistry    = codecTypes.NewInterfaceRegistry()
	testMarshaller       = codec.NewProtoCodec(interfaceRegistry)
	testStreamingService *StreamingService
	testingCtx           sdk.Context

	// test abci message types
	mockHash          = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	testBeginBlockReq = abci.RequestBeginBlock{
		Header: types1.Header{
			Height: 1,
		},
		ByzantineValidators: []abci.Evidence{},
		Hash:                mockHash,
		LastCommitInfo: abci.LastCommitInfo{
			Round: 1,
			Votes: []abci.VoteInfo{},
		},
	}
	testBeginBlockRes = abci.ResponseBeginBlock{
		Events: []abci.Event{
			{
				Type: "testEventType1",
			},
			{
				Type: "testEventType2",
			},
		},
	}
	testEndBlockReq = abci.RequestEndBlock{
		Height: 1,
	}
	testEndBlockRes = abci.ResponseEndBlock{
		Events:                []abci.Event{},
		ConsensusParamUpdates: &abci.ConsensusParams{},
		ValidatorUpdates:      []abci.ValidatorUpdate{},
	}

	// Kafka stuff
	topics = []string{
		string(BeginBlockReqTopic),
		BeginBlockResTopic,
		EndBlockReqTopic,
		EndBlockResTopic,
	}
	bootstrapServers = "localhost:9092"
	timeoutMs        = 5000
	appOpts          = loadAppOptions()
	topicPrefix      = cast.ToString(appOpts.Get(fmt.Sprintf("%s.%s", TomlKey, TopicPrefixParam)))
)

// change this to write to in-memory io.Writer (e.g. bytes.Buffer)
func TestStreamingService(t *testing.T) {
	testingCtx = sdk.NewContext(nil, types1.Header{}, false, log.TestingLogger())
	testStreamingService = NewStreamingService(appOpts, testMarshaller)
	require.NotNil(t, testStreamingService)
	require.IsType(t, &StreamingService{}, testStreamingService)
	deleteTopics(t, topics, bootstrapServers)
	createTopics(t, topics, bootstrapServers)
	testListenBeginBlocker(t)
	testListenEndBlocker(t)
	testStreamingService.Close(timeoutMs)
}

func testListenBeginBlocker(t *testing.T) {
	expectedBeginBlockReqBytes, err := marshal(&testBeginBlockReq)
	require.Nil(t, err)
	expectedBeginBlockResBytes, err := marshal(&testBeginBlockRes)
	require.Nil(t, err)

	// send the ABCI messages
	testStreamingService.ListenBeginBlocker(testingCtx, testBeginBlockReq, testBeginBlockRes)

	// consume stored messages
	// expectedMsgCnt must equal messages sent count so the consumer can close (not stay alive for more messages).
	expectedMsgCnt := 2
	topics := []string{string(BeginBlockReqTopic), BeginBlockResTopic}
	msgs, err := poll(bootstrapServers, topics, expectedMsgCnt)
	require.Nil(t, err)

	// validate data stored in Kafka
	require.Equal(t, expectedBeginBlockReqBytes, getMessageValueForTopic(msgs, string(BeginBlockReqTopic), 0))
	require.Equal(t, expectedBeginBlockResBytes, getMessageValueForTopic(msgs, BeginBlockResTopic, 0))
}

func testListenEndBlocker(t *testing.T) {
	expectedEndBlockReqBytes, err := marshal(&testEndBlockReq)
	require.Nil(t, err)
	expectedEndBlockResBytes, err := marshal(&testEndBlockRes)
	require.Nil(t, err)

	// send the ABCI messages
	testStreamingService.ListenEndBlocker(testingCtx, testEndBlockReq, testEndBlockRes)

	// consume stored messages
	// expectedMsgCnt must equal messages sent count so the consumer can close (not stay alive for more messages).
	expectedMsgCnt := 2
	topics := []string{EndBlockReqTopic, EndBlockResTopic}
	msgs, err := poll(bootstrapServers, topics, expectedMsgCnt)
	require.Nil(t, err)

	// validate data stored in Kafka
	require.Equal(t, expectedEndBlockReqBytes, getMessageValueForTopic(msgs, EndBlockReqTopic, 0))
	require.Equal(t, expectedEndBlockResBytes, getMessageValueForTopic(msgs, EndBlockResTopic, 0))
}

func marshal(o codec.ProtoMarshaler) ([]byte, error) {
	return testMarshaller.Marshal(o)
}

func getMessageValueForTopic(msgs []*kafka.Message, topic string, offset int64) []byte {
	topic = fmt.Sprintf("%s-%s", topicPrefix, topic)
	for _, m := range msgs {
		t := *m.TopicPartition.Topic
		o := int64(m.TopicPartition.Offset)
		if t == topic && o == offset {
			return m.Value
		}
	}
	return []byte{0}
}

func poll(bootstrapServers string, topics []string, expectedMsgCnt int) ([]*kafka.Message, error) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		// Avoid connecting to IPv6 brokers:
		// This is needed for the ErrAllBrokersDown show-case below
		// when using localhost brokers on OSX, since the OSX resolver
		// will return the IPv6 addresses first.
		// You typically don't need to specify this configuration property.
		"broker.address.family": "v4",
		"group.id":              fmt.Sprintf("testGroup-%d", os.Process{}.Pid),
		"auto.offset.reset":     "earliest"})

	if err != nil {
		panic(fmt.Sprintf("Failed to create consumer: %s\n", err))
	}

	fmt.Printf("Created Consumer %v\n", c)

	var _topics []string
	for _, t := range topics {
		_topics = append(_topics, fmt.Sprintf("%s-%s", topicPrefix, t))
	}

	if err = c.SubscribeTopics(_topics, nil); err != nil {
		panic(fmt.Sprintf("Failed to subscribe to consumer: %s\n", err))
	}

	msgs := make([]*kafka.Message, 0)

	run := true

	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				msgs = append(msgs, e)
			case kafka.Error:
				// Errors should generally be considered
				// informational, the client will try to
				// automatically recover.
				// But in this example we choose to terminate
				// the application if all brokers are down.
				fmt.Printf("%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("Ignored %v\n", e)

				// Workaround so our tests pass.
				// Wait for the expected messages to be delivered before closing the consumer
				if expectedMsgCnt == len(msgs) {
					run = false
				}
			}
		}
	}

	fmt.Printf("Closing consumer\n")
	if err := c.Close(); err != nil {
		return nil, err
	}

	return msgs, nil
}

func createTopics(t *testing.T, topics []string, bootstrapServers string) {

	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers":       bootstrapServers,
		"broker.version.fallback": "0.10.0.0",
		"api.version.fallback.ms": 0,
	})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		t.Fail()
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create topics on cluster.
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDuration, err := time.ParseDuration("60s")
	if err != nil {
		fmt.Printf("time.ParseDuration(60s)")
		t.Fail()
	}

	var _topics []kafka.TopicSpecification
	for _, s := range topics {
		_topics = append(_topics,
			kafka.TopicSpecification{
				Topic:             fmt.Sprintf("%s-%s", topicPrefix, s),
				NumPartitions:     1,
				ReplicationFactor: 1})
	}
	results, err := adminClient.CreateTopics(ctx, _topics, kafka.SetAdminOperationTimeout(maxDuration))
	if err != nil {
		fmt.Printf("Problem during the topicPrefix creation: %v\n", err)
		t.Fail()
	}

	// Check for specific topicPrefix errors
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError &&
			result.Error.Code() != kafka.ErrTopicAlreadyExists {
			fmt.Printf("Topic creation failed for %s: %v",
				result.Topic, result.Error.String())
			t.Fail()
		}
	}

	adminClient.Close()
}

func deleteTopics(t *testing.T, topics []string, bootstrapServers string) {
	// Create a new AdminClient.
	// AdminClient can also be instantiated using an existing
	// Producer or Consumer instance, see NewAdminClientFromProducer and
	// NewAdminClientFromConsumer.
	a, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": bootstrapServers})
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		t.Fail()
	}

	// Contexts are used to abort or limit the amount of time
	// the Admin call blocks waiting for a result.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Delete topics on cluster
	// Set Admin options to wait for the operation to finish (or at most 60s)
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		fmt.Printf("ParseDuration(60s)")
		t.Fail()
	}

	var _topics []string
	for _, s := range topics {
		_topics = append(_topics, fmt.Sprintf("%s-%s", topicPrefix, s))
	}

	results, err := a.DeleteTopics(ctx, _topics, kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Failed to delete topics: %v\n", err)
		t.Fail()
	}

	// Print results
	for _, result := range results {
		fmt.Printf("%s\n", result)
	}

	a.Close()
}

func loadAppOptions() types2.AppOptions {
	m := make(map[string]interface{})
	m["kafka.enabled"] = true
	m["kafka.topic_prefix"] = "unittest"
	// Kafka plugin producer
	m["kafka.producer.bootstrap_servers"] = "localhost:9092"
	m["kafka.producer.client_id"] = "kafka_test.go"
	m["kafka.producer.acks"] = "all"
	m["kafka.producer.enable_idempotence"] = true

	vpr := viper.New()
	for key, value := range m {
		vpr.SetDefault(key, value)
	}

	return vpr
}
