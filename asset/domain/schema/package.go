package schema

import (
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/samber/lo"
)

type Package struct {
	schema            *Schema
	metaSchema        *Schema
	groupSchemas      map[id.GroupID]*Schema
	referencedSchemas List
}

func NewPackage(
	s *Schema,
	meta *Schema,
	groupSchemas map[id.GroupID]*Schema,
	referencedSchemas List,
) *Package {
	return &Package{
		schema:            s,
		metaSchema:        meta,
		groupSchemas:      groupSchemas,
		referencedSchemas: referencedSchemas,
	}
}

func (p *Package) Schema() *Schema {
	if p == nil {
		return nil
	}
	return p.schema
}

func (p *Package) MetaSchema() *Schema {
	if p == nil {
		return nil
	}
	return p.metaSchema
}

func (p *Package) GroupSchemas() List {
	if p == nil {
		return nil
	}
	return lo.FilterMap(lo.Values(p.groupSchemas), func(s *Schema, _ int) (*Schema, bool) {
		if s == nil {
			return nil, false
		}
		return s, true
	})
}

func (p *Package) GroupSchema(gid id.GroupID) *Schema {
	if p == nil || p.groupSchemas == nil {
		return nil
	}
	s, ok := p.groupSchemas[gid]
	if !ok {
		return nil
	}
	return s
}

func (p *Package) ReferencedSchemas() List {
	if p == nil {
		return nil
	}
	return p.referencedSchemas
}

func (p *Package) ReferencedSchema(fieldID id.FieldID) *Schema {
	if p == nil {
		return nil
	}
	f := p.schema.Field(fieldID)
	if f == nil {
		return nil
	}
	return p.referencedSchemas.Schema(f.TypeProperty().reference.Schema().Ref())
}

func (p *Package) Field(fieldID id.FieldID) *Field {
	if p == nil {
		return nil
	}
	f := p.schema.Field(fieldID)
	if f != nil {
		return f
	}
	f = p.metaSchema.Field(fieldID)
	if f != nil {
		return f
	}
	for _, s := range p.groupSchemas {
		f = s.Field(fieldID)
		if f != nil {
			return f
		}
	}
	return nil
}

func (p *Package) FieldByIDOrKey(fID *id.FieldID, k *id.Key) *Field {
	if p == nil {
		return nil
	}
	f := p.schema.FieldByIDOrKey(fID, k)
	if f != nil {
		return f
	}
	f = p.metaSchema.FieldByIDOrKey(fID, k)
	if f != nil {
		return f
	}
	for _, s := range p.groupSchemas {
		f = s.FieldByIDOrKey(fID, k)
		if f != nil {
			return f
		}
	}
	return nil
}
