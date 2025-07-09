package v0

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nametaginc/cli/directory/dirbyid/byidclient"
)

type Identity struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
}

// GetIdentity returns the identity with the given ID.
// https://api.byndid.com/v2/users/{user_id}
// https://docs.beyondidentity.com/api/v0#tag/Users/operation/GetUser.
func (c *V0Client) GetIdentity(ctx context.Context, id string) (*byidclient.Identity, error) {
	joinedURL, err := url.JoinPath(c.baseURL.String(), "v2", "users", id)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, joinedURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
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
		ID:          raw.ID,
		DisplayName: raw.DisplayName,
		Username:    raw.Username,
	}, nil
}
