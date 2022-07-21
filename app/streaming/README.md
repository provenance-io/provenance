# Provenance Streaming Services

This package contains a map of supported streaming services that push data out to external systems. 
In addition, this package contains service implementations that build on top of `StreamService` and `StreamServiceInitializer` interfaces.
These two interfaces are defined in `internal/streaming/streaming.go`.

## `StreamServiceInitializer`

This interface is defined as:

```go
// internal/streaming/streaming.go

// StreamServiceInitializer interface initializes StreamService
// implementations which are then registered with the App
type StreamServiceInitializer interface {
	// Init configures and initializes the streaming service
	Init(opts servertypes.AppOptions, marshaller codec.BinaryCodec) StreamService
}
```

Stream Services need to extend this interface to enable them to be initialized and loaded by the `App`. 
In addition, the service must be defined in the `StreamServiceIntializers` map in `app/streaming/streaming.go` 
to allow the App to load a service through `config` properties. 
Take a look at [app/streaming/trace/trace.go](./trace/trace.go) for an implementation example of this interface.

Implementations must be added `app/streaming/streaming.go` for the App to be able to load the service when enabled. 
See the [configuration](#configuration) section for how to configure a service.

```go
// app/streaming/streaming.go

// StreamServiceInitializers contains a map of supported StreamServiceInitializer implementations
var StreamServiceInitializers = map[string]streaming.StreamServiceInitializer{
	"trace": trace.StreamServiceInitializer,
}
```

## `StreamService`

This interface is defined as:

```go
// internal/streaming/streaming.go

// StreamService interface used to hook into the ABCI message processing of the BaseApp
type StreamService interface {
    // StreamBeginBlocker updates the streaming service with the latest BeginBlock messages
    StreamBeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock)
    // StreamEndBlocker updates the steaming service with the latest EndBlock messages
    StreamEndBlocker(ctx sdk.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock)
}
```

Take a look at [app/streaming/trace/service/service.go](./trace/service/service.go) for an implementation example of this interface.

## Configuration

Streaming services are configured in the `streaming` TOML mapping in the App's `app.toml` file. There are two parameters
for configuring a service: `streaming.enabled` and `streaming.service`. `streaming.enabled` is bool that turns on or of a streaming service.
`streaming.service` specifies the service name that is registered with the App. 

```toml
[streaming]
    enabled = true
    # The streaming service name that ABCI BeginBlocker and EndBlocker request and response will be sent to.
    # Supported services are: trace
    service = "service name to stream ABCI data"
```

This provides node operates with the ability to `opt-in` and enable streaming to external systems.

At this time, the only pre-defined service is the [trace](./trace) streaming service.
AS mentioned above, service can be added by adding the service to the `StreamServiceInitializers` 
in [app/streaming/streaming.go](./streaming.go) and adding configuration properties in `app.toml`. 
See [Trace streaming service configuration](#trace-streaming-service-configuration) for an example.

### Trace streaming service configuration

The configuration for the `trace` service is defined as:

```toml
[streaming]
enabled = true
# The streaming service name that ABCI BeginBlocker and EndBlocker request and response will be sent to.
# Supported services are: trace
service = "trace"

[streaming.trace]
# When true, it will print ABCI BeginBlocker and EndBlocker request and response to stdout.
print_data_to_stdout = false
```
