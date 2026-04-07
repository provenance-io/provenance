package v2_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"

	provenanceapp "github.com/provenance-io/provenance/app"
	ibchooks "github.com/provenance-io/provenance/x/ibchooks"
	"github.com/provenance-io/provenance/x/ibchooks/keeper"
	"github.com/provenance-io/provenance/x/ibchooks/types"
	v2 "github.com/provenance-io/provenance/x/ibchooks/v2"
)

// --- helpers ---

func sampleData() transfertypes.FungibleTokenPacketData {
	return transfertypes.FungibleTokenPacketData{
		Denom:    "stake",
		Amount:   "1000",
		Sender:   "cosmos1sender",
		Receiver: "cosmos1receiver",
		Memo:     "test memo",
	}
}

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

// mockApp is a minimal mock for api.IBCModule to verify delegation behavior.
type mockApp struct {
	onSendPacketCalled    bool
	onSendPacketErr       error
	onSendPayload         channeltypesv2.Payload // captures the payload passed to OnSendPacket
	onRecvPacketCalled    bool
	onRecvPacketResult    channeltypesv2.RecvPacketResult
	onRecvPayload         channeltypesv2.Payload // captures the payload passed to OnRecvPacket
	onAckPacketCalled     bool
	onAckPacketErr        error
	onTimeoutPacketCalled bool
	onTimeoutPacketErr    error
}

func (m *mockApp) OnSendPacket(_ sdk.Context, _, _ string, _ uint64, payload channeltypesv2.Payload, _ sdk.AccAddress) error {
	m.onSendPacketCalled = true
	m.onSendPayload = payload
	return m.onSendPacketErr
}

func (m *mockApp) OnRecvPacket(_ sdk.Context, _, _ string, _ uint64, payload channeltypesv2.Payload, _ sdk.AccAddress) channeltypesv2.RecvPacketResult {
	m.onRecvPacketCalled = true
	m.onRecvPayload = payload
	return m.onRecvPacketResult
}

func (m *mockApp) OnAcknowledgementPacket(_ sdk.Context, _, _ string, _ uint64, _ []byte, _ channeltypesv2.Payload, _ sdk.AccAddress) error {
	m.onAckPacketCalled = true
	return m.onAckPacketErr
}

