// Copyright 2026 Nametag Inc.
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
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

const directoryHTTPHeadersEnvVar = "NAMETAG_DIRECTORY_HTTP_HEADERS"

func addDirectoryHTTPHeaderFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArray(
		"directory-http-header",
		splitCommaSeparatedEnv(os.Getenv(directoryHTTPHeadersEnvVar)),
		"Additional HTTP header to inject into outbound directory API requests (repeat flag or comma-separated, $NAMETAG_DIRECTORY_HTTP_HEADERS)",
	)
}

func getDirectoryHTTPHeaders(cmd *cobra.Command) (http.Header, error) {
	pairs, err := cmd.Flags().GetStringArray("directory-http-header")
	if err != nil {
		return nil, err
	}
	return parseDirectoryHTTPHeaders(pairs)
}

func directoryHTTPHeaderWorkerEnv(cmd *cobra.Command) (map[string]string, error) {
	headers, err := getDirectoryHTTPHeaders(cmd)
	if err != nil {
		return nil, err
	}
	if len(headers) == 0 {
		return nil, nil
	}
	return map[string]string{
		directoryHTTPHeadersEnvVar: serializeDirectoryHTTPHeaders(headers),
	}, nil
}

func parseDirectoryHTTPHeaders(values []string) (http.Header, error) {
	headers := http.Header{}
	for _, pair := range values {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid directory HTTP header %q: expected key=value", pair)
		}

		key := http.CanonicalHeaderKey(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		if key == "" || value == "" {
			return nil, fmt.Errorf("invalid directory HTTP header %q: key and value must be non-empty", pair)
		}

		headers.Add(key, value)
	}
	return headers, nil
}

func serializeDirectoryHTTPHeaders(headers http.Header) string {
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := []string{}
	for _, key := range keys {
		for _, value := range headers.Values(key) {
			pairs = append(pairs, key+"="+value)
		}
	}
	return strings.Join(pairs, ",")
}
