package mongo

import (
	"context"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/usecase/repo"
	"github.com/reearth/reearthx/log"
	"github.com/reearth/reearthx/mongox"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func New(
	ctx context.Context,
	mc *mongo.Client,
	databaseName string,
	useTransaction bool,
	acRepo *accountrepo.Container,
) (*repo.Container, error) {
	if databaseName == "" {
		databaseName = "reearth_cms"
	}

	lock, err := NewLock(mc.Database(databaseName).Collection("locks"))
	if err != nil {
		return nil, err
	}

	client := mongox.NewClient(databaseName, mc)
	if useTransaction {
		client = client.WithTransaction()
	}

	c := &repo.Container{
		Asset:             NewAsset(client),
		AssetFile:         NewAssetFile(client),
		AssetUpload:       NewAssetUpload(client),
		User:              acRepo.User,
		Project:           NewProject(client),
		Workspace:         acRepo.Workspace,
		Transaction:       client.Transaction(),
		Lock:              lock,
		Request:           NewRequest(client),
		Item:              NewItem(client),
		View:              NewView(client),
		Model:             NewModel(client),
		Schema:            NewSchema(client),
		Thread:            NewThread(client),
		Integration:       NewIntegration(client),
		Group:             NewGroup(client),
		Event:             NewEvent(client),
		WorkspaceSettings: NewWorkspaceSettings(client),
	}

	return c, nil
}

func NewWithDB(
	ctx context.Context,
	db *mongo.Database,
	useTransaction bool,
	acRepo *accountrepo.Container,
) (*repo.Container, error) {
	return New(ctx, db.Client(), db.Name(), useTransaction, acRepo)
}

func logIndexResult(name string, r mongox.IndexResult) {
	d := r.DeletedNames()
	u := r.UpdatedNames()
	a := r.AddedNames()
	if len(d) == 0 && len(u) == 0 && len(a) == 0 {
		return
	}
	log.Infof("mongo: %s: index deleted: %v, updated: %v, created: %v", name, d, u, a)
}

func applyWorkspaceFilter(filter interface{}, ids accountdomain.WorkspaceIDList) interface{} {
	if ids == nil {
		return filter
	}
	return mongox.And(filter, "workspace", bson.M{"$in": ids.Strings()})
}

func applyProjectFilter(filter interface{}, ids id.ProjectIDList) interface{} {
	if ids == nil {
		return filter
	}
	return mongox.And(filter, "project", bson.M{"$in": ids.Strings()})
}

func applyProjectFilterToPipeline(pipeline []any, ids id.ProjectIDList) []any {
	if ids == nil {
		return pipeline
	}
	return append(
		[]any{bson.M{"$match": bson.M{"project": bson.M{"$in": ids.Strings()}}}},
		pipeline...)
}

// func applyWorkspaceFilterToPipeline(pipeline []any, ids accountdomain.WorkspaceIDList) []any {
// 	if ids == nil {
// 		return pipeline
// 	}
// 	return append([]any{bson.M{"$match": bson.M{"workspace": bson.M{"$in": ids.Strings()}}}}, pipeline...)
// }