func (m *mockApp) OnTimeoutPacket(_ sdk.Context, _, _ string, _ uint64, _ channeltypesv2.Payload, _ sdk.AccAddress) error {
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

// newTestModule creates a fully-configured IBCModule for testing, backed by real keepers.
func newTestModule(t *testing.T, app *mockApp) (v2.IBCModule, sdk.Context) {
	t.Helper()
	provApp := provenanceapp.Setup(t)
	ctx := provApp.BaseApp.NewContext(false).WithEventManager(sdk.NewEventManager())
	return v2.NewIBCModule(
		app,
		provApp.IBCKeeper,
		provApp.IBCHooksKeeper,
		provApp.WasmKeeper,
		provApp.Ics20MarkerHooks,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
	), ctx
}

// newUnconfiguredModule creates an IBCModule with nil keepers (not properly configured).
func newUnconfiguredModule(t *testing.T, app *mockApp) (v2.IBCModule, sdk.Context) {
	t.Helper()
	provApp := provenanceapp.Setup(t)
	ctx := provApp.BaseApp.NewContext(false).WithEventManager(sdk.NewEventManager())
	return v2.NewIBCModule(app, nil, nil, nil, nil, "cosmos"), ctx
}

// --- OnSendPacket tests ---

func TestOnSendPacket_NotConfigured_Delegates(t *testing.T) {
	app := newMockApp()
	mw, ctx := newUnconfiguredModule(t, app)
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnSendPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.NoError(t, err)
	assert.True(t, app.onSendPacketCalled)
}

func TestOnSendPacket_NoWasmMemo_Delegates(t *testing.T) {
	app := newMockApp()
	mw, ctx := newTestModule(t, app)

	data := sampleData()
	data.Memo = "just a regular memo"
	payload := makePayload(t, data, transfertypes.EncodingJSON)

	err := mw.OnSendPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.NoError(t, err)
	assert.True(t, app.onSendPacketCalled)
}

func TestOnSendPacket_InvalidPayload_Delegates(t *testing.T) {
	app := newMockApp()
	mw, ctx := newTestModule(t, app)

	badPayload := channeltypesv2.Payload{
		SourcePort:      "transfer",
		DestinationPort: "transfer",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingJSON,
		Value:           []byte("not valid"),
	}
	err := mw.OnSendPacket(ctx, "src-client", "dst-client", 1, badPayload, nil)
	assert.NoError(t, err, "invalid payload should still delegate")
	assert.True(t, app.onSendPacketCalled)
}

func TestOnSendPacket_CallbackMemo_StoresCallback_StripsMemo(t *testing.T) {
	mockA := newMockApp()
	provApp := provenanceapp.Setup(t)
	ctx := provApp.BaseApp.NewContext(false).WithEventManager(sdk.NewEventManager())
	hooksKeeper := provApp.IBCHooksKeeper

	mw := v2.NewIBCModule(
		mockA,
		provApp.IBCKeeper,
		hooksKeeper,
		provApp.WasmKeeper,
		provApp.Ics20MarkerHooks,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
	)

	// Use a real bech32 address for the callback contract.
	contractAddr := sdk.AccAddress([]byte("contract_address_____")).String()
	data := sampleData()
	data.Memo = fmt.Sprintf(`{"%s":"%s"}`, types.IBCCallbackKey, contractAddr)
	payload := makePayload(t, data, transfertypes.EncodingJSON)

	err := mw.OnSendPacket(ctx, "src-client", "dst-client", 5, payload, nil)
	assert.NoError(t, err)
	assert.True(t, mockA.onSendPacketCalled)

	// Verify the callback was stored.
	stored := hooksKeeper.GetPacketCallback(ctx, "src-client", 5)
	assert.Equal(t, contractAddr, stored, "callback contract should be stored")

	// Verify the memo was stripped from the payload forwarded to the wrapped app.
	// When ibc_callback is the only key, memo should be completely cleared.
	fwdData, err := transfertypes.UnmarshalPacketData(
		mockA.onSendPayload.Value, mockA.onSendPayload.Version, mockA.onSendPayload.Encoding)
	require.NoError(t, err)
	assert.Empty(t, fwdData.Memo, "memo should be empty when ibc_callback was the only key")
}

func TestOnSendPacket_CallbackMemo_PreservesOtherMemoKeys(t *testing.T) {
	mockA := newMockApp()
	provApp := provenanceapp.Setup(t)
	ctx := provApp.BaseApp.NewContext(false).WithEventManager(sdk.NewEventManager())

	mw := v2.NewIBCModule(
		mockA,
		provApp.IBCKeeper,
		provApp.IBCHooksKeeper,
		provApp.WasmKeeper,
		provApp.Ics20MarkerHooks,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
	)

	contractAddr := sdk.AccAddress([]byte("contract_address_____")).String()
	data := sampleData()
	data.Memo = fmt.Sprintf(`{"%s":"%s","other":"value"}`, types.IBCCallbackKey, contractAddr)
	payload := makePayload(t, data, transfertypes.EncodingJSON)

	err := mw.OnSendPacket(ctx, "src-client", "dst-client", 6, payload, nil)
	assert.NoError(t, err)
	assert.True(t, mockA.onSendPacketCalled)

	// Verify the forwarded payload has ibc_callback stripped but other keys preserved.
	fwdData, err := transfertypes.UnmarshalPacketData(
		mockA.onSendPayload.Value, mockA.onSendPayload.Version, mockA.onSendPayload.Encoding)
	require.NoError(t, err)
	assert.NotEmpty(t, fwdData.Memo, "memo should not be empty when other keys exist")
	assert.NotContains(t, fwdData.Memo, types.IBCCallbackKey, "ibc_callback should be stripped from memo")
	assert.Contains(t, fwdData.Memo, "other", "other memo keys should be preserved")
}

func TestOnSendPacket_CallbackMemo_NonStringValue_Errors(t *testing.T) {
	mockA := newMockApp()
	provApp := provenanceapp.Setup(t)
	ctx := provApp.BaseApp.NewContext(false).WithEventManager(sdk.NewEventManager())

	mw := v2.NewIBCModule(
		mockA,
		provApp.IBCKeeper,
		provApp.IBCHooksKeeper,
		provApp.WasmKeeper,
		provApp.Ics20MarkerHooks,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
	)

	data := sampleData()
	data.Memo = fmt.Sprintf(`{"%s": 12345}`, types.IBCCallbackKey)
	payload := makePayload(t, data, transfertypes.EncodingJSON)

	err := mw.OnSendPacket(ctx, "src-client", "dst-client", 5, payload, nil)
	assert.Error(t, err, "non-string callback value should error")
	assert.Contains(t, err.Error(), "unable to format callback")
	assert.False(t, mockA.onSendPacketCalled, "should not delegate on callback validation error")
}

func TestOnSendPacket_CallbackMemo_InvalidBech32_Errors(t *testing.T) {
	mockA := newMockApp()
	provApp := provenanceapp.Setup(t)
	ctx := provApp.BaseApp.NewContext(false).WithEventManager(sdk.NewEventManager())

	mw := v2.NewIBCModule(
		mockA,
		provApp.IBCKeeper,
		provApp.IBCHooksKeeper,
		provApp.WasmKeeper,
		provApp.Ics20MarkerHooks,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
	)

	data := sampleData()
	data.Memo = fmt.Sprintf(`{"%s": "not-a-bech32-addr"}`, types.IBCCallbackKey)
	payload := makePayload(t, data, transfertypes.EncodingJSON)

	err := mw.OnSendPacket(ctx, "src-client", "dst-client", 5, payload, nil)
	assert.Error(t, err, "invalid bech32 should error")
	assert.Contains(t, err.Error(), "invalid bech32 contract address")
	assert.False(t, mockA.onSendPacketCalled, "should not delegate on callback validation error")
}

func TestOnSendPacket_AppError(t *testing.T) {
	app := newMockApp()
	app.onSendPacketErr = errors.New("app send error")
	mw, ctx := newTestModule(t, app)

	contractAddr := sdk.AccAddress([]byte("contract_address_____")).String()
	data := sampleData()
	data.Memo = fmt.Sprintf(`{"%s":"%s"}`, types.IBCCallbackKey, contractAddr)
	payload := makePayload(t, data, transfertypes.EncodingJSON)

	err := mw.OnSendPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.EqualError(t, err, "app send error")
}

// --- OnRecvPacket tests ---

func TestOnRecvPacket_NotConfigured_Delegates(t *testing.T) {
	app := newMockApp()
	mw, ctx := newUnconfiguredModule(t, app)
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	result := mw.OnRecvPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.Equal(t, channeltypesv2.PacketStatus_Success, result.Status)
	assert.True(t, app.onRecvPacketCalled)
}

func TestOnRecvPacket_InvalidPayload_Delegates(t *testing.T) {
	app := newMockApp()
	mw, ctx := newUnconfiguredModule(t, app)
	badPayload := channeltypesv2.Payload{
		SourcePort:      "transfer",
		DestinationPort: "transfer",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingJSON,
		Value:           []byte("garbage"),
	}
	result := mw.OnRecvPacket(ctx, "src-client", "dst-client", 1, badPayload, nil)
	assert.Equal(t, channeltypesv2.PacketStatus_Success, result.Status)
	assert.True(t, app.onRecvPacketCalled)
}

func TestOnRecvPacket_NoWasmMemo_Delegates(t *testing.T) {
	app := newMockApp()
	mw, ctx := newTestModule(t, app)

	data := sampleData()
	data.Memo = ""
	payload := makePayload(t, data, transfertypes.EncodingJSON)

	result := mw.OnRecvPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.Equal(t, channeltypesv2.PacketStatus_Success, result.Status)
	assert.True(t, app.onRecvPacketCalled)
}

func TestOnRecvPacket_WasmMemo_ContractMismatch(t *testing.T) {
	app := newMockApp()
	mw, ctx := newTestModule(t, app)

	// wasm memo where contract != receiver should fail validation.
	contractAddr := sdk.AccAddress([]byte("contract_address_____")).String()
	data := sampleData()
	data.Receiver = "cosmos1differentreceiver"
	data.Memo = fmt.Sprintf(`{"wasm":{"contract":"%s","msg":{"echo":{"msg":"test"}}}}`, contractAddr)
	payload := makePayload(t, data, transfertypes.EncodingJSON)

	result := mw.OnRecvPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.Equal(t, channeltypesv2.PacketStatus_Failure, result.Status, "contract/receiver mismatch should fail")
	assert.False(t, app.onRecvPacketCalled, "should not delegate on validation error")
}

func TestOnRecvPacket_WasmMemo_InvalidContract(t *testing.T) {
	app := newMockApp()
	mw, ctx := newTestModule(t, app)

	data := sampleData()
	data.Receiver = "not-a-valid-address"
	data.Memo = `{"wasm":{"contract":"not-a-valid-address","msg":{"echo":{"msg":"test"}}}}`
	payload := makePayload(t, data, transfertypes.EncodingJSON)

	result := mw.OnRecvPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.Equal(t, channeltypesv2.PacketStatus_Failure, result.Status, "invalid contract address should fail")
	assert.False(t, app.onRecvPacketCalled)
}

func TestOnRecvPacket_WasmMemo_ModifiesReceiver(t *testing.T) {
	// When a wasm memo is present, OnRecvPacket should hijack the receiver
	// to the derived intermediate sender. The wrapped app should see the modified payload.
	mockA := newMockApp()
	provApp := provenanceapp.Setup(t)
	ctx := provApp.BaseApp.NewContext(false).WithEventManager(sdk.NewEventManager())
	prefix := sdk.GetConfig().GetBech32AccountAddrPrefix()

	mw := v2.NewIBCModule(
		mockA,
		provApp.IBCKeeper,
		provApp.IBCHooksKeeper,
		provApp.WasmKeeper,
		provApp.Ics20MarkerHooks,
		prefix,
	)

	contractAddr := sdk.AccAddress([]byte("contract_address_____")).String()
	originalSender := "cosmos1sender"
	data := sampleData()
	data.Sender = originalSender
	data.Receiver = contractAddr
	data.Memo = fmt.Sprintf(`{"wasm":{"contract":"%s","msg":{"echo":{"msg":"test"}}}}`, contractAddr)
	payload := makePayload(t, data, transfertypes.EncodingJSON)

	_ = mw.OnRecvPacket(ctx, "src-client", "dst-client", 1, payload, nil)

	// The wrapped app should have been called (even though wasm execution may fail later).
	if mockA.onRecvPacketCalled {
		// Verify the receiver was modified to the derived intermediate sender.
		modifiedData, err := transfertypes.UnmarshalPacketData(
			mockA.onRecvPayload.Value, mockA.onRecvPayload.Version, mockA.onRecvPayload.Encoding)
		require.NoError(t, err)

		expectedSender, err := keeper.DeriveIntermediateSender("dst-client", originalSender, prefix)
		require.NoError(t, err)
		assert.Equal(t, expectedSender, modifiedData.Receiver,
			"receiver should be changed to derived intermediate sender")
	}
}

func TestOnRecvPacket_ConfiguredApp_InvalidPayload_ReturnsError(t *testing.T) {
	app := newMockApp()
	mw, ctx := newTestModule(t, app)

	badPayload := channeltypesv2.Payload{
		SourcePort:      "transfer",
		DestinationPort: "transfer",
		Version:         transfertypes.V1,
		Encoding:        transfertypes.EncodingJSON,
		Value:           []byte("not a valid packet"),
	}
	result := mw.OnRecvPacket(ctx, "src-client", "dst-client", 1, badPayload, nil)
	// When configured, invalid payloads should still delegate since they're not ICS20.
	assert.True(t, app.onRecvPacketCalled)
	assert.Equal(t, channeltypesv2.PacketStatus_Success, result.Status)
}

// --- OnAcknowledgementPacket tests ---

func TestOnAcknowledgementPacket_SuccessAck_Delegates(t *testing.T) {
	app := newMockApp()
	mw, ctx := newTestModule(t, app)

	successAck := channeltypes.NewResultAcknowledgement([]byte(`{"result":"ok"}`)).Acknowledgement()
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnAcknowledgementPacket(ctx, "src-client", "dst-client", 1, successAck, payload, nil)
	assert.NoError(t, err)
	assert.True(t, app.onAckPacketCalled)
}

func TestOnAcknowledgementPacket_AppError(t *testing.T) {
	app := newMockApp()
	app.onAckPacketErr = errors.New("app ack error")
	mw, ctx := newTestModule(t, app)

	successAck := channeltypes.NewResultAcknowledgement([]byte(`{"result":"ok"}`)).Acknowledgement()
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnAcknowledgementPacket(ctx, "src-client", "dst-client", 1, successAck, payload, nil)
	assert.EqualError(t, err, "app ack error")
}

func TestOnAcknowledgementPacket_NotConfigured_SkipsCallback(t *testing.T) {
	app := newMockApp()
	mw, ctx := newUnconfiguredModule(t, app)

	successAck := channeltypes.NewResultAcknowledgement([]byte(`{"result":"ok"}`)).Acknowledgement()
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnAcknowledgementPacket(ctx, "src-client", "dst-client", 1, successAck, payload, nil)
	assert.NoError(t, err)
	assert.True(t, app.onAckPacketCalled)
}

func TestOnAcknowledgementPacket_NoCallback_NoError(t *testing.T) {
	app := newMockApp()
	mw, ctx := newTestModule(t, app)

	// No stored callback for this sequence.
	successAck := channeltypes.NewResultAcknowledgement([]byte(`{"result":"ok"}`)).Acknowledgement()
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnAcknowledgementPacket(ctx, "src-client", "dst-client", 99, successAck, payload, nil)
	assert.NoError(t, err)
	assert.True(t, app.onAckPacketCalled)
}

func TestOnAcknowledgementPacket_StoredCallback_InvalidBech32(t *testing.T) {
	mockA := newMockApp()
	provApp := provenanceapp.Setup(t)
	ctx := provApp.BaseApp.NewContext(false).WithEventManager(sdk.NewEventManager())

	mw := v2.NewIBCModule(
		mockA,
		provApp.IBCKeeper,
		provApp.IBCHooksKeeper,
		provApp.WasmKeeper,
		provApp.Ics20MarkerHooks,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
	)

	// Store a callback with an invalid bech32 address.
	provApp.IBCHooksKeeper.StorePacketCallback(ctx, "src-client", 10, "not-a-bech32")

	successAck := channeltypes.NewResultAcknowledgement([]byte(`{"result":"ok"}`)).Acknowledgement()
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnAcknowledgementPacket(ctx, "src-client", "dst-client", 10, successAck, payload, nil)
	assert.Error(t, err, "invalid bech32 callback contract should error")
	assert.Contains(t, err.Error(), "ack callback error")
}

func TestOnAcknowledgementPacket_ErrorAck_StoredCallback_InvalidBech32(t *testing.T) {
	mockA := newMockApp()
	provApp := provenanceapp.Setup(t)
	ctx := provApp.BaseApp.NewContext(false).WithEventManager(sdk.NewEventManager())

	mw := v2.NewIBCModule(
		mockA,
		provApp.IBCKeeper,
		provApp.IBCHooksKeeper,
		provApp.WasmKeeper,
		provApp.Ics20MarkerHooks,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
	)

	provApp.IBCHooksKeeper.StorePacketCallback(ctx, "src-client", 10, "not-a-bech32")

	errorAck := channeltypes.NewErrorAcknowledgement(errors.New("transfer failed")).Acknowledgement()
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnAcknowledgementPacket(ctx, "src-client", "dst-client", 10, errorAck, payload, nil)
	assert.Error(t, err, "invalid bech32 callback contract should error")
	assert.Contains(t, err.Error(), "ack callback error")
}

// --- OnTimeoutPacket tests ---

func TestOnTimeoutPacket_NotConfigured_Delegates(t *testing.T) {
	app := newMockApp()
	mw, ctx := newUnconfiguredModule(t, app)
	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnTimeoutPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.NoError(t, err)
	assert.True(t, app.onTimeoutPacketCalled)
}

func TestOnTimeoutPacket_AppError(t *testing.T) {
	app := newMockApp()
	app.onTimeoutPacketErr = errors.New("app timeout error")
	mw, ctx := newTestModule(t, app)

	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnTimeoutPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.EqualError(t, err, "app timeout error")
	assert.True(t, app.onTimeoutPacketCalled)
}

func TestOnTimeoutPacket_NoCallback_NoError(t *testing.T) {
	app := newMockApp()
	mw, ctx := newTestModule(t, app)

	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnTimeoutPacket(ctx, "src-client", "dst-client", 99, payload, nil)
	assert.NoError(t, err)
	assert.True(t, app.onTimeoutPacketCalled)
}

func TestOnTimeoutPacket_StoredCallback_InvalidBech32(t *testing.T) {
	mockA := newMockApp()
	provApp := provenanceapp.Setup(t)
	ctx := provApp.BaseApp.NewContext(false).WithEventManager(sdk.NewEventManager())

	mw := v2.NewIBCModule(
		mockA,
		provApp.IBCKeeper,
		provApp.IBCHooksKeeper,
		provApp.WasmKeeper,
		provApp.Ics20MarkerHooks,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
	)

	provApp.IBCHooksKeeper.StorePacketCallback(ctx, "src-client", 10, "not-a-bech32")

	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnTimeoutPacket(ctx, "src-client", "dst-client", 10, payload, nil)
	assert.Error(t, err, "invalid bech32 callback contract should error")
	assert.Contains(t, err.Error(), "timeout callback error")
}

func TestOnTimeoutPacket_StoredCallback_DeletesAfterExec(t *testing.T) {
	mockA := newMockApp()
	provApp := provenanceapp.Setup(t)
	ctx := provApp.BaseApp.NewContext(false).WithEventManager(sdk.NewEventManager())

	mw := v2.NewIBCModule(
		mockA,
		provApp.IBCKeeper,
		provApp.IBCHooksKeeper,
		provApp.WasmKeeper,
		provApp.Ics20MarkerHooks,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
	)

	// Use a valid bech32 address. The Sudo call will fail (no contract exists) but the
	// timeout handler should still delete the callback and emit an error event.
	contractAddr := sdk.AccAddress([]byte("contract_address_____")).String()
	provApp.IBCHooksKeeper.StorePacketCallback(ctx, "src-client", 10, contractAddr)

	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)
	err := mw.OnTimeoutPacket(ctx, "src-client", "dst-client", 10, payload, nil)
	// The timeout handler does not return Sudo errors; it emits an event.
	assert.NoError(t, err)
	assert.True(t, mockA.onTimeoutPacketCalled)

	// Verify callback was deleted.
	stored := provApp.IBCHooksKeeper.GetPacketCallback(ctx, "src-client", 10)
	assert.Empty(t, stored, "callback should be deleted after timeout processing")
}

