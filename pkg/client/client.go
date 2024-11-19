package client

import (
	openfga "github.com/openfga/go-sdk/client"
)

const LocalApiURL = "http://host.docker.internal:8080"

// Client ...
type Client struct {
	fga *openfga.OpenFgaClient
}

// NewClient ...
func NewClient(apiURL string) (*Client, error) {
	cfg := &openfga.ClientConfiguration{
		ApiUrl: apiURL,
	}

	fga, err := openfga.NewSdkClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		fga: fga,
	}, nil
}
