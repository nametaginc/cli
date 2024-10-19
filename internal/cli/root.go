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
	"strings"

	"github.com/spf13/cobra"
)

//go:embed "VERSION"
var versionBuf []byte

// Version is the version of the command line tool.
var Version = strings.TrimSpace(string(versionBuf))

// Root is the main command.
var Root = &cobra.Command{
	Use:           "nametag",
	Short:         "Nametag command line interface",
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func subcmd(parent *cobra.Command, child *cobra.Command) *cobra.Command {
	parent.AddCommand(child)
	return child
}

func init() {
	Root.CompletionOptions.HiddenDefaultCmd = true
}
