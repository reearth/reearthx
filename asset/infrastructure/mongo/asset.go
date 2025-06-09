package mongo

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/infrastructure/mongo/mongodoc"
	"github.com/reearth/reearthx/asset/usecase/gateway"
	"github.com/reearth/reearthx/asset/usecase/interfaces"
	"github.com/reearth/reearthx/asset/usecase/repo"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// HostAdapter defines an interface for getting the current host from context
type HostAdapter interface {
	CurrentHost(ctx context.Context) string
}

var (
	assetIndexes = []string{
		"project,!createdat,!id",
		"project,createdat,id",
		"project,!size,!id",
		"project,size,id",
		"!createdat,!id",
	}
	assetUniqueIndexes = []string{"id", "uuid"}
)

type Asset struct {
	hostAdapter     HostAdapter
	client          *mongox.Collection
	projectFilter   repo.ProjectFilter
	workspaceFilter repo.WorkspaceFilter
}

func NewAsset(client *mongox.Client) repo.Asset {
	return &Asset{client: client.WithCollection("asset")}
}

func NewAssetWithHostAdapter(client *mongox.Client, hostAdapter HostAdapter) repo.Asset {
	return &Asset{
		client:      client.WithCollection("asset"),
		hostAdapter: hostAdapter,
	}
}

func (r *Asset) Init() error {
	return createIndexes2(
		context.Background(),
		r.client,
		append(
			mongox.IndexFromKeys(assetUniqueIndexes, true),
			mongox.IndexFromKeys(assetIndexes, false)...,
		)...,
	)
}

func (r *Asset) Filtered(f repo.ProjectFilter) repo.Asset {
	return &Asset{
		client:          r.client,
		projectFilter:   r.projectFilter.Merge(f),
		workspaceFilter: r.workspaceFilter,
		hostAdapter:     r.hostAdapter,
	}
}

func (r *Asset) FindByID(ctx context.Context, id id.AssetID) (*asset.Asset, error) {
	return r.findOne(ctx, bson.M{
		"id": id.String(),
	})
}

func (r *Asset) FindByUUID(ctx context.Context, uuid string) (*asset.Asset, error) {
	return r.findOne(ctx, bson.M{
		"uuid": uuid,
	})
}

