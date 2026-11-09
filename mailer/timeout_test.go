package mailer

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
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

// A relay that rejects EHLO and accepts HELO must still get mail when no
// credentials are configured.
func TestSMTP_SendMail_HELOOnlyRelayWithoutCredentials(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()

	got := make(chan string, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = c.Close() }()
		r := bufio.NewReader(c)
		w := func(s string) { _, _ = c.Write([]byte(s + "\r\n")) }
		w("220 relay")
		var body strings.Builder
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			cmd := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(cmd, "EHLO"):
				w("502 not implemented")
			case strings.HasPrefix(cmd, "HELO"), strings.HasPrefix(cmd, "MAIL"), strings.HasPrefix(cmd, "RCPT"):
				w("250 ok")
			case cmd == "DATA":
				w("354 go ahead")
				for {
					l, err := r.ReadString('\n')
					if err != nil || l == ".\r\n" {
						break
					}
					body.WriteString(l)
				}
				w("250 queued")
				got <- body.String()
			case cmd == "QUIT":
				w("221 bye")
				return
			default:
				w("500 unknown")
			}
		}
	}()

	host, port, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)
	m := NewSMTP(host, port, "", "from@example.com", "")

	err = m.SendMail(context.Background(), []Contact{{Email: "to@example.com"}}, "subject", "plain", "html")
	require.NoError(t, err)
	assert.Contains(t, <-got, "subject")
}
