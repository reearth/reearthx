package mongodoc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Expected values were computed with reearth-accounts
// mongodoc.ComputeWorkspaceMembersHash (server @ 305a60c).
func TestComputeWorkspaceMembersHash(t *testing.T) {
	const (
		a  = "01hz7k3q8v9w2x4y6z8a0b2c4d"
		c  = "01hz7k3q8v9w2x4y6z8a0b2c4f"
		i1 = "01hz7k3q8v9w2x4y6z8a0b2c4g"
	)

	tests := []struct {
		name                  string
		members, integrations map[string]WorkspaceMemberDocument
		want                  string
	}{
		{
			name: "users and integrations; external ID is ignored",
			members: map[string]WorkspaceMemberDocument{
				a: {Role: "owner", InvitedBy: a},
				c: {Role: "writer", InvitedBy: a},
			},
			integrations: map[string]WorkspaceMemberDocument{i1: {Role: "writer", InvitedBy: a}},
			want:         "8ccb595518129aa289892b3932537958963471f10792568d3bddd45998f04089",
		},
		{
			name:         "empty maps hash JSON null",
			members:      map[string]WorkspaceMemberDocument{},
			integrations: map[string]WorkspaceMemberDocument{},
			want:         "74234e98afe7498fb5daf1f36ac2d78acc339464f950703b8c019892f982b90b",
		},
		{
			name:    "nil integrations, disabled member",
			members: map[string]WorkspaceMemberDocument{a: {Role: "owner", InvitedBy: a, Disabled: true}},
			want:    "5362eb62f948d470418c2b3bc9ab751233a9a98c7be851d8c7d133fac9f9ff7b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ComputeWorkspaceMembersHash(tt.members, tt.integrations)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
