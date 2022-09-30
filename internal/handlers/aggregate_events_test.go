package handlers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	piohandlers "github.com/provenance-io/provenance/internal/handlers"
)

func TestAggregateEvents(tt *testing.T) {

	originalFeeEvent := NewEvent(sdk.EventTypeTx,
		NewAttribute(sdk.AttributeKeyFee, "100111stake"),
		NewAttribute(sdk.AttributeKeyFeePayer, "payer"))

	anteEvents := []abci.Event{
		originalFeeEvent,
	}

	actualAnte, actualResultEvents := piohandlers.AggregateEvents(anteEvents, nil)
	assert.Nil(tt, actualResultEvents, "should not have any resultEvents since this is a failed tx case")
	assert.Equal(tt, anteEvents, actualAnte, "should return original anteevents since first spent_event not found")

	minFeeEvent := []abci.Event{NewEvent(sdk.EventTypeTx,
		NewAttribute(antewrapper.AttributeKeyMinFeeCharged, "100000stake"),
		NewAttribute(sdk.AttributeKeyFeePayer, "payer")),
	}

	expectedFeeEvent := []abci.Event{NewEvent(sdk.EventTypeTx,
		NewAttribute(sdk.AttributeKeyFee, "100000stake"),
		NewAttribute(sdk.AttributeKeyFeePayer, "payer")),
	}

	expectedAnteEvents := append(expectedFeeEvent, minFeeEvent...)
	anteEvents = append(anteEvents, minFeeEvent...)

	actualAnte, actualResultEvents = piohandlers.AggregateEvents(anteEvents, nil)
	assert.Nil(tt, actualResultEvents, "should not have any resultEvents since this is a failed tx case")
	assert.Equal(tt, expectedAnteEvents, actualAnte, "should return new ante events with fee amount from first send coins")

	// test when the result events are sent in...this should be a no-op for now
	actualAnte, actualResultEvents = piohandlers.AggregateEvents(anteEvents, anteEvents)
	assert.NotNil(tt, actualResultEvents, "should have resultEvents since this is a successful tx case")
	assert.Equal(tt, anteEvents, actualAnte, "should return original anteevents since successful tx is a noop")
	assert.Equal(tt, anteEvents, actualResultEvents, "should return original anteevents since successful tx is a noop")
}
