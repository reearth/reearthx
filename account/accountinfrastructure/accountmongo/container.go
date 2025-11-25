package accountmongo

import (
	"context"

	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/mongox"
	"go.mongodb.org/mongo-driver/mongo"
)

func New(ctx context.Context, mc *mongo.Client, databaseName string, useTransaction, needCompat bool, users []accountrepo.User) (*accountrepo.Container, error) {
	if databaseName == "" {
		databaseName = "reearth_cms"
	}

	client := mongox.NewClient(databaseName, mc)
	if useTransaction {
		client = client.WithTransaction()
	}
	var ws accountrepo.Workspace

	if needCompat {
		ws = NewWorkspaceCompat(client)
	} else {
		ws = NewWorkspace(client)
	}
	c := &accountrepo.Container{
		Workspace:   ws,
		User:        NewUser(client),
		Transaction: client.Transaction(),
		Users:       users,
		Role:        NewRole(client),
		Permittable: NewPermittable(client),
	}

	return c, nil
}

func NewWithDB(ctx context.Context, db *mongo.Database, useTransaction, needCompat bool, users []accountrepo.User) (*accountrepo.Container, error) {
	return New(ctx, db.Client(), db.Name(), useTransaction, needCompat, users)
}
