package types

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AttributeTestSuite struct {
	suite.Suite
}

func TestAttributeTestSuite(t *testing.T) {
	suite.Run(t, new(AttributeTestSuite))
}

func (s *AttributeTestSuite) TestAttributeValidateBasic() {
	cases := map[string]struct {
		attribute Attribute
		expectErr bool
		errValue  string
	}{
		"should fail to validate basic attribute empty name": {
			attribute: Attribute{
				Name:          "",
				Value:         []byte("string value"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_String,
			},
			expectErr: true,
			errValue: "invalid name: empty",
		},
		"should fail to validate basic attribute value is nil": {
			attribute: Attribute{
				Name:          "value",
				Value:         nil,
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_String,
			},
			expectErr: true,
			errValue: "invalid value: nil",
		},
		"should fail to validate basic attribute invalid address": {
			attribute: Attribute{
				Name:          "name",
				Value:         []byte("string value"),
				Address:       "not an address",
				AttributeType: AttributeType_String,
			},
			expectErr: true,
			errValue: "invalid attribute address: must be either an account address or scope metadata address: \"not an address\"",
		},
		"should fail to validate basic attribute invalid type": {
			attribute: Attribute{
				Name:          "type",
				Value:         []byte("string value"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: -100,
			},
			expectErr: true,
			errValue: "invalid attribute type",
		},
		"should fail to validate basic attribute invalid type unspecified": {
			attribute: Attribute{
				Name:          "type",
				Value:         []byte("string value"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_Unspecified,
			},
			expectErr: true,
			errValue: "invalid attribute type",
		},
		"should fail to validate basic attribute invalid value for type uuid": {
			attribute: Attribute{
				Name:          "uuid",
				Value:         []byte("string value"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_UUID,
			},
			expectErr: true,
			errValue: "invalid attribute value for assigned type: ATTRIBUTE_TYPE_UUID",
		},
		"should succeed to validate basic attribute for type uuid": {
			attribute: Attribute{
				Name:          "uuid",
				Value:         []byte("91978ba2-5f35-459a-86a7-feca1b0512e0"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_UUID,
			},
			expectErr: false,
			errValue: "",
		},
		"should fail to validate basic attribute invalid value for type json": {
			attribute: Attribute{
				Name:          "json",
				Value:         []byte("{"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_JSON,
			},
			expectErr: true,
			errValue: "invalid attribute value for assigned type: ATTRIBUTE_TYPE_JSON",
		},
		"should succeed to validate basic attribute for type json": {
			attribute: Attribute{
				Name:          "json",
				Value:         []byte("{}"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_JSON,
			},
			expectErr: false,
			errValue: "",
		},
		"should fail to validate basic attribute invalid value for type uri": {
			attribute: Attribute{
				Name:          "uri",
				Value:         []byte("not uri"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_Uri,
			},
			expectErr: true,
			errValue: "invalid attribute value for assigned type: ATTRIBUTE_TYPE_URI",
		},
		"should succeed to validate basic attribute for type uri": {
			attribute: Attribute{
				Name:          "uri",
				Value:         []byte("ftp://www.this-is-uri.com"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_Uri,
			},
			expectErr: false,
			errValue: "",
		},
		"should fail to validate basic attribute invalid value for type int": {
			attribute: Attribute{
				Name:          "int",
				Value:         []byte("not an int"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_Int,
			},
			expectErr: true,
			errValue: "invalid attribute value for assigned type: ATTRIBUTE_TYPE_INT",
		},
		"should succeed to validate basic attribute for type int": {
			attribute: Attribute{
				Name:          "int",
				Value:         []byte("406"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_Int,
			},
			expectErr: false,
			errValue: "",
		},
		"should fail to validate basic attribute invalid value for type float": {
			attribute: Attribute{
				Name:          "float",
				Value:         []byte("not a float"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_Float,
			},
			expectErr: true,
			errValue: "invalid attribute value for assigned type: ATTRIBUTE_TYPE_FLOAT",
		},
		"should succeed to validate basic attribute for type float": {
			attribute: Attribute{
				Name:          "float",
				Value:         []byte("3.14"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_Float,
			},
			expectErr: false,
			errValue: "",
		},
		"should succeed to validate basic attribute for type proto": {
			attribute: Attribute{
				Name:          "proto",
				Value:         []byte("Treat proto as just a special tag for bytes"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_Proto,
			},
			expectErr: false,
			errValue: "",
		},
		"should succeed to validate basic attribute for type bytes": {
			attribute: Attribute{
				Name:          "bytes",
				Value:         []byte("there will be bytes"),
				Address:       "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				AttributeType: AttributeType_Bytes,
			},
			expectErr: false,
			errValue: "",
		},
	}

	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := tc.attribute.ValidateBasic()
			if tc.expectErr {
				s.Error(err)
				s.Equal(err.Error(), tc.errValue)
			} else {
				s.NoError(err)
			}

		})
	}
}
