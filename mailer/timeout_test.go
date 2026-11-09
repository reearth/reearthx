package mailer

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSMTP_SendMail_StalledServerHonorsContext(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()

	// Accept the connection and never send the SMTP greeting.
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = c.Close() }()
		time.Sleep(5 * time.Second)
	}()

	host, port, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)
	m := NewSMTP(host, port, "user", "from@example.com", "pass")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	err = m.SendMail(ctx, []Contact{{Email: "to@example.com"}}, "s", "p", "h")
	assert.Error(t, err)
	assert.Less(t, time.Since(start), 2*time.Second)
}

func TestSendGrid_SendMail_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"message":"invalid key"}]}`))
	}))
	defer srv.Close()

	m := NewSendGrid("name", "from@example.com", "key").(*sendgridMailer)
	m.client.BaseURL = srv.URL + "/v3/mail/send"

	err := m.SendMail(context.Background(), []Contact{{Email: "to@example.com"}}, "s", "p", "h")
	assert.ErrorContains(t, err, "401")
}
