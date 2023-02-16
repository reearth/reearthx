package accountmongo

import (
	"context"

	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/log"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/util"
	"go.mongodb.org/mongo-driver/mongo"
)

func New(ctx context.Context, mc *mongo.Client, databaseName string) (*accountrepo.Container, error) {
	if databaseName == "" {
		databaseName = "reearth_cms"
	}

	client := mongox.NewClient(databaseName, mc)
	c := &accountrepo.Container{
		Workspace:   NewWorkspace(client),
		User:        NewUser(client),
		Transaction: mongox.NewTransaction(client),
	}

	// init
	if err := Init(c); err != nil {
		return nil, err
	}

	return c, nil
}

func NewWithDB(ctx context.Context, db *mongo.Database) (*accountrepo.Container, error) {
	return New(ctx, db.Client(), db.Name())
}

func Init(r *accountrepo.Container) error {
	if r == nil {
		return nil
	}

	return util.Try(
		r.Workspace.(*Workspace).Init,
		r.User.(*User).Init,
	)
}

func createIndexes(ctx context.Context, c *mongox.ClientCollection, keys, uniqueKeys []string) error {
	created, deleted, err := c.Indexes(ctx, keys, uniqueKeys)
	if len(created) > 0 || len(deleted) > 0 {
		log.Infof("mongo: %s: index deleted: %v, created: %v", c.Client().Name(), deleted, created)
	}
	return err
}
