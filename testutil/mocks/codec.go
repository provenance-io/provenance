package mocks

import (
	"errors"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdksim "cosmossdk.io/simapp"
)

// This doesn't yet have injectable errors for all the possible things because they weren't needed yet.
// Feel free to add them if you need them.

// MockCodec is a wrapper on a codec that allows injection of errors on these functions:
// MarshalJSON, MustMarshalJSON, UnmarshalJSON, MustUnmarshalJSON.
type MockCodec struct {
	Base              codec.Codec
	MarshalJSONErrs   []string
	UnmarshalJSONErrs []string
}

var _ codec.Codec = (*MockCodec)(nil)

// NewMockCodec creates a new mock codec based on the standard test encoding config codec.
func NewMockCodec() *MockCodec {
	return &MockCodec{Base: sdksim.MakeTestEncodingConfig().Codec}
}

// WithMarshalJSONErrs adds the given errors to be returned from MarshalJSON or MustMarshalJSON.
// Each entry is used once in the order they are provided.
// An empty string indicates no error (do the normal thing).
// The receiver is both updated and returned.
func (c *MockCodec) WithMarshalJSONErrs(errMsgs ...string) *MockCodec {
	c.MarshalJSONErrs = append(c.MarshalJSONErrs, errMsgs...)
	return c
}

// WithUnmarshalJSONErrs adds the given errors to be returned from UnmarshalJSON or MustUnmarshalJSON.
// Each entry is used once in the order they are provided.
// An empty string indicates no error (do the normal thing).
// The receiver is both updated and returned.
func (c *MockCodec) WithUnmarshalJSONErrs(errMsgs ...string) *MockCodec {
	c.UnmarshalJSONErrs = append(c.UnmarshalJSONErrs, errMsgs...)
	return c
}

func (c *MockCodec) Marshal(o codec.ProtoMarshaler) ([]byte, error) {
	return c.Base.Marshal(o)
}

func (c *MockCodec) MustMarshal(o codec.ProtoMarshaler) []byte {
	return c.Base.MustMarshal(o)
}

func (c *MockCodec) MarshalLengthPrefixed(o codec.ProtoMarshaler) ([]byte, error) {
	return c.Base.MarshalLengthPrefixed(o)
}

func (c *MockCodec) MustMarshalLengthPrefixed(o codec.ProtoMarshaler) []byte {
	return c.Base.MustMarshalLengthPrefixed(o)
}

func (c *MockCodec) Unmarshal(bz []byte, ptr codec.ProtoMarshaler) error {
	return c.Base.Unmarshal(bz, ptr)
}

func (c *MockCodec) MustUnmarshal(bz []byte, ptr codec.ProtoMarshaler) {
	c.Base.MustUnmarshal(bz, ptr)
}

func (c *MockCodec) UnmarshalLengthPrefixed(bz []byte, ptr codec.ProtoMarshaler) error {
	return c.Base.UnmarshalLengthPrefixed(bz, ptr)
}

func (c *MockCodec) MustUnmarshalLengthPrefixed(bz []byte, ptr codec.ProtoMarshaler) {
	c.Base.MustUnmarshalLengthPrefixed(bz, ptr)
}

func (c *MockCodec) MarshalInterface(i proto.Message) ([]byte, error) {
	return c.Base.MarshalInterface(i)
}

func (c *MockCodec) UnmarshalInterface(bz []byte, ptr interface{}) error {
	return c.Base.UnmarshalInterface(bz, ptr)
}

func (c *MockCodec) UnpackAny(a *codectypes.Any, iface interface{}) error {
	return c.Base.UnpackAny(a, iface)
}

func (c *MockCodec) MarshalJSON(o proto.Message) ([]byte, error) {
	if len(c.MarshalJSONErrs) > 0 {
		errMsg := c.MarshalJSONErrs[0]
		c.MarshalJSONErrs = c.MarshalJSONErrs[1:]
		if len(errMsg) > 0 {
			return nil, errors.New(errMsg)
		}
	}
	return c.Base.MarshalJSON(o)
}

func (c *MockCodec) MustMarshalJSON(o proto.Message) []byte {
	if len(c.MarshalJSONErrs) > 0 {
		errMsg := c.MarshalJSONErrs[0]
		c.MarshalJSONErrs = c.MarshalJSONErrs[1:]
		if len(errMsg) > 0 {
			panic(errors.New(errMsg))
		}
	}
	return c.Base.MustMarshalJSON(o)
}

func (c *MockCodec) MarshalInterfaceJSON(i proto.Message) ([]byte, error) {
	return c.Base.MarshalInterfaceJSON(i)
}

func (c *MockCodec) UnmarshalInterfaceJSON(bz []byte, ptr interface{}) error {
	return c.Base.UnmarshalInterfaceJSON(bz, ptr)
}

func (c *MockCodec) UnmarshalJSON(bz []byte, ptr proto.Message) error {
	if len(c.UnmarshalJSONErrs) > 0 {
		errMsg := c.UnmarshalJSONErrs[0]
		c.UnmarshalJSONErrs = c.UnmarshalJSONErrs[1:]
		if len(errMsg) > 0 {
			return errors.New(errMsg)
		}
	}
	return c.Base.UnmarshalJSON(bz, ptr)
}

func (c *MockCodec) MustUnmarshalJSON(bz []byte, ptr proto.Message) {
	if len(c.UnmarshalJSONErrs) > 0 {
		errMsg := c.UnmarshalJSONErrs[0]
		c.UnmarshalJSONErrs = c.UnmarshalJSONErrs[1:]
		if len(errMsg) > 0 {
			panic(errors.New(errMsg))
		}
	}
	c.Base.MustUnmarshalJSON(bz, ptr)
}
