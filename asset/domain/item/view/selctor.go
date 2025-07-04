package view

import "github.com/reearth/reearthx/asset/domain/id"

type FieldType string

const (
	FieldTypeId               FieldType = "ID"
	FieldTypeCreationDate     FieldType = "CREATIONDATE"
	FieldTypeCreationUser     FieldType = "CREATIONUSER"
	FieldTypeModificationDate FieldType = "MODIFICATIONDATE"
	FieldTypeModificationUser FieldType = "MODIFICATIONUSER"
	FieldTypeStatus           FieldType = "STATUS"

	FieldTypeField     FieldType = "FIELD"
	FieldTypeMetaField FieldType = "METAFIELD"
)

type FieldSelector struct {
	ID   *id.FieldID
	Type FieldType
}

type FieldSelectorList []FieldSelector
