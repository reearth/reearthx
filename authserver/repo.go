package authserver

import "context"

type RequestRepo interface {
	FindByID(context.Context, RequestID) (*Request, error)
	FindByCode(context.Context, string) (*Request, error)
	FindBySubject(context.Context, string) (*Request, error)
	Save(context.Context, *Request) error
	Remove(context.Context, RequestID) error
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
