package mongox

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// Tx implements usecasex.Tx, but note that it's not goroutine-safe.
type Tx struct {
	ctx     mongo.SessionContext
	session mongo.Session
	commit  bool
}

func newTx(ctx context.Context, session mongo.Session) *Tx {
	return &Tx{
		ctx:     mongo.NewSessionContext(ctx, session),
		session: session,
		commit:  false,
	}
}

func (t *Tx) Context() context.Context {
	return t.ctx
}

func (t *Tx) Commit() {
	if t == nil {
		return
	}
	t.commit = true
}

func (t *Tx) End(ctx context.Context) error {
	if t == nil {
		return nil
	}

	// the session must be returned to the pool even when the commit or the
	// abort fails, otherwise the server holds its locks until the transaction
	// lifetime limit expires
	defer t.session.EndSession(ctx)

	if t.commit {
		return t.session.CommitTransaction(ctx)
	}
	return t.session.AbortTransaction(ctx)
}

func (t *Tx) IsCommitted() bool {
	return t.commit
}
