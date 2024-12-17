// Copyright 2024 Nametag Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cli contains implementation of the Nametag cli subcommands
package cli

import (
	_ "embed"
	"io"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed "VERSION"
var versionBuf []byte

// Version is the version of the command line tool.
var Version = strings.TrimSpace(string(versionBuf))

// Log is the logger used by the CLI
var Log = log.New(io.Discard, "", 0)

// New creates a new root command for the Nametag CLI
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "nametag",
		Short:         "Nametag command line interface",
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cachedConfig = nil // for testing

			Log.SetOutput(cmd.OutOrStdout())
			return nil
		},
	}
	cmd.CompletionOptions.HiddenDefaultCmd = true
	cmd.SetUsageTemplate(usageTemplate)
	cmd.SetHelpTemplate(helpTemplate)

	cmd.PersistentFlags().StringP("auth-token", "t", "", "Nametag API authentication token")
	cmd.PersistentFlags().StringP("config", "c", "", "Path to Nametag CLI configuration file")

	cmd.AddCommand(newAuthCmd())
	cmd.AddCommand(newDirCmd())
	cmd.AddCommand(newEnvCmd())
	cmd.AddCommand(newSelfServiceCmd())

	return cmd
}
