package v2_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"

	provenanceapp "github.com/provenance-io/provenance/app"
	v2 "github.com/provenance-io/provenance/x/ibcratelimit/module/v2"
	"github.com/provenance-io/provenance/x/ibcratelimit/keeper"
)

// --- helpers ---

// makePayload creates a v2 Payload for the given FungibleTokenPacketData using the given encoding.
func makePayload(t *testing.T, data transfertypes.FungibleTokenPacketData, encoding string) channeltypesv2.Payload {
	t.Helper()
	bz, err := transfertypes.MarshalPacketData(data, transfertypes.V1, encoding)
	require.NoError(t, err, "MarshalPacketData")
	return channeltypesv2.Payload{
		SourcePort:      "transfer",
		DestinationPort: "transfer",
		Version:         transfertypes.V1,
		Encoding:        encoding,
		Value:           bz,
	}
}

// sampleData returns a simple FungibleTokenPacketData.
func sampleData() transfertypes.FungibleTokenPacketData {
	return transfertypes.FungibleTokenPacketData{
		Denom:    "stake",
		Amount:   "1000",
		Sender:   "cosmos1sender",
		Receiver: "cosmos1receiver",
		Memo:     "test memo",
	}
}

// mockApp is a minimal mock for api.IBCModule used to verify delegation.
type mockApp struct {
	onSendPacketCalled    bool
	onSendPacketErr       error
	onRecvPacketCalled    bool
	onRecvPacketResult    channeltypesv2.RecvPacketResult
	onAckPacketCalled     bool
	onAckPacketErr        error
	onTimeoutPacketCalled bool
	onTimeoutPacketErr    error
}

func (m *mockApp) OnSendPacket(sdk.Context, string, string, uint64, channeltypesv2.Payload, sdk.AccAddress) error {
	m.onSendPacketCalled = true
	return m.onSendPacketErr
}

func (m *mockApp) OnRecvPacket(sdk.Context, string, string, uint64, channeltypesv2.Payload, sdk.AccAddress) channeltypesv2.RecvPacketResult {
	m.onRecvPacketCalled = true
	return m.onRecvPacketResult
}

func (m *mockApp) OnAcknowledgementPacket(sdk.Context, string, string, uint64, []byte, channeltypesv2.Payload, sdk.AccAddress) error {
	m.onAckPacketCalled = true
	return m.onAckPacketErr
}

func (m *mockApp) OnTimeoutPacket(sdk.Context, string, string, uint64, channeltypesv2.Payload, sdk.AccAddress) error {
	m.onTimeoutPacketCalled = true
	return m.onTimeoutPacketErr
}

func newMockApp() *mockApp {
	return &mockApp{
		onRecvPacketResult: channeltypesv2.RecvPacketResult{
			Status:          channeltypesv2.PacketStatus_Success,
			Acknowledgement: []byte(`{"result":"ok"}`),
		},
	}
}

// newTestKeeperAndCtx creates a Keeper with no contract configured and a context with a working KV store.
func newTestKeeperAndCtx(t *testing.T) (*keeper.Keeper, sdk.Context) {
	t.Helper()
	provApp := provenanceapp.Setup(t)
	ctx := provApp.BaseApp.NewContext(false).WithEventManager(sdk.NewEventManager())
	return provApp.RateLimitingKeeper, ctx
}

// --- OnSendPacket tests ---

func TestOnSendPacket_NoContractConfigured(t *testing.T) {
	k, ctx := newTestKeeperAndCtx(t)
	app := newMockApp()
	mw := v2.NewIBCMiddleware(k, app)

	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnSendPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.NoError(t, err)
	assert.True(t, app.onSendPacketCalled, "should delegate to wrapped app")
}

func TestOnSendPacket_AppError(t *testing.T) {
	k, ctx := newTestKeeperAndCtx(t)
	app := newMockApp()
	app.onSendPacketErr = errors.New("app send error")
	mw := v2.NewIBCMiddleware(k, app)

	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnSendPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.EqualError(t, err, "app send error")
}

// --- OnRecvPacket tests ---

func TestOnRecvPacket_NoContractConfigured(t *testing.T) {
	k, ctx := newTestKeeperAndCtx(t)
	app := newMockApp()
	mw := v2.NewIBCMiddleware(k, app)

	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	result := mw.OnRecvPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.Equal(t, channeltypesv2.PacketStatus_Success, result.Status)
	assert.True(t, app.onRecvPacketCalled, "should delegate to wrapped app")
}

func TestOnRecvPacket_InvalidPayload_NoContract(t *testing.T) {
	k, ctx := newTestKeeperAndCtx(t)
	app := newMockApp()
	mw := v2.NewIBCMiddleware(k, app)

	// Without a contract configured, even invalid payloads are passed through.
	badPayload := channeltypesv2.Payload{
		SourcePort:      "transfer",
		DestinationPort: "transfer",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingJSON,
		Value:           []byte("garbage"),
	}
	result := mw.OnRecvPacket(ctx, "src-client", "dst-client", 1, badPayload, nil)
	assert.Equal(t, channeltypesv2.PacketStatus_Success, result.Status,
		"no contract configured, so garbage payload is just passed through")
	assert.True(t, app.onRecvPacketCalled)
}

