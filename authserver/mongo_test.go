package authserver

import (
	"context"
	"testing"
	"time"

	"github.com/reearth/reearthx/mongox"
	"github.com/reearth/reearthx/mongox/mongotest"
	"github.com/reearth/reearthx/rerror"
	"github.com/stretchr/testify/assert"
	"github.com/zitadel/oidc/pkg/oidc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	mongotest.Env = "REEARTH_DB"
}

func TestNewMongo(t *testing.T) {
	c := mongotest.Connect(t)(t)
	col := mongox.NewClientCollection(c.Collection("auth_request"))
	assert.Equal(t, &Mongo{
		client: col,
	}, NewMongo(col))
}

func TestMongo_FindByID(t *testing.T) {
	c := mongotest.Connect(t)(t)
	col := c.Collection("auth_request")
	m := &Mongo{client: mongox.NewClientCollection(col)}

	ctx := context.Background()
	id := NewRequestID()
	r := NewRequest().ID(id).MustBuild()

	got, err := m.FindByID(ctx, id)
	assert.Nil(t, got)
	assert.Same(t, rerror.ErrNotFound, err)

	_, _ = col.InsertOne(ctx, bson.M{
		"id": r.ID().String(),
	})

	got, err = m.FindByID(ctx, id)
	assert.Equal(t, r, got)
	assert.NoError(t, err)
}

func TestMongo_FindByCode(t *testing.T) {
	c := mongotest.Connect(t)(t)
	col := c.Collection("auth_request")
	m := &Mongo{client: mongox.NewClientCollection(col)}

	ctx := context.Background()
	r := NewRequest().NewID().Code("aaa").MustBuild()

	got, err := m.FindByCode(ctx, "aaa")
	assert.Nil(t, got)
	assert.Same(t, rerror.ErrNotFound, err)

	_, _ = col.InsertOne(ctx, bson.M{
		"id":   r.ID().String(),
		"code": "aaa",
	})

	got, err = m.FindByCode(ctx, "aaa")
	assert.Equal(t, r, got)
	assert.NoError(t, err)
}

func TestMongo_FindBySubject(t *testing.T) {
	c := mongotest.Connect(t)(t)
	col := c.Collection("auth_request")
	m := &Mongo{client: mongox.NewClientCollection(col)}

	ctx := context.Background()
	r := NewRequest().NewID().Subject("sss").MustBuild()

	got, err := m.FindBySubject(ctx, "sss")
	assert.Nil(t, got)
	assert.Same(t, rerror.ErrNotFound, err)

	_, _ = col.InsertOne(ctx, bson.M{
		"id":      r.ID().String(),
		"subject": "sss",
	})

	got, err = m.FindBySubject(ctx, "sss")
	assert.Equal(t, r, got)
	assert.NoError(t, err)
}

func TestMongo_Save(t *testing.T) {
	c := mongotest.Connect(t)(t)
	col := c.Collection("auth_request")
	m := &Mongo{client: mongox.NewClientCollection(col)}

	ctx := context.Background()
	aa := time.Now()
	r := NewRequest().NewID().
		ClientID("client").
		Subject("sub").
		Code("code").
		State("state").
		ResponseType("rt").
		Scopes([]string{"a", "openid"}).
		Audiences([]string{"aud"}).
		RedirectURI("ru").
		Nonce("nonce").
		CodeChallenge(&oidc.CodeChallenge{
			Challenge: "xxx",
			Method:    oidc.CodeChallengeMethodPlain,
		}).
		AuthorizedAt(&aa).
		MustBuild()

	assert.NoError(t, m.Save(ctx, r))
	cur := col.FindOne(ctx, bson.M{"id": r.ID().String()})
	var data bson.M
	assert.NoError(t, cur.Decode(&data))
	assert.Equal(t, bson.M{
		"_id":          data["_id"],
		"id":           r.ID().String(),
		"clientid":     "client",
		"subject":      "sub",
		"code":         "code",
		"state":        "state",
		"responsetype": "rt",
		"scopes":       bson.A{"a", "openid", "profile", "email"},
		"audiences":    bson.A{"aud"},
		"redirecturi":  "ru",
		"nonce":        "nonce",
		"codechallenge": bson.M{
			"challenge": "xxx",
			"method":    "plain",
		},
		"authorizedat": primitive.NewDateTimeFromTime(aa),
	}, data)
}

func TestMongo_Remove(t *testing.T) {
	c := mongotest.Connect(t)(t)
	col := c.Collection("auth_request")
	m := &Mongo{client: mongox.NewClientCollection(col)}

	ctx := context.Background()
	r := NewRequest().NewID().MustBuild()

	err := m.Remove(ctx, r.ID())
	assert.Same(t, rerror.ErrNotFound, err)

	_, _ = col.InsertOne(ctx, bson.M{
		"id": r.ID().String(),
	})

	err = m.Remove(ctx, r.ID())
	assert.NoError(t, err)
	assert.Equal(t, mongo.ErrNoDocuments, col.FindOne(ctx, bson.M{"id": r.ID().String()}).Err())
}