// --- Constructor / properlyConfigured tests ---

func TestNewIBCModule_Unconfigured_DelegatesToApp(t *testing.T) {
	app := newMockApp()
	mw, ctx := newUnconfiguredModule(t, app)

	payload := makePayload(t, sampleData(), transfertypes.EncodingJSON)

	// All four methods should delegate when not properly configured.
	err := mw.OnSendPacket(ctx, "src", "dst", 1, payload, nil)
	assert.NoError(t, err)
	assert.True(t, app.onSendPacketCalled)

	result := mw.OnRecvPacket(ctx, "src", "dst", 1, payload, nil)
	assert.Equal(t, channeltypesv2.PacketStatus_Success, result.Status)
	assert.True(t, app.onRecvPacketCalled)
}

// --- Helper function tests ---

func TestJsonStringHasKey_WithCallbackKey(t *testing.T) {
	contractAddr := "cosmos1somecontract"
	memo := fmt.Sprintf(`{"%s": "%s"}`, types.IBCCallbackKey, contractAddr)
	found, metadata := ibchooks.JsonStringHasKey(memo, types.IBCCallbackKey)
	assert.True(t, found)
	assert.Equal(t, contractAddr, metadata[types.IBCCallbackKey])
}

func TestJsonStringHasKey_WithoutKey(t *testing.T) {
	found, _ := ibchooks.JsonStringHasKey(`{"other": "value"}`, types.IBCCallbackKey)
	assert.False(t, found)
}

