package accountmongo

// Tests for sharing the user/workspace collections with reearth-accounts-api.
//
// accounts-api writes fields (see reearth-accounts
// server/internal/infrastructure/mongo/mongodoc) that reearthx used to drop.
// reearthx now carries them in its documents and domain models, so a user or
// workspace that was read from the DB keeps them when it is saved again.

import (
	"context"
	"testing"
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/account/accountinfrastructure/accountmongo/mongodoc"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	compatUserA = "01hz7k3q8v9w2x4y6z8a0b2c4d"
	compatUserB = "01hz7k3q8v9w2x4y6z8a0b2c4e"
	compatUserC = "01hz7k3q8v9w2x4y6z8a0b2c4f"
	compatInteg = "01hz7k3q8v9w2x4y6z8a0b2c4g"
	compatWS    = "01hz7k3q8v9w2x4y6z8a0b2c4h"
	compatWS2   = "01hz7k3q8v9w2x4y6z8a0b2c4j"

	// Golden values computed with reearth-accounts
	// mongodoc.ComputeWorkspaceMembersHash (server @ 305a60c).
	//
	// members {A: owner/A, C: writer/A}, integrations {I: writer/A}
	compatHashACI = "8ccb595518129aa289892b3932537958963471f10792568d3bddd45998f04089"
	// no members and no integrations (hash of JSON "null")
	compatHashEmpty = "74234e98afe7498fb5daf1f36ac2d78acc339464f950703b8c019892f982b90b"
)

var compatPast = time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

func compatDB(t *testing.T) *mongo.Database {
	t.Helper()
	return mongotest.Connect(t)(t)
}

// accountsAPIUser returns a user document in the shape accounts-api writes.
func accountsAPIUser(id, wsID, email string) bson.M {
	return bson.M{
		"id":             id,
		"name":           "name-" + id,
		"alias":          "alias-" + id,
		"email":          email,
		"subs":           bson.A{"sub-" + id},
		"workspace":      wsID,
		"lang":           "",
		"theme":          "",
		"metadata":       bson.M{"description": "", "website": "", "photourl": "", "lang": "ja", "theme": "dark"},
		"latestlogoutat": compatPast,
		"createdat":      compatPast,
		"updatedat":      compatPast,
		"verification":   bson.M{"code": "code-" + id, "expiration": compatPast, "verified": true},
		"passwordreset":  bson.M{"token": "token-" + id, "createdat": compatPast},
	}
}

// accountsAPIWorkspace returns a workspace document in the shape accounts-api writes.
func accountsAPIWorkspace(id string) bson.M {
	return bson.M{
		"id":    id,
		"name":  "ws-" + id,
		"alias": "ws-alias-" + id,
		"email": "",
		"members": bson.M{
			compatUserA: bson.M{"disabled": false, "externalid": "ext-a", "invitedby": compatUserA, "role": "owner"},
			compatUserB: bson.M{"disabled": false, "externalid": "ext-b", "invitedby": compatUserA, "role": "reader"},
		},
		"integrations": bson.M{
			compatInteg: bson.M{"disabled": false, "invitedby": compatUserA, "role": "writer"},
		},
		"members_hash": "stale-hash",
		"metadata":     bson.M{"description": "", "website": "", "location": "", "billingemail": "", "photourl": ""},
		"personal":     false,
		"policy":       "policy-1",
		"scimconfig": bson.M{
			"enabled":          true,
			"grouprolemapping": bson.M{"admins": "owner"},
			"tokenhash":        "bcrypt-hash",
		},
		"createdat": compatPast,
		"createdby": compatUserA,
		"updatedat": compatPast,
	}
}

func rawFind(t *testing.T, db *mongo.Database, col, id string) bson.M {
	t.Helper()
	var doc bson.M
	require.NoError(t, db.Collection(col).FindOne(context.Background(), bson.M{"id": id}).Decode(&doc))
	return doc
}

