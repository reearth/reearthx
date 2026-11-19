package mailer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCloudflare(t *testing.T, h http.HandlerFunc) *cloudflareMailer {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		h(w, r)
	}))
	t.Cleanup(srv.Close)
	return NewCloudflare("Re:Earth", "noreply@example.com", "acc123", "token", option.WithBaseURL(srv.URL+"/")).(*cloudflareMailer)
}

func TestCloudflare_SendMail(t *testing.T) {
	var reqs []map[string]any
	m := newTestCloudflare(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/accounts/acc123/email/sending/send", r.URL.Path)
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		reqs = append(reqs, body)

		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"delivered":["` + body["to"].(string) + `"],"permanent_bounces":[],"queued":[]}}`))
	})

	err := m.SendMail(context.Background(), []Contact{
		{Email: "a@example.com", Name: "A"},
		{Email: "b@example.com"},
	}, "subject", "plain", "<p>html</p>")
	require.NoError(t, err)

	require.Len(t, reqs, 2)
	assert.Equal(t, map[string]any{
		"from":    map[string]any{"address": "noreply@example.com", "name": "Re:Earth"},
		"to":      "a@example.com",
		"subject": "subject",
		"text":    "plain",
		"html":    "<p>html</p>",
	}, reqs[0])
	assert.Equal(t, "b@example.com", reqs[1]["to"])
}

func TestCloudflare_SendMail_APIError(t *testing.T) {
	m := newTestCloudflare(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"success":false,"errors":[{"code":10001,"message":"email.sending.error.invalid_request_schema"}],"messages":[],"result":null}`))
	})

	err := m.SendMail(context.Background(), []Contact{{Email: "to@example.com"}}, "s", "p", "h")
	assert.ErrorContains(t, err, "400")
	assert.ErrorContains(t, err, "10001")
}

func TestCloudflare_SendMail_NonJSONError(t *testing.T) {
	m := newTestCloudflare(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway"))
	})

	err := m.SendMail(context.Background(), []Contact{{Email: "to@example.com"}}, "s", "p", "h")
	assert.ErrorContains(t, err, "502")
}

func TestCloudflare_SendMail_PermanentBounce(t *testing.T) {
	m := newTestCloudflare(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"delivered":[],"permanent_bounces":["to@example.com"],"queued":[]}}`))
	})

	err := m.SendMail(context.Background(), []Contact{{Email: "to@example.com"}}, "s", "p", "h")
	assert.ErrorContains(t, err, "bounced")
}

func TestCloudflare_SendMail_InvalidEmail(t *testing.T) {
	called := false
	m := newTestCloudflare(t, func(w http.ResponseWriter, _ *http.Request) {
		called = true
	})

	err := m.SendMail(context.Background(), []Contact{{Email: "not-an-email"}}, "s", "p", "h")
	assert.Error(t, err)
	assert.False(t, called)
}

func TestCloudflare_SendMail_HonorsContext(t *testing.T) {
	m := newTestCloudflare(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := m.SendMail(ctx, []Contact{{Email: "to@example.com"}}, "s", "p", "h")
	assert.Error(t, err)
	assert.Less(t, time.Since(start), 2*time.Second)
}

// A failed send must not be retried: the first attempt may already have been delivered.
func TestCloudflare_SendMail_NoRetry(t *testing.T) {
	calls := 0
	m := newTestCloudflare(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	err := m.SendMail(context.Background(), []Contact{{Email: "to@example.com"}}, "s", "p", "h")
	assert.ErrorContains(t, err, "503")
	assert.Equal(t, 1, calls)
}
