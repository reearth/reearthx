package appx

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

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

	res := lo.Must(http.Get(ts.URL))
	body := string(lo.Must(io.ReadAll(res.Body)))
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

	res := lo.Must(http.Get(ts.URL))
	body := string(lo.Must(io.ReadAll(res.Body)))
	assert.Equal(t, "", body)

	res = lo.Must(http.Post(ts.URL, "", nil))
	body = string(lo.Must(io.ReadAll(res.Body)))
	assert.Equal(t, "aaa", body)
}

func TestRequestIDMiddleware(t *testing.T) {
	ts := httptest.NewServer(RequestIDMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(GetRequestID(r.Context())))
	})))
	defer ts.Close()

	req := lo.Must(http.NewRequest(http.MethodGet, ts.URL, nil))
	req.Header.Set("X-Request-ID", "aaa")
	res := lo.Must(http.DefaultClient.Do(req))
	body := string(lo.Must(io.ReadAll(res.Body)))
	assert.Equal(t, "aaa", body)

	req = lo.Must(http.NewRequest(http.MethodGet, ts.URL, nil))
	req.Header.Set("x-cloud-trace-context", "xxx")
	res = lo.Must(http.DefaultClient.Do(req))
	body = string(lo.Must(io.ReadAll(res.Body)))
	assert.Equal(t, "xxx", body)
}

func TestGetAuthInfo(t *testing.T) {
	trueValue := true
	falseValue := false
	type authInfoKey struct{}
	type otherKey struct{}

	testCases := []struct {
		name     string
		ctx      context.Context
		key      any
		expected *AuthInfo
	}{
		{
			name: "Valid auth info",
			ctx: context.WithValue(context.Background(), authInfoKey{}, &AuthInfo{
				Token:         "test_token",
				Sub:           "test_sub",
				Iss:           "test_iss",
				Name:          "test_name",
				Email:         "test_email",
				EmailVerified: &trueValue,
			}),
			key: authInfoKey{},
			expected: &AuthInfo{
				Token:         "test_token",
				Sub:           "test_sub",
				Iss:           "test_iss",
				Name:          "test_name",
				Email:         "test_email",
				EmailVerified: &trueValue,
			},
		},
		{
			name:     "Nil context",
			ctx:      nil,
			key:      "auth_key",
			expected: nil,
		},
		{
			name: "Invalid key",
			ctx: context.WithValue(context.Background(), otherKey{}, &AuthInfo{
				Token:         "wrong_token",
				Sub:           "wrong_sub",
				Iss:           "wrong_iss",
				Name:          "wrong_name",
				Email:         "wrong_email",
				EmailVerified: &falseValue,
			}),
			key:      authInfoKey{},
			expected: nil,
		},
		{
			name:     "Value is not *AuthInfo",
			ctx:      context.WithValue(context.Background(), authInfoKey{}, "not_auth_info"),
			key:      authInfoKey{},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetAuthInfo(tc.ctx, tc.key)

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Test case '%s' failed: expected %+v, got %+v", tc.name, tc.expected, result)
			}
		})
	}
}
