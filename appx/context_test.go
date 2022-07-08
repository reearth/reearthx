package appx

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/reearth/reearthx/util"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestContextMiddleware(t *testing.T) {
	key := struct{}{}
	value := lo.ToPtr("a")
	ts := httptest.NewServer(ContextMiddleware(key, value)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(*(r.Context().Value(key).(*string))))
	})))
	defer ts.Close()

	res := util.Unwrap(http.Get(ts.URL))
	body := string(util.Unwrap(io.ReadAll(res.Body)))
	assert.Equal(t, "a", body)
}

func TestContextMiddlewareBy(t *testing.T) {
	key := struct{}{}
	ts := httptest.NewServer(ContextMiddlewareBy(func(r *http.Request) context.Context {
		if r.Method == http.MethodPost {
			return context.WithValue(r.Context(), key, "aaa")
		}
		return nil
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v, ok := r.Context().Value(key).(string); ok {
			_, _ = w.Write([]byte(v))
		}
	})))
	defer ts.Close()

	res := util.Unwrap(http.Get(ts.URL))
	body := string(util.Unwrap(io.ReadAll(res.Body)))
	assert.Equal(t, "", body)

	res = util.Unwrap(http.Post(ts.URL, "", nil))
	body = string(util.Unwrap(io.ReadAll(res.Body)))
	assert.Equal(t, "aaa", body)
}
