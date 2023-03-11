//go:generate go run github.com/Khan/genqlient

package accountgqlclient

import (
	"context"

	_ "github.com/Khan/genqlient/generate"
	"github.com/Khan/genqlient/graphql"
)

type Client struct {
	client graphql.Client
}

func New(endpoint string, httpClient graphql.Doer) *Client {
	return &Client{client: graphql.NewClient(endpoint, httpClient)}
}

func (c *Client) Me(ctx context.Context) (*MeResponse, error) {
	return Me(ctx, c.client)
}
