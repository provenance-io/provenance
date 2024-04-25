package cmd_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"

	"github.com/provenance-io/provenance/app"
	provenancecmd "github.com/provenance-io/provenance/cmd/provenanced/cmd"
	"github.com/provenance-io/provenance/testutil/assertions"
)

func TestDocGen(t *testing.T) {
	appCodec := app.MakeTestEncodingConfig(t).Marshaler
	tests := []struct {
		name         string
		target       string
		createTarget bool
		flags        []string
		err          string
		extensions   []string
	}{
		{
			name:         "failure - no flags specified",
			target:       "tmp",
			createTarget: true,
			err:          "at least one doc type must be specified",
		},
		{
			name:         "failure - unsupported flag format",
			target:       "tmp",
			flags:        []string{"--bad"},
			createTarget: true,
			err:          "unknown flag: --bad",
		},
		{
			name:         "failure - invalid target directory",
			target:       "/tmp/tmp2/tmp3",
			flags:        []string{"--yaml"},
			createTarget: false,
			err:          "mkdir %s: no such file or directory",
		},
		{
			name:         "failure - bad yaml value",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--yaml=xyz"},
			err:          "invalid argument \"xyz\" for \"--yaml\" flag: strconv.ParseBool: parsing \"xyz\": invalid syntax",
		},
		{
			name:         "failure - bad rst value",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--rst=xyz"},
			err:          "invalid argument \"xyz\" for \"--rst\" flag: strconv.ParseBool: parsing \"xyz\": invalid syntax",
		},
		{
			name:         "failure - bad markdown value",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--markdown=xyz"},
			err:          "invalid argument \"xyz\" for \"--markdown\" flag: strconv.ParseBool: parsing \"xyz\": invalid syntax",
		},
		{
			name:         "failure - bad manpage value",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--manpage=xyz"},
			err:          "invalid argument \"xyz\" for \"--manpage\" flag: strconv.ParseBool: parsing \"xyz\": invalid syntax",
		},
		{
			name:         "success - yaml is generated",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--yaml"},
			extensions:   []string{".yaml"},
		},
		{
			name:         "success - rst is generated",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--rst"},
			extensions:   []string{".rst"},
		},
		{
			name:         "success - manpage is generated",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--manpage"},
			extensions:   []string{".1"},
		},
		{
			name:         "success - markdown is generated",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--markdown"},
			extensions:   []string{".md"},
		},
		{
			name:         "success - multiple types supported",
			target:       "tmp",
			createTarget: true,
			flags:        []string{"--markdown", "--yaml"},
			extensions:   []string{".md", ".yaml"},
		},
		{
			name:         "success - generates a new directory",
			target:       "tmp2",
			createTarget: false,
			flags:        []string{"--yaml"},
			extensions:   []string{".yaml"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()

			targetPath := filepath.Join(home, tc.target)
			if tc.createTarget {
				require.NoError(t, os.Mkdir(targetPath, 0755), "Mkdir successfully created directory")
			}

			cfg, err := genutiltest.CreateDefaultCometConfig(home)
			require.NoError(t, err, "Created default tendermint config")

			logger := log.NewNopLogger()
			serverCtx := server.NewContext(viper.New(), cfg, logger)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home)

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			cmd := provenancecmd.GetDocGenCmd()
			args := append([]string{targetPath}, tc.flags...)
			cmd.SetArgs(args)
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			if strings.Contains(tc.err, "%s") {
				tc.err = fmt.Sprintf(tc.err, targetPath)
			}

			err = cmd.ExecuteContext(ctx)
			assertions.AssertErrorValue(t, err, tc.err, "cmd %q %q error", cmd.Name(), args)

			if len(tc.err) > 0 {
				files, _ := os.ReadDir(targetPath)
				assert.Equal(t, 0, len(files), "should not generate files when failed")
			} else {
				files, err := os.ReadDir(targetPath)
				require.NoError(t, err, "ReadDir should not return an error")
				require.NotEmpty(t, files, "should generate files when successful")

				filenames := make([]string, len(files))
				exts := make([]string, len(files))
				for i, file := range files {
					filenames[i] = file.Name()
					exts[i] = filepath.Ext(filenames[i])
				}

				missingExt := false
				for _, extension := range tc.extensions {
					contains := false
					for _, ext := range exts {
						if ext == extension {
							contains = true
							break
						}
					}
					if !assert.True(t, contains, "should generate a file with the extension %q", extension) {
						missingExt = true
					}
				}
				if missingExt {
					t.Logf("Files in %s\n%s", targetPath, strings.Join(filenames, "\n"))
				}
			}

			if _, err = os.Stat(targetPath); err != nil {
				require.NoError(t, os.RemoveAll(targetPath), "RemoveAll should be able to remove the temporary target directory")
			}
		})
	}
}