func rawTime(t *testing.T, v any) time.Time {
	t.Helper()
	dt, ok := v.(primitive.DateTime)
	require.Truef(t, ok, "expected a BSON datetime, got %T (%v)", v, v)
	return dt.Time().UTC()
}

// ulidTime is the creation time encoded in a ULID, at the millisecond precision Mongo stores.
func ulidTime(id string) time.Time {
	return accountdomain.MustUserID(id).Timestamp().UTC()
}

// rawUserModel loads a user through mongodoc without the repo finders (which skip soft-deleted users).
func rawUserModel(t *testing.T, db *mongo.Database, id string) *user.User {
	t.Helper()
	var d mongodoc.UserDocument
	require.NoError(t, db.Collection("user").FindOne(context.Background(), bson.M{"id": id}).Decode(&d))
	u, err := d.Model()
	require.NoError(t, err)
	return u
}

// rawWorkspaceModel loads a workspace through mongodoc without the repo finders.
func rawWorkspaceModel(t *testing.T, db *mongo.Database, id string) *workspace.Workspace {
	t.Helper()
	var d mongodoc.WorkspaceDocument
	require.NoError(t, db.Collection("workspace").FindOne(context.Background(), bson.M{"id": id}).Decode(&d))
	w, err := d.Model()
	require.NoError(t, err)
	return w
}

// editWorkspace applies the edits CMS typically makes: rename, remove B, add C, clear the policy.
func editWorkspace(t *testing.T, ws *workspace.Workspace) {
	t.Helper()
	a := accountdomain.MustUserID(compatUserA)
	c := user.New().ID(accountdomain.MustUserID(compatUserC)).Email("c@example.com").
		Workspace(accountdomain.MustWorkspaceID(compatWS)).MustBuild()

	ws.Rename("renamed-" + ws.ID().String())
	require.NoError(t, ws.Members().Leave(accountdomain.MustUserID(compatUserB)))
	require.NoError(t, ws.Members().Join(c, workspace.RoleWriter, a))
	ws.SetPolicy(nil)
}

// ---------------------------------------------------------------------------
// User writes
// ---------------------------------------------------------------------------

func TestUser_Save_KeepsAccountsAPIFields(t *testing.T) {
	ctx := context.Background()
	db := compatDB(t)
	repo := NewUser(mongox.NewClientWithDatabase(db))

	seed := accountsAPIUser(compatUserA, compatWS, "a@example.com")
	seed["team"] = compatWS // legacy field; reearthx always writes "workspace"
	_, err := db.Collection("user").InsertOne(ctx, seed)
	require.NoError(t, err)

	u, err := repo.FindByID(ctx, accountdomain.MustUserID(compatUserA))
	require.NoError(t, err)
	assert.Equal(t, compatPast, *u.CreatedAt())
	assert.Equal(t, compatPast, u.LatestLogoutAt())

	u.UpdateName("renamed")
	before := time.Now().Add(-time.Second)
	require.NoError(t, repo.Save(ctx, u))

	doc := rawFind(t, db, "user", compatUserA)
	assert.Equal(t, "renamed", doc["name"])
	assert.Equal(t, compatPast, rawTime(t, doc["createdat"]), "createdat must survive Save")
	assert.Equal(t, compatPast, rawTime(t, doc["latestlogoutat"]), "latestlogoutat must survive Save")
	if assert.Contains(t, doc, "updatedat") {
		assert.True(t, rawTime(t, doc["updatedat"]).After(before), "updatedat must be bumped on Save")
	}
	assert.NotContains(t, doc, "team", "empty omitempty field team must be unset")
	assert.NotContains(t, doc, "deletedat")
	assert.Equal(t, bson.A{"sub-" + compatUserA}, doc["subs"])
}

