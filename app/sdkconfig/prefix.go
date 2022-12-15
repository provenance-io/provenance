package sdkconfig

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	AccountAddressPrefixMainNet = "pb"
	AccountAddressPrefixTestNet = "tp"
	CoinTypeMainNet             = 505
	CoinTypeTestNet             = 1
	Purpose                     = 44
)

var (
	AccountAddressPrefix   string
	AccountPubKeyPrefix    string
	ValidatorAddressPrefix string
	ValidatorPubKeyPrefix  string
	ConsNodeAddressPrefix  string
	ConsNodePubKeyPrefix   string
	CoinType               int
)

func setVars(hrp string, coinType int) {
	AccountAddressPrefix = hrp
	AccountPubKeyPrefix = AccountAddressPrefix + "pub"
	ValidatorAddressPrefix = AccountAddressPrefix + "valoper"
	ValidatorPubKeyPrefix = AccountAddressPrefix + "valoperpub"
	ConsNodeAddressPrefix = AccountAddressPrefix + "valcons"
	ConsNodePubKeyPrefix = AccountAddressPrefix + "valconspub"
	CoinType = coinType
}

func init() {
	// Defaults are for mainnet
	setVars(AccountAddressPrefixMainNet, CoinTypeMainNet)
}

// SetConfig sets the configuration for the network using mainnet or testnet
func SetConfig(testnet bool, seal bool) {
	if testnet {
		// not the default (mainnet) so reset with testnet config
		setVars(AccountAddressPrefixTestNet, CoinTypeTestNet)
	}

	config := sdk.GetConfig()
	config.SetCoinType(uint32(CoinType))
	config.SetPurpose(Purpose)
	config.SetBech32PrefixForAccount(AccountAddressPrefix, AccountPubKeyPrefix)
	config.SetBech32PrefixForValidator(ValidatorAddressPrefix, ValidatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(ConsNodeAddressPrefix, ConsNodePubKeyPrefix)
	if seal {
		config.Seal()
	}
}
