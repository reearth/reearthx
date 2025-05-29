package mongo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/infrastructure/mongo/mongodoc"
	"github.com/reearth/reearthx/asset/usecase/repo"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/idx"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	assetIndexes = []string{
		"groupid,!createdat,!id",
		"groupid,createdat,id",
		"groupid,!size,!id",
		"groupid,size,id",
		"!createdat,!id",
	}
	assetUniqueIndexes = []string{"id", "uuid"}
)

var _ repo.AssetRepository = &AssetRepository{}

type AssetRepository struct {
	client *mongox.Collection
	f      *asset.GroupFilter
}

func NewAssetRepository(db *mongo.Database) *AssetRepository {
	return &AssetRepository{
		client: mongox.NewCollection(db.Collection("assets")),
		f:      &asset.GroupFilter{},
	}
}

func (r *AssetRepository) Init() error {
	return createIndexes(
		context.Background(),
		r.client,
		assetIndexes,
		assetUniqueIndexes,
	)
}

func (r *AssetRepository) Filtered(filter asset.GroupFilter) repo.AssetRepository {
	return &AssetRepository{
		client: r.client,
		f:      r.f.Merge(&filter),
	}
}

func (r *AssetRepository) SaveCMS(ctx context.Context, asset *asset.Asset) error {
	if !r.f.CanWrite(*asset.GroupID()) {
		return errors.New("operation denied")
	}

	doc, id := mongodoc.NewAsset(asset)
	_, err := r.client.Client().UpdateOne(ctx, bson.M{
		"id": id,
	}, bson.M{
		"$set": doc,
	}, options.Update().SetUpsert(true))
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}

func (r *AssetRepository) Search(ctx context.Context, pID asset.GroupID, filter asset.Filter) (asset.List, *usecasex.PageInfo, error) {
	if !r.f.CanRead(pID) {
		return nil, usecasex.EmptyPageInfo(), nil
	}

	filters := bson.M{
		"groupid": pID.String(),
	}

	if filter.Keyword != nil && *filter.Keyword != "" {
		filters["filename"] = bson.M{
			"$regex": primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", regexp.QuoteMeta(*filter.Keyword)), Options: "i"},
		}
	}

	if len(filter.ContentTypes) > 0 {
		filters["contenttype"] = bson.M{
			"$in": filter.ContentTypes,
		}
	}

	pagination := filter.Pagination
	if pagination == nil {
		pagination = &usecasex.Pagination{
			Offset: &usecasex.OffsetPagination{
				Offset: 0,
				Limit:  50,
			},
		}
	}

	result, pageInfo, err := r.paginate(ctx, filters, filter.Sort, pagination)

	return result, pageInfo, err
}

func (r *AssetRepository) FindByID(ctx context.Context, id asset.ID) (*asset.Asset, error) {
	return r.findOne(ctx, bson.M{
		"id": id.String(),
	})
}

func (r *AssetRepository) FindByUUID(ctx context.Context, uuid string) (*asset.Asset, error) {
	return r.findOne(ctx, bson.M{
		"uuid": uuid,
	})
}

func (r *AssetRepository) FindByIDs(ctx context.Context, ids asset.IDList) ([]*asset.Asset, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	filter := bson.M{
		"id": bson.M{"$in": ids.Strings()},
	}
	res, err := r.find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return filterAssets(ids, res), nil
}

func (r *AssetRepository) UpdateProject(ctx context.Context, from, to asset.GroupID) error {
	if !r.f.CanWrite(from) || !r.f.CanWrite(to) {
		return errors.New("operation denied")
	}

	return r.client.UpdateMany(ctx, bson.M{
		"groupid": from.String(),
	}, bson.M{
		"groupid": to.String(),
	})
}

func (r *AssetRepository) FindByIDList(ctx context.Context, ids asset.IDList) (asset.List, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	filter := bson.M{
		"id": bson.M{"$in": ids.Strings()},
	}

	res, err := r.find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}

	return asset.List(res), nil
}

