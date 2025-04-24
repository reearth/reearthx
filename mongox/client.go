package mongox

import (
	"net/url"
	"strings"

	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

type Client struct {
	db          *mongo.Database
	transaction usecasex.Transaction
}

func NewClient(database string, c *mongo.Client) *Client {
	rp := readpref.Nearest()

	sessionOpts := options.Database().SetReadPreference(rp)

	return &Client{
		db:          c.Database(database, sessionOpts),
		transaction: &usecasex.NopTransaction{},
	}
}

func NewClientWithDatabase(db *mongo.Database) *Client {
	return &Client{
		db:          db,
		transaction: &usecasex.NopTransaction{},
	}
}

func (c *Client) WithTransaction() *Client {
	c.transaction = NewTransaction(c.db.Client())
	return c
}

func (c *Client) Collection(col string) *Collection {
	return NewCollection(c.db.Collection(col))
}

// WithCollection is deprecated
func (c *Client) WithCollection(col string) *Collection {
	return c.Collection(col)
}

func (c *Client) Database() *mongo.Database {
	return c.db
}

func (c *Client) Transaction() usecasex.Transaction {
	return c.transaction
}

func IsTransactionAvailable(original string) bool {
	u, _ := url.Parse(original)
	return u.Scheme == connstring.SchemeMongoDBSRV || u.Scheme == connstring.SchemeMongoDB && strings.Contains(u.Host, ",")
}
