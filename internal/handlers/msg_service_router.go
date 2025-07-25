package handlers

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/runtime/protoiface"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/protocompat"
)

// This file is basically a copy of the SDK's baseapp/msg_service_router.go file with the following modifications:
//  - Different package and re-ordered imports plus a couple of our own.
//  - We use the SDK's definition of MessageRouter, IMsgServiceRouter and MsgServiceHandler (and don't redefine them here).
//  - We've named the struct PioMsgServiceRouter (from just MsgServiceRouter).
//  - In registerMsgServiceHandler, at the start of the handler, we consume the msg.
//  - We added the consumeMsgFees method.

// PioMsgServiceRouter routes fully-qualified Msg service methods to their handler with additional fee processing of msgs.
type PioMsgServiceRouter struct {
	interfaceRegistry codectypes.InterfaceRegistry
	routes            map[string]baseapp.MsgServiceHandler
	hybridHandlers    map[string]protocompat.Handler
	circuitBreaker    baseapp.CircuitBreaker
}

var (
	_ baseapp.IMsgServiceRouter = (*PioMsgServiceRouter)(nil)
	_ gogogrpc.Server           = (*PioMsgServiceRouter)(nil)
)

// NewPioMsgServiceRouter creates a new PioMsgServiceRouter.
func NewPioMsgServiceRouter() *PioMsgServiceRouter {
	return &PioMsgServiceRouter{
		routes:         make(map[string]baseapp.MsgServiceHandler),
		hybridHandlers: make(map[string]protocompat.Handler),
	}
}

func (msr *PioMsgServiceRouter) SetCircuit(cb baseapp.CircuitBreaker) {
	msr.circuitBreaker = cb
}

// Handler returns the MsgServiceHandler for a given msg or nil if not found.
func (msr *PioMsgServiceRouter) Handler(msg sdk.Msg) baseapp.MsgServiceHandler {
	return msr.routes[sdk.MsgTypeURL(msg)]
}

// HandlerByTypeURL returns the MsgServiceHandler for a given query route path or nil
// if not found.
func (msr *PioMsgServiceRouter) HandlerByTypeURL(typeURL string) baseapp.MsgServiceHandler {
	return msr.routes[typeURL]
}

// RegisterService implements the gRPC Server.RegisterService method. sd is a gRPC
// service description, handler is an object which implements that gRPC service.
//
// This function PANICs:
//   - if it is called before the service `Msg`s have been registered using
//     RegisterInterfaces,
//   - or if a service is being registered twice.
func (msr *PioMsgServiceRouter) RegisterService(sd *grpc.ServiceDesc, handler interface{}) {
	// Adds a top-level query handler based on the gRPC service name.
	for _, method := range sd.Methods {
		err := msr.registerMsgServiceHandler(sd, method, handler)
		if err != nil {
			panic(err)
		}
		err = msr.registerHybridHandler(sd, method, handler)
		if err != nil {
			panic(err)
		}
	}
}

func (msr *PioMsgServiceRouter) HybridHandlerByMsgName(msgName string) func(ctx context.Context, req, resp protoiface.MessageV1) error {
	return msr.hybridHandlers[msgName]
}

func (msr *PioMsgServiceRouter) registerHybridHandler(sd *grpc.ServiceDesc, method grpc.MethodDesc, handler interface{}) error {
	inputName, err := protocompat.RequestFullNameFromMethodDesc(sd, method)
	if err != nil {
		return err
	}
	cdc := codec.NewProtoCodec(msr.interfaceRegistry)
	hybridHandler, err := protocompat.MakeHybridHandler(cdc, sd, method, handler)
	if err != nil {
		return err
	}
	// if circuit breaker is not nil, then we decorate the hybrid handler with the circuit breaker
	if msr.circuitBreaker == nil {
		msr.hybridHandlers[string(inputName)] = hybridHandler
		return nil
	}
	// decorate the hybrid handler with the circuit breaker
	circuitBreakerHybridHandler := func(ctx context.Context, req, resp protoiface.MessageV1) error {
		messageName := codectypes.MsgTypeURL(req)
		allowed, err := msr.circuitBreaker.IsAllowed(ctx, messageName)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("circuit breaker disallows execution of message %s", messageName)
		}
		return hybridHandler(ctx, req, resp)
	}
	msr.hybridHandlers[string(inputName)] = circuitBreakerHybridHandler
	return nil
}