func TestUser_Save_KeepsDeletedAt(t *testing.T) {
	ctx := context.Background()
	db := compatDB(t)
	repo := NewUser(mongox.NewClientWithDatabase(db))

	seed := accountsAPIUser(compatUserA, compatWS, "a@example.com")
	seed["deletedat"] = compatPast
	_, err := db.Collection("user").InsertOne(ctx, seed)
	require.NoError(t, err)

	u := rawUserModel(t, db, compatUserA)
	assert.Equal(t, compatPast, *u.DeletedAt())
	u.UpdateName("renamed")
	require.NoError(t, repo.Save(ctx, u))

	doc := rawFind(t, db, "user", compatUserA)
	assert.Equal(t, compatPast, rawTime(t, doc["deletedat"]), "deletedat must survive Save")
	assert.Equal(t, compatPast, rawTime(t, doc["createdat"]))
}

func TestUser_Save_InsertSetsTimestamps(t *testing.T) {
	ctx := context.Background()
	db := compatDB(t)
	repo := NewUser(mongox.NewClientWithDatabase(db))

	u := user.New().
		ID(accountdomain.MustUserID(compatUserA)).
		Name("new").
		Email("a@example.com").
		Workspace(accountdomain.MustWorkspaceID(compatWS)).
		MustBuild()
	before := time.Now().Add(-time.Second)
	require.NoError(t, repo.Save(ctx, u))

	doc := rawFind(t, db, "user", compatUserA)
	assert.Equal(t, ulidTime(compatUserA), rawTime(t, doc["createdat"]), "createdat comes from the ULID")
	if assert.Contains(t, doc, "updatedat") {
		assert.True(t, rawTime(t, doc["updatedat"]).After(before))
	}
	assert.NotContains(t, doc, "latestlogoutat", "a zero latestlogoutat is not written")

	// a second save of a reloaded user keeps createdat
	got, err := repo.FindByID(ctx, u.ID())
	require.NoError(t, err)
	got.UpdateName("again")
	require.NoError(t, repo.Save(ctx, got))
	doc = rawFind(t, db, "user", compatUserA)
	assert.Equal(t, ulidTime(compatUserA), rawTime(t, doc["createdat"]))
	assert.Equal(t, "again", doc["name"])
}

func TestUser_Create_SetsTimestamps(t *testing.T) {
	ctx := context.Background()
	db := compatDB(t)
	repo := NewUser(mongox.NewClientWithDatabase(db))

	u := user.New().
		NewID().
		Name("new").
		Email("a@example.com").
		Workspace(accountdomain.MustWorkspaceID(compatWS)).
		MustBuild()
	before := time.Now().Add(-time.Second)
	require.NoError(t, repo.Create(ctx, u))

	doc := rawFind(t, db, "user", u.ID().String())
	if assert.Contains(t, doc, "createdat") {
		assert.True(t, rawTime(t, doc["createdat"]).After(before))
	}
	if assert.Contains(t, doc, "updatedat") {
		assert.True(t, rawTime(t, doc["updatedat"]).After(before))
	}
}

// ---------------------------------------------------------------------------
// Workspace writes
// ---------------------------------------------------------------------------

func assertWorkspaceSaved(t *testing.T, doc bson.M, id string, before time.Time) {
	t.Helper()

	assert.Equal(t, "renamed-"+id, doc["name"])

	// accounts-api-only fields
	assert.Equal(t, bson.M{
		"enabled":          true,
		"grouprolemapping": bson.M{"admins": "owner"},
		"tokenhash":        "bcrypt-hash",
	}, doc["scimconfig"], "scimconfig must survive")
	assert.Equal(t, compatPast, rawTime(t, doc["createdat"]), "createdat must survive")
	assert.Equal(t, compatUserA, doc["createdby"], "createdby must survive")
	if assert.Contains(t, doc, "updatedat") {
		assert.True(t, rawTime(t, doc["updatedat"]).After(before), "updatedat must be bumped")
	}

	// members: A kept (with externalid), B removed, C added
	members, ok := doc["members"].(bson.M)
	require.True(t, ok, "members must be a document, got %T", doc["members"])
	assert.ElementsMatch(t, []string{compatUserA, compatUserC}, keys(members))
	assert.Equal(t, bson.M{"disabled": false, "externalid": "ext-a", "invitedby": compatUserA, "role": "owner"},
		members[compatUserA], "members.<id>.externalid must survive")
	assert.Equal(t, bson.M{"disabled": false, "invitedby": compatUserA, "role": "writer"}, members[compatUserC])

	integrations, ok := doc["integrations"].(bson.M)
	require.True(t, ok)
	assert.Equal(t, bson.M{compatInteg: bson.M{"disabled": false, "invitedby": compatUserA, "role": "writer"}}, integrations)

	// accounts-api's ComputeWorkspaceMembersHash for the new member set
	assert.Equal(t, compatHashACI, doc["members_hash"], "members_hash must be recomputed")

	// policy cleared
	assert.NotContains(t, doc, "policy", "cleared policy must be unset")
	assert.NotContains(t, doc, "deletedat")
}

