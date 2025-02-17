package spotify

import "github.com/zmb3/spotify"

type Client struct {
	spotify.Client
}

func NewClient(client *spotify.Client) *Client {
	return &Client{
		Client: *client,
	}
}
