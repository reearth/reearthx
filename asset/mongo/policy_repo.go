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

var _ asset.PolicyRepository = &PolicyRepository{}

type PolicyRepository struct {
	client *mongox.Collection
}

func NewPolicyRepository(db *mongo.Database) asset.PolicyRepository {
	return &PolicyRepository{
		client: mongox.NewCollection(db.Collection("asset_policies")),
	}
}

func (r *PolicyRepository) Save(ctx context.Context, policy *asset.Policy) error {
	doc := policyToDoc(policy)

	_, err := r.client.Client().UpdateOne(
		ctx,
		bson.M{"id": policy.ID.String()},
		bson.M{"$set": doc},
		options.Update().SetUpsert(true),
	)

	return err
}

func (r *PolicyRepository) FindByID(ctx context.Context, id asset.PolicyID) (*asset.Policy, error) {
	var doc policyDocument

	err := r.client.FindOne(ctx, bson.M{"id": id.String()}, mongox.FuncConsumer(func(raw bson.Raw) error {
		return bson.Unmarshal(raw, &doc)
	}))

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return docToPolicy(&doc)
}

func (r *PolicyRepository) Delete(ctx context.Context, id asset.PolicyID) error {
	return r.client.RemoveOne(ctx, bson.M{"id": id.String()})
}

type policyDocument struct {
	ID           string    `bson:"id"`
	Name         string    `bson:"name"`
	StorageLimit int64     `bson:"storagelimit"`
	CreatedAt    time.Time `bson:"createdat"`
	UpdatedAt    time.Time `bson:"updatedat"`
}

func policyToDoc(p *asset.Policy) *policyDocument {
	return &policyDocument{
		ID:           p.ID.String(),
		Name:         p.Name,
		StorageLimit: p.StorageLimit,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func docToPolicy(doc *policyDocument) (*asset.Policy, error) {
	policyID, err := asset.PolicyIDFrom(doc.ID)
	if err != nil {
		return nil, err
	}

	policy := &asset.Policy{
		ID:           policyID,
		Name:         doc.Name,
		StorageLimit: doc.StorageLimit,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}

	return policy, nil
}
