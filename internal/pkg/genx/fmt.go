// Copyright 2024 Nametag Inc.
//
// All information contained herein is the property of Nametag Inc.. The
// intellectual and technical concepts contained herein are proprietary, trade
// secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
// and Foreign Patents, patents in process, and are protected by trade secret or
// copyright law. Reproduction or distribution, in whole or in part, is
// forbidden except by express written permission of Nametag, Inc.

package genx

import (
	"os/exec"
	"path/filepath"

	"github.com/nametaginc/cli/internal/pkg/lox"
)

// AddLicense adds a license to the given paths.
func AddLicense(paths []string) error {
	sourceRoot, err := SourceRoot()
	if err != nil {
		return err
	}
	args := []string{
		"go", "tool", "addlicense",
		"-c", "Nametag Inc.",
		"-f", filepath.Join(sourceRoot, "hack/license.tmpl"),
	}
	args = append(args, paths...)
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec  // we the arguments
	return cmd.Run()
}

// FormatGo formats go source code according to our rules
func FormatGo(paths []string) error {
	if err := AddLicense(paths); err != nil {
		return err
	}
	if err := exec.Command("gofmt", append([]string{"-s", "-w"}, paths...)...).Run(); err != nil { //nolint:gosec  // we the arguments
		return err
	}
	if err := exec.Command("go", append([]string{"tool", //nolint:gosec  // we the arguments
		"goimports",
		"-local", "github.com/nametaginc/nt",
		"-w"}, paths...)...).Run(); err != nil {
		return err
	}
	if err := exec.Command("go", append([]string{"tool", //nolint:gosec  // we the arguments
		"gogroup",
		"-order", "std,other,prefix=github.com/nametaginc/nt",
		"-rewrite"}, paths...)...).Run(); err != nil {
		return err
	}
	return nil
}

// Prettier runs prettier on the given paths.
func Prettier(paths []string) error {
	absPaths := lox.Map(paths, func(p string) string {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return p
		}
		return absPath
	})

	sourceRoot, err := SourceRoot()
	if err != nil {
		return err
	}

	args := []string{
		"go", "run", "./pkg/prettier",
		"--config", filepath.Join(sourceRoot, "ntmake/fmt/prettier.json"),
		"--write", "-l",
	}
	args = append(args, absPaths...)
	cmd := exec.Command(args[0], args[1:]...) //nolint:gosec  // we the arguments
	cmd.Dir = sourceRoot

	return cmd.Run()
}
