package schema

import (
	"errors"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/value"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
)

var ErrInvalidTitleField = errors.New("title field must be one of schema fields")

type Schema struct {
	titleField *FieldID
	fields     []*Field
	id         ID
	project    ProjectID
	workspace  accountdomain.WorkspaceID
}

func (s *Schema) ID() ID {
	return s.id
}

func (s *Schema) Workspace() accountdomain.WorkspaceID {
	return s.workspace
}

func (s *Schema) Project() ProjectID {
	return s.project
}

func (s *Schema) SetWorkspace(workspace accountdomain.WorkspaceID) {
	s.workspace = workspace
}

func (s *Schema) ReferencedSchemas() IDList {
	return lo.Map(s.FieldsByType(value.TypeReference), func(f *Field, _ int) ID {
		var sID ID
		f.TypeProperty().Match(TypePropertyMatch{
			Reference: func(rf *FieldReference) {
				sID = rf.Schema()
			},
		})
		return sID
	})
}

func (s *Schema) Groups() GroupIDList {
	return lo.Map(s.FieldsByType(value.TypeGroup), func(f *Field, _ int) id.GroupID {
		var gID id.GroupID
		f.TypeProperty().Match(TypePropertyMatch{
			Group: func(f *FieldGroup) {
				gID = f.Group()
			},
		})
		return gID
	})
}

func (s *Schema) HasField(f FieldID) bool {
	return lo.SomeBy(s.fields, func(g *Field) bool { return g.ID() == f })
}

func (s *Schema) HasFieldByKey(k string) bool {
	return lo.SomeBy(s.fields, func(g *Field) bool { return g.Key().String() == k })
}

func (s *Schema) AddField(f *Field) {
	if s.HasField(f.ID()) {
		return
	}
	if s.Fields().Count() == 0 {
		f.order = 0
	} else {
		// get the biggest order
		f.order = s.Fields().Ordered()[s.Fields().Count()-1].Order() + 1
	}
	s.fields = append(s.fields, f)
}

func (s *Schema) Field(fId FieldID) *Field {
	f, _ := lo.Find(s.fields, func(f *Field) bool { return f.id == fId })
	return f
}

func (s *Schema) FieldByIDOrKey(fId *FieldID, key *id.Key) *Field {
	if s == nil || s.fields == nil {
		return nil
	}
	f, _ := lo.Find(s.fields, func(f *Field) bool {
		return fId != nil && f.id == *fId || key != nil && key.IsValid() && f.key == *key
	})
	return f
}

func (s *Schema) Fields() FieldList {
	var fl FieldList = slices.Clone(s.fields)
	return fl.Ordered()
}

func (s *Schema) FieldsByType(t value.Type) FieldList {
	return lo.Filter(s.Fields(), func(f *Field, _ int) bool {
		return f.Type() == t
	})
}

func (s *Schema) RemoveField(fid FieldID) {
	for i, field := range s.fields {
		if field.id == fid {
			s.fields = slices.Delete(s.fields, i, i+1)
			if lo.FromPtr(s.titleField) == fid {
				s.titleField = nil
			}
			return
		}
	}
}

func (s *Schema) TitleField() *FieldID {
	if s.Fields() == nil || len(s.Fields()) == 0 {
		return nil
	}
	return s.titleField.CloneRef()
}

func (s *Schema) SetTitleField(tf *FieldID) error {
	// unsetting title
	if tf == nil {
		s.titleField = nil
		return nil
	}

	if !s.HasField(*tf) || s.Fields() == nil || len(s.Fields()) == 0 {
		return ErrInvalidTitleField
	}
	s.titleField = tf.CloneRef()
	return nil
}

func (s *Schema) Clone() *Schema {
	if s == nil {
		return nil
	}

	return &Schema{
		id:         s.ID(),
		project:    s.Project().Clone(),
		workspace:  s.Workspace().Clone(),
		fields:     slices.Clone(s.fields),
		titleField: s.TitleField().CloneRef(),
	}
}

func (s *Schema) HasGeometryFields() bool {
	if s == nil {
		return false
	}
	return len(s.FieldsByType(value.TypeGeometryObject)) > 0 ||
		len(s.FieldsByType(value.TypeGeometryEditor)) > 0
}

func (s *Schema) IsPointFieldSupported() bool {
	if s == nil {
		return false
	}
	for _, f := range s.Fields() {
		if f.SupportsPointField() {
			return true
		}
	}
	return false
}

func (s *Schema) CopyFrom(s2 *Schema) {
	if s == nil || s2 == nil {
		return
	}
	s.fields = slices.Clone(s2.fields)
	s.titleField = s2.TitleField().CloneRef()
}
