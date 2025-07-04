package model

import (
	"fmt"
	"time"

	"github.com/reearth/reearthx/asset/domain/id"

	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"golang.org/x/exp/slices"
)

var (
	ErrInvalidKey = rerror.NewE(i18n.T("invalid key"))
	ngKeys        = []string{"assets", "schemas", "models", "items"}
)

type Model struct {
	updatedAt   time.Time
	metadata    *SchemaID
	name        string
	description string
	key         id.Key
	order       int
	id          ID
	project     ProjectID
	schema      SchemaID
	public      bool
}

func (p *Model) ID() ID {
	return p.id
}

func (p *Model) Schema() SchemaID {
	return p.schema
}

func (p *Model) Metadata() *SchemaID {
	return p.metadata
}

func (p *Model) Project() ProjectID {
	return p.project
}

func (p *Model) Name() string {
	return p.name
}

func (p *Model) SetName(name string) {
	p.name = name
}

func (p *Model) SetMetadata(id id.SchemaID) {
	p.metadata = id.Ref()
}

func (p *Model) Description() string {
	return p.description
}

func (p *Model) SetDescription(description string) {
	p.description = description
}

func (p *Model) Key() id.Key {
	return p.key
}

func (p *Model) SetKey(key id.Key) error {
	if !validateModelKey(key) {
		return &rerror.Error{
			Label: ErrInvalidKey,
			Err:   fmt.Errorf("%s", key.String()),
		}
	}
	p.key = key
	return nil
}

func (p *Model) Public() bool {
	return p.public
}

func (p *Model) SetPublic(public bool) {
	p.public = public
}

func (p *Model) UpdatedAt() time.Time {
	if p.updatedAt.IsZero() {
		return p.id.Timestamp()
	}
	return p.updatedAt
}

func (p *Model) SetUpdatedAt(updatedAt time.Time) {
	p.updatedAt = updatedAt
}

func (p *Model) CreatedAt() time.Time {
	return p.id.Timestamp()
}

func (p *Model) Order() int {
	return p.order
}

func (p *Model) SetOrder(order int) {
	p.order = order
}

func (p *Model) Clone() *Model {
	if p == nil {
		return nil
	}

	return &Model{
		id:          p.id.Clone(),
		project:     p.project.Clone(),
		schema:      p.schema.Clone(),
		name:        p.name,
		description: p.description,
		key:         p.Key(),
		public:      p.public,
		updatedAt:   p.updatedAt,
		order:       p.order,
	}
}

func validateModelKey(k id.Key) bool {
	// assets is used as an API endpoint
	return k.IsURLCompatible() && len(k.String()) > 2 && !slices.Contains(ngKeys, k.String())
}
