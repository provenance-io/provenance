# State
The attribute module inserts all attributes into a basic state store.

<!-- TOC -->
  - [Attribute KV-Store](#attribute-kv-store)
    - [Key layout](#key-layout)
    - [Attribute Record](#attribute-record)
    - [Attribute Type](#attribute-type)



## Attribute KV-Store

The attribute module takes in an attribute supplied by the user and generates a key for it. This key is generated
by combinining the attribute prefix, address, attribute name, and a hashed value of the attribute value. This
can then be used to either store a marshalled attribute record, or retrieve the value it points to in the store.

### Key layout
[0x02][address][attribute name][hashvalue]

### Attribute Record
```
// Attribute holds a typed key/value structure for data associated with an account
type Attribute struct {
	// The attribute name.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`

	// The attribute value.
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`

	// The attribute value type.
	AttributeType AttributeType `protobuf:"varint,3,opt,name=attribute_type,json=attributeType,proto3,enum=provenance.attribute.v1.AttributeType" json:"attribute_type,omitempty"`

	// The address the attribute is bound to
	Address string `protobuf:"bytes,4,opt,name=address,proto3" json:"address,omitempty"`
}
```

### Attribute Type
```
// AttributeType defines the type of the data stored in the attribute value
type AttributeType int32

const (
	// ATTRIBUTE_TYPE_UNSPECIFIED defines an unknown/invalid type
	AttributeType_Unspecified AttributeType = 0
	// ATTRIBUTE_TYPE_UUID defines an attribute value that contains a string value representation of a V4 uuid
	AttributeType_UUID AttributeType = 1
	// ATTRIBUTE_TYPE_JSON defines an attribute value that contains a byte string containing json data
	AttributeType_JSON AttributeType = 2
	// ATTRIBUTE_TYPE_STRING defines an attribute value that contains a generic string value
	AttributeType_String AttributeType = 3
	// ATTRIBUTE_TYPE_URI defines an attribute value that contains a URI
	AttributeType_Uri AttributeType = 4
	// ATTRIBUTE_TYPE_INT defines an attribute value that contains an integer (cast as int64)
	AttributeType_Int AttributeType = 5
	// ATTRIBUTE_TYPE_FLOAT defines an attribute value that contains a float
	AttributeType_Float AttributeType = 6
	// ATTRIBUTE_TYPE_PROTO defines an attribute value that contains a serialized proto value in bytes
	AttributeType_Proto AttributeType = 7
	// ATTRIBUTE_TYPE_BYTES defines an attribute value that contains an untyped array of bytes
	AttributeType_Bytes AttributeType = 8
)
```
