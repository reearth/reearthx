package authserver

import (
	"context"

	"github.com/zitadel/oidc/pkg/oidc"
)

type UserRepo interface {
	Sub(ctx context.Context, email, password, authRequestID string) (string, error)
	Info(context.Context, string, []string, oidc.UserInfoSetter) error
}

type RequestRepo interface {
	FindByID(context.Context, RequestID) (*Request, error)
	FindByCode(context.Context, string) (*Request, error)
	FindBySubject(context.Context, string) (*Request, error)
	Save(context.Context, *Request) error
	Remove(context.Context, RequestID) error
}

type TokenRepo interface {
	FindByToken(context.Context, string) (*Token, error)
	Save(context.Context, *Token) error
	Remove(context.Context, string) error
}

type Config struct {
	Cert string
	Key  string
}

type ConfigRepo interface {
	Load(context.Context) (*Config, error)
	Save(context.Context, *Config) error
	Unlock(context.Context) error
}
