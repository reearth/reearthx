package mongo

import (
	"context"
	"errors"
	asset2 "github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/infrastructure/mongo/mongodoc"
	"github.com/reearth/reearthx/asset/usecase/repo"

	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/rerror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ repo.AssetFile = &AssetFile{}

type AssetFile struct {
	client           *mongox.Collection
	assetFilesClient *mongox.Collection
}

func NewAssetFile(db *mongo.Database) repo.AssetFile {
	return &AssetFile{
		client:           mongox.NewCollection(db.Collection("assets")),
		assetFilesClient: mongox.NewCollection(db.Collection("asset_files")),
	}
}

func (r *AssetFile) FindByID(ctx context.Context, id asset2.AssetID) (*asset2.File, error) {
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

	f := c.Result[0].File.Model()
	if f == nil {
		return nil, rerror.ErrNotFound
	}
	if c.Result[0].FlatFiles {
		var afc mongodoc.AssetFilesConsumer
		if err := r.assetFilesClient.Find(ctx, bson.M{
			"assetid": id.String(),
		}, &afc, options.Find().SetSort(bson.D{
			{Key: "page", Value: 1},
		})); err != nil {
			return nil, err
		}
		files := afc.Result().Model()
		// f = asset.FoldFiles(files, f)
		f.SetFiles(files)
	} else if len(f.Children()) > 0 {
		f.SetFiles(f.FlattenChildren())
	}
	return f, nil
}

func (r *AssetFile) FindByIDs(ctx context.Context, ids repo.AssetIDList) (map[asset2.AssetID]*asset2.File, error) {
	filesMap := make(map[asset2.AssetID]*asset2.File)

	c := &mongodoc.AssetAndFileConsumer{}
	if err := r.client.Find(ctx, bson.M{
		"id": bson.M{"$in": ids.Strings()},
	}, c, options.Find().SetProjection(bson.M{
		"id":        1,
		"file":      1,
		"flatfiles": 1,
	})); err != nil {
		return nil, err
	}

	for _, result := range c.Result {
		assetID := result.ID
		f := result.File.Model()
		if f == nil {
			return nil, rerror.ErrNotFound
		}

		if result.FlatFiles {
			var afc mongodoc.AssetFilesConsumer
			if err := r.assetFilesClient.Find(ctx, bson.M{
				"assetid": assetID,
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

		aId, err := asset2.AssetIDFrom(assetID)
		if err != nil {
			return nil, err
		}
		filesMap[aId] = f
	}

	return filesMap, nil
}

func (r *AssetFile) Save(ctx context.Context, id asset2.AssetID, file *asset2.File) error {
	doc := mongodoc.NewFile(file)
	_, err := r.client.Client().UpdateOne(ctx, bson.M{
		"id": id.String(),
	}, bson.M{
		"$set": bson.M{
			"id":   id.String(),
			"file": doc,
		},
	})
	if errors.Is(err, mongo.ErrNoDocuments) {
		return rerror.ErrNotFound
	}
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}

func (r *AssetFile) SaveFlat(ctx context.Context, id asset2.AssetID, parent *asset2.File, files []*asset2.File) error {
	doc := mongodoc.NewFile(parent)
	_, err := r.client.Client().UpdateOne(ctx, bson.M{
		"id": id.String(),
	}, bson.M{
		"$set": bson.M{
			"id":        id.String(),
			"flatfiles": true,
			"file":      doc,
		},
	})
	if errors.Is(err, mongo.ErrNoDocuments) {
		return rerror.ErrNotFound
	}
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	if err := r.assetFilesClient.RemoveAll(ctx, bson.M{"assetid": id.String()}); err != nil {
		return rerror.ErrInternalBy(err)
	}
	if len(files) == 0 {
		return nil
	}
	filesDoc := mongodoc.NewFiles(id, files)
	writeModels := make([]mongo.WriteModel, 0, len(filesDoc))
	for _, pageDoc := range filesDoc {
		writeModels = append(writeModels, mongo.NewInsertOneModel().SetDocument(pageDoc))
	}
	if _, err := r.assetFilesClient.Client().BulkWrite(ctx, writeModels); err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}
