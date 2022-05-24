package keeper_test

import (
	sdkTypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

/*func (suite *KeeperTestSuite) TestNewShare() {
	suite.SetupTest()

	time := timestamppb.Now().AsTime()
	share := types.NewShare(
		1,
		2,
		"test",
		true,
		time,
		5,
	)

	suite.Assert().Equal(uint64(1), share.GetRewardProgramId(), "reward program id must match")
	suite.Assert().Equal(uint64(2), share.GetEpochId(), "epoch id must match")
	suite.Assert().Equal("test", share.GetAddress(), "address must match")
	suite.Assert().Equal(true, share.GetClaimed(), "claim status must match")
	suite.Assert().Equal(time, share.GetExpireTime(), "expiration time must match")
	suite.Assert().Equal(int64(5), share.GetAmount(), "share amount must match")
}*/

func setupEventHistory(suite *KeeperTestSuite) {
	attributes1 := []sdkTypes.Attribute{
		sdkTypes.NewAttribute("key1", "value1"),
		sdkTypes.NewAttribute("key2", "value2"),
		sdkTypes.NewAttribute("key3", "value3"),
	}
	attributes2 := []sdkTypes.Attribute{
		sdkTypes.NewAttribute("key1", "value1"),
		sdkTypes.NewAttribute("key3", "value2"),
		sdkTypes.NewAttribute("key4", "value3"),
	}
	event1 := sdkTypes.NewEvent("event1", attributes1...)
	event2 := sdkTypes.NewEvent("event2", attributes2...)
	event3 := sdkTypes.NewEvent("event1", attributes1...)
	loggedEvents := sdkTypes.Events{
		event1,
		event2,
		event3,
	}
	eventManagerStub := sdkTypes.NewEventManagerWithHistory(loggedEvents.ToABCIEvents())
	suite.ctx = suite.ctx.WithEventManager(eventManagerStub)
}

func (suite *KeeperTestSuite) TestIterateABCIEventsWildcard() {
	suite.SetupTest()
	setupEventHistory(suite)
	events := []types.ABCIEvent{}
	criteria := types.NewEventCriteria(events)
	counter := 0
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(3, counter, "should iterate for each event")
}

func (suite *KeeperTestSuite) TestIterateABCIEventsByEventType() {
	suite.SetupTest()
	setupEventHistory(suite)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type: "event1",
		},
	}
	criteria := types.NewEventCriteria(events)
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(2, counter, "should iterate only for specified event types")
}

func (suite *KeeperTestSuite) TestIterateABCIEventsByEventTypeAndAttributeName() {
	suite.SetupTest()
	setupEventHistory(suite)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type: "event1",
			Attributes: map[string][]byte{
				"key1": nil,
			},
		},
	}
	criteria := types.NewEventCriteria(events)
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(2, counter, "should iterate only for specified event types with matching attributes")
}

func (suite *KeeperTestSuite) TestIterateABCIEventsByEventTypeAndAttributeNameAndValue() {
	suite.SetupTest()
	setupEventHistory(suite)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type: "event1",
			Attributes: map[string][]byte{
				"key1": []byte("value1"),
			},
		},
	}
	criteria := types.NewEventCriteria(events)
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(2, counter, "should iterate only for specified event types with matching attributes")
}

func (suite *KeeperTestSuite) TestIterateABCIEventsNonExistantEventType() {
	suite.SetupTest()
	setupEventHistory(suite)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type:       "event5",
			Attributes: map[string][]byte{},
		},
	}
	criteria := types.NewEventCriteria(events)
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(0, counter, "should not iterate if event doesn't exist")
}

func (suite *KeeperTestSuite) TestIterateABCIEventsNonExistantAttributeName() {
	suite.SetupTest()
	setupEventHistory(suite)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type: "event1",
			Attributes: map[string][]byte{
				"blah": []byte("value5"),
			},
		},
	}
	criteria := types.NewEventCriteria(events)
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(0, counter, "should not iterate if attribute doesn't match")
}

func (suite *KeeperTestSuite) TestIterateABCIEventsNonAttributeValueMatch() {
	suite.SetupTest()
	setupEventHistory(suite)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type: "event1",
			Attributes: map[string][]byte{
				"key1": []byte("value5"),
			},
		},
	}
	criteria := types.NewEventCriteria(events)
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(0, counter, "should not iterate if attribute doesn't match")
}
