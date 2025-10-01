package config

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	txmodule "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
)

// This is similar to the content in the SDK's client/config/config.go file.
// Key Differences:
//   - Our version allows us to change the defaults for testnets.
//   - We have a validation method (like the other configs).
//   - We can't use the SDK's ReadFromClientConfig:
//   - That is restricted to only the client.toml file, which doesn't work with a packed config.
//   - That writes the client.toml file if it doesn't exist, which we don't want.
//   - We have the ApplyClientConfigToContext function which is mostly extracted from
//     the SDK's ReadFromClientConfig with the addition of the signing mode.
//   - The SDK's file writing stuff is all private, and they don't give a way to just write the file.
//   - The way the SDK loads the config into viper stomps on previously loades stuff.

var (
	DefaultChainID        = ""
	DefaultKeyringBackend = "os"
	DefaultOutput         = "text"
	DefaultNode           = "tcp://localhost:26657"
	DefaultBroadcastMode  = "sync"
)

type ClientConfig struct {
	ChainID        string `mapstructure:"chain-id" json:"chain-id"`
	KeyringBackend string `mapstructure:"keyring-backend" json:"keyring-backend"`
	Output         string `mapstructure:"output" json:"output"`
	Node           string `mapstructure:"node" json:"node"`
	BroadcastMode  string `mapstructure:"broadcast-mode" json:"broadcast-mode"`
}

// DefaultClientConfig returns the reference to ClientConfig with default values.
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		ChainID:        DefaultChainID,
		KeyringBackend: DefaultKeyringBackend,
		Output:         DefaultOutput,
		Node:           DefaultNode,
		BroadcastMode:  DefaultBroadcastMode,
	}
}

func (c *ClientConfig) SetChainID(chainID string) {
	c.ChainID = chainID
}

func (c *ClientConfig) SetKeyringBackend(keyringBackend string) {
	c.KeyringBackend = keyringBackend
}

func (c *ClientConfig) SetOutput(output string) {
	c.Output = output
}

func (c *ClientConfig) SetNode(node string) {
	c.Node = node
}

func (c *ClientConfig) SetBroadcastMode(broadcastMode string) {
	c.BroadcastMode = broadcastMode
}

func (c ClientConfig) ValidateBasic() error {
	switch c.KeyringBackend {
	case "os", "file", "kwallet", "pass", "test", "memory":
	default:
		return errors.New("unknown keyring-backend (must be one of 'os', 'file', 'kwallet', 'pass', 'test', 'memory')")
	}
	switch c.Output {
	case "text", "json":
	default:
		return errors.New("unknown output (must be 'text' or 'json')")
	}
	switch c.BroadcastMode {
	case "sync", "async":
	default:
		return errors.New("unknown broadcast-mode (must be one of 'sync' 'async' 'block')")
	}
	return nil
}

// ApplyClientConfigToContext returns a client context after having the client config values applied to it.
func ApplyClientConfigToContext(ctx client.Context, config *ClientConfig) (client.Context, error) {
	// we need to update KeyringDir field on Client Context first cause it is used in NewKeyringFromBackend
	ctx = ctx.WithOutputFormat(config.Output).
		WithChainID(config.ChainID).
		WithKeyringDir(ctx.HomeDir)

	keyring, err := client.NewKeyringFromBackend(ctx, config.KeyringBackend)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get keyring with backend %q: %w", config.KeyringBackend, err)
	}

	ctx = ctx.WithKeyring(keyring)

	// https://github.com/cosmos/cosmos-sdk/issues/8986
	clnt, err := client.NewClientFromNode(config.Node)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get client from nodeURI: %w", err)
	}

	ctx = ctx.WithNodeURI(config.Node).
		WithClient(clnt).
		WithBroadcastMode(config.BroadcastMode)

	// This needs to go after the client has been set in the context.
	// That client is needed for SIGN_MODE_TEXTUAL.
	// This sign mode is only available if the client is online.
	if !ctx.Offline {
		enabledSignModes := tx.DefaultSignModes
		enabledSignModes = append(enabledSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
		txConfigOpts := tx.ConfigOptions{
			EnabledSignModes:           enabledSignModes,
			TextualCoinMetadataQueryFn: txmodule.NewGRPCCoinMetadataQueryFn(ctx),
		}

		var txConfig client.TxConfig
		txConfig, err = tx.NewTxConfigWithOptions(
			ctx.Codec,
			txConfigOpts,
		)
		if err != nil {
			return ctx, err
		}

		ctx = ctx.WithTxConfig(txConfig)
	}

	return ctx, nil
}
