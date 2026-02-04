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

// Package dirauthentik implements an Authentik directory provider.
package dirauthentik

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory"
)

const (
	defaultPageSize = 250
	requestTimeout  = 120 * time.Second
)

// Provider represents the configuration details required to connect to Authentik.
type Provider struct {
	URL        string
	Token      string
	HTTPClient *http.Client
}

// Configure returns static information about the integration.
func (p *Provider) Configure(ctx context.Context, req diragentapi.DirAgentConfigureRequest) (*diragentapi.DirAgentConfigureResponse, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}

	return &diragentapi.DirAgentConfigureResponse{
		Traits: diragentapi.DirAgentTraits{
			Name:                  p.displayName(),
			CanGetPasswordLink:    lo.ToPtr(true),
			CanRemoveAllMFA:       lo.ToPtr(true),
			CanUpdateAccountsList: lo.ToPtr(true),
		},
		ImmutableID: fmt.Sprintf("urn:agent:authentik:%s", p.URL),
	}, nil
}

func (p *Provider) validate() error {
	if strings.TrimSpace(p.URL) == "" {
		return directory.CodedError{
			Code:    diragentapi.ConfigurationError,
			Message: "authentik URL is required",
		}
	}
	if strings.TrimSpace(p.Token) == "" {
		return directory.CodedError{
			Code:    diragentapi.ConfigurationError,
			Message: "authentik token is required",
		}
	}
	_, err := p.apiBaseURL()
	if err != nil {
		return directory.CodedError{
			Code:    diragentapi.ConfigurationError,
			Message: err.Error(),
		}
	}
	return nil
}

func (p *Provider) apiBaseURL() (*url.URL, error) {
	if strings.TrimSpace(p.URL) == "" {
		return nil, fmt.Errorf("authentik URL is required")
	}

	base, err := url.Parse(p.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid authentik URL: %w", err)
	}
	if base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("invalid authentik URL: %s", p.URL)
	}

	path := strings.TrimSuffix(base.Path, "/")
	if !strings.HasSuffix(path, "/api/v3") {
		path += "/api/v3"
	}
	base.Path = path + "/"
	base.RawQuery = ""
	base.Fragment = ""
	return base, nil
}

func (p *Provider) displayName() string {
	base, err := p.apiBaseURL()
	if err != nil {
		return "authentik"
	}
	host := base.Hostname()
	if host == "" {
		return "authentik"
	}
	return fmt.Sprintf("authentik (%s)", host)
}

func (p *Provider) client() (*http.Client, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	if p.HTTPClient == nil {
		p.HTTPClient = &http.Client{Timeout: requestTimeout}
	}
	return p.HTTPClient, nil
}

func (p *Provider) doJSON(ctx context.Context, method string, path string, query url.Values, payload any, out any) error {
	client, err := p.client()
	if err != nil {
		return err
	}
	base, err := p.apiBaseURL()
	if err != nil {
		return err
	}

	endpoint := strings.TrimPrefix(path, "/")
	urlRef := &url.URL{Path: endpoint}
	endpointURL := base.ResolveReference(urlRef)
	if query != nil {
		endpointURL.RawQuery = query.Encode()
	}

	var body io.Reader
	if payload != nil {
		buf := &bytes.Buffer{}
		if err := json.NewEncoder(buf).Encode(payload); err != nil {
			return fmt.Errorf("authentik: encode request: %w", err)
		}
		body = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, endpointURL.String(), body)
	if err != nil {
		return err
	}
	token := strings.TrimSpace(p.Token)
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = strings.TrimSpace(token[len("bearer "):])
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return p.parseError(resp)
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("authentik: decode response: %w", err)
	}
	return nil
}

func (p *Provider) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	message := strings.TrimSpace(string(body))

	if len(body) > 0 {
		var apiErr struct {
			Detail string `json:"detail"`
			Code   string `json:"code"`
		}
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Detail != "" {
			message = apiErr.Detail
		}
	}

	if message == "" {
		message = resp.Status
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return directory.CodedError{
			Code:    diragentapi.ServiceAuthenticationFailed,
			Message: message,
		}
	case http.StatusForbidden:
		return directory.CodedError{
			Code:    diragentapi.PermissionDenied,
			Message: message,
		}
	case http.StatusNotFound:
		return directory.CodedError{
			Code:    diragentapi.AccountNotFound,
			Message: message,
		}
	default:
		return fmt.Errorf("authentik: %s", message)
	}
}

func (p *Provider) fetchUsers(ctx context.Context, query url.Values) ([]apiUser, error) {
	page := 1
	users := []apiUser{}
	for {
		query.Set("page", strconv.Itoa(page))

		var resp userListResponse
		if err := p.doJSON(ctx, http.MethodGet, "core/users/", query, nil, &resp); err != nil {
			return nil, err
		}
		users = append(users, resp.Results...)

		if resp.Pagination.Next == nil || *resp.Pagination.Next <= 0 {
			break
		}
		page = *resp.Pagination.Next
	}
	return users, nil
}

func (p *Provider) fetchGroups(ctx context.Context, query url.Values) (*groupListResponse, error) {
	var resp groupListResponse
	if err := p.doJSON(ctx, http.MethodGet, "core/groups/", query, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *Provider) fetchUserByPK(ctx context.Context, pk string) (*apiUser, error) {
	var resp apiUser
	if err := p.doJSON(ctx, http.MethodGet, fmt.Sprintf("core/users/%s/", pk), nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
