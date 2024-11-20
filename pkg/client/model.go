package client

import (
	"context"
	"encoding/json"

	openfga "github.com/openfga/go-sdk/client"
	"github.com/openfga/language/pkg/go/transformer"
	"github.com/zeiss/pkg/cast"
)

// AuthorizationModel ...
type AuthorizationModel struct {
	ID   string `json:"id,omitempty"`
	Spec string `json:"spec,omitempty"`
}

// CreateModel ...
func (c *Client) CreateModel(ctx context.Context, id, spec string) (*AuthorizationModel, error) {
	s, err := transformer.TransformDSLToJSON(spec)
	if err != nil {
		return nil, err
	}

	var body openfga.ClientWriteAuthorizationModelRequest
	if err := json.Unmarshal([]byte(s), &body); err != nil {
		return nil, err
	}

	resp, err := c.fga.WriteAuthorizationModel(ctx).Options(openfga.ClientWriteAuthorizationModelOptions{StoreId: cast.Ptr(id)}).Body(body).Execute()
	if err != nil {
		return nil, err
	}

	model := AuthorizationModel{
		ID: resp.AuthorizationModelId,
	}

	return cast.Ptr(model), nil
}

// UpdateModel ...
func (c *Client) UpdateModel(ctx context.Context, id, spec string) (*AuthorizationModel, error) {
	s, err := transformer.TransformDSLToJSON(spec)
	if err != nil {
		return nil, err
	}

	var body openfga.ClientWriteAuthorizationModelRequest
	if err := json.Unmarshal([]byte(s), &body); err != nil {
		return nil, err
	}

	resp, err := c.fga.WriteAuthorizationModel(ctx).Options(openfga.ClientWriteAuthorizationModelOptions{StoreId: cast.Ptr(id)}).Body(body).Execute()
	if err != nil {
		return nil, err
	}

	model := AuthorizationModel{
		ID: resp.AuthorizationModelId,
	}

	return cast.Ptr(model), nil
}

// GetAuthorizationModel ...
func (c *Client) GetAuthorizationModel(ctx context.Context, store, model string) (*AuthorizationModel, error) {
	resp, err := c.fga.ReadAuthorizationModel(ctx).Options(openfga.ClientReadAuthorizationModelOptions{StoreId: cast.Ptr(store), AuthorizationModelId: cast.Ptr(model)}).Execute()
	if err != nil {
		return nil, err
	}

	authModel := AuthorizationModel{
		ID: resp.AuthorizationModel.GetId(),
	}

	return cast.Ptr(authModel), nil
}

// DeleteAuthorizationModel ...
func (c *Client) DeleteAuthorizationModel(ctx context.Context, id string) error {
	return nil
}
