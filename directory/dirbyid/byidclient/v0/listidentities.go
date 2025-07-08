package v0

import (
	"context"

	"github.com/nametaginc/cli/directory/dirbyid/byidclient"
)

func (c *V0Client) ListIdentities(ctx context.Context, filter, pageToken *string) (*byidclient.ListIdentitiesResponse, error) {
	return nil, nil
}
