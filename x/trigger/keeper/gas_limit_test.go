package keeper_test

import (
	"github.com/provenance-io/provenance/x/trigger/types"
)

func (s *KeeperTestSuite) TestGetAllGasLimits() {
	tests := []struct {
		name      string
		gasLimits []types.GasLimit
	}{
		{
			name:      "valid - no gas limits",
			gasLimits: []types.GasLimit(nil),
		},
		{
			name: "valid - one gas limit",
			gasLimits: []types.GasLimit{
				{
					TriggerId: 1,
					Amount:    5,
				},
			},
		},
		{
			name: "valid - multiple gas limits",
			gasLimits: []types.GasLimit{
				{
					TriggerId: 1,
					Amount:    5,
				},
				{
					TriggerId: 2,
					Amount:    7,
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, gasLimit := range tc.gasLimits {
				s.app.TriggerKeeper.SetGasLimit(s.ctx, gasLimit.TriggerId, gasLimit.Amount)
			}
			gasLimits, err := s.app.TriggerKeeper.GetAllGasLimits(s.ctx)
			s.NoError(err)
			s.Equal(tc.gasLimits, gasLimits)
		})
	}
}

func (s *KeeperTestSuite) TestGetAndSetGasLimits() {
	s.app.TriggerKeeper.SetGasLimit(s.ctx, 1, 5)

	tests := []struct {
		name     string
		id       types.TriggerID
		expected uint64
		panic    bool
	}{
		{
			name:     "valid - gas limit",
			id:       1,
			expected: 5,
			panic:    false,
		},
		{
			name:     "invalid - gas limit doesn't exist",
			id:       999,
			expected: 0,
			panic:    true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.panic {
				s.Panics(func() {
					s.app.TriggerKeeper.GetGasLimit(s.ctx, tc.id)
				})
			} else {
				gasLimit := s.app.TriggerKeeper.GetGasLimit(s.ctx, tc.id)
				s.Equal(tc.expected, gasLimit)
			}
		})
	}
}

func (s *KeeperTestSuite) TestIterateGasLimits() {
	tests := []struct {
		name      string
		shortStop bool
		add       []types.GasLimit
		expected  []uint64
	}{
		{
			name:      "valid - no gas limits",
			shortStop: false,
			add:       []types.GasLimit{},
			expected:  []uint64{},
		},
		{
			name:      "valid - one gas limit",
			shortStop: false,
			add: []types.GasLimit{
				{
					TriggerId: 1,
					Amount:    5,
				},
			},
			expected: []uint64{1},
		},
		{
			name:      "valid - multiple gas limits",
			shortStop: false,
			add: []types.GasLimit{
				{
					TriggerId: 1,
					Amount:    5,
				},
				{
					TriggerId: 2,
					Amount:    7,
				},
			},
			expected: []uint64{1, 2},
		},
		{
			name:      "valid - do not iterate through all",
			shortStop: true,
			add:       []types.GasLimit{},
			expected:  []uint64{1},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, gasLimit := range tc.add {
				s.app.TriggerKeeper.SetGasLimit(s.ctx, gasLimit.TriggerId, gasLimit.Amount)
			}
			ids := []uint64{}
			err := s.app.TriggerKeeper.IterateGasLimits(s.ctx, func(gasLimit types.GasLimit) (stop bool, err error) {

				ids = append(ids, gasLimit.TriggerId)
				return tc.shortStop, nil
			})
			s.NoError(err)
			s.Equal(tc.expected, ids)
		})
	}
}

func (s *KeeperTestSuite) TestRemoveGasLimit() {
	s.app.TriggerKeeper.SetGasLimit(s.ctx, 1, 5)

	tests := []struct {
		name     string
		id       types.TriggerID
		expected bool
	}{
		{
			name:     "valid - gas limit",
			id:       1,
			expected: true,
		},
		{
			name:     "invalid - gas limit doesn't exist",
			id:       2,
			expected: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			success := s.app.TriggerKeeper.RemoveGasLimit(s.ctx, tc.id)
			s.Equal(tc.expected, success)
		})
	}
}
