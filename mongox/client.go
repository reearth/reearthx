package mongox

import (
	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	Client *mongo.Database
}

func NewClient(database string, c *mongo.Client) *Client {
	return &Client{Client: c.Database(database)}
}

func NewClientWithDatabase(c *mongo.Database) *Client {
	return &Client{Client: c}
}

func (c *Client) WithCollection(col string) *Collection {
	return NewCollection(c.Client.Collection(col))
}

func (c *Client) BeginTransaction() (usecasex.Tx, error) {
	s, err := c.Client.Client().StartSession()
	if err != nil {
		return nil, err
	}

	if err := s.StartTransaction(options.Transaction()); err != nil {
		return nil, err
	}

	return &Tx{session: s, commit: false}, nil
}
