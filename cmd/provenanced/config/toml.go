package config

import (
	"bytes"
	"os"
	"text/template"
)

// This is similar to the content in the SDK's client/config/toml.go file.
// The only difference in the defaultConfigTemplate is the fixed header border.
// This primarily exists, though, because the SDK's version isn't public,
// so we can't use that stuff to just write the file.
// Our WriteConfigToFile also panics instead of returning an error because
// the other two config file writers behave that way too.

const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

###############################################################################
###                           Client Configuration                          ###
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

	if err := os.WriteFile(configFilePath, buffer.Bytes(), 0o600); err != nil {
		panic(err)
	}
}
