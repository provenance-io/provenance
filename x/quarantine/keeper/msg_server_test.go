package keeper_test

import (
	"github.com/provenance-io/provenance/x/quarantine"
)

func (s *TestSuite) TestOptIn() {
	s.Run("quarantine module removed", func() {
		resp, err := s.keeper.OptIn(s.stdlibCtx, &quarantine.MsgOptIn{ToAddress: s.addr1.String()})
		s.Assert().Nil(resp, "response")
		s.Assert().ErrorContains(err, "quarantine module has been removed")
	})
}

func (s *TestSuite) TestOptOut() {
	s.Run("quarantine module removed", func() {
		resp, err := s.keeper.OptOut(s.stdlibCtx, &quarantine.MsgOptOut{ToAddress: s.addr1.String()})
		s.Assert().Nil(resp, "response")
		s.Assert().ErrorContains(err, "quarantine module has been removed")
	})
}

func (s *TestSuite) TestAccept() {
	s.Run("quarantine module removed", func() {
		resp, err := s.keeper.Accept(s.stdlibCtx, &quarantine.MsgAccept{
			ToAddress:     s.addr1.String(),
			FromAddresses: []string{s.addr2.String()},
		})
		s.Assert().Nil(resp, "response")
		s.Assert().ErrorContains(err, "quarantine module has been removed")
	})
}

func (s *TestSuite) TestDecline() {
	s.Run("quarantine module removed", func() {
		resp, err := s.keeper.Decline(s.stdlibCtx, &quarantine.MsgDecline{
			ToAddress:     s.addr1.String(),
			FromAddresses: []string{s.addr2.String()},
		})
		s.Assert().Nil(resp, "response")
		s.Assert().ErrorContains(err, "quarantine module has been removed")
	})
}

func (s *TestSuite) TestUpdateAutoResponses() {
	s.Run("quarantine module removed", func() {
		resp, err := s.keeper.UpdateAutoResponses(s.stdlibCtx, &quarantine.MsgUpdateAutoResponses{
			ToAddress: s.addr1.String(),
		})
		s.Assert().Nil(resp, "response")
		s.Assert().ErrorContains(err, "quarantine module has been removed")
	})
}
