package queries

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func GetLastGovProp(val *network.Validator) (*govv1.Proposal, error) {
	url := fmt.Sprintf("%s/cosmos/gov/v1/proposals?limit=1&reverse=true", val.APIAddress)
	resp, err := GetRequest(val, url, &govv1.QueryProposalsResponse{})
	if err != nil {
		return nil, err
	}
	if len(resp.Proposals) == 0 {
		return nil, errors.New("no governance proposals found")
	}
	return resp.Proposals[0], nil
}

func GetGovProp(val *network.Validator, propID string) (*govv1.Proposal, error) {
	url := fmt.Sprintf("%s/cosmos/gov/v1/proposals/%s", val.APIAddress, propID)
	resp, err := GetRequest(val, url, &govv1.QueryProposalResponse{})
	if err != nil {
		return nil, err
	}
	if resp.Proposal == nil {
		return nil, fmt.Errorf("governance proposal %d not found", propID)
	}
	return resp.Proposal, nil
}
