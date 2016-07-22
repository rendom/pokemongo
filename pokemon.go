package pokemongo

type Client struct {
	token string
}

func New() *Client {
	p := &Client{}
	return p
}

// Authenticate to PTC
func (api *Client) Authenticate(username string, password string) error {
	var err error
	api.token, err = api.Login(username, password)
	return err
}
