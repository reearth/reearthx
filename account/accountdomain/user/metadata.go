package user

type Metadata struct {
	description string
	website     string
}

func NewMetadata() *Metadata {
	return &Metadata{}
}

func MetadataFrom(description, website string) *Metadata {
	return &Metadata{
		description: description,
		website:     website,
	}
}

func (m *Metadata) Description() string {
	return m.description
}

func (m *Metadata) Website() string {
	return m.website
}
