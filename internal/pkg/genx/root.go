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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SourceRoot returns the source root directory by looking for .git
func SourceRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		// Note: if you use os.Stat() here, it causes the go test cache to be busted
		// by any change to .git, e.g. when the commit changes. We don't want this
		// because it causes lots of otherwise cached tests to be re-run.
		//
		// os.Readlink returns not found when the path doesn't exist and EINVAL when
		// it does (because .git is not a symlink). It does this all without depending
		// on the modification time of any files in .git.
		_, err := os.Readlink(filepath.Join(dir, ".git"))
		if !os.IsNotExist(err) {
			return dir, nil
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			return "", os.ErrNotExist
		}
		dir = parentDir
		continue
	}
}

// GitCommonDir returns the common git directory for the repository at sourceRoot.
// This handles git worktrees where .git is a file instead of a directory.
func GitCommonDir(sourceRoot string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	cmd.Dir = sourceRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git common dir: %w", err)
	}
	gitDir := strings.TrimSpace(string(output))
	// If the path is relative, make it absolute relative to sourceRoot
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(sourceRoot, gitDir)
	}
	return gitDir, nil
}
