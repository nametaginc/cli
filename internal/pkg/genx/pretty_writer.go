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
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nametaginc/cli/internal/pkg/must"
)

// PrettyWriter is a writer than passes the contents through prettier
// before writing them to disk.
func PrettyWriter(path string) io.WriteCloser {
	return &prettyWriter{path: path}
}

type prettyWriter struct {
	path   string
	buffer bytes.Buffer
}

func (w *prettyWriter) Write(p []byte) (n int, err error) {
	return w.buffer.Write(p)
}

func (w *prettyWriter) Close() error {
	sourceRoot, err := SourceRoot()
	if err != nil {
		return err
	}
	cmd := exec.Command( //nolint:gosec  // we the arguments
		"go", "run", "./pkg/prettier",
		"--config", filepath.Join(must.Return(SourceRoot()), "ntmake/fmt/prettier.json"),
		"--stdin-filepath="+w.path,
	)
	cmd.Dir = sourceRoot
	cmd.Stdin = &w.buffer

	var formatted bytes.Buffer
	cmd.Stdout = &formatted

	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.MkdirAll(filepath.Dir(w.path), 0755)
		_ = os.WriteFile(w.path, w.buffer.Bytes(), 0644)
		return err
	}

	current, _ := os.ReadFile(w.path)
	if bytes.Equal(current, formatted.Bytes()) {
		return nil
	}

	return os.WriteFile(w.path, formatted.Bytes(), 0644)
}
