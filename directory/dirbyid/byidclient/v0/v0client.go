package v0

type V0Client struct {
	BaseURL string
}

func NewV0Client(clientID, clientSecret, baseURL string) (*V0Client, error) {
	return &V0Client{
		BaseURL: baseURL,
	}, nil
}