func (msr *PioMsgServiceRouter) registerMsgServiceHandler(sd *grpc.ServiceDesc, method grpc.MethodDesc, handler interface{}) error {
	fqMethod := fmt.Sprintf("/%s/%s", sd.ServiceName, method.MethodName)
	methodHandler := method.Handler

	var requestTypeName string

	// NOTE: This is how we pull the concrete request type for each handler for registering in the InterfaceRegistry.
	// This approach is maybe a bit hacky, but less hacky than reflecting on the handler object itself.
	// We use a no-op interceptor to avoid actually calling into the handler itself.
	_, _ = methodHandler(nil, context.Background(), func(i interface{}) error {
		msg, ok := i.(sdk.Msg)
		if !ok {
			// We panic here because there is no other alternative and the app cannot be initialized correctly
			// this should only happen if there is a problem with code generation in which case the app won't
			// work correctly anyway.
			panic(fmt.Errorf("unable to register service method %s: %T does not implement sdk.Msg", fqMethod, i))
		}

		requestTypeName = sdk.MsgTypeURL(msg)
		return nil
	}, noopInterceptor)

	// Check that the service Msg fully-qualified method name has already
	// been registered (via RegisterInterfaces). If the user registers a
	// service without registering according service Msg type, there might be
	// some unexpected behavior down the road. Since we can't return an error
	// (`Server.RegisterService` interface restriction) we panic (at startup).
	reqType, err := msr.interfaceRegistry.Resolve(requestTypeName)
	if err != nil || reqType == nil {
		return fmt.Errorf(
			"type_url %s has not been registered yet. "+
				"Before calling RegisterService, you must register all interfaces by calling the `RegisterInterfaces` "+
				"method on module.BasicManager. Each module should call `msgservice.RegisterMsgServiceDesc` inside its "+
				"`RegisterInterfaces` method with the `_Msg_serviceDesc` generated by proto-gen",
			requestTypeName,
		)
	}

	// Check that each service is only registered once. If a service is
	// registered more than once, then we should error. Since we can't
	// return an error (`Server.RegisterService` interface restriction) we
	// panic (at startup).
	_, found := msr.routes[requestTypeName]
	if found {
		return fmt.Errorf(
			"msg service %s has already been registered. Please make sure to only register each service once. "+
				"This usually means that there are conflicting modules registering the same msg service",
			fqMethod,
		)
	}

	msr.routes[requestTypeName] = func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		// Provenance specific modification to msg service router that handles x/flatfees.
		msr.consumeMsgFees(ctx, msg)
		// End of Provenanced specific modification.

		ctx = ctx.WithEventManager(sdk.NewEventManager())
		interceptor := func(goCtx context.Context, _ interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			goCtx = context.WithValue(goCtx, sdk.SdkContextKey, ctx)
			return handler(goCtx, msg)
		}

		if m, ok := msg.(sdk.HasValidateBasic); ok {
			if err := m.ValidateBasic(); err != nil {
				return nil, err
			}
		}

		if msr.circuitBreaker != nil {
			msgURL := sdk.MsgTypeURL(msg)

			isAllowed, err := msr.circuitBreaker.IsAllowed(ctx, msgURL)
			if err != nil {
				return nil, err
			}

			if !isAllowed {
				return nil, fmt.Errorf("circuit breaker disables execution of this message: %s", msgURL)
			}
		}

		// Call the method handler from the service description with the handler object.
		// We don't do any decoding here because the decoding was already done.
		res, err := methodHandler(handler, ctx, noopDecoder, interceptor)
		if err != nil {
			return nil, err
		}

		resMsg, ok := res.(proto.Message)
		if !ok {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidType, "Expecting proto.Message, got %T", res)
		}

		return sdk.WrapServiceResult(ctx, resMsg, err)
	}

	return nil
}

// SetInterfaceRegistry sets the interface registry for the router.
func (msr *PioMsgServiceRouter) SetInterfaceRegistry(interfaceRegistry codectypes.InterfaceRegistry) {
	msr.interfaceRegistry = interfaceRegistry
}

func noopDecoder(_ interface{}) error { return nil }
func noopInterceptor(_ context.Context, _ interface{}, _ *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (interface{}, error) {
	return nil, nil
}

// consumeMsgFees consumes any flat fees for the provided msg.
func (msr *PioMsgServiceRouter) consumeMsgFees(ctx sdk.Context, msg sdk.Msg) {
	// The x/gov module calls the message service router for proposal messages that have passed.
	// In such cases, the antehandler is not run, so the gas meter will not be a fee gas meter.
	// But those messages were voted on and have passed, so they should be processed regardless of msg fees.
	// So in here, if there's an error getting the fee gas meter, we skip all this msg fee consumption.
	gasMeter, _ := antewrapper.GetFlatFeeGasMeter(ctx)
	if gasMeter != nil {
		gasMeter.ConsumeMsg(msg)
	}
}
