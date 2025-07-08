package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nametaginc/cli/directory/dirbyid/byidclient"
)

// ListIdentityGroupsResponse is the response from the Beyond Identity API.
type ListIdentityGroupsResponse struct {
	Groups        []*Group `json:"groups"`
	NextPageToken *string  `json:"next_page_token"`
}

func (c *V1Client) ListIdentityGroups(ctx context.Context, id string) (*byidclient.ListGroupsResponse, error) {
	baseURLStr := c.baseURL.String()
	joinedURL, err := url.JoinPath(baseURLStr, "identities", id, ":listGroups")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, joinedURL, nil)
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

	var raw ListIdentityGroupsResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	groups := make([]*byidclient.Group, len(raw.Groups))
	for i, group := range raw.Groups {
		groups[i] = &byidclient.Group{
			ID:          group.ID,
			DisplayName: group.DisplayName,
		}
	}
	return &byidclient.ListGroupsResponse{
		Groups:        groups,
		NextPageToken: raw.NextPageToken,
	}, nil
}
