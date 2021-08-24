package config

import (
	"bytes"
	"os"
	"text/template"

	"github.com/spf13/viper"

	tmos "github.com/tendermint/tendermint/libs/os"
)

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                           Client Configuration                            ###
###############################################################################

# The network chain ID
chain-id = "{{ .ChainID }}"
# The keyring's backend, where the keys are stored (os|file|kwallet|pass|test|memory)
keyring-backend = "{{ .KeyringBackend }}"
# CLI output format (text|json)
output = "{{ .Output }}"
# <host>:<port> to Tendermint RPC interface for this chain
node = "{{ .Node }}"
# Transaction broadcasting mode (sync|async|block)
broadcast-mode = "{{ .BroadcastMode }}"
`

var configTemplate *template.Template

func init() {
	var err error
	tmpl := template.New("clientConfigFileTemplate")
	if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

// WriteConfigToFile creates the file contents using a template and the provided config
// then writes the contents to the provided configFilePath.
func WriteConfigToFile(configFilePath string, config *ClientConfig) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	tmos.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}

// ensureConfigPath creates a directory configPath if it does not exist
func ensureConfigPath(configPath string) error {
	// TODO: Remove ensureConfigPath. Shouldn't be needed after the overhaul of ReadFromClientConfig.
	return os.MkdirAll(configPath, os.ModePerm)
}

// GetClientConfig reads values from client.toml file and unmarshalls them into ClientConfig
func GetClientConfig(configPath string, v *viper.Viper) (*ClientConfig, error) {
	// TODO: Remove GetClientConfig in favor of stuff in the config manager.
	v.AddConfigPath(configPath)
	v.SetConfigName("client")
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	conf := new(ClientConfig)
	if err := v.Unmarshal(conf); err != nil {
		return nil, err
	}

	return conf, nil
}
