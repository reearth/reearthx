package domain

import "errors"

var (
	// ErrEmptyWorkspaceID Asset errors
	ErrEmptyWorkspaceID = errors.New("workspace id is required")
	ErrEmptyURL         = errors.New("url is required")
	ErrEmptySize        = errors.New("size must be greater than 0")
	ErrAssetNotFound    = errors.New("asset not found")
	ErrInvalidAsset     = errors.New("invalid asset")

	// ErrEmptyGroupName Group errors
	ErrEmptyGroupName = errors.New("group name is required")
	ErrEmptyPolicy    = errors.New("policy is required")
	ErrGroupNotFound  = errors.New("group not found")
	ErrInvalidGroup   = errors.New("invalid group")

	// ErrUploadFailed Storage errors
	ErrUploadFailed   = errors.New("failed to upload asset")
	ErrDownloadFailed = errors.New("failed to download asset")
	ErrDeleteFailed   = errors.New("failed to delete asset")

	// ErrExtractionFailed Extraction errors
	ErrExtractionFailed = errors.New("failed to extract asset")
	ErrNotExtractable   = errors.New("asset is not extractable")
)
