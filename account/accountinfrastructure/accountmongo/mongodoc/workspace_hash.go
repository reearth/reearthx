package mongodoc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// The workspace collection is shared with reearth-accounts, which has a unique index on
// (alias, members_hash) with a case-insensitive collation. The hash below must stay byte-identical to
// reearth-accounts server/internal/infrastructure/mongo/mongodoc.ComputeWorkspaceMembersHash.
// Do not change the struct field order, JSON tags or the sort.

type workspaceMemberForHash struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	InvitedBy string `json:"invited_by"`
	Disabled  bool   `json:"disabled"`
	Type      string `json:"type"` // "user" or "integration"
}

// ComputeWorkspaceMembersHash returns the members_hash of a workspace in the format reearth-accounts uses.
// Member external IDs are not part of the hash.
func ComputeWorkspaceMembersHash(members, integrations map[string]WorkspaceMemberDocument) (string, error) {
	// nil when there are no members: reearth-accounts hashes the JSON "null" in that case.
	var all []workspaceMemberForHash

	for id, m := range members {
		all = append(all, workspaceMemberForHash{
			ID:        id,
			Role:      m.Role,
			InvitedBy: m.InvitedBy,
			Disabled:  m.Disabled,
			Type:      "user",
		})
	}

	for id, m := range integrations {
		all = append(all, workspaceMemberForHash{
			ID:        id,
			Role:      m.Role,
			InvitedBy: m.InvitedBy,
			Disabled:  m.Disabled,
			Type:      "integration",
		})
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].ID < all[j].ID
	})

	jsonData, err := json.Marshal(all)
	if err != nil {
		return "", fmt.Errorf("failed to marshal members for hash: %w", err)
	}
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}
