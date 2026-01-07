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
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/samber/lo"

	"github.com/nametaginc/cli/internal/pkg/lox"
	"github.com/nametaginc/cli/internal/pkg/must"
)

// Cached invokes generator, a function that produces outputs from inputGlobs,
// but only if the inputs and outputs have changed since the last invocation.
func Cached(
	inputGlobs []string,
	outputs []string,
	generator func() error,
) error {
	var inputPaths []string
	for _, inputGlob := range inputGlobs {
		v, err := filepath.Glob(inputGlob)
		if err != nil {
			return err
		}
		inputPaths = append(inputPaths, v...)
	}
	inputPaths = lo.Uniq(inputPaths)
	sort.Strings(inputPaths)

	for _, outputPath := range outputs {
		if lo.Contains(inputPaths, outputPath) {
			inputPaths = lo.Without(inputPaths, outputPath)
		}
	}

	// inputHash represents the state of all the input files
	inputHash := sha256.New()
	for _, inputPath := range inputPaths {
		fmt.Fprintln(inputHash, "input")
		fmt.Fprintln(inputHash, inputPath)
		inputFile, err := os.Open(inputPath) //nolint:gosec  // we control this path
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(inputHash, inputFile)
		_ = inputFile.Close()
		if copyErr != nil {
			return copyErr
		}
	}

	// outputHash represents the state of all the output files
	outputFilesMissing := false
	outputHash := sha256.New()
	for _, outputPath := range outputs {
		fmt.Fprintln(outputHash, "output")
		fmt.Fprintln(outputHash, outputPath)

		outputFile, err := os.Open(outputPath) //nolint:gosec  // we control this path
		if os.IsNotExist(err) {
			fmt.Printf("generate: output not present: %s\n", outputPath)
			outputFilesMissing = true
			break
		}
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(outputHash, outputFile)
		_ = outputFile.Close()
		if copyErr != nil {
			return copyErr
		}
	}

	sourceRoot, err := SourceRoot()
	if err != nil {
		return err
	}

	cacheRoot := os.Getenv("GENX_CACHE_ROOT")
	if cacheRoot == "" {
		cacheRoot = filepath.Join(must.Return(os.UserCacheDir()), "genx")
	}
	cachePath := filepath.Join(cacheRoot, fmt.Sprintf("%x", inputHash.Sum(nil)))

	if !outputFilesMissing {
		cacheContents, err := os.ReadFile(cachePath) //nolint:gosec  // we control this path
		if os.IsNotExist(err) {
			fmt.Printf("generate: input not cached: [%s]\n",
				lox.Elide(strings.Join(inputPaths, ", "), 128))
			// ok
		} else if err != nil {
			return err
		} else if string(cacheContents) == fmt.Sprintf("%x", outputHash.Sum(nil)) {
			fmt.Printf("generate: cache hit\n")
			return nil
		}
	}

	if err := generator(); err != nil {
		return err
	}

	outputHash = sha256.New()
	for _, outputPath := range outputs {
		fmt.Println(must.Return(filepath.Rel(sourceRoot, must.Return(filepath.Abs(outputPath)))))

		fmt.Fprintln(outputHash, "output")
		fmt.Fprintln(outputHash, outputPath)

		outputFile, err := os.Open(outputPath) //nolint:gosec  // we control this path
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(outputHash, outputFile)
		_ = outputFile.Close()
		if copyErr != nil {
			return copyErr
		}
	}

	_ = os.MkdirAll(filepath.Dir(cachePath), 0755)
	if err := os.WriteFile(cachePath,
		[]byte(fmt.Sprintf("%x", outputHash.Sum(nil))),
		0644); err != nil {
		return err
	}

	return nil
}
