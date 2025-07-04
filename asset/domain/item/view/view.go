package view

import (
	"time"

	"github.com/samber/lo"
)

type View struct {
	updatedAt time.Time
	sort      *Sort
	filter    *Condition
	columns   *ColumnList
	name      string
	order     int
	id        ID
	schema    SchemaID
	model     ModelID
	project   ProjectID
	user      UserID
}

func (v *View) Model() ModelID {
	return v.model
}

func (v *View) ID() ID {
	return v.id
}

func (g *View) Order() int {
	return g.order
}

func (g *View) SetOrder(order int) {
	g.order = order
}

func (v *View) Clone() *View {
	if v == nil {
		return nil
	}
	return &View{
		id:        v.id.Clone(),
		name:      v.name,
		schema:    v.schema.Clone(),
		model:     v.model.Clone(),
		project:   v.project.Clone(),
		sort:      v.sort,
		filter:    v.filter,
		columns:   v.columns,
		order:     v.order,
		user:      v.user.Clone(),
		updatedAt: lo.FromPtr(&v.updatedAt),
	}
}

func (v *View) Project() ProjectID {
	return v.project
}

func (v *View) SetName(name string) {
	v.name = name
}

func (v *View) SetFilter(condition *Condition) {
	v.filter = condition
}

func (v *View) SetSort(sort *Sort) {
	v.sort = sort
}

func (v *View) SetColumns(columns *ColumnList) {
	v.columns = columns
}

func (v *View) SetUpdatedAt(now time.Time) {
	v.updatedAt = now
}

func (v *View) Name() string {
	return v.name
}

func (v *View) Sort() *Sort {
	return v.sort
}

func (v *View) Columns() *ColumnList {
	return v.columns
}

func (v *View) Filter() *Condition {
	return v.filter
}

func (v *View) User() UserID {
	return v.user
}

func (v *View) Schema() SchemaID {
	return v.schema
}

func (v *View) UpdatedAt() time.Time {
	return v.updatedAt
}
