package dirbyid

import (
	"github.com/nametaginc/cli/diragentapi"
	"github.com/nametaginc/cli/directory/dirbyid/byidclient"
)

func toDirAgentAccount(identity byidclient.Identity) diragentapi.DirAgentAccount {
	return diragentapi.DirAgentAccount{
		ImmutableID: identity.ID,
		IDs:         []string{identity.Username}, // TODO: What happens if we put email address here too?
		Name:        identity.DisplayName,
	}
}

func toDirAgentGroup(group byidclient.Group) diragentapi.DirAgentGroup {
	return diragentapi.DirAgentGroup{
		ImmutableID: group.ID,
		Name:        group.DisplayName,
		Kind:        group.Type,
	}
}