func (r *AssetRepository) FindByGroup(
	ctx context.Context,
	groupID asset.GroupID,
	filter asset.Filter,
	sort asset.Sort,
	pagination asset.Pagination,
) ([]*asset.Asset, int64, error) {
	if !r.f.CanRead(groupID) {
		return nil, 0, nil
	}

	query := bson.M{"groupid": groupID.String()}

	if filter.Keyword != nil {
		query["filename"] = bson.M{
			"$regex": primitive.Regex{
				Pattern: fmt.Sprintf(".*%s.*", regexp.QuoteMeta(*filter.Keyword)),
				Options: "i",
			},
		}
	}

	sortOptions := bson.D{}
	switch sort.By {
	case asset.SortTypeDate:

		if sort.Direction == asset.SortDirectionAsc {
			sortOptions = append(sortOptions, bson.E{Key: "createdat", Value: 1})
		} else {
			sortOptions = append(sortOptions, bson.E{Key: "createdat", Value: -1})
		}
	case asset.SortTypeSize:
		if sort.Direction == asset.SortDirectionAsc {
			sortOptions = append(sortOptions, bson.E{Key: "size", Value: 1})
		} else {
			sortOptions = append(sortOptions, bson.E{Key: "size", Value: -1})
		}
	case asset.SortTypeName:
		if sort.Direction == asset.SortDirectionAsc {
			sortOptions = append(sortOptions, bson.E{Key: "filename", Value: 1})
		} else {
			sortOptions = append(sortOptions, bson.E{Key: "filename", Value: -1})
		}
	default:
		sortOptions = append(sortOptions, bson.E{Key: "createdat", Value: -1})
	}

	count, err := r.client.Count(ctx, r.readFilter(query))
	if err != nil {
		return nil, 0, rerror.ErrInternalBy(err)
	}

	findOptions := options.Find().
		SetSort(sortOptions).
		SetSkip(pagination.Offset).
		SetLimit(pagination.Limit)

	consumer := mongox.NewSliceFuncConsumer(func(doc assetDocument) (*assetDocument, error) {
		return &doc, nil
	})

	err = r.client.Find(ctx, r.readFilter(query), consumer, findOptions)
	if err != nil {
		return nil, 0, rerror.ErrInternalBy(err)
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

func (r *AssetRepository) FindByProject(ctx context.Context, groupID asset.GroupID, filter asset.Filter) (asset.List, *usecasex.PageInfo, error) {
	if !r.f.CanRead(groupID) {
		return nil, usecasex.EmptyPageInfo(), nil
	}

	var query interface{} = bson.M{"groupid": groupID.String()}

	if filter.Keyword != nil {
		query = mongox.And(query, "", bson.M{
			"filename": bson.M{
				"$regex": primitive.Regex{
					Pattern: fmt.Sprintf(".*%s.*", regexp.QuoteMeta(*filter.Keyword)),
					Options: "i",
				},
			},
		})
	}

	pagination := filter.Pagination
	if pagination == nil {
		pagination = &usecasex.Pagination{
			Offset: &usecasex.OffsetPagination{
				Offset: 0,
				Limit:  50,
			},
		}
	}

	return r.paginate(ctx, query, filter.Sort, pagination)
}

func (r *AssetRepository) Delete(ctx context.Context, id asset.ID) error {
	return r.client.RemoveOne(ctx, r.writeFilter(bson.M{
		"id": id.String(),
	}))
}

func (r *AssetRepository) DeleteMany(ctx context.Context, ids []asset.ID) error {
	idStrings := make([]string, len(ids))
	for i, id := range ids {
		idStrings[i] = id.String()
	}

	filter := r.writeFilter(bson.M{
		"id": bson.M{"$in": idStrings},
	})

	return r.client.RemoveAll(ctx, filter)
}

func (r *AssetRepository) BatchDelete(ctx context.Context, ids asset.IDList) error {
	if len(ids) == 0 {
		return nil
	}

	filter := r.writeFilter(bson.M{
		"id": bson.M{"$in": ids.Strings()},
	})

	return r.client.RemoveAll(ctx, filter)
}

func (r *AssetRepository) UpdateExtractionStatus(ctx context.Context, id asset.ID, status asset.ExtractionStatus) error {
	filter := r.writeFilter(bson.M{"id": id.String()})
	_, err := r.client.Client().UpdateOne(
		ctx,
		filter,
		bson.M{"$set": bson.M{"archiveextractionstatus": string(status)}},
	)
	if err != nil {
		return rerror.ErrInternalBy(err)
	}
	return nil
}

type FileRemover interface {
	RemoveAsset(context.Context, *url.URL) error
}

// Helper methods for document operations
func (r *AssetRepository) find(ctx context.Context, filter any) ([]*asset.Asset, error) {
	c := mongodoc.NewAssetConsumer()
	if err := r.client.Find(ctx, r.readFilter(filter), c, options.Find().SetProjection(bson.M{"file": 0})); err != nil {
		return nil, rerror.ErrInternalBy(err)
	}
	return c.Result, nil
}

func (r *AssetRepository) findOne(ctx context.Context, filter any) (*asset.Asset, error) {
	c := mongodoc.NewAssetConsumer()
	if err := r.client.FindOne(ctx, r.readFilter(filter), c, options.FindOne().SetProjection(bson.M{"file": 0})); err != nil {
		return nil, err
	}
	return c.Result[0], nil
}

// cms
func filterAssets(ids []asset.ID, rows []*asset.Asset) []*asset.Asset {
	res := make([]*asset.Asset, 0, len(ids))
	for _, id := range ids {
		var r2 *asset.Asset
		for _, r := range rows {
			if r.ID() == id {
				r2 = r
				break
			}
		}
		res = append(res, r2)
	}
	return res
}

func applyGroupFilter(filter interface{}, groups []asset.GroupID) interface{} {
	if len(groups) == 0 {
		return filter
	}

	groupStrings := make([]string, len(groups))
	for i, g := range groups {
		groupStrings[i] = g.String()
	}

	groupFilter := bson.M{
		"groupid": bson.M{"$in": groupStrings},
	}

	return mongox.And(filter, "", groupFilter)
}

func createIndexes(ctx context.Context, c *mongox.Collection, indexes, uniqueIndexes []string) error {
	_, _, err := c.Indexes(ctx, indexes, uniqueIndexes)
	return err
}

type assetDocument struct {
	ID                      string    `bson:"id"`
	GroupID                 string    `bson:"groupid"`
	CreatedAt               time.Time `bson:"createdat"`
	Size                    int64     `bson:"size"`
	ContentType             string    `bson:"contenttype"`
	ContentEncoding         string    `bson:"contentencoding,omitempty"`
	PreviewType             string    `bson:"previewtype"`
	UUID                    string    `bson:"uuid"`
	URL                     string    `bson:"url"`
	FileName                string    `bson:"filename"`
	ArchiveExtractionStatus string    `bson:"archiveextractionstatus,omitempty"`
	IntegrationID           string    `bson:"integrationid"`
}

// func assetToDoc(a *asset.Asset) *assetDocument {
// 	doc := &assetDocument{
// 		ID:            a.ID().String(),
// 		GroupID:       a.GroupID().String(),
// 		CreatedAt:     a.CreatedAt(),
// 		Size:          a.Size(),
// 		ContentType:   a.ContentType(),
// 		UUID:          a.UUID(),
// 		URL:           a.URL(),
// 		FileName:      a.FileName(),
// 		IntegrationID: a.Integration().String(),
// 	}

// 	if a.PreviewType() != nil {
// 		doc.PreviewType = string(*a.PreviewType())
// 	}

// 	if a.ContentEncoding() != "" {
// 		doc.ContentEncoding = a.ContentEncoding()
// 	}

// 	if a.ArchiveExtractionStatus() != nil {
// 		doc.ArchiveExtractionStatus = string(*a.ArchiveExtractionStatus())
// 	}

// 	return doc
// }

func docToAsset(doc *assetDocument) (*asset.Asset, error) {
	assetID, err := asset.IdFrom(doc.ID)
	if err != nil {
		return nil, err
	}

	groupID, err := asset.GroupIDFrom(doc.GroupID)
	if err != nil {
		return nil, err
	}

	var integration asset.IntegrationID
	if doc.IntegrationID != "" {
		integration, err = idx.From[asset.IntegrationIDType](doc.IntegrationID)
		if err != nil {
			return nil, err
		}
	}

	a := asset.NewAsset(assetID, &groupID, doc.CreatedAt, doc.Size, doc.ContentType)
	a.SetContentEncoding(doc.ContentEncoding)
	a.SetPreviewType(asset.PreviewType(doc.PreviewType))
	a.SetUUID(doc.UUID)
	a.SetURL(doc.URL)
	a.SetFileName(doc.FileName)
	a.AddIntegration(integration)

	if doc.ArchiveExtractionStatus != "" {
		status := asset.ExtractionStatus(doc.ArchiveExtractionStatus)
		a.SetArchiveExtractionStatus(&status)
	}

	return a, nil
}

// cms
func (r *AssetRepository) paginate(ctx context.Context, filter interface{}, sort *usecasex.Sort, pagination *usecasex.Pagination) (asset.List, *usecasex.PageInfo, error) {
	c := mongodoc.NewAssetConsumer()

	actualFilter := r.readFilter(filter)

	pageInfo, err := r.client.Paginate(ctx, actualFilter, sort, pagination, c, options.Find().SetProjection(bson.M{"file": 0}))
	if err != nil {
		return nil, nil, rerror.ErrInternalBy(err)
	}

	return c.Result, pageInfo, nil
}

func (r *AssetRepository) readFilter(filter any) any {
	if r.f == nil || r.f.Readable == nil {
		return filter
	}
	return applyGroupFilter(filter, r.f.Readable)
}

func (r *AssetRepository) writeFilter(filter any) any {
	if r.f == nil || r.f.Writable == nil {
		return filter
	}
	return applyGroupFilter(filter, r.f.Writable)
}

// viz
func (r *AssetRepository) FindByURL(ctx context.Context, path string) (*asset.Asset, error) {
	return r.findOne(ctx, bson.M{
		"url": path,
	})
}

// viz
func (r *AssetRepository) FindByWorkspaceProject(ctx context.Context, workspaceID accountdomain.WorkspaceID, groupID *asset.GroupID, filter asset.Filter) ([]*asset.Asset, *usecasex.PageInfo, error) {
	if !r.f.CanRead(asset.GroupID(workspaceID)) {
		return nil, usecasex.EmptyPageInfo(), nil
	}

	query := bson.M{
		"coresupport": true,
	}

	if groupID != nil {
		query["groupid"] = groupID.String()
	} else {
		query["groupid"] = workspaceID.String()
	}

	if filter.Keyword != nil {
		keyword := fmt.Sprintf(".*%s.*", regexp.QuoteMeta(*filter.Keyword))
		query["filename"] = bson.M{"$regex": primitive.Regex{Pattern: keyword, Options: "i"}}
	}

	if filter.Keyword != nil {
		keyword := fmt.Sprintf(".*%s.*", regexp.QuoteMeta(*filter.Keyword))
		query["filename"] = bson.M{"$regex": primitive.Regex{Pattern: keyword, Options: "i"}}
	}

	bucketPattern := "localhost"
	if strings.Contains(bucketPattern, "localhost") {
		bucketPattern = "localhost"
	} else {
		bucketPattern = "visualizer"
	}

	if andFilter, ok := mongox.And(query, "url", bson.M{
		"$regex": primitive.Regex{Pattern: bucketPattern, Options: "i"},
	}).(bson.M); ok {
		query = andFilter
	}

	pagination := filter.Pagination
	if pagination == nil {
		pagination = &usecasex.Pagination{
			Offset: &usecasex.OffsetPagination{
				Offset: 0,
				Limit:  50,
			},
		}
	}

	return r.paginate(ctx, query, filter.Sort, pagination)
}

// viz and flow
func (r *AssetRepository) TotalSizeByWorkspace(ctx context.Context, wid accountdomain.WorkspaceID) (int64, error) {
	if !r.f.CanRead(asset.GroupID(wid)) {
		return 0, rerror.ErrInvalidParams
	}

	// Use MongoDB aggregation to sum up asset sizes for the workspace
	pipeline := []bson.M{
		{"$match": bson.M{"groupid": wid.String()}},
		{"$group": bson.M{"_id": nil, "totalSize": bson.M{"$sum": "$size"}}},
	}

	cursor, err := r.client.Client().Aggregate(ctx, pipeline)
	if err != nil {
		return 0, rerror.ErrInternalBy(err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			slog.Error("error closing cursor")
		}

	}(cursor, ctx)

	type result struct {
		TotalSize int64 `bson:"totalSize"`
	}

	if cursor.Next(ctx) {
		var res result
		if err := cursor.Decode(&res); err != nil {
			return 0, rerror.ErrInternalBy(err)
		}
		return res.TotalSize, nil
	}

	return 0, nil
}

// viz
func (r *AssetRepository) RemoveByProjectWithFile(ctx context.Context, groupID asset.GroupID, fileInterface any) error {
	if !r.f.CanWrite(groupID) {
		return rerror.ErrInvalidParams
	}

	assets, err := r.find(ctx, bson.M{"groupid": groupID.String()})
	if err != nil {
		return err
	}

	var fileRemover FileRemover
	if fr, ok := fileInterface.(FileRemover); ok {
		fileRemover = fr
	}

	for _, a := range assets {

		if !r.f.CanWrite(*a.GroupID()) {
			return errors.New("operation denied")
		}

		if fileRemover != nil && a.URL() != "" {
			assetURL, err := url.Parse(a.URL())
			if err != nil {
				continue
			}

			if err := fileRemover.RemoveAsset(ctx, assetURL); err != nil {
				continue
			}
		}

		if err := r.Delete(ctx, a.ID()); err != nil {
			continue
		}
	}

	return nil
}

// viz and flow
func (r *AssetRepository) Save(ctx context.Context, asset *asset.Asset) error {
	if !r.f.CanWrite(*asset.GroupID()) {
		return errors.New("operation denied")
	}
	doc, id := mongodoc.NewAsset(asset)
	return r.client.SaveOne(ctx, id, doc)
}
