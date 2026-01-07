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

// Package dirldap implements an LDAP agent
package dirldap

import (
	"context"
	"fmt"
	"log"

	"github.com/go-ldap/ldap/v3"
	"github.com/samber/lo"

	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/internal/config"
)

const timeFormat = "20060102150405Z"

// Provider represents the directory provider
type Provider struct {
	_client Client
	Config  *config.LDAPConfig
}

// Client defines an interface for client operations.
type Client interface {
	Bind(username, password string) error
	Search(request *ldap.SearchRequest) (*ldap.SearchResult, error)
	Modify(request *ldap.ModifyRequest) error
	Close() error
}

// LDAPClient defines an interface for LDAP operations.
type LDAPClient struct {
	conn *ldap.Conn
}

// Bind creates a connection with the provided credentials.
func (r *LDAPClient) Bind(username, password string) error {
	return r.conn.Bind(username, password)
}

// Search runs the ldap search request and returns the result
func (r *LDAPClient) Search(request *ldap.SearchRequest) (*ldap.SearchResult, error) {
	return r.conn.Search(request)
}

// Modify applies the ldap modify request
func (r *LDAPClient) Modify(request *ldap.ModifyRequest) error {
	return r.conn.Modify(request)
}

// Close cleans up ant client related connections
func (r *LDAPClient) Close() error {
	return r.conn.Close()
}

func (p *Provider) client() (Client, error) {
	if p._client == nil {
		// Connect to LDAP server
		client, err := ldap.DialURL(p.Config.LDAPUrl)
		if err != nil {
			return nil, err
		}

		// Bind with credentials
		err = client.Bind(p.Config.BindDN, p.Config.BindPassword)
		if err != nil {
			return nil, err
		}

		p._client = &LDAPClient{
			conn: client,
		}
	}
	return p._client, nil
}

// Close cleans up resources associated with the Provider
func (p *Provider) Close() error {
	return p._client.Close()
}

// Configure returns static information about the integration
func (p *Provider) Configure(ctx context.Context, req diragentapi.DirAgentConfigureRequest) (*diragentapi.DirAgentConfigureResponse, error) {
	client, err := p.client()
	if err != nil {
		return nil, err
	}

	var name string
	// If the config specifies a base dn to use, we use that. Or else we assume the root
	if p.Config.BaseDN != "" {
		name = p.Config.BaseDN
	} else {
		// Search for Root DSE
		searchRequest := ldap.NewSearchRequest(
			"", // Empty DN for Root DSE
			ldap.ScopeBaseObject,
			ldap.NeverDerefAliases,
			0,
			0,
			false,
			"(objectClass=*)",
			[]string{"namingContexts"},
			nil,
		)

		result, err := client.Search(searchRequest)
		if err != nil {
			return nil, err
		}

		if len(result.Entries) > 0 {
			entry := result.Entries[0]
			name = entry.GetAttributeValue("namingContexts")
		} else {
			return nil, fmt.Errorf("unable to find Root DSE")
		}

		log.Printf("Found: %s", name)
		p.Config.BaseDN = name
	}

	return &diragentapi.DirAgentConfigureResponse{
		Traits: diragentapi.DirAgentTraits{
			Name:                    name,
			CanGetTemporaryPassword: lo.ToPtr(true),
			CanUnlock:               lo.ToPtr(true),
			CanUpdateAccountsList:   lo.ToPtr(true),
		},
	}, nil
}
