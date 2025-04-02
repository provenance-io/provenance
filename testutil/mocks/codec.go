package mocks

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/gogoproto/proto"
)

// This doesn't yet have injectable errors for all the possible things because they weren't needed yet.
// Feel free to add them if you need them.

// MockCodec is a wrapper on a codec that allows injection of errors on these functions:
// MarshalJSON, MustMarshalJSON, UnmarshalJSON, MustUnmarshalJSON.
type MockCodec struct {
	codec.Codec
	MarshalJSONErrs   []string
	UnmarshalJSONErrs []string
	UnpackAnyErrs     []string
}

var (
	_ codec.Codec            = (*MockCodec)(nil)
	_ codec.JSONCodec        = (*MockCodec)(nil)
	_ codectypes.AnyUnpacker = (*MockCodec)(nil)
)

// NewMockCodec creates a new mock codec based on the standard test encoding config codec.
func NewMockCodec(baseCdc codec.Codec) *MockCodec {
	return &MockCodec{Codec: baseCdc}
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

// WithUnpackAnyErrs adds the given errors to be returned from UnpackAny.
// Each entry is used once in the order they are provided.
// An empty string indicates no error (do the normal thing).
// The receiver is both updated and returned.
func (c *MockCodec) WithUnpackAnyErrs(errMsgs ...string) *MockCodec {
	c.UnpackAnyErrs = append(c.UnpackAnyErrs, errMsgs...)
	return c
}

func (c *MockCodec) MarshalJSON(o proto.Message) ([]byte, error) {
	if len(c.MarshalJSONErrs) > 0 {
		errMsg := c.MarshalJSONErrs[0]
		c.MarshalJSONErrs = c.MarshalJSONErrs[1:]
		if len(errMsg) > 0 {
			return nil, errors.New(errMsg)
		}
	}
	return c.Codec.MarshalJSON(o)
}

func (c *MockCodec) MustMarshalJSON(o proto.Message) []byte {
	if len(c.MarshalJSONErrs) > 0 {
		errMsg := c.MarshalJSONErrs[0]
		c.MarshalJSONErrs = c.MarshalJSONErrs[1:]
		if len(errMsg) > 0 {
			panic(errors.New(errMsg))
		}
	}
	return c.Codec.MustMarshalJSON(o)
}

func (c *MockCodec) UnmarshalJSON(bz []byte, ptr proto.Message) error {
	if len(c.UnmarshalJSONErrs) > 0 {
		errMsg := c.UnmarshalJSONErrs[0]
		c.UnmarshalJSONErrs = c.UnmarshalJSONErrs[1:]
		if len(errMsg) > 0 {
			return errors.New(errMsg)
		}
	}
	return c.Codec.UnmarshalJSON(bz, ptr)
}

func (c *MockCodec) MustUnmarshalJSON(bz []byte, ptr proto.Message) {
	if len(c.UnmarshalJSONErrs) > 0 {
		errMsg := c.UnmarshalJSONErrs[0]
		c.UnmarshalJSONErrs = c.UnmarshalJSONErrs[1:]
		if len(errMsg) > 0 {
			panic(errors.New(errMsg))
		}
	}
	c.Codec.MustUnmarshalJSON(bz, ptr)
}

func (c *MockCodec) UnpackAny(any *codectypes.Any, iface interface{}) error {
	if len(c.UnpackAnyErrs) > 0 {
		errMsg := c.UnpackAnyErrs[0]
		c.UnpackAnyErrs = c.UnpackAnyErrs[1:]
		if len(errMsg) > 0 {
			return errors.New(errMsg)
		}
	}
	return c.Codec.UnpackAny(any, iface)
}
