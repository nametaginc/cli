package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nametaginc/cli/directory/dirbyid/byidclient"
)

// ListIdentitiesResponse is the response from the Beyond Identity API.
type ListIdentitiesResponse struct {
	Identities    []*Identity `json:"identities"`
	TotalSize     int         `json:"total_size"`
	NextPageToken *string     `json:"next_page_token"`
}

// ListIdentities returns the identities with the given page token.
// https://docs.beyondidentity.com/api/v1#tag/Identities/operation/ListIdentities.
func (c *V1Client) ListIdentities(ctx context.Context, filter, pageToken *string) (*byidclient.ListIdentitiesResponse, error) {
	q := url.Values{}
	if filter != nil {
		q.Add("filter", *filter)
	}
	if pageToken != nil {
		q.Add("page_token", *pageToken)
	}

	baseURLStr := c.baseURL.String()
	joinedURL, err := url.JoinPath(baseURLStr, "identities")
	if err != nil {
		return nil, err
	}

	fullURL := joinedURL + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
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

	var raw ListIdentitiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	identities := make([]*byidclient.Identity, len(raw.Identities))
	for i, identity := range raw.Identities {
		identities[i] = &byidclient.Identity{
			ID:          identity.ID,
			DisplayName: identity.DisplayName,
			Username:    identity.Traits.Username,
		}
	}
	return &byidclient.ListIdentitiesResponse{
		Identities:    identities,
		NextPageToken: raw.NextPageToken,
	}, nil
}
