package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type ActionBuilderTestSuite struct {
	suite.Suite
}

func TestActionBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(ActionBuilderTestSuite))
}

func (s *ActionBuilderTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *ActionBuilderTestSuite) TestDelegateActionGetEventCriteria() {
	builder := DelegateActionBuilder{}
	criteria := builder.GetEventCriteria()

	s.Assert().Contains(criteria.Events, "message", "must contain message event")
	s.Assert().Contains(criteria.Events["message"].Attributes, "module", "must contain module attribute")
	s.Assert().Contains(criteria.Events, "delegate", "must contain delegate event")
	s.Assert().Contains(criteria.Events, "create_validator", "must contain create_validator event")
}

func (s *ActionBuilderTestSuite) TestDelegateActionAddEventDelegate() {
	builder := DelegateActionBuilder{}

	err := builder.AddEvent("delegate", &map[string][]byte{
		"validator": []byte("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun"),
	})
	s.Assert().NoError(err, "add event should not return an error")
	validator_address, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	s.Assert().Equal(validator_address, builder.Validator, "validator address should be set")

	err = builder.AddEvent("delegate", &map[string][]byte{
		"validator": []byte("blah"),
	})
	s.Assert().Error(err, "add event should return error on invalid validator address")
}

func (s *ActionBuilderTestSuite) TestDelegateActionAddEventCreateValidator() {
	builder := DelegateActionBuilder{}

	err := builder.AddEvent("create_validator", &map[string][]byte{
		"validator": []byte("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun"),
	})
	s.Assert().NoError(err, "add event should not return an error")
	validator_address, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	s.Assert().Equal(validator_address, builder.Validator, "validator address should be set")

	err = builder.AddEvent("create_validator", &map[string][]byte{
		"validator": []byte("blah"),
	})
	s.Assert().Error(err, "add event should return error on invalid validator address")
}

func (s *ActionBuilderTestSuite) TestDelegateActionAddEventMessage() {
	builder := DelegateActionBuilder{}

	err := builder.AddEvent("message", &map[string][]byte{
		"sender": []byte("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
	})
	s.Assert().NoError(err, "add event should not return an error")
	delegator_address, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	s.Assert().Equal(delegator_address, builder.Delegator, "delegator address should be set")

	err = builder.AddEvent("message", &map[string][]byte{
		"sender": []byte("blah"),
	})
	s.Assert().Error(err, "add event should return error on invalid delegator address")
}

func (s *ActionBuilderTestSuite) TestDelegateActionAddEventInvalid() {
	builder := DelegateActionBuilder{}

	err := builder.AddEvent("blah", &map[string][]byte{})
	s.Assert().NoError(err, "add event should not return an error")
	s.Assert().True(builder.Delegator.Empty(), "delegator should not be set")
	s.Assert().True(builder.Validator.Empty(), "validator should not be set")
}

func (s *ActionBuilderTestSuite) TestReset() {
	builder := DelegateActionBuilder{}
	builder.Delegator, _ = sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	builder.Validator, _ = sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	builder.Reset()
	s.Assert().True(builder.Delegator.Empty(), "reset should remove delegator address")
	s.Assert().True(builder.Validator.Empty(), "reset should remove validator address")
}

func (s *ActionBuilderTestSuite) TestCanBuildSuccess() {
	builder := DelegateActionBuilder{}
	builder.Delegator, _ = sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	builder.Validator, _ = sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	s.Assert().True(builder.CanBuild())
}

func (s *ActionBuilderTestSuite) TestCanBuildFailure() {
	builder := DelegateActionBuilder{}
	s.Assert().False(builder.CanBuild())

	builder.Validator, _ = sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	s.Assert().False(builder.CanBuild())

	builder.Reset()
	builder.Delegator, _ = sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	s.Assert().False(builder.CanBuild())
}

func (s *ActionBuilderTestSuite) TestBuildActionValid() {
	builder := DelegateActionBuilder{}
	builder.Delegator, _ = sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	builder.Validator, _ = sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	result, err := builder.BuildAction()
	s.Assert().NoError(err, "builder should not return an error on successful build")
	s.Assert().Equal(int64(1), result.Shares, "1 share should be returned")
	s.Assert().Equal(builder.Delegator, result.Address, "address should be set to delegator address")
	s.Assert().Equal(builder.Delegator, result.Delegator, "delegator should be set to delegator address")
	s.Assert().Equal(builder.Validator, result.Validator, "validator should be set to validator address")
}

func (s *ActionBuilderTestSuite) TestBuildActionInvalid() {
	builder := DelegateActionBuilder{}
	_, err := builder.BuildAction()
	s.Assert().Error(err, "builder should return error on unsuccessful build")
}