func keys(m bson.M) []string {
	res := make([]string, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	return res
}

func TestWorkspace_Save_KeepsAccountsAPIFields(t *testing.T) {
	ctx := context.Background()
	db := compatDB(t)
	repo := NewWorkspace(mongox.NewClientWithDatabase(db))

	_, err := db.Collection("workspace").InsertOne(ctx, accountsAPIWorkspace(compatWS))
	require.NoError(t, err)

	ws, err := repo.FindByID(ctx, accountdomain.MustWorkspaceID(compatWS))
	require.NoError(t, err)
	editWorkspace(t, ws)
	before := time.Now().Add(-time.Second)
	require.NoError(t, repo.Save(ctx, ws))

	assertWorkspaceSaved(t, rawFind(t, db, "workspace", compatWS), compatWS, before)

	// and it still reads back through the repo
	got, err := repo.FindByID(ctx, accountdomain.MustWorkspaceID(compatWS))
	require.NoError(t, err)
	assert.Equal(t, "renamed-"+compatWS, got.Name())
	assert.Nil(t, got.Policy())
	assert.Equal(t, "ext-a", got.Members().User(accountdomain.MustUserID(compatUserA)).ExternalID)
	assert.Equal(t, &workspace.ScimConfig{
		Enabled:          true,
		GroupRoleMapping: map[string]workspace.Role{"admins": workspace.RoleOwner},
		TokenHash:        "bcrypt-hash",
	}, got.ScimConfig())
	assert.Equal(t, accountdomain.MustUserID(compatUserA), *got.CreatedBy())
}

func TestWorkspace_SaveAll_KeepsAccountsAPIFields(t *testing.T) {
	ctx := context.Background()
	db := compatDB(t)
	repo := NewWorkspace(mongox.NewClientWithDatabase(db))

	for _, id := range []string{compatWS, compatWS2} {
		_, err := db.Collection("workspace").InsertOne(ctx, accountsAPIWorkspace(id))
		require.NoError(t, err)
	}

	wss, err := repo.FindByIDs(ctx, accountdomain.WorkspaceIDList{
		accountdomain.MustWorkspaceID(compatWS), accountdomain.MustWorkspaceID(compatWS2),
	})
	require.NoError(t, err)
	require.Len(t, wss, 2)
	for _, ws := range wss {
		editWorkspace(t, ws)
	}
	before := time.Now().Add(-time.Second)
	require.NoError(t, repo.SaveAll(ctx, wss))

	for _, id := range []string{compatWS, compatWS2} {
		assertWorkspaceSaved(t, rawFind(t, db, "workspace", id), id, before)
	}
}

func TestWorkspace_Save_KeepsDeletedAt(t *testing.T) {
	ctx := context.Background()
	db := compatDB(t)
	repo := NewWorkspace(mongox.NewClientWithDatabase(db))

	seed := accountsAPIWorkspace(compatWS)
	seed["deletedat"] = compatPast
	_, err := db.Collection("workspace").InsertOne(ctx, seed)
	require.NoError(t, err)

	ws := rawWorkspaceModel(t, db, compatWS)
	assert.Equal(t, compatPast, *ws.DeletedAt())
	ws.Rename("renamed")
	require.NoError(t, repo.Save(ctx, ws))

	doc := rawFind(t, db, "workspace", compatWS)
	assert.Equal(t, compatPast, rawTime(t, doc["deletedat"]), "deletedat must survive Save")
}

func TestWorkspace_Save_InsertSetsTimestampsAndHash(t *testing.T) {
	ctx := context.Background()
	db := compatDB(t)
	repo := NewWorkspace(mongox.NewClientWithDatabase(db))

	a := accountdomain.MustUserID(compatUserA)
	ws := workspace.New().
		ID(accountdomain.MustWorkspaceID(compatWS)).
		Name("new").
		Members(map[accountdomain.UserID]workspace.Member{
			a:                                     {Role: workspace.RoleOwner, InvitedBy: a},
			accountdomain.MustUserID(compatUserC): {Role: workspace.RoleWriter, InvitedBy: a},
		}).
		Integrations(map[accountdomain.IntegrationID]workspace.Member{
			accountdomain.MustIntegrationID(compatInteg): {Role: workspace.RoleWriter, InvitedBy: a},
		}).
		MustBuild()
	before := time.Now().Add(-time.Second)
	require.NoError(t, repo.Save(ctx, ws))

	doc := rawFind(t, db, "workspace", compatWS)
	assert.Equal(t, ulidTime(compatWS), rawTime(t, doc["createdat"]), "createdat comes from the ULID")
	if assert.Contains(t, doc, "updatedat") {
		assert.True(t, rawTime(t, doc["updatedat"]).After(before))
	}
	assert.Equal(t, compatHashACI, doc["members_hash"])
	assert.Equal(t, bson.M{"disabled": false, "invitedby": compatUserA, "role": "owner"},
		doc["members"].(bson.M)[compatUserA], "members without an external ID have no externalid")
	assert.NotContains(t, doc, "createdby")
	assert.NotContains(t, doc, "scimconfig")
}

func TestWorkspace_Create_SetsTimestampsAndHash(t *testing.T) {
	ctx := context.Background()
	db := compatDB(t)
	repo := NewWorkspace(mongox.NewClientWithDatabase(db))

	ws := workspace.New().NewID().Name("n").MustBuild()
	before := time.Now().Add(-time.Second)
	require.NoError(t, repo.Create(ctx, ws))

	doc := rawFind(t, db, "workspace", ws.ID().String())
	if assert.Contains(t, doc, "createdat") {
		assert.True(t, rawTime(t, doc["createdat"]).After(before))
	}
	if assert.Contains(t, doc, "updatedat") {
		assert.True(t, rawTime(t, doc["updatedat"]).After(before))
	}
	assert.Equal(t, compatHashEmpty, doc["members_hash"])
}

func TestSave_BackfillsCreatedAtFromID(t *testing.T) {
	// Documents written before createdat existed get the creation time encoded in their ULID.
	ctx := context.Background()
	db := compatDB(t)
	client := mongox.NewClientWithDatabase(db)

	u := accountsAPIUser(compatUserA, compatWS, "a@example.com")
	delete(u, "createdat")
	ws := accountsAPIWorkspace(compatWS)
	delete(ws, "createdat")
	_, err := db.Collection("user").InsertOne(ctx, u)
	require.NoError(t, err)
	_, err = db.Collection("workspace").InsertOne(ctx, ws)
	require.NoError(t, err)

	require.NoError(t, NewUser(client).Save(ctx, rawUserModel(t, db, compatUserA)))
	require.NoError(t, NewWorkspace(client).Save(ctx, rawWorkspaceModel(t, db, compatWS)))

	assert.Equal(t, ulidTime(compatUserA), rawTime(t, rawFind(t, db, "user", compatUserA)["createdat"]))
	assert.Equal(t, ulidTime(compatWS), rawTime(t, rawFind(t, db, "workspace", compatWS)["createdat"]))
}
