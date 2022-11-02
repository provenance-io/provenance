package keeper

import (
	"github.com/provenance-io/provenance/x/marker/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFilterAccessList(t *testing.T) {
	adminAddress := "cosmos1sh49f6ze3vn7cdl2amh2gnc70z5mten3y08xck"

	testCases := []struct {
		name          string
		accessList    []types.AccessGrant
		administrator string
		want          []types.AccessGrant
		errorMsg      string
	}{
		{
			name: "fail - administrator does not have transfer rights",
			accessList: []types.AccessGrant{
				{
					"non admin address",
					[]types.Access{
						types.Access_Deposit,
						types.Access_Transfer,
						types.Access_Mint,
						types.Access_Burn,
					},
				},
			},
			administrator: adminAddress,
			want: nil,
			errorMsg: "marker does not have valid access rights",
		},
		{
			name: "success - filtered access list should not have mint or burn permissions",
			accessList: []types.AccessGrant{
				{
					adminAddress,
					[]types.Access{
						types.Access_Deposit,
						types.Access_Transfer,
						types.Access_Mint,
						types.Access_Burn,
					},
				},
			},
			administrator: adminAddress,
			want: []types.AccessGrant{
				{
					adminAddress,
					[]types.Access{
						types.Access_Deposit,
						types.Access_Transfer,
					},
				},
			},
			errorMsg: "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := filterAccessList(tc.accessList, tc.administrator)

			if len(tc.errorMsg) > 0 {
				require.Error(t, err)
				require.Equal(t, tc.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}
