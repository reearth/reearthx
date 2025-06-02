package integrationapi

import (
	"github.com/reearth/reearthx/asset/domain/group"
	"github.com/reearth/reearthx/asset/domain/schema"
)

func NewGroup(g *group.Group, s *schema.Schema) Group {
	return Group{
		Id:          g.ID(),
		Key:         g.Key().String(),
		Name:        g.Name(),
		Description: g.Description(),
		ProjectId:   g.Project(),
		SchemaId:    g.Schema(),
		Schema:      NewSchema(s),
	}
}
