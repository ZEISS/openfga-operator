package client

import (
	"context"

	openfga "github.com/openfga/go-sdk/client"
	"github.com/zeiss/pkg/cast"
)

// Store ...
type Store struct {
	// ID ...
	ID string `json:"id,omitempty"`
	// Name ...
	Name string `json:"name,omitempty"`
}

// CreateStore ...
func (c *Client) CreateStore(ctx context.Context, name string) (*Store, error) {
	resp, err := c.fga.CreateStore(ctx).Body(openfga.ClientCreateStoreRequest{Name: name}).Execute()
	if err != nil {
		return nil, err
	}

	store := Store{
		ID:   resp.GetId(),
		Name: resp.GetName(),
	}

	return cast.Ptr(store), nil
}

// GetStore ...
func (c *Client) GetStore(ctx context.Context, id string) (*Store, error) {
	resp, err := c.fga.GetStore(ctx).Options(openfga.ClientGetStoreOptions{StoreId: cast.Ptr(id)}).Execute()
	if err != nil {
		return nil, err
	}

	store := Store{
		ID:   resp.GetId(),
		Name: resp.GetName(),
	}

	return cast.Ptr(store), nil
}

// DeleteStore ...
func (c *Client) DeleteStore(ctx context.Context, id string) error {
	_, err := c.fga.DeleteStore(ctx).Options(openfga.ClientDeleteStoreOptions{StoreId: cast.Ptr(id)}).Execute()
	if err != nil {
		return err
	}

	return nil
}
