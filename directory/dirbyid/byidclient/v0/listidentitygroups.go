package v0

import (
	"context"

	"github.com/nametaginc/cli/directory/dirbyid/byidclient"
)

// ListIdentityGroups returns the groups for the given identity.
// TODO: We need to implement this on the Beyond Identity side.
func (c *V0Client) ListIdentityGroups(ctx context.Context, id string) (*byidclient.ListGroupsResponse, error) {
	return nil, nil
}
