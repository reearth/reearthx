package mongox

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
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

	if t.commit {
		// Retry commit on UnknownTransactionCommitResult — the transaction may
		// have committed but the driver didn't receive confirmation.
		for {
			err := t.session.CommitTransaction(ctx)
			if err == nil {
				break
			}
			if !errorHasLabel(err, driver.UnknownTransactionCommitResult) {
				return err
			}
		}
	} else if err := t.session.AbortTransaction(ctx); err != nil {
		return err
	}

	t.session.EndSession(ctx)
	return nil
}

func (t *Tx) IsCommitted() bool {
	return t.commit
}
