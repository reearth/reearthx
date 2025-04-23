package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/reearth/reearthx/appx"
	"github.com/stretchr/testify/assert"
)

func TestClient_NewClient(t *testing.T) {
	dashboardURL := "http://test-dashboard"
	client := NewClient(dashboardURL)

	assert.NotNil(t, client)
	assert.Equal(t, dashboardURL, client.dashboardURL)
	assert.NotNil(t, client.httpClient)
}

func TestClient_CheckPermission(t *testing.T) {
	tests := []struct {
		name         string
		authInfo     *appx.AuthInfo
		input        CheckPermissionInput
		serverStatus int
		serverResp   CheckPermissionResponse
		wantAllowed  bool
		wantErr      string
	}{
		{
			name: "success - permission allowed",
			authInfo: &appx.AuthInfo{
				Token: "test-token",
			},
			input: CheckPermissionInput{
				Service:  "flow",
				Resource: "project",
				Action:   "read",
			},
			serverStatus: http.StatusOK,
			serverResp: CheckPermissionResponse{
				Data: struct {
					CheckPermission struct {
						Allowed bool "json:\"allowed\""
					} "json:\"checkPermission\""
				}{
					CheckPermission: struct {
						Allowed bool "json:\"allowed\""
					}{
						Allowed: true,
					},
				},
			},
			wantAllowed: true,
		},
		{
			name: "success - permission denied",
			authInfo: &appx.AuthInfo{
				Token: "test-token",
			},
			input: CheckPermissionInput{
				Service:  "flow",
				Resource: "project",
				Action:   "write",
			},
			serverStatus: http.StatusOK,
			serverResp: CheckPermissionResponse{
				Data: struct {
					CheckPermission struct {
						Allowed bool "json:\"allowed\""
					} "json:\"checkPermission\""
				}{
					CheckPermission: struct {
						Allowed bool "json:\"allowed\""
					}{
						Allowed: false,
					},
				},
			},
			wantAllowed: false,
		},
		{
			name:     "error - nil auth info",
			authInfo: nil,
			input: CheckPermissionInput{
				Service:  "flow",
				Resource: "project",
				Action:   "read",
			},
			wantErr: "auth info is required",
		},
		{
			name: "error - server error",
			authInfo: &appx.AuthInfo{
				Token: "test-token",
			},
			input: CheckPermissionInput{
				Service:  "flow",
				Resource: "project",
				Action:   "read",
			},
			serverStatus: http.StatusInternalServerError,
			wantErr:      fmt.Sprint("server returned non-OK status: ", http.StatusInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/api/graphql", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				if tt.authInfo != nil {
					assert.Equal(t, "Bearer "+tt.authInfo.Token, r.Header.Get("Authorization"))
				}

				var gqlRequest GraphQLQuery
				err := json.NewDecoder(r.Body).Decode(&gqlRequest)
				assert.NoError(t, err)
				assert.Contains(t, gqlRequest.Query, "query CheckPermission")
				assert.Contains(t, gqlRequest.Query, "checkPermission")

				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					if err := json.NewEncoder(w).Encode(tt.serverResp); err != nil {
						t.Fatalf("failed to encode response: %v", err)
						return
					}
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			allowed, err := client.CheckPermission(context.Background(), tt.authInfo, tt.input)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantAllowed, allowed)
		})
	}
}