func (r *Asset) FindByIDs(ctx context.Context, ids id.AssetIDList) (asset.List, error) {
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

func (r *Asset) Search(
	ctx context.Context,
	pID id.ProjectID,
	filter repo.AssetFilter,
) (asset.List, *usecasex.PageInfo, error) {
	if !r.projectFilter.CanRead(pID) {
		return nil, usecasex.EmptyPageInfo(), nil
	}

	filters := bson.M{
		"project": pID.String(),
	}

	if filter.Keyword != nil && *filter.Keyword != "" {
		filters["filename"] = bson.M{
			"$regex": primitive.Regex{
				Pattern: fmt.Sprintf(".*%s.*", regexp.QuoteMeta(*filter.Keyword)),
				Options: "i",
			},
		}
	}

	if len(filter.ContentTypes) > 0 {
		filters["file.contenttype"] = bson.M{
			"$in": filter.ContentTypes,
		}
	}

	return r.paginate(ctx, filters, filter.Sort, filter.Pagination)
}

func (r *Asset) UpdateProject(ctx context.Context, from, to id.ProjectID) error {
	if !r.projectFilter.CanWrite(from) || !r.projectFilter.CanWrite(to) {
		return repo.ErrOperationDenied
	}

	return r.client.UpdateMany(ctx, bson.M{
		"project": from.String(),
	}, bson.M{
		"project": to.String(),
	})
}

func (r *Asset) Save(ctx context.Context, asset *asset.Asset) error {
	if !r.projectFilter.CanWrite(asset.Project()) {
		return repo.ErrOperationDenied
	}

	doc, i := mongodoc.NewAsset(asset)
	_, err := r.client.Client().UpdateOne(ctx, bson.M{
		"id": i,
	}, bson.M{
		"$set": doc,
	}, options.Update().SetUpsert(true))
	if err != nil {
		return rerror.ErrInternalBy(err)
	}

	return nil
}

func (r *Asset) Delete(ctx context.Context, id id.AssetID) error {
	return r.client.RemoveOne(ctx, r.writeFilter(bson.M{
		"id": id.String(),
	}))
}

// BatchDelete deletes assets in batch based on multiple asset IDs
func (r *Asset) BatchDelete(ctx context.Context, ids id.AssetIDList) error {
	filter := bson.M{
		"id": bson.M{"$in": ids.Strings()},
	}
	return r.client.RemoveAll(ctx, r.writeFilter(filter))
}

func (r *Asset) paginate(
	ctx context.Context,
	filter any,
	sort *usecasex.Sort,
	pagination *usecasex.Pagination,
) ([]*asset.Asset, *usecasex.PageInfo, error) {
	c := mongodoc.NewAssetConsumer()
	pageInfo, err := r.client.Paginate(
		ctx,
		r.readFilter(filter),
		sort,
		pagination,
		c,
		options.Find().SetProjection(bson.M{"file": 0}),
	)
	if err != nil {
		return nil, nil, rerror.ErrInternalBy(err)
	}
	return c.Result, pageInfo, nil
}

func (r *Asset) find(ctx context.Context, filter interface{}) ([]*asset.Asset, error) {
	c := mongodoc.NewAssetConsumer()
	if err := r.client.Find(ctx, r.readFilter(filter), c, options.Find().SetProjection(bson.M{"file": 0})); err != nil {
		return nil, rerror.ErrInternalBy(err)
	}
	return c.Result, nil
}

func (r *Asset) findOne(ctx context.Context, filter interface{}) (*asset.Asset, error) {
	c := mongodoc.NewAssetConsumer()
	if err := r.client.FindOne(ctx, r.readFilter(filter), c, options.FindOne().SetProjection(bson.M{"file": 0})); err != nil {
		return nil, err
	}
	return c.Result[0], nil
}

func filterAssets(ids []id.AssetID, rows []*asset.Asset) []*asset.Asset {
	res := make([]*asset.Asset, 0, len(ids))
	for _, i := range ids {
		var r2 *asset.Asset
		for _, r := range rows {
			if r.ID() == i {
				r2 = r
				break
			}
		}
		res = append(res, r2)
	}
	return res
}

func (r *Asset) FindByWorkspaceProject(
	ctx context.Context,
	id accountdomain.WorkspaceID,
	projectId *id.ProjectID,
	uFilter repo.AssetFilter,
) ([]*asset.Asset, *usecasex.PageInfo, error) {
	if !r.workspaceFilter.CanRead(id) {
		return nil, usecasex.EmptyPageInfo(), nil
	}

	filter := bson.M{
		"coresupport": true,
	}

	if projectId != nil {
		filter["project"] = projectId.String()
	} else {
		filter["team"] = id.String()
	}

	if uFilter.Keyword != nil {
		keyword := fmt.Sprintf(".*%s.*", regexp.QuoteMeta(*uFilter.Keyword))
		filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: keyword, Options: "i"}}
	}

	bucketPattern := ""
	if r.hostAdapter != nil {
		bucketPattern = r.hostAdapter.CurrentHost(ctx)
	}

	switch {
	case bucketPattern == "":
		bucketPattern = "example.com"
	case strings.Contains(bucketPattern, "localhost"):
		bucketPattern = "localhost"
	default:
		bucketPattern = "visualizer"
	}

	if andFilter, ok := mongox.And(filter, "url", bson.M{
		"$regex": primitive.Regex{Pattern: bucketPattern, Options: "i"},
	}).(bson.M); ok {
		filter = andFilter
	}

	return r.paginate(ctx, filter, uFilter.Sort, uFilter.Pagination)
}

