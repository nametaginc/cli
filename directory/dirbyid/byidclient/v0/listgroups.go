package v0

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/nametaginc/cli/directory/dirbyid/byidclient"
)

// ListGroupsResponse is the response from the Beyond Identity API.
type ListGroupsResponse struct {
	Groups    []*Group `json:"groups"`
	TotalSize int      `json:"total_size"`
}

// Group is the response from the Beyond Identity API.
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListGroups returns the groups with the given page token.
// https://api.byndid.com/v2/groups
// https://docs.beyondidentity.com/api/v0#tag/Groups/operation/ListGroups.
func (c *V0Client) ListGroups(ctx context.Context, pageToken *string) (*byidclient.ListGroupsResponse, error) {
	q := url.Values{}
	if pageToken != nil {
		q.Add("skip", *pageToken)
	}

	listURL := *c.baseURL
	listURL.Path = path.Join("v2", "groups")
	listURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL.String(), nil)
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

	var raw ListGroupsResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	groups := make([]*byidclient.Group, len(raw.Groups))
	for i, group := range raw.Groups {
		groups[i] = &byidclient.Group{
			ID:          group.ID,
			DisplayName: group.Name,
			Type:        "group",
		}
	}

	// If the total size is greater than the number of identities, there are more identities to fetch.
	// We return the number of identities fetched so far as the next page token.
	// This will be piped into the `skip` parameter of the next request.
	var nextPageToken *string
	if raw.TotalSize > len(groups) {
		token := strconv.Itoa(len(groups))
		nextPageToken = &token
	}

	return &byidclient.ListGroupsResponse{
		Groups:        groups,
		NextPageToken: nextPageToken,
	}, nil
}
