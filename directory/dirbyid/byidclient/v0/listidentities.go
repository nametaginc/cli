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

type ListIdentitiesResponse struct {
	Identities []*Identity `json:"users"`
	TotalSize  int         `json:"total_size"`
}

// ListIdentities returns the identities with the given page token.
// https://api.byndid.com/v2/users
// https://docs.beyondidentity.com/api/v0#tag/Users/operation/ListUsers.
func (c *V0Client) ListIdentities(ctx context.Context, filter, pageToken *string) (*byidclient.ListIdentitiesResponse, error) {
	q := url.Values{}
	if filter != nil {
		q.Add("filter", *filter)
	}
	// For v0, there is no page token. Instead, we use the skip parameter to paginate.
	if pageToken != nil {
		q.Add("skip", *pageToken)
	}
	q.Add("page_size", "100")

	listURL := *c.baseURL
	listURL.Path = path.Join("v2", "users")
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

	var raw ListIdentitiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	identities := make([]*byidclient.Identity, len(raw.Identities))
	for i, identity := range raw.Identities {
		identities[i] = &byidclient.Identity{
			ID:           identity.ID,
			DisplayName:  identity.DisplayName,
			Username:     identity.Username,
			EmailAddress: identity.EmailAddress,
			UpdateTime:   identity.UpdateTime,
		}
	}

	// If the total size is greater than the number of identities, there are more identities to fetch.
	// We return the number of identities fetched so far as the next page token.
	// This will be piped into the `skip` parameter of the next request.
	var nextPageToken *string
	if raw.TotalSize > len(identities) {
		token := strconv.Itoa(len(identities))
		nextPageToken = &token
	}

	return &byidclient.ListIdentitiesResponse{
		Identities:    identities,
		NextPageToken: nextPageToken,
	}, nil
}
