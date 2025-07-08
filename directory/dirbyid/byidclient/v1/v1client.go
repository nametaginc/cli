package v1

type V1Client struct {
	BaseURL  string
	TenantID string
	RealmID  string
}

func NewV1Client(clientID, clientSecret, baseURL, tenantID, realmID string) (*V1Client, error) {
	return &V1Client{
		BaseURL:  baseURL,
		TenantID: tenantID,
		RealmID:  realmID,
	}, nil
}
