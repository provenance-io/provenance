package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// AccountAddressPrefixMainNet is the mainnet address prefix.
	AccountAddressPrefixMainNet = "pb"
	// AccountAddressPrefixTestNet is the testnet address prefix.
	AccountAddressPrefixTestNet = "tp"
	// CoinTypeMainNet is the coin type for mainnet.
	CoinTypeMainNet = uint32(505)
	// CoinTypeTestNet is the coin type for testnet.
	CoinTypeTestNet = uint32(1)
	// Purpose is a constant for prefix purpose.
	Purpose = 44

	// EnvPrefix is the prefix added to config/flag names to get its environment variable name.
	EnvPrefix = "PIO"
)

// Defaults are for mainnet
var (
	// AccountAddressPrefix is the prefix for account addresses.
	AccountAddressPrefix = AccountAddressPrefixMainNet
	// AccountPubKeyPrefix is the mainnet account public key prefix.
	AccountPubKeyPrefix = AccountAddressPrefix + "pub"
	// ValidatorPubKeyPrefix is the prefix for validator public keys.
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
