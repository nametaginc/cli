// Copyright 2024 Nametag Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/nametaginc/cli/internal/config"
)

const defaultServer = "https://nametag.co"

func getServer(cmd *cobra.Command) string {
	if s := os.Getenv("NAMETAG_SERVER"); s != "" {
		return s
	}

	config, err := config.ReadConfig(cmd)
	if err == nil && config.Server != "" {
		return config.Server
	}

	return defaultServer
}
