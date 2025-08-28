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

// ListIdentityGroups returns the groups for the given identity.
func (c *V0Client) ListIdentityGroups(ctx context.Context, id string, pageToken *string) (*byidclient.ListGroupsResponse, error) {
	q := url.Values{}
	if pageToken != nil {
		q.Add("skip", *pageToken)
	}
	q.Add("page_size", "100")

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

	// If the total size is greater than the number of groups, there are more groups to fetch.
	// We return the number of groups fetched so far as the next page token.
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