func (r *Asset) TotalSizeByWorkspace(
	ctx context.Context,
	wid accountdomain.WorkspaceID,
) (int64, error) {
	if !r.workspaceFilter.CanRead(wid) {
		return 0, repo.ErrOperationDenied
	}

	c, err := r.client.Client().Aggregate(ctx, []bson.M{
		{"$match": bson.M{"team": wid.String()}},
		{"$group": bson.M{"_id": nil, "size": bson.M{"$sum": "$size"}}},
	})
	if err != nil {
		return 0, rerror.ErrInternalByWithContext(ctx, err)
	}
	defer func() {
		_ = c.Close(ctx)
	}()

	if !c.Next(ctx) {
		return 0, nil
	}

	type resp struct {
		Size int64
	}
	var res resp
	if err := c.Decode(&res); err != nil {
		return 0, rerror.ErrInternalByWithContext(ctx, err)
	}
	return res.Size, nil
}

func (r *Asset) RemoveByProjectWithFile(
	ctx context.Context,
	pid id.ProjectID,
	f gateway.File,
) error {
	projectAssets, err := r.find(ctx, bson.M{
		"coresupport": true,
		"project":     pid.String(),
	})
	if err != nil {
		return err
	}

	for _, a := range projectAssets {

		if !r.workspaceFilter.CanWrite(a.Workspace()) {
			return repo.ErrOperationDenied
		}

		aPath, err := url.Parse(a.URL())
		if err != nil {
			continue
		}

		err = f.RemoveAsset(ctx, aPath)
		if err != nil {
			log.Print(err.Error())
		}

		err = r.Delete(ctx, a.ID())
		if err != nil {
			log.Print(err.Error())
		}

	}

	return nil
}

func (r *Asset) FindByWorkspace(
	ctx context.Context,
	id accountdomain.WorkspaceID,
	uFilter repo.AssetFilter,
) ([]*asset.Asset, *interfaces.PageBasedInfo, error) {
	if !r.workspaceFilter.CanRead(id) {
		return nil, interfaces.NewPageBasedInfo(0, 1, 1), nil
	}

	var filter any = bson.M{
		"workspace": id.String(),
	}

	if uFilter.Keyword != nil {
		filter = mongox.And(filter, "name", bson.M{
			"$regex": primitive.Regex{
				Pattern: fmt.Sprintf(".*%s.*", regexp.QuoteMeta(*uFilter.Keyword)),
				Options: "i",
			},
		})
	}

	return r.paginateFlow(ctx, filter, uFilter.SortType, uFilter.Pagination)
}

func (r *Asset) readFilter(filter interface{}) interface{} {
	return applyProjectFilter(filter, r.projectFilter.Readable)
}

func (r *Asset) writeFilter(filter interface{}) interface{} {
	return applyProjectFilter(filter, r.projectFilter.Writable)
}

func (r *Asset) paginateFlow(
	ctx context.Context,
	filter any,
	sort *asset.SortType,
	pagination *usecasex.Pagination,
) ([]*asset.Asset, *interfaces.PageBasedInfo, error) {
	c := mongodoc.NewAssetConsumer()

	if pagination != nil && pagination.Offset != nil {
		skip := pagination.Offset.Offset
		limit := pagination.Offset.Limit

		total, err := r.client.Count(ctx, filter)
		if err != nil {
			return nil, nil, rerror.ErrInternalByWithContext(ctx, err)
		}

		opts := options.Find()
		if sort != nil {
			opts.SetSort(bson.D{{Key: string(sort.Key), Value: 1}})
		}

		opts.SetSkip(skip).SetLimit(limit)

		if err := r.client.Find(ctx, filter, c, opts); err != nil {
			return nil, nil, rerror.ErrInternalByWithContext(ctx, err)
		}

		page := int(skip/limit) + 1
		pageSize := int(limit)

		return c.Result, interfaces.NewPageBasedInfo(total, page, pageSize), nil
	}

	return c.Result, interfaces.NewPageBasedInfo(int64(len(c.Result)), 1, len(c.Result)), nil
}
