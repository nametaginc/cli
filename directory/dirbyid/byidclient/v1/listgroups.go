package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/nametaginc/cli/directory/dirbyid/byidclient"
)

// ListGroupsResponse is the response from the Beyond Identity API.
type ListGroupsResponse struct {
	Groups        []*Group `json:"groups"`
	NextPageToken *string  `json:"next_page_token"`
}

// Group is the response from the Beyond Identity API.
type Group struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// ListGroups returns the groups with the given page token.
// https://api-us.beyondidentity.com/v1/tenants/{tenant_id}/realms/{realm_id}/groups
// https://docs.beyondidentity.com/api/v1#tag/Groups/operation/ListGroups.
func (c *V1Client) ListGroups(ctx context.Context, pageToken *string) (*byidclient.ListGroupsResponse, error) {
	query := url.Values{}
	if pageToken != nil {
		query.Add("page_token", *pageToken)
	}

	listURL := *c.baseURL
	listURL.Path = path.Join("v1", "tenants", c.tenantID, "realms", c.realmID, "groups")
	listURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL.String(), nil)
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

	var raw ListGroupsResponse
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
