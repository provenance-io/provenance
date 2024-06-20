package types

import fmt "fmt"

// WasmAny represents the type with raw bytes value for codectypes.Any
type WasmAny struct {
	Value []byte
}

func (*WasmAny) ProtoMessage()             {}
func (*WasmAny) XXX_WellKnownType() string { return "BytesValue" } //nolint:revive
func (m *WasmAny) Reset()                  { *m = WasmAny{} }
func (m *WasmAny) String() string {
	return fmt.Sprintf("%x", m.Value) // not compatible w/ pb oct
}
func (m *WasmAny) Unmarshal(b []byte) error {
	m.Value = append([]byte(nil), b...)
	return nil
}
