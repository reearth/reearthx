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

type AssetRepository struct {
	client *mongox.Collection
}

func NewAssetRepository(db *mongo.Database) asset.AssetRepository {
	return &AssetRepository{
		client: mongox.NewCollection(db.Collection("assets")),
	}
}

func (r *AssetRepository) Save(ctx context.Context, asset *asset.Asset) error {
	doc := assetToDoc(asset)

	_, err := r.client.Client().UpdateOne(
		ctx,
		bson.M{"id": asset.ID.String()},
		bson.M{"$set": doc},
		options.Update().SetUpsert(true),
	)

	return err
}

func (r *AssetRepository) FindByID(ctx context.Context, id asset.AssetID) (*asset.Asset, error) {
	var doc assetDocument

	err := r.client.FindOne(ctx, bson.M{"id": id.String()}, mongox.FuncConsumer(func(raw bson.Raw) error {
		return bson.Unmarshal(raw, &doc)
	}))

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return docToAsset(&doc)
}

func (r *AssetRepository) FindByIDs(ctx context.Context, ids []asset.AssetID) ([]*asset.Asset, error) {
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.String()
	}

	consumer := mongox.NewSliceFuncConsumer(func(doc assetDocument) (*assetDocument, error) {
		return &doc, nil
	})

	err := r.client.Find(ctx, bson.M{"id": bson.M{"$in": idStrings}}, consumer)
	if err != nil {
		return nil, err
	}

	assets := make([]*asset.Asset, 0, len(consumer.Result))
	for _, doc := range consumer.Result {
		a, err := docToAsset(doc)
		if err != nil {
			continue
		}
		assets = append(assets, a)
	}

	return assets, nil
}

func (r *AssetRepository) FindByGroup(
	ctx context.Context,
	groupID asset.GroupID,
	filter asset.AssetFilter,
	sort asset.AssetSort,
	pagination asset.Pagination,
) ([]*asset.Asset, int64, error) {
	query := bson.M{"groupid": groupID.String()}

	if filter.Keyword != "" {
		query["filename"] = bson.M{"$regex": filter.Keyword, "$options": "i"}
	}

	sortOptions := bson.D{}
	switch sort.By {
	case asset.AssetSortTypeDate:
		if sort.Direction == asset.SortDirectionAsc {
			sortOptions = append(sortOptions, bson.E{Key: "createdat", Value: 1})
		} else {
			sortOptions = append(sortOptions, bson.E{Key: "createdat", Value: -1})
		}
	case asset.AssetSortTypeSize:
		if sort.Direction == asset.SortDirectionAsc {
			sortOptions = append(sortOptions, bson.E{Key: "size", Value: 1})
		} else {
			sortOptions = append(sortOptions, bson.E{Key: "size", Value: -1})
		}
	case asset.AssetSortTypeName:
		if sort.Direction == asset.SortDirectionAsc {
			sortOptions = append(sortOptions, bson.E{Key: "filename", Value: 1})
		} else {
			sortOptions = append(sortOptions, bson.E{Key: "filename", Value: -1})
		}
	default:
		sortOptions = append(sortOptions, bson.E{Key: "createdat", Value: -1})
	}

	count, err := r.client.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	findOptions := options.Find().
		SetSort(sortOptions).
		SetSkip(pagination.Offset).
		SetLimit(pagination.Limit)

	consumer := mongox.NewSliceFuncConsumer(func(doc assetDocument) (*assetDocument, error) {
		return &doc, nil
	})

	err = r.client.Find(ctx, query, consumer, findOptions)
	if err != nil {
		return nil, 0, err
	}

	assets := make([]*asset.Asset, 0, len(consumer.Result))
	for _, doc := range consumer.Result {
		a, err := docToAsset(doc)
		if err != nil {
			continue
		}
		assets = append(assets, a)
	}

	return assets, count, nil
}

func (r *AssetRepository) Delete(ctx context.Context, id asset.AssetID) error {
	return r.client.RemoveOne(ctx, bson.M{"id": id.String()})
}

func (r *AssetRepository) DeleteMany(ctx context.Context, ids []asset.AssetID) error {
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.String()
	}

	return r.client.RemoveAll(ctx, bson.M{"id": bson.M{"$in": idStrings}})
}

func (r *AssetRepository) UpdateExtractionStatus(ctx context.Context, id asset.AssetID, status asset.ExtractionStatus) error {
	_, err := r.client.Client().UpdateOne(
		ctx,
		bson.M{"id": id.String()},
		bson.M{"$set": bson.M{"archiveextractionstatus": string(status)}},
	)
	return err
}

type assetDocument struct {
	ID                      string    `bson:"id"`
	GroupID                 string    `bson:"groupid"`
	CreatedAt               time.Time `bson:"createdat"`
	CreatedByType           string    `bson:"createdbytype"`
	CreatedByID             string    `bson:"createdbyid"`
	Size                    int64     `bson:"size"`
	ContentType             string    `bson:"contenttype"`
	ContentEncoding         string    `bson:"contentencoding,omitempty"`
	PreviewType             string    `bson:"previewtype"`
	UUID                    string    `bson:"uuid"`
	URL                     string    `bson:"url"`
	FileName                string    `bson:"filename"`
	ArchiveExtractionStatus string    `bson:"archiveextractionstatus,omitempty"`
}

func assetToDoc(a *asset.Asset) *assetDocument {
	doc := &assetDocument{
		ID:            a.ID.String(),
		GroupID:       a.GroupID.String(),
		CreatedAt:     a.CreatedAt,
		CreatedByType: string(a.CreatedBy.Type),
		CreatedByID:   a.CreatedBy.ID,
		Size:          a.Size,
		ContentType:   a.ContentType,
		PreviewType:   string(a.PreviewType),
		UUID:          a.UUID,
		URL:           a.URL,
		FileName:      a.FileName,
	}

	if a.ContentEncoding != "" {
		doc.ContentEncoding = a.ContentEncoding
	}

	if a.ArchiveExtractionStatus != nil {
		doc.ArchiveExtractionStatus = string(*a.ArchiveExtractionStatus)
	}

	return doc
}

func docToAsset(doc *assetDocument) (*asset.Asset, error) {
	assetID, err := asset.AssetIDFrom(doc.ID)
	if err != nil {
		return nil, err
	}

	groupID, err := asset.GroupIDFrom(doc.GroupID)
	if err != nil {
		return nil, err
	}

	a := &asset.Asset{
		ID:        assetID,
		GroupID:   groupID,
		CreatedAt: doc.CreatedAt,
		CreatedBy: asset.OperatorInfo{
			Type: asset.OperatorType(doc.CreatedByType),
			ID:   doc.CreatedByID,
		},
		Size:            doc.Size,
		ContentType:     doc.ContentType,
		ContentEncoding: doc.ContentEncoding,
		PreviewType:     asset.PreviewType(doc.PreviewType),
		UUID:            doc.UUID,
		URL:             doc.URL,
		FileName:        doc.FileName,
	}

	if doc.ArchiveExtractionStatus != "" {
		status := asset.ExtractionStatus(doc.ArchiveExtractionStatus)
		a.ArchiveExtractionStatus = &status
	}

	return a, nil
}
