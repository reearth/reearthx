package mongo

import (
	"context"
	"time"

	"github.com/reearth/reearthx/asset"
	"github.com/reearth/reearthx/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ asset.GroupRepository = &GroupRepository{}

type GroupRepository struct {
	client *mongox.Collection
}

func NewGroupRepository(db *mongo.Database) asset.GroupRepository {
	return &GroupRepository{
		client: mongox.NewCollection(db.Collection("asset_groups")),
	}
}

func (r *GroupRepository) Save(ctx context.Context, group *asset.Group) error {
	doc := groupToDoc(group)

	_, err := r.client.Client().UpdateOne(
		ctx,
		bson.M{"id": group.ID.String()},
		bson.M{"$set": doc},
		options.Update().SetUpsert(true),
	)

	return err
}

func (r *GroupRepository) FindByID(ctx context.Context, id asset.GroupID) (*asset.Group, error) {
	var doc groupDocument

	err := r.client.FindOne(ctx, bson.M{"id": id.String()}, mongox.FuncConsumer(func(raw bson.Raw) error {
		return bson.Unmarshal(raw, &doc)
	}))

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return docToGroup(&doc)
}

func (r *GroupRepository) Delete(ctx context.Context, id asset.GroupID) error {
	return r.client.RemoveOne(ctx, bson.M{"id": id.String()})
}

func (r *GroupRepository) UpdatePolicy(ctx context.Context, id asset.GroupID, policyID *asset.PolicyID) error {
	var update bson.M

	if policyID == nil {
		update = bson.M{"$unset": bson.M{"policyid": ""}}
	} else {
		update = bson.M{"$set": bson.M{"policyid": policyID.String()}}
	}

	_, err := r.client.Client().UpdateOne(
		ctx,
		bson.M{"id": id.String()},
		update,
	)

	return err
}

type groupDocument struct {
	ID        string    `bson:"id"`
	Name      string    `bson:"name"`
	CreatedAt time.Time `bson:"createdat"`
	UpdatedAt time.Time `bson:"updatedat"`
	PolicyID  string    `bson:"policyid,omitempty"`
}

func groupToDoc(g *asset.Group) *groupDocument {
	doc := &groupDocument{
		ID:        g.ID.String(),
		Name:      g.Name,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}

	if g.PolicyID != nil {
		doc.PolicyID = g.PolicyID.String()
	}

	return doc
}

func docToGroup(doc *groupDocument) (*asset.Group, error) {
	groupID, err := asset.GroupIDFrom(doc.ID)
	if err != nil {
		return nil, err
	}

	group := &asset.Group{
		ID:        groupID,
		Name:      doc.Name,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}

	if doc.PolicyID != "" {
		policyID, err := asset.PolicyIDFrom(doc.PolicyID)
		if err != nil {
			return nil, err
		}
		group.PolicyID = &policyID
	}

	return group, nil
}
