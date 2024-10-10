package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	AccountAddressPrefixMainNet = "pb"
	AccountAddressPrefixTestNet = "tp"
	CoinTypeMainNet             = uint32(505)
	CoinTypeTestNet             = uint32(1)
	Purpose                     = 44

	// EnvPrefix is the prefix added to config/flag names to get its environment variable name.
	EnvPrefix = "PIO"
)

var (
	// Defaults are for mainnet

	AccountAddressPrefix   = AccountAddressPrefixMainNet
	AccountPubKeyPrefix    = AccountAddressPrefix + "pub"
	ValidatorAddressPrefix = AccountAddressPrefix + "valoper"
	ValidatorPubKeyPrefix  = AccountAddressPrefix + "valoperpub"
	ConsNodeAddressPrefix  = AccountAddressPrefix + "valcons"
	ConsNodePubKeyPrefix   = AccountAddressPrefix + "valconspub"
	CoinType               = CoinTypeMainNet
)

// SetConfig sets the configuration for the network using mainnet or testnet
func SetConfig(testnet bool, seal bool) {
	AccountAddressPrefix = AccountAddressPrefixMainNet
	CoinType = CoinTypeMainNet
	if testnet {
		AccountAddressPrefix = AccountAddressPrefixTestNet
		CoinType = CoinTypeTestNet
	}
	AccountPubKeyPrefix = AccountAddressPrefix + "pub"
	ValidatorAddressPrefix = AccountAddressPrefix + "valoper"
	ValidatorPubKeyPrefix = AccountAddressPrefix + "valoperpub"
	ConsNodeAddressPrefix = AccountAddressPrefix + "valcons"
	ConsNodePubKeyPrefix = AccountAddressPrefix + "valconspub"

	config := sdk.GetConfig()
	config.SetCoinType(CoinType)
	config.SetPurpose(Purpose)
	config.SetBech32PrefixForAccount(AccountAddressPrefix, AccountPubKeyPrefix)
	config.SetBech32PrefixForValidator(ValidatorAddressPrefix, ValidatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(ConsNodeAddressPrefix, ConsNodePubKeyPrefix)

	if seal {
		config.Seal()
	}
}
