package model

import (
	"time"

	"github.com/99designs/gqlgen/graphql"
)

type Node interface {
	IsNode()
	GetID() string
}

type PreviewType string

const (
	PreviewTypeImage      PreviewType = "IMAGE"
	PreviewTypeImageSvg   PreviewType = "IMAGE_SVG"
	PreviewTypeGeo        PreviewType = "GEO"
	PreviewTypeGeo3DTiles PreviewType = "GEO_3D_TILES"
	PreviewTypeGeoMvt     PreviewType = "GEO_MVT"
	PreviewTypeModel3D    PreviewType = "MODEL_3D"
	PreviewTypeCsv        PreviewType = "CSV"
	PreviewTypeUnknown    PreviewType = "UNKNOWN"
)

type ArchiveExtractionStatus string

const (
	ArchiveExtractionStatusSkipped    ArchiveExtractionStatus = "SKIPPED"
	ArchiveExtractionStatusPending    ArchiveExtractionStatus = "PENDING"
	ArchiveExtractionStatusInProgress ArchiveExtractionStatus = "IN_PROGRESS"
	ArchiveExtractionStatusDone       ArchiveExtractionStatus = "DONE"
	ArchiveExtractionStatusFailed     ArchiveExtractionStatus = "FAILED"
)

type OperatorType string

const (
	OperatorTypeUser        OperatorType = "USER"
	OperatorTypeIntegration OperatorType = "INTEGRATION"
)

type Asset struct {
	ID                      string                   `json:"id"`
	GroupID                 string                   `json:"groupId"`
	Group                   *Group                   `json:"group"`
	CreatedAt               time.Time                `json:"createdAt"`
	CreatedByType           OperatorType             `json:"createdByType"`
	CreatedByID             string                   `json:"createdById"`
	Size                    int                      `json:"size"`
	ContentType             string                   `json:"contentType"`
	ContentEncoding         *string                  `json:"contentEncoding"`
	PreviewType             *PreviewType             `json:"previewType"`
	UUID                    string                   `json:"uuid"`
	URL                     string                   `json:"url"`
	FileName                string                   `json:"fileName"`
	ArchiveExtractionStatus *ArchiveExtractionStatus `json:"archiveExtractionStatus"`
}

func (Asset) IsNode()         {}
func (a Asset) GetID() string { return a.ID }

type Group struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Policy    *Policy   `json:"policy"`
}

func (Group) IsNode()         {}
func (g Group) GetID() string { return g.ID }

type Policy struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	StorageLimit int       `json:"storageLimit"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (Policy) IsNode()         {}
func (p Policy) GetID() string { return p.ID }

type AssetFile struct {
	Name            string  `json:"name"`
	Size            int     `json:"size"`
	ContentType     string  `json:"contentType"`
	ContentEncoding *string `json:"contentEncoding"`
	Data            string  `json:"data"`
}

type CreateAssetInput struct {
	GroupID           string          `json:"groupId"`
	File              *graphql.Upload `json:"file"`
	URL               *string         `json:"url"`
	Token             *string         `json:"token"`
	SkipDecompression *bool           `json:"skipDecompression"`
	ContentEncoding   *string         `json:"contentEncoding"`
}

type UpdateAssetInput struct {
	ID          string       `json:"id"`
	PreviewType *PreviewType `json:"previewType"`
}
