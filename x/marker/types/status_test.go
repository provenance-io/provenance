package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {

}

func TestStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status MarkerStatus
		valid  bool
		expErr error
	}{
		{
			name: "undefined",
			status: StatusUndefined,
			valid: false,
			expErr: nil,
		},
		{
			name: "proposed",
			status: StatusProposed,
			valid: true,
			expErr: nil,
		},
		{
			name: "finalized",
			status: StatusFinalized,
			valid: true,
			expErr: nil,
		},
		{
			name: "active",
			status: StatusActive,
			valid: true,
			expErr: nil,
		},
		{
			name: "cancelled",
			status: StatusCancelled,
			valid: true,
			expErr: nil,
		},
		{
			name: "destroyed",
			status: StatusDestroyed,
			valid: true,
			expErr: nil,
		},
		{
			name: "not-defined",
			status: StatusUndefined,
			valid: false,
			expErr: fmt.Errorf("'%s' is not a valid marker status", "not-defined"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// check parse from string
			s, err := MarkerStatusFromString(tt.name)
			require.Equal(t, tt.expErr, err)
			// check marker is valid/or not
			require.Equal(t, ValidMarkerStatus(s), tt.valid)
			// if valid, check that parsed marker status matches expected marker status
			if err == nil {
				// no error so this shouldn't panic either.
				require.Equal(t, MustGetMarkerStatus(tt.name), tt.status)
				require.Equal(t, s, tt.status)
				require.Equal(t, s.String(), tt.name)
				require.Equal(t, s.String(), tt.name)
			}
		})
	}
}
