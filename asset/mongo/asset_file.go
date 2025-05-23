package mongo

import (
	"context"

	"github.com/reearth/reearthx/asset"
	"github.com/reearth/reearthx/asset/mongo/mongodoc"
	"github.com/reearth/reearthx/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ asset.AssetFile = &AssetFile{}

type AssetFile struct {
	client           *mongox.Collection
	assetFilesClient *mongox.Collection
}

func NewAssetFile(db *mongo.Database) asset.AssetFile {
	return &AssetFile{
		client:           mongox.NewCollection(db.Collection("assets")),
		assetFilesClient: mongox.NewCollection(db.Collection("asset_files")),
	}
}

func (r *AssetFile) FindByID(ctx context.Context, id asset.AssetID) (*asset.File, error) {
	c := &mongodoc.AssetAndFileConsumer{}
	if err := r.client.FindOne(ctx, bson.M{
		"id": id.String(),
	}, c, options.FindOne().SetProjection(bson.M{
		"id":        1,
		"file":      1,
		"flatfiles": 1,
	})); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	results := c.Result()
	if len(results) == 0 {
		return nil, nil
	}

	result := results[0]
	f := result.File.Model()
	if f == nil {
		return nil, nil
	}

	if result.FlatFiles {
		var afc mongodoc.AssetFilesConsumer
		if err := r.assetFilesClient.Find(ctx, bson.M{
			"assetid": id.String(),
		}, &afc, options.Find().SetSort(bson.D{
			{Key: "page", Value: 1},
		})); err != nil {
			return nil, err
		}
		files := afc.Result().Model()
		f.SetFiles(files)
	} else if len(f.Children()) > 0 {
		f.SetFiles(f.FlattenChildren())
	}

	return f, nil
}

func (r *AssetFile) FindByIDs(ctx context.Context, ids asset.AssetIDList) (map[asset.AssetID]*asset.File, error) {
	filesMap := make(map[asset.AssetID]*asset.File)

	idStrings := ids.Strings()

	c := &mongodoc.AssetAndFileConsumer{}
	if err := r.client.Find(ctx, bson.M{
		"id": bson.M{"$in": idStrings},
	}, c, options.Find().SetProjection(bson.M{
		"id":        1,
		"file":      1,
		"flatfiles": 1,
	})); err != nil {
		return nil, err
	}

	results := c.Result()
	for _, result := range results {
		f := result.File.Model()
		if f == nil {
			continue
		}

		if result.FlatFiles {
			var afc mongodoc.AssetFilesConsumer
			if err := r.assetFilesClient.Find(ctx, bson.M{
				"assetid": result.ID,
			}, &afc, options.Find().SetSort(bson.D{
				{Key: "page", Value: 1},
			})); err != nil {
				return nil, err
			}
			files := afc.Result().Model()
			f.SetFiles(files)
		} else if len(f.Children()) > 0 {
			f.SetFiles(f.FlattenChildren())
		}

		aId, err := asset.AssetIDFrom(result.ID)
		if err != nil {
			return nil, err
		}
		filesMap[aId] = f
	}

	return filesMap, nil
}

func (r *AssetFile) Save(ctx context.Context, id asset.AssetID, file *asset.File) error {
	doc := mongodoc.NewFile(file)
	_, err := r.client.Client().UpdateOne(ctx, bson.M{
		"id": id.String(),
	}, bson.M{
		"$set": bson.M{
			"id":   id.String(),
			"file": doc,
		},
	}, options.Update().SetUpsert(true))

	return err
}

func (r *AssetFile) SaveFlat(ctx context.Context, id asset.AssetID, parent *asset.File, files []*asset.File) error {
	doc := mongodoc.NewFile(parent)
	_, err := r.client.Client().UpdateOne(ctx, bson.M{
		"id": id.String(),
	}, bson.M{
		"$set": bson.M{
			"id":        id.String(),
			"flatfiles": true,
			"file":      doc,
		},
	}, options.Update().SetUpsert(true))

	if err != nil {
		return err
	}

	// Remove existing asset files
	if err := r.assetFilesClient.RemoveAll(ctx, bson.M{"assetid": id.String()}); err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	// Insert new files in pages
	filesDoc := mongodoc.NewFiles(id, files)
	writeModels := make([]mongo.WriteModel, 0, len(filesDoc))
	for _, pageDoc := range filesDoc {
		writeModels = append(writeModels, mongo.NewInsertOneModel().SetDocument(pageDoc))
	}

	if _, err := r.assetFilesClient.Client().BulkWrite(ctx, writeModels); err != nil {
		return err
	}

	return nil
}
