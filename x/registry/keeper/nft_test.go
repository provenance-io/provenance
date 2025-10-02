package keeper_test

import (
	"cosmossdk.io/x/nft"
)

func (s *KeeperTestSuite) TestHasNFT() {
	tests := []struct {
		name     string
		setup    func() (string, string)
		expected bool
	}{
		{
			name: "nft exists in nft module",
			setup: func() (string, string) {
				return s.validNFTClass.Id, s.validNFT.Id
			},
			expected: true,
		},
		{
			name: "nft does not exist in nft module",
			setup: func() (string, string) {
				return "nonexistent-class", "nonexistent-nft"
			},
			expected: false,
		},
		{
			name: "different class, nft exists",
			setup: func() (string, string) {
				// Create another class and nft
				class := nft.Class{Id: "another-class"}
				s.nftKeeper.SaveClass(s.ctx, class)
				nftItem := nft.NFT{ClassId: class.Id, Id: "another-nft"}
				s.nftKeeper.Mint(s.ctx, nftItem, s.user1Addr)
				return class.Id, nftItem.Id
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			assetClassID, nftID := tc.setup()

			var result bool
			testFunc := func() {
				result = s.app.RegistryKeeper.HasNFT(s.ctx, &assetClassID, &nftID)
			}

			s.Require().NotPanics(testFunc, "HasNFT")
			s.Require().Equal(tc.expected, result, "HasNFT result")
		})
	}
}

func (s *KeeperTestSuite) TestAssetClassExists() {
	tests := []struct {
		name     string
		setup    func() string
		expected bool
	}{
		{
			name: "class exists in nft module",
			setup: func() string {
				return s.validNFTClass.Id
			},
			expected: true,
		},
		{
			name: "class does not exist",
			setup: func() string {
				return "nonexistent-class"
			},
			expected: false,
		},
		{
			name: "new class created",
			setup: func() string {
				class := nft.Class{Id: "new-class"}
				s.nftKeeper.SaveClass(s.ctx, class)
				return class.Id
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			assetClassID := tc.setup()

			var result bool
			testFunc := func() {
				result = s.app.RegistryKeeper.AssetClassExists(s.ctx, &assetClassID)
			}

			s.Require().NotPanics(testFunc, "AssetClassExists")
			s.Require().Equal(tc.expected, result, "AssetClassExists result")
		})
	}
}

func (s *KeeperTestSuite) TestGetNFTOwner() {
	tests := []struct {
		name         string
		setup        func() (string, string)
		expectedAddr string
	}{
		{
			name: "get owner from nft module",
			setup: func() (string, string) {
				return s.validNFTClass.Id, s.validNFT.Id
			},
			expectedAddr: s.user1,
		},
		{
			name: "get owner of newly minted nft",
			setup: func() (string, string) {
				class := nft.Class{Id: "owner-test-class"}
				s.nftKeeper.SaveClass(s.ctx, class)
				nftItem := nft.NFT{ClassId: class.Id, Id: "owner-test-nft"}
				s.nftKeeper.Mint(s.ctx, nftItem, s.user2Addr)
				return class.Id, nftItem.Id
			},
			expectedAddr: s.user2,
		},
		{
			name: "nonexistent nft returns nil",
			setup: func() (string, string) {
				return "nonexistent-class", "nonexistent-nft"
			},
			expectedAddr: "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			assetClassID, nftID := tc.setup()

			var owner string
			testFunc := func() {
				ownerAddr := s.app.RegistryKeeper.GetNFTOwner(s.ctx, &assetClassID, &nftID)
				if ownerAddr != nil {
					owner = ownerAddr.String()
				}
			}

			s.Require().NotPanics(testFunc, "GetNFTOwner")
			s.Require().Equal(tc.expectedAddr, owner, "GetNFTOwner result")
		})
	}
}
