package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/provenance-io/provenance/x/reward/types"
)


type EvaluationResult struct {
	eventTypeToSearch string
	attributeKey string
	shares int64
	address sdk.AccAddress // shares to attribute to this address
}

// EvaluateRules takes in a Eligibility criteria and measure it against the events in the context
func (k Keeper) EvaluateRules(ctx sdk.Context, epochIdentifier string, epochNumber uint64, criteria *types.EligibilityCriteria) {
	// get the events from the context history
	switch criteria.Action.TypeUrl  {
	case  proto.MessageName(&types.ActionTransferDelegations{}):{
		ctx.Logger().Info(fmt.Sprintf("The Action type is %s",proto.MessageName(&types.ActionTransferDelegations{})))
		// check the event history
		// for transfers
		ctx.EventManager().GetABCIEventHistory()

	}
	case  proto.MessageName(&types.ActionDelegate{}):{
		ctx.Logger().Info(fmt.Sprintf("The Action type is %s",proto.MessageName(&types.ActionDelegate{})))
	}
	default:
		// TODO throw an error or just log it? Leaning towards just logging it for now
		ctx.Logger().Error(fmt.Sprintf("The Action type %s, cannot be evaluated",criteria.Action.TypeUrl))
	}
}

func GetEvent(ctx sdk.Context, eventTypeToSearch string, attributeKey string, ) ([]EvaluationResult,error) {

	result := ([]EvaluationResult)(nil)
	for _, s := range ctx.EventManager().GetABCIEventHistory() {
		ctx.Logger().Info(fmt.Sprintf("events type is %s", s.Type))
		if s.Type == eventTypeToSearch {
			// now look for the attribute
			for _, y := range s.Attributes {
				ctx.Logger().Info(fmt.Sprintf("event attribute is %s attribute_key:attribute_value  %s:%s", s.Type, y.Key, y.Value))
				//4:24PM INF events type is transfer
				//4:24PM INF event attribute is transfer attribute_key:attribute_value  recipient:tp17xpfvakm2amg962yls6f84z3kell8c5l2udfyt
				//4:24PM INF event attribute is transfer attribute_key:attribute_value  sender:tp1sha7e07l5knw4vdw2vgc3k06gd0fscz9r32yv6
				//4:24PM INF event attribute is transfer attribute_key:attribute_value  amount:76200000000000nhash
				if attributeKey == string(y.Key) {
					// really not possible to get an error but could happen i guess
					address,err:=sdk.AccAddressFromBech32(string(y.Value))
					if err!=nil{
						return nil,err
					}
					result = append(result, EvaluationResult{
						eventTypeToSearch: eventTypeToSearch,
						attributeKey:      string(y.Key),
						shares:            1,
						address:          address,
					})
				}
			}
		}
	}

	return result,nil

}
