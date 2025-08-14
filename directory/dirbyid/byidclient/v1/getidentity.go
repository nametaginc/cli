package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nametaginc/cli/directory/dirbyid/byidclient"
)

// Identity is the response from the Beyond Identity API.
type Identity struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Traits      Traits `json:"traits"`
}

// Traits is the traits of the identity.
type Traits struct {
	Username            string `json:"username"`
	PrimaryEmailAddress string `json:"primary_email_address"`
}

// GetIdentity returns the identity with the given ID.
// https://api-us.beyondidentity.com/v1/tenants/{tenant_id}/realms/{realm_id}/identities/{identity_id}
// https://docs.beyondidentity.com/api/v1#tag/Identities/operation/GetIdentity.
func (c *V1Client) GetIdentity(ctx context.Context, id string) (*byidclient.Identity, error) {
	identityURL, err := url.JoinPath(c.baseURL.String(), "v1", "tenants", c.tenantID, "realms", c.realmID, "identities", id)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, identityURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var raw Identity

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	return &byidclient.Identity{
		ID:           raw.ID,
		DisplayName:  raw.DisplayName,
		Username:     raw.Traits.Username,
		EmailAddress: raw.Traits.PrimaryEmailAddress,
	}, nil
}
