package workspace

import "maps"

// ScimConfig is the SCIM provisioning configuration of a workspace. It is managed by reearth-accounts;
// reearthx only carries it so that saving a workspace does not erase it.
type ScimConfig struct {
	Enabled          bool
	GroupRoleMapping map[string]Role
	TokenHash        string
}

func (c *ScimConfig) Clone() *ScimConfig {
	if c == nil {
		return nil
	}
	return &ScimConfig{
		Enabled:          c.Enabled,
		GroupRoleMapping: maps.Clone(c.GroupRoleMapping),
		TokenHash:        c.TokenHash,
	}
}
