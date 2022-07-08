package appx

import (
	"context"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/ravilushqa/otelgqlgen"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type GraphQLHandlerConfig struct {
	Schema          graphql.ExecutableSchema
	Dev             bool
	Context         func(r *http.Request) context.Context
	ComplexityLimit int
}

func GraphQLHandler(c GraphQLHandlerConfig) http.Handler {
	srv := handler.NewDefaultServer(c.Schema)
	srv.Use(otelgqlgen.Middleware())

	if c.ComplexityLimit > 0 {
		srv.Use(extension.FixedComplexityLimit(c.ComplexityLimit))
	}

	if c.Dev {
		srv.Use(extension.Introspection{})
	}

	srv.SetErrorPresenter(
		// show more detailed error messgage in debug mode
		func(ctx context.Context, e error) *gqlerror.Error {
			if c.Dev {
				return gqlerror.ErrorPathf(graphql.GetFieldContext(ctx).Path(), e.Error())
			}
			return graphql.DefaultErrorPresenter(ctx, e)
		},
	)

	return srv
}