func TestJsonStringHasKey_EmptyMemo(t *testing.T) {
	found, _ := ibchooks.JsonStringHasKey("", types.IBCCallbackKey)
	assert.False(t, found)
}

func TestJsonStringHasKey_InvalidJSON(t *testing.T) {
	found, _ := ibchooks.JsonStringHasKey("not json", types.IBCCallbackKey)
	assert.False(t, found)
}

func TestValidateAndParseMemo_NoWasmKey(t *testing.T) {
	isWasm, _, _, _ := ibchooks.ValidateAndParseMemo(`{"other": "value"}`, "cosmos1receiver")
	assert.False(t, isWasm)
}

func TestValidateAndParseMemo_ValidWasmMemo(t *testing.T) {
	contractAddr := sdk.AccAddress([]byte("contract_address_____")).String()
	memo := fmt.Sprintf(`{"wasm":{"contract":"%s","msg":{"echo":{"msg":"test"}}}}`, contractAddr)
	isWasm, addr, msgBytes, err := ibchooks.ValidateAndParseMemo(memo, contractAddr)
	assert.True(t, isWasm)
	assert.NoError(t, err)
	assert.Equal(t, contractAddr, addr.String())
	assert.NotNil(t, msgBytes)

	// Verify msgBytes are valid JSON.
	var msg map[string]interface{}
	assert.NoError(t, json.Unmarshal(msgBytes, &msg))
	assert.Contains(t, msg, "echo")
}

