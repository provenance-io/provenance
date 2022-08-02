package service

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/provenance-io/provenance/internal/streaming"
	abci "github.com/tendermint/tendermint/abci/types"
)

var _ streaming.StreamService = (*TraceStreamingService)(nil)

// TraceStreamingService is a lightweight streaming service for showing how streaming is plugged in.
type TraceStreamingService struct {
	printDataToStdout bool              // Print request response data to stdout.
	codec             codec.BinaryCodec // binary marshaller used for re-marshaling the ABCI messages to write them out to the destination files
}

func NewTraceStreamingService(
	printDataToStdout bool,
	marshaller codec.BinaryCodec,
) *TraceStreamingService {
	tss := &TraceStreamingService{
		printDataToStdout: printDataToStdout,
		codec:             marshaller,
	}

	return tss
}

// StreamBeginBlocker writes out the received BeginBlockEvent request and response
func (tss *TraceStreamingService) StreamBeginBlocker(
	ctx sdk.Context,
	req abci.RequestBeginBlock,
	res abci.ResponseBeginBlock,
) {
	errMsg := "BeginBlocker listening hook failed"

	// write req
	if err := tss.write(ctx, &req); err != nil {
		ctx.Logger().Error(errMsg, "action", "write request", "height", req.Header.Height, "err", err)
		panic(err)
	}

	// write res
	if err := tss.write(ctx, &res); err != nil {
		ctx.Logger().Error(errMsg, "action", "write response", "height", ctx.BlockHeight(), "err", err)
		panic(err)
	}
}

// StreamEndBlocker writes out the received app.EndBlocker request and response
func (tss *TraceStreamingService) StreamEndBlocker(
	ctx sdk.Context,
	req abci.RequestEndBlock,
	res abci.ResponseEndBlock,
) {
	errMsg := "EndBlocker listening hook failed"

	// write req
	if err := tss.write(ctx, &req); err != nil {
		ctx.Logger().Error(errMsg, "action", "write request", "height", ctx.BlockHeight(), "err", err)
		panic(err)
	}

	// write res
	if err := tss.write(ctx, &res); err != nil {
		ctx.Logger().Error(errMsg, "action", "write response", "height", ctx.BlockHeight(), "err", err)
		panic(err)
	}
}

// nolint: unparam
func (tss *TraceStreamingService) write(
	ctx sdk.Context,
	data proto.Message,
) error {
	var m = "omitted"
	if tss.printDataToStdout {
		m = fmt.Sprintf("%v", data)
	}
	ctx.Logger().Debug("streaming ABCI data to external service", "service", "trace", "data", m)

	return nil
}