// --- OnAcknowledgementPacket tests ---

func TestOnAcknowledgementPacket_SuccessAck(t *testing.T) {
	k, ctx := newTestKeeperAndCtx(t)
	app := newMockApp()
	mw := v2.NewIBCMiddleware(k, app)

	successAck := channeltypes.NewResultAcknowledgement([]byte(`{"result":"ok"}`)).Acknowledgement()
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnAcknowledgementPacket(ctx, "src-client", "dst-client", 1, successAck, payload, nil)
	assert.NoError(t, err)
	assert.True(t, app.onAckPacketCalled)
}

func TestOnAcknowledgementPacket_ErrorAck_NoContract(t *testing.T) {
	k, ctx := newTestKeeperAndCtx(t)
	app := newMockApp()
	mw := v2.NewIBCMiddleware(k, app)

	// Error ack with no contract configured: RevertSentPacket is a no-op.
	errorAck := channeltypes.NewErrorAcknowledgement(errors.New("transfer failed")).Acknowledgement()
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnAcknowledgementPacket(ctx, "src-client", "dst-client", 1, errorAck, payload, nil)
	assert.NoError(t, err, "error ack with no contract should not error")
	assert.True(t, app.onAckPacketCalled)
}

func TestOnAcknowledgementPacket_ErrorAck_BadPayload(t *testing.T) {
	k, ctx := newTestKeeperAndCtx(t)
	app := newMockApp()
	mw := v2.NewIBCMiddleware(k, app)

	errorAck := channeltypes.NewErrorAcknowledgement(errors.New("transfer failed")).Acknowledgement()
	badPayload := channeltypesv2.Payload{
		SourcePort:      "transfer",
		DestinationPort: "transfer",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingJSON,
		Value:           []byte("bad data"),
	}
	err := mw.OnAcknowledgementPacket(ctx, "src-client", "dst-client", 1, errorAck, badPayload, nil)
	assert.Error(t, err, "should error on bad payload with error ack")
	assert.False(t, app.onAckPacketCalled, "should not delegate when conversion fails")
}

func TestOnAcknowledgementPacket_AppError(t *testing.T) {
	k, ctx := newTestKeeperAndCtx(t)
	app := newMockApp()
	app.onAckPacketErr = errors.New("app ack error")
	mw := v2.NewIBCMiddleware(k, app)

	successAck := channeltypes.NewResultAcknowledgement([]byte(`{"result":"ok"}`)).Acknowledgement()
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnAcknowledgementPacket(ctx, "src-client", "dst-client", 1, successAck, payload, nil)
	assert.EqualError(t, err, "app ack error")
}

// --- OnTimeoutPacket tests ---

func TestOnTimeoutPacket_NoContractConfigured(t *testing.T) {
	k, ctx := newTestKeeperAndCtx(t)
	app := newMockApp()
	mw := v2.NewIBCMiddleware(k, app)

	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnTimeoutPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.NoError(t, err, "no contract configured, revert is a no-op")
	assert.True(t, app.onTimeoutPacketCalled)
}

func TestOnTimeoutPacket_BadPayload(t *testing.T) {
	k, ctx := newTestKeeperAndCtx(t)
	app := newMockApp()
	mw := v2.NewIBCMiddleware(k, app)

	badPayload := channeltypesv2.Payload{
		SourcePort:      "transfer",
		DestinationPort: "transfer",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingJSON,
		Value:           []byte("bad data"),
	}
	err := mw.OnTimeoutPacket(ctx, "src-client", "dst-client", 1, badPayload, nil)
	assert.Error(t, err, "should error on bad payload")
	assert.False(t, app.onTimeoutPacketCalled, "should not delegate when conversion fails")
}

func TestOnTimeoutPacket_AppError(t *testing.T) {
	k, ctx := newTestKeeperAndCtx(t)
	app := newMockApp()
	app.onTimeoutPacketErr = errors.New("app timeout error")
	mw := v2.NewIBCMiddleware(k, app)

	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnTimeoutPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.EqualError(t, err, "app timeout error")
	assert.True(t, app.onTimeoutPacketCalled)
}

// --- Constructor test ---

func TestNewIBCMiddleware(t *testing.T) {
	k, ctx := newTestKeeperAndCtx(t)
	app := newMockApp()
	// Just verify it doesn't panic and returns a usable middleware.
	mw := v2.NewIBCMiddleware(k, app)

	// Verify the middleware works by calling a method.
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnSendPacket(ctx, "src", "dst", 1, payload, nil)
	assert.NoError(t, err)
	assert.True(t, app.onSendPacketCalled, "middleware should delegate to wrapped app")
}