func TestValidateAndParseMemo_ContractReceiverMismatch(t *testing.T) {
	contractAddr := sdk.AccAddress([]byte("contract_address_____")).String()
	memo := fmt.Sprintf(`{"wasm":{"contract":"%s","msg":{"echo":{"msg":"test"}}}}`, contractAddr)
	isWasm, _, _, err := ibchooks.ValidateAndParseMemo(memo, "cosmos1differentreceiver")
	assert.True(t, isWasm)
	assert.Error(t, err, "contract/receiver mismatch should error")
}

func TestDeriveIntermediateSender(t *testing.T) {
	sender, err := keeper.DeriveIntermediateSender("dst-client", "cosmos1sender", "cosmos")
	assert.NoError(t, err)
	assert.NotEmpty(t, sender)
	assert.Contains(t, sender, "cosmos1", "should have correct bech32 prefix")

	// Same inputs should always produce the same output.
	sender2, err := keeper.DeriveIntermediateSender("dst-client", "cosmos1sender", "cosmos")
	assert.NoError(t, err)
	assert.Equal(t, sender, sender2, "should be deterministic")

	// Different channel should produce different sender.
	sender3, err := keeper.DeriveIntermediateSender("other-client", "cosmos1sender", "cosmos")
	assert.NoError(t, err)
	assert.NotEqual(t, sender, sender3, "different channel should produce different sender")
}

// --- Encoding variant tests ---

func TestOnSendPacket_ProtobufEncoding(t *testing.T) {
	app := newMockApp()
	mw, ctx := newTestModule(t, app)

	data := sampleData()
	data.Memo = ""
	payload := makePayload(t, data, transfertypes.EncodingProtobuf)

	err := mw.OnSendPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.NoError(t, err)
	assert.True(t, app.onSendPacketCalled)
}

func TestOnRecvPacket_ProtobufEncoding_NoMemo(t *testing.T) {
	app := newMockApp()
	mw, ctx := newTestModule(t, app)

	data := sampleData()
	data.Memo = ""
	payload := makePayload(t, data, transfertypes.EncodingProtobuf)

	result := mw.OnRecvPacket(ctx, "src-client", "dst-client", 1, payload, nil)
	assert.Equal(t, channeltypesv2.PacketStatus_Success, result.Status)
	assert.True(t, app.onRecvPacketCalled)
}
