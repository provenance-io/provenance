package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
)

// Default constants
const (
	chainID        = ""
	keyringBackend = "test"
	output         = "text"
	node           = "tcp://localhost:26657"
	broadcastMode  = "block"
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
	return &ClientConfig{chainID, keyringBackend, output, node, broadcastMode}
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
	case "sync", "async", "block":
	default:
		return errors.New("unknown broadcast-mode (must be one of 'sync' 'async' 'block')")
	}
	return nil
}

// ReadFromClientConfig reads values from client.toml file and updates them in client Context
func ReadFromClientConfig(ctx client.Context) (client.Context, error) {
	// TODO: Overhaul ReadFromClientConfig. Split out context setup and config loading.
	configPath := filepath.Join(ctx.HomeDir, "config")
	configFilePath := filepath.Join(configPath, "client.toml")
	conf := DefaultClientConfig()

	// if client.toml file does not exist we create it and write default ClientConfig values into it.
	if _, ferr := os.Stat(configFilePath); os.IsNotExist(ferr) {
		if err := ensureConfigPath(configPath); err != nil {
			return ctx, fmt.Errorf("couldn't make client config: %v", err)
		}

		WriteConfigToFile(configFilePath, conf)
	}

	conf, err := GetClientConfig(configPath, ctx.Viper)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get client config: %v", err)
	}
	// we need to update KeyringDir field on Client Context first cause it is used in NewKeyringFromBackend
	ctx = ctx.WithOutputFormat(conf.Output).
		WithChainID(conf.ChainID).
		WithKeyringDir(ctx.HomeDir)

	keyring, err := client.NewKeyringFromBackend(ctx, conf.KeyringBackend)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get key ring: %v", err)
	}

	ctx = ctx.WithKeyring(keyring)

	// https://github.com/cosmos/cosmos-sdk/issues/8986
	clnt, err := client.NewClientFromNode(conf.Node)
	if err != nil {
		return ctx, fmt.Errorf("couldn't get client from nodeURI: %v", err)
	}

	ctx = ctx.WithNodeURI(conf.Node).
		WithClient(clnt).
		WithBroadcastMode(conf.BroadcastMode)

	return ctx, nil
}
