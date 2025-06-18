// TODO: Delete this file once the permission check migration is complete.

package accountmongo

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/role"
	"github.com/reearth/reearthx/account/accountinfrastructure/accountmongo/mongodoc"
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/mongox"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	roleIndexes       = []string{}
	roleUniqueIndexes = []string{"id", "name"}
)

type Role struct {
	client *mongox.Collection
}

func NewRole(client *mongox.Client) accountrepo.Role {
	return &Role{
		client: client.WithCollection("role"),
	}
}

func (r *Role) Init() error {
	return createIndexes(context.Background(), r.client, roleIndexes, roleUniqueIndexes)
}

func (r *Role) FindAll(ctx context.Context) (role.List, error) {
	filter := bson.M{}
	return r.find(ctx, filter)
}

func (r *Role) FindByID(ctx context.Context, id accountdomain.RoleID) (*role.Role, error) {
	return r.findOne(ctx, bson.M{
		"id": id.String(),
	})
}

func (r *Role) FindByIDs(ctx context.Context, ids accountdomain.RoleIDList) (role.List, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	filter := bson.M{
		"id": bson.M{
			"$in": ids.Strings(),
		},
	}
	return r.find(ctx, filter)
}

func (r *Role) Save(ctx context.Context, role role.Role) error {
	doc, gId := mongodoc.NewRole(role)
	return r.client.SaveOne(ctx, gId, doc)
}

func (r *Role) Remove(ctx context.Context, id accountdomain.RoleID) error {
	return r.client.RemoveOne(ctx, bson.M{"id": id.String()})
}

func (r *Role) find(ctx context.Context, filter any) (role.List, error) {
	c := mongodoc.NewRoleConsumer()
	if err := r.client.Find(ctx, filter, c); err != nil {
		return nil, err
	}
	if len(c.Result) == 0 {
		return role.List{}, nil
	}
	return (role.List)(c.Result), nil
}

func (r *Role) findOne(ctx context.Context, filter any) (*role.Role, error) {
	c := mongodoc.NewRoleConsumer()
	if err := r.client.FindOne(ctx, filter, c); err != nil {
		return nil, err
	}
	return c.Result[0], nil
}
