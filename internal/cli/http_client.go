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
	"fmt"
	"net/http"
	"runtime"
	"sync"
)

// HTTPClient is the HTTP client used by the CLI.
var HTTPClient = &http.Client{
	Transport: &httpTransport{},
}

type httpTransport struct {
	init      sync.Once
	next      http.RoundTripper
	userAgent string
}

// RoundTrip implements http.RoundTripper.
func (t *httpTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.init.Do(func() {
		t.next = http.DefaultTransport
		t.userAgent = fmt.Sprintf("nametag-cli/%s; %s-%s", Version,
			runtime.GOOS, runtime.GOARCH)
	})
	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", t.userAgent)
	}
	return t.next.RoundTrip(r)
}
