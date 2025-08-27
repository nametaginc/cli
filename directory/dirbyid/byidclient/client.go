package byidclient

import (
	"context"
	"time"
)

// Client is the interface for the Beyond Identity API.
type Client interface {
	GetIdentity(ctx context.Context, immutableID string) (*Identity, error)
	ListIdentities(ctx context.Context, filter, pageToken *string) (*ListIdentitiesResponse, error)
	ListGroups(ctx context.Context, pageToken *string) (*ListGroupsResponse, error)
	ListIdentityGroups(ctx context.Context, id string, pageToken *string) (*ListGroupsResponse, error)
}

// Identity represents an identity in Beyond Identity.
type Identity struct {
	ID           string
	DisplayName  string
	Username     string
	EmailAddress string
	UpdateTime   *time.Time
}

type ListIdentitiesResponse struct {
	Identities    []*Identity
	NextPageToken *string
}

// Group represents a group in Beyond Identity.
type Group struct {
	ID          string
	DisplayName string
	Type        string
}

type ListGroupsResponse struct {
	Groups        []*Group
	NextPageToken *string
}
