// Copyright 2025 Nametag Inc.
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

//go:build generate
// +build generate

package main

import (
	"os"
	"os/exec"

	"github.com/nametaginc/cli/internal/pkg/genx"
	"github.com/nametaginc/cli/internal/pkg/must"
)

//go:generate go run generate.go

func main() {
	defer genx.Log()()

	must.NotFail(genx.Cached(
		[]string{"diragent.yaml", "config.yaml", "generate.go"},
		[]string{"api.gen.go"},
		func() error {
			cmd := exec.Command("go", "run",
				"github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen",
				"--config=config.yaml",
				"diragent.yaml",
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}))
}
