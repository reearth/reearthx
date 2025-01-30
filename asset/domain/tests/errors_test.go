package domain_test

import (
	"testing"

	"github.com/reearth/reearthx/asset/domain"
	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrEmptyWorkspaceID",
			err:  domain.ErrEmptyWorkspaceID,
			want: "workspace id is required",
		},
		{
			name: "ErrEmptyURL",
			err:  domain.ErrEmptyURL,
			want: "url is required",
		},
		{
			name: "ErrEmptySize",
			err:  domain.ErrEmptySize,
			want: "size must be greater than 0",
		},
		{
			name: "ErrAssetNotFound",
			err:  domain.ErrAssetNotFound,
			want: "asset not found",
		},
		{
			name: "ErrInvalidAsset",
			err:  domain.ErrInvalidAsset,
			want: "invalid asset",
		},
		{
			name: "ErrEmptyGroupName",
			err:  domain.ErrEmptyGroupName,
			want: "group name is required",
		},
		{
			name: "ErrEmptyPolicy",
			err:  domain.ErrEmptyPolicy,
			want: "policy is required",
		},
		{
			name: "ErrGroupNotFound",
			err:  domain.ErrGroupNotFound,
			want: "group not found",
		},
		{
			name: "ErrInvalidGroup",
			err:  domain.ErrInvalidGroup,
			want: "invalid group",
		},
		{
			name: "ErrUploadFailed",
			err:  domain.ErrUploadFailed,
			want: "failed to upload asset",
		},
		{
			name: "ErrDownloadFailed",
			err:  domain.ErrDownloadFailed,
			want: "failed to download asset",
		},
		{
			name: "ErrDeleteFailed",
			err:  domain.ErrDeleteFailed,
			want: "failed to delete asset",
		},
		{
			name: "ErrExtractionFailed",
			err:  domain.ErrExtractionFailed,
			want: "failed to extract asset",
		},
		{
			name: "ErrNotExtractable",
			err:  domain.ErrNotExtractable,
			want: "asset is not extractable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}
