package byidclient

import (
	"context"
)

type Client interface {
	GetIdentity(ctx context.Context, id string) error
	ListIdentities(ctx context.Context) error
	ListGroups(ctx context.Context) error
}
