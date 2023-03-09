package authserver

import (
	"context"
	"time"

	"github.com/reearth/reearthx/log"
	"github.com/reearth/reearthx/mongox"
	"github.com/zitadel/oidc/pkg/oidc"
	"go.mongodb.org/mongo-driver/bson"
)

type Mongo struct {
	client *mongox.Collection
}

var _ RequestRepo = (*Mongo)(nil)

func NewMongo(client *mongox.Collection) *Mongo {
	r := &Mongo{client: client}
	return r
}

func (r *Mongo) Init() error {
	added, deleted, err := r.client.Indexes(context.Background(), []string{"code", "subject"}, []string{"id"})
	if err != nil {
		return err
	}
	if len(added) > 0 || len(deleted) > 0 {
		log.Infof("mongo: authRequest: index: deleted: %v, created: %v", deleted, added)
	}
	return nil
}

func (r *Mongo) FindByID(ctx context.Context, id2 RequestID) (*Request, error) {
	return r.findOne(ctx, bson.M{"id": id2.String()})
}

func (r *Mongo) FindByCode(ctx context.Context, s string) (*Request, error) {
	return r.findOne(ctx, bson.M{"code": s})
}

func (r *Mongo) FindBySubject(ctx context.Context, s string) (*Request, error) {
	return r.findOne(ctx, bson.M{"subject": s})
}

func (r *Mongo) Save(ctx context.Context, request *Request) error {
	doc, id1 := newMongoDocument(request)
	return r.client.SaveOne(ctx, id1, doc)
}

func (r *Mongo) Remove(ctx context.Context, requestID RequestID) error {
	return r.client.RemoveOne(ctx, bson.M{"id": requestID.String()})
}

func (r *Mongo) findOne(ctx context.Context, filter any) (*Request, error) {
	c := newMongoConsumer()
	if err := r.client.FindOne(ctx, filter, c); err != nil {
		return nil, err
	}
	return c.Result[0], nil
}

func newMongoConsumer() *mongox.SliceFuncConsumer[*mongoDocument, *Request] {
	return mongox.NewSliceFuncConsumer(func(d *mongoDocument) (*Request, error) {
		return d.Model()
	})
}

type mongoDocument struct {
	ID            string                      `bson:"id"`
	ClientID      string                      `bson:"clientid"`
	Subject       string                      `bson:"subject"`
	Code          string                      `bson:"code"`
	State         string                      `bson:"state"`
	ResponseType  string                      `bson:"responsetype"`
	Scopes        []string                    `bson:"scopes"`
	Audiences     []string                    `bson:"audiences"`
	RedirectURI   string                      `bson:"redirecturi"`
	Nonce         string                      `bson:"nonce"`
	CodeChallenge *mongoCodeChallengeDocument `bson:"codechallenge"`
	AuthorizedAt  *time.Time                  `bson:"authorizedat"`
}

type mongoCodeChallengeDocument struct {
	Challenge string
	Method    string
}

func newMongoDocument(req *Request) (*mongoDocument, string) {
	if req == nil {
		return nil, ""
	}
	reqID := req.GetID()
	var cc *mongoCodeChallengeDocument
	if req.GetCodeChallenge() != nil {
		cc = &mongoCodeChallengeDocument{
			Challenge: req.GetCodeChallenge().Challenge,
			Method:    string(req.GetCodeChallenge().Method),
		}
	}
	return &mongoDocument{
		ID:            reqID,
		ClientID:      req.GetClientID(),
		Subject:       req.GetSubject(),
		Code:          req.GetCode(),
		State:         req.GetState(),
		ResponseType:  string(req.GetResponseType()),
		Scopes:        req.GetScopes(),
		Audiences:     req.GetAudience(),
		RedirectURI:   req.GetRedirectURI(),
		Nonce:         req.GetNonce(),
		CodeChallenge: cc,
		AuthorizedAt:  req.AuthorizedAt(),
	}, reqID
}

func (d *mongoDocument) Model() (*Request, error) {
	if d == nil {
		return nil, nil
	}

	ulid, err := RequestIDFrom(d.ID)
	if err != nil {
		return nil, err
	}

	var cc *oidc.CodeChallenge
	if d.CodeChallenge != nil {
		cc = &oidc.CodeChallenge{
			Challenge: d.CodeChallenge.Challenge,
			Method:    oidc.CodeChallengeMethod(d.CodeChallenge.Method),
		}
	}

	var req = NewRequest().
		ID(ulid).
		ClientID(d.ClientID).
		Subject(d.Subject).
		Code(d.Code).
		State(d.State).
		ResponseType(oidc.ResponseType(d.ResponseType)).
		Scopes(d.Scopes).
		Audiences(d.Audiences).
		RedirectURI(d.RedirectURI).
		Nonce(d.Nonce).
		CodeChallenge(cc).
		AuthorizedAt(d.AuthorizedAt).
		MustBuild()
	return req, nil
}
