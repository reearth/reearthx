package accountmongo

import (
	"context"
	"testing"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountdomain/workspace"
	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

// ---------------------------------------------------------------------------
// Reads ignore soft-deleted documents
// ---------------------------------------------------------------------------

func TestUser_Finders_IgnoreSoftDeleted(t *testing.T) {
	ctx := context.Background()
	db := compatDB(t)
	repo := NewUser(mongox.NewClientWithDatabase(db))

	// A: active (accounts-api shape), B: soft-deleted, C: OSS single-DB shape (no deletedat)
	active := accountsAPIUser(compatUserA, compatWS, "a@example.com")
	deleted := accountsAPIUser(compatUserB, compatWS, "b@example.com")
	deleted["deletedat"] = compatPast
	oss := user.New().
		ID(accountdomain.MustUserID(compatUserC)).
		Name("name-" + compatUserC).
		Alias("alias-" + compatUserC).
		Email("c@example.com").
		Workspace(accountdomain.MustWorkspaceID(compatWS)).
		Auths([]user.Auth{user.AuthFrom("sub-" + compatUserC)}).
		MustBuild()
	_, err := db.Collection("user").InsertMany(ctx, []any{active, deleted})
	require.NoError(t, err)
	require.NoError(t, repo.Save(ctx, oss))

	a, b, c := accountdomain.MustUserID(compatUserA), accountdomain.MustUserID(compatUserB), accountdomain.MustUserID(compatUserC)

	type finder struct {
		name string
		find func(id, email string) (*user.User, error)
	}
	finders := []finder{
		{"FindByID", func(id, _ string) (*user.User, error) { return repo.FindByID(ctx, accountdomain.MustUserID(id)) }},
		{"FindBySub", func(id, _ string) (*user.User, error) { return repo.FindBySub(ctx, "sub-"+id) }},
		{"FindByEmail", func(_, email string) (*user.User, error) { return repo.FindByEmail(ctx, email) }},
		{"FindByName", func(id, _ string) (*user.User, error) { return repo.FindByName(ctx, "name-"+id) }},
		{"FindByAlias", func(id, _ string) (*user.User, error) { return repo.FindByAlias(ctx, "alias-"+id) }},
		{"FindByNameOrEmail/email", func(_, email string) (*user.User, error) { return repo.FindByNameOrEmail(ctx, email) }},
		{"FindByNameOrEmail/name", func(id, _ string) (*user.User, error) { return repo.FindByNameOrEmail(ctx, "name-"+id) }},
	}
	for _, f := range finders {
		t.Run(f.name, func(t *testing.T) {
			got, err := f.find(compatUserA, "a@example.com")
			if assert.NoError(t, err, "active accounts-api user") {
				assert.Equal(t, a, got.ID())
			}
			got, err = f.find(compatUserC, "c@example.com")
			if assert.NoError(t, err, "OSS user without deletedat") {
				assert.Equal(t, c, got.ID())
			}
			_, err = f.find(compatUserB, "b@example.com")
			assert.ErrorIs(t, err, rerror.ErrNotFound, "soft-deleted user must not be found")
		})
	}

	t.Run("FindByVerification", func(t *testing.T) {
		_, err := repo.FindByVerification(ctx, "code-"+compatUserA)
		assert.NoError(t, err)
		_, err = repo.FindByVerification(ctx, "code-"+compatUserB)
		assert.ErrorIs(t, err, rerror.ErrNotFound)
	})

	t.Run("FindByPasswordResetRequest", func(t *testing.T) {
		_, err := repo.FindByPasswordResetRequest(ctx, "token-"+compatUserA)
		assert.NoError(t, err)
		_, err = repo.FindByPasswordResetRequest(ctx, "token-"+compatUserB)
		assert.ErrorIs(t, err, rerror.ErrNotFound)
	})

	t.Run("FindByIDs", func(t *testing.T) {
		got, err := repo.FindByIDs(ctx, user.IDList{a, b, c})
		require.NoError(t, err)
		// FindByIDs keeps positions and returns nil for missing ids
		require.Len(t, got, 3)
		assert.Equal(t, a, got[0].ID())
		assert.Nil(t, got[1], "soft-deleted user must not be returned")
		assert.Equal(t, c, got[2].ID())
	})

	t.Run("FindAll", func(t *testing.T) {
		got, err := repo.FindAll(ctx)
		require.NoError(t, err)
		assert.ElementsMatch(t, []accountdomain.UserID{a, c}, ids(got))
	})

	t.Run("SearchByKeyword", func(t *testing.T) {
		got, err := repo.SearchByKeyword(ctx, "name-")
		require.NoError(t, err)
		assert.ElementsMatch(t, []accountdomain.UserID{a, c}, ids(got))
	})

	t.Run("FindByIDsWithPagination", func(t *testing.T) {
		got, info, err := repo.FindByIDsWithPagination(ctx, user.IDList{a, b, c},
			usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap())
		require.NoError(t, err)
		assert.ElementsMatch(t, []accountdomain.UserID{a, c}, ids(got))
		assert.Equal(t, int64(2), info.TotalCount)
	})

	t.Run("FindBySubOrCreate does not resurrect a soft-deleted user", func(t *testing.T) {
		in := user.New().NewID().Name("x").Email("b@example.com").
			Workspace(accountdomain.MustWorkspaceID(compatWS)).MustBuild()
		got, err := repo.FindBySubOrCreate(ctx, in, "sub-"+compatUserB)
		assert.ErrorIs(t, err, rerror.ErrNotFound, "a deactivated user must not be returned")
		assert.Nil(t, got)
		n, err := db.Collection("user").CountDocuments(ctx, bson.M{"subs": "sub-" + compatUserB})
		require.NoError(t, err)
		assert.Equal(t, int64(1), n, "no second user may be created for the same sub")
	})
}

func ids(l user.List) []accountdomain.UserID {
	res := make([]accountdomain.UserID, 0, len(l))
	for _, u := range l {
		res = append(res, u.ID())
	}
	return res
}

func TestWorkspace_Finders_IgnoreSoftDeleted(t *testing.T) {
	ctx := context.Background()
	db := compatDB(t)
	repo := NewWorkspace(mongox.NewClientWithDatabase(db))

	// WS: active accounts-api workspace, WS2: soft-deleted one (same members), WS3: OSS shape
	active := accountsAPIWorkspace(compatWS)
	deleted := accountsAPIWorkspace(compatWS2)
	deleted["deletedat"] = compatPast
	deleted["name"] = "deleted-name"
	_, err := db.Collection("workspace").InsertMany(ctx, []any{active, deleted})
	require.NoError(t, err)

	ws3ID := accountdomain.NewWorkspaceID()
	a := accountdomain.MustUserID(compatUserA)
	require.NoError(t, repo.Save(ctx, workspace.New().ID(ws3ID).Name("oss").Alias("oss-alias").
		Members(map[accountdomain.UserID]workspace.Member{a: {Role: workspace.RoleOwner, InvitedBy: a}}).
		Integrations(map[accountdomain.IntegrationID]workspace.Member{
			accountdomain.MustIntegrationID(compatInteg): {Role: workspace.RoleWriter, InvitedBy: a},
		}).
		MustBuild()))

	wsID, ws2ID := accountdomain.MustWorkspaceID(compatWS), accountdomain.MustWorkspaceID(compatWS2)
	want := []accountdomain.WorkspaceID{wsID, ws3ID}

	t.Run("FindByID", func(t *testing.T) {
		_, err := repo.FindByID(ctx, wsID)
		assert.NoError(t, err)
		_, err = repo.FindByID(ctx, ws3ID)
		assert.NoError(t, err)
		_, err = repo.FindByID(ctx, ws2ID)
		assert.ErrorIs(t, err, rerror.ErrNotFound)
	})

	t.Run("FindByIDs", func(t *testing.T) {
		got, err := repo.FindByIDs(ctx, accountdomain.WorkspaceIDList{wsID, ws2ID, ws3ID})
		require.NoError(t, err)
		assert.ElementsMatch(t, want, got.IDs())
	})

	t.Run("FindByUser", func(t *testing.T) {
		got, err := repo.FindByUser(ctx, a)
		require.NoError(t, err)
		assert.ElementsMatch(t, want, got.IDs())
	})

	t.Run("FindByUserWithPagination", func(t *testing.T) {
		got, info, err := repo.FindByUserWithPagination(ctx, a, usecasex.OffsetPagination{Offset: 0, Limit: 10}.Wrap())
		require.NoError(t, err)
		assert.ElementsMatch(t, want, got.IDs())
		assert.Equal(t, int64(2), info.TotalCount)
	})

	t.Run("FindByIntegration", func(t *testing.T) {
		got, err := repo.FindByIntegration(ctx, accountdomain.MustIntegrationID(compatInteg))
		require.NoError(t, err)
		assert.ElementsMatch(t, want, got.IDs())
	})

	t.Run("FindByIntegrations", func(t *testing.T) {
		got, err := repo.FindByIntegrations(ctx, workspace.IntegrationIDList{accountdomain.MustIntegrationID(compatInteg)})
		require.NoError(t, err)
		assert.ElementsMatch(t, want, got.IDs())
	})

	t.Run("FindByName", func(t *testing.T) {
		_, err := repo.FindByName(ctx, "ws-"+compatWS)
		assert.NoError(t, err)
		_, err = repo.FindByName(ctx, "deleted-name")
		assert.ErrorIs(t, err, rerror.ErrNotFound)
	})

	t.Run("FindByAlias", func(t *testing.T) {
		_, err := repo.FindByAlias(ctx, "ws-alias-"+compatWS)
		assert.NoError(t, err)
		_, err = repo.FindByAlias(ctx, "oss-alias")
		assert.NoError(t, err)
		_, err = repo.FindByAlias(ctx, "ws-alias-"+compatWS2)
		assert.ErrorIs(t, err, rerror.ErrNotFound)
	})

	t.Run("FindByIDOrAlias", func(t *testing.T) {
		_, err := repo.FindByIDOrAlias(ctx, workspace.IDOrAlias(compatWS))
		assert.NoError(t, err)
		_, err = repo.FindByIDOrAlias(ctx, workspace.IDOrAlias("WS-ALIAS-"+compatWS))
		assert.NoError(t, err)
		_, err = repo.FindByIDOrAlias(ctx, workspace.IDOrAlias(compatWS2))
		assert.ErrorIs(t, err, rerror.ErrNotFound)
		_, err = repo.FindByIDOrAlias(ctx, workspace.IDOrAlias("ws-alias-"+compatWS2))
		assert.ErrorIs(t, err, rerror.ErrNotFound)
	})
}
