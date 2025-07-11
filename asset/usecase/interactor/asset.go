package interactor

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/asset/domain/project"

	"github.com/reearth/reearthx/asset/domain/asset"
	"github.com/reearth/reearthx/asset/domain/event"
	"github.com/reearth/reearthx/asset/domain/file"
	"github.com/reearth/reearthx/asset/domain/id"
	"github.com/reearth/reearthx/asset/domain/task"
	"github.com/reearth/reearthx/asset/usecase"
	"github.com/reearth/reearthx/asset/usecase/gateway"
	"github.com/reearth/reearthx/asset/usecase/interfaces"
	"github.com/reearth/reearthx/asset/usecase/repo"

	"github.com/google/uuid"
	"github.com/reearth/reearthx/log"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/samber/lo"
)

// HostAdapter defines an interface for getting the current host from context
type HostAdapter interface {
	CurrentHost(ctx context.Context) string
}

// DefaultHostAdapter is a simple implementation of HostAdapter
// Example implementation:
/*
type DefaultHostAdapter struct {
	host string
}

func NewDefaultHostAdapter(host string) HostAdapter {
	return &DefaultHostAdapter{host: host}
}

func (h *DefaultHostAdapter) CurrentHost(ctx context.Context) string {
	return h.host
}
*/

type Asset struct {
	repos       *repo.Container
	gateways    *gateway.Container
	hostAdapter HostAdapter
	ignoreEvent bool
}

func NewAsset(r *repo.Container, g *gateway.Container) interfaces.Asset {
	return &Asset{
		repos:    r,
		gateways: g,
	}
}

func NewAssetWithHostAdapter(
	r *repo.Container,
	g *gateway.Container,
	hostAdapter HostAdapter,
) interfaces.Asset {
	return &Asset{
		repos:       r,
		gateways:    g,
		hostAdapter: hostAdapter,
	}
}

func (i *Asset) FindByID(
	ctx context.Context,
	aid id.AssetID,
	_ *usecase.Operator,
) (*asset.Asset, error) {
	a, err := i.repos.Asset.FindByID(ctx, aid)
	if err != nil {
		return nil, err
	}
	if a != nil {
		a.SetAccessInfoResolver(i.gateways.File.GetAccessInfoResolver())
	}
	return a, nil
}

func (i *Asset) FindByUUID(
	ctx context.Context,
	uuid string,
	_ *usecase.Operator,
) (*asset.Asset, error) {
	a, err := i.repos.Asset.FindByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	if a != nil {
		a.SetAccessInfoResolver(i.gateways.File.GetAccessInfoResolver())
	}
	return a, nil
}

func (i *Asset) FindByIDs(
	ctx context.Context,
	assets []id.AssetID,
	_ *usecase.Operator,
) (asset.List, error) {
	al, err := i.repos.Asset.FindByIDs(ctx, assets)
	if err != nil {
		return nil, err
	}
	if al != nil {
		al.SetAccessInfoResolver(i.gateways.File.GetAccessInfoResolver())
	}
	return al, nil
}

func (i *Asset) Search(
	ctx context.Context,
	projectID id.ProjectID,
	filter interfaces.AssetFilter,
	_ *usecase.Operator,
) (asset.List, *usecasex.PageInfo, error) {
	al, pi, err := i.repos.Asset.Search(ctx, projectID, repo.AssetFilter{
		Sort:         filter.Sort,
		Keyword:      filter.Keyword,
		Pagination:   filter.Pagination,
		ContentTypes: filter.ContentTypes,
	})
	if err != nil {
		return nil, nil, err
	}
	if al != nil {
		al.SetAccessInfoResolver(i.gateways.File.GetAccessInfoResolver())
	}
	return al, pi, nil
}

func (i *Asset) FindFileByID(
	ctx context.Context,
	aid id.AssetID,
	_ *usecase.Operator,
) (*asset.File, error) {
	_, err := i.repos.Asset.FindByID(ctx, aid)
	if err != nil {
		return nil, err
	}

	files, err := i.repos.AssetFile.FindByID(ctx, aid)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (i *Asset) FindFilesByIDs(
	ctx context.Context,
	ids id.AssetIDList,
	_ *usecase.Operator,
) (map[id.AssetID]*asset.File, error) {
	_, err := i.repos.Asset.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	files, err := i.repos.AssetFile.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (i *Asset) DownloadByID(
	ctx context.Context,
	aid id.AssetID,
	headers map[string]string,
	_ *usecase.Operator,
) (io.ReadCloser, map[string]string, error) {
	a, err := i.repos.Asset.FindByID(ctx, aid)
	if err != nil {
		return nil, nil, err
	}

	f, headers, err := i.gateways.File.ReadAsset(ctx, a.UUID(), a.FileName(), headers)
	if err != nil {
		return nil, nil, err
	}

	return f, headers, nil
}

func (i *Asset) Create(
	ctx context.Context,
	inp interfaces.CreateAssetParam,
	op *usecase.Operator,
) (result *asset.Asset, afile *asset.File, err error) {
	if op.AcOperator.User == nil && op.Integration == nil {
		return nil, nil, interfaces.ErrInvalidOperator
	}

	if inp.File == nil && inp.Token == "" {
		return nil, nil, interfaces.ErrFileNotIncluded
	}

	prj, err := i.repos.Project.FindByID(ctx, inp.ProjectID)
	if err != nil {
		return nil, nil, err
	}

	if !op.IsWritableWorkspace(prj.Workspace()) {
		return nil, nil, interfaces.ErrOperationDenied
	}

	var uuid string
	var file *file.File
	if inp.File != nil {
		if inp.File.ContentEncoding == "gzip" {
			inp.File.Name = strings.TrimSuffix(inp.File.Name, ".gz")
		}

		var size int64
		file = inp.File
		uuid, size, err = i.gateways.File.UploadAsset(ctx, inp.File)
		if err != nil {
			return nil, nil, err
		}

		file.Size = size
	}

	a, f, err := Run2(
		ctx, op, i.repos,
		Usecase().Transaction(),
		func(ctx context.Context) (*asset.Asset, *asset.File, error) {
			if inp.Token != "" {
				uuid = inp.Token
				u, err := i.repos.AssetUpload.FindByID(ctx, uuid)
				if err != nil {
					return nil, nil, err
				}
				if u.Expired(time.Now()) {
					return nil, nil, rerror.ErrInternalBy(
						fmt.Errorf("expired upload token: %s", uuid),
					)
				}
				file, err = i.gateways.File.UploadedAsset(ctx, u)
				if err != nil {
					return nil, nil, err
				}
			}

			needDecompress := false
			if ext := strings.ToLower(path.Ext(file.Name)); ext == ".zip" || ext == ".7z" {
				needDecompress = true
			}

			es := lo.ToPtr(asset.ArchiveExtractionStatusDone)
			if needDecompress {
				if inp.SkipDecompression {
					es = lo.ToPtr(asset.ArchiveExtractionStatusSkipped)
				} else {
					es = lo.ToPtr(asset.ArchiveExtractionStatusPending)
				}
			}

			ab := asset.New().
				NewID().
				Project(inp.ProjectID).
				FileName(path.Base(file.Name)).
				Size(uint64(file.Size)).
				Type(asset.DetectPreviewType(file)).
				UUID(uuid).
				ArchiveExtractionStatus(es)

			if op.AcOperator.User != nil {
				ab.CreatedByUser(*op.AcOperator.User)
			}
			if op.Integration != nil {
				ab.CreatedByIntegration(*op.Integration)
			}

			a, err := ab.Build()
			if err != nil {
				return nil, nil, err
			}

			a.SetAccessInfoResolver(i.gateways.File.GetAccessInfoResolver())

			f := asset.NewFile().
				Name(file.Name).
				Path(file.Name).
				Size(uint64(file.Size)).
				ContentType(file.ContentType).
				GuessContentTypeIfEmpty().
				ContentEncoding(file.ContentEncoding).
				Build()

			if err := i.repos.Asset.Save(ctx, a); err != nil {
				return nil, nil, err
			}

			if err := i.repos.AssetFile.Save(ctx, a.ID(), f); err != nil {
				return nil, nil, err
			}

			if needDecompress && !inp.SkipDecompression {
				if err := i.triggerDecompressEvent(ctx, a, f); err != nil {
					return nil, nil, err
				}
			}
			return a, f, nil
		})
	if err != nil {
		return nil, nil, err
	}

	// In AWS, extraction is done in very short time when a zip file is small, so it often results in an error because an asset is not saved yet in MongoDB. So an event should be created after commtting the transaction.
	if err := i.event(ctx, Event{
		Project:   prj,
		Workspace: prj.Workspace(),
		Type:      event.AssetCreate,
		Object:    a,
		Operator:  op.Operator(),
	}); err != nil {
		return nil, nil, err
	}

	return a, f, nil
}

func (i *Asset) CreateWithWorkspace(ctx context.Context, inp interfaces.CreateAssetParam, operator *usecase.Operator) (result *asset.Asset, afile *asset.File, err error) {
	if inp.File == nil {
		return nil, nil, interfaces.ErrFileNotIncluded
	}

	ws, err := i.repos.Workspace.FindByID(ctx, inp.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}

	if !operator.IsWritableWorkspace(ws.ID()) {
		return nil, nil, interfaces.ErrOperationDenied
	}

	uploadURL, size, err := i.gateways.File.UploadAsset(ctx, inp.File)
	if err != nil {
		return nil, nil, err
	}

	// enforce policy
	if policyID := operator.Policy(ws.Policy()); policyID != nil {
		p, err := i.repos.Policy.FindByID(ctx, *policyID)
		if err != nil {
			return nil, nil, err
		}
		s, err := i.repos.Asset.TotalSizeByWorkspace(ctx, ws.ID())
		if err != nil {
			return nil, nil, err
		}
		if err := p.EnforceAssetStorageSize(s + size); err != nil {
			if parsedURL, parseErr := url.Parse(uploadURL); parseErr == nil {
				_ = i.gateways.File.RemoveAsset(ctx, parsedURL)
			}
			return nil, nil, err
		}
	}

	a, err := asset.New().
		NewID().
		Workspace(inp.WorkspaceID).
		Project(inp.ProjectID).
		Name(path.Base(inp.File.Path)).
		Size(uint64(size)).
		URL(uploadURL).
		CreatedByUser(*operator.AcOperator.User).
		Build()
	if err != nil {
		return nil, nil, err
	}

	f := asset.NewFile().
		Name(inp.File.Name).
		Path(inp.File.Name).
		Size(uint64(inp.File.Size)).
		ContentType(inp.File.ContentType).
		GuessContentTypeIfEmpty().
		ContentEncoding(inp.File.ContentEncoding).
		Build()

	if err := i.repos.Asset.Save(ctx, a); err != nil {
		return nil, nil, err
	}

	return a, f, nil
}

func (i *Asset) Decompress(
	ctx context.Context,
	aId id.AssetID,
	operator *usecase.Operator,
) (*asset.Asset, error) {
	if operator.AcOperator.User == nil && operator.Integration == nil {
		return nil, interfaces.ErrInvalidOperator
	}

	return Run1(
		ctx, operator, i.repos,
		Usecase().Transaction(),
		func(ctx context.Context) (*asset.Asset, error) {
			a, err := i.repos.Asset.FindByID(ctx, aId)
			if err != nil {
				return nil, err
			}

			a.SetAccessInfoResolver(i.gateways.File.GetAccessInfoResolver())

			if !operator.CanUpdate(a) {
				return nil, interfaces.ErrOperationDenied
			}

			f, err := i.repos.AssetFile.FindByID(ctx, aId)
			if err != nil {
				return nil, err
			}

			if err := i.triggerDecompressEvent(ctx, a, f); err != nil {
				return nil, err
			}

			a.UpdateArchiveExtractionStatus(lo.ToPtr(asset.ArchiveExtractionStatusPending))

			if err := i.repos.Asset.Save(ctx, a); err != nil {
				return nil, err
			}

			return a, nil
		},
	)
}

func (i *Asset) Publish(
	ctx context.Context,
	aId id.AssetID,
	operator *usecase.Operator,
) (*asset.Asset, error) {
	if operator.AcOperator.User == nil && operator.Integration == nil {
		return nil, interfaces.ErrInvalidOperator
	}

	return Run1(
		ctx,
		operator,
		i.repos,
		Usecase().Transaction(),
		func(ctx context.Context) (*asset.Asset, error) {
			a, err := i.repos.Asset.FindByID(ctx, aId)
			if err != nil {
				return nil, err
			}

			if a != nil {
				a.SetAccessInfoResolver(i.gateways.File.GetAccessInfoResolver())
			}

			if !operator.CanUpdate(a) {
				return nil, interfaces.ErrOperationDenied
			}

			err = i.gateways.File.PublishAsset(ctx, a.UUID(), a.FileName())
			if err != nil {
				return nil, err
			}

			a.UpdatePublic(true)

			if err := i.repos.Asset.Save(ctx, a); err != nil {
				return nil, err
			}

			return a, nil
		},
	)
}

func (i *Asset) Unpublish(
	ctx context.Context,
	aId id.AssetID,
	operator *usecase.Operator,
) (*asset.Asset, error) {
	if operator.AcOperator.User == nil && operator.Integration == nil {
		return nil, interfaces.ErrInvalidOperator
	}

	return Run1(
		ctx,
		operator,
		i.repos,
		Usecase().Transaction(),
		func(ctx context.Context) (*asset.Asset, error) {
			a, err := i.repos.Asset.FindByID(ctx, aId)
			if err != nil {
				return nil, err
			}

			if a != nil {
				a.SetAccessInfoResolver(i.gateways.File.GetAccessInfoResolver())
			}

			if !operator.CanUpdate(a) {
				return nil, interfaces.ErrOperationDenied
			}

			err = i.gateways.File.UnpublishAsset(ctx, a.UUID(), a.FileName())
			if err != nil {
				return nil, err
			}

			a.UpdatePublic(false)

			if err := i.repos.Asset.Save(ctx, a); err != nil {
				return nil, err
			}

			return a, nil
		},
	)
}

type wrappedUploadCursor struct {
	UUID   string
	Cursor string
}

func (c wrappedUploadCursor) String() string {
	return c.UUID + "_" + c.Cursor
}

func parseWrappedUploadCursor(c string) (*wrappedUploadCursor, error) {
	uuid, cursor, found := strings.Cut(c, "_")
	if !found {
		return nil, fmt.Errorf("separator not found")
	}
	return &wrappedUploadCursor{
		UUID:   uuid,
		Cursor: cursor,
	}, nil
}

func wrapUploadCursor(uuid, cursor string) string {
	if cursor == "" {
		return ""
	}
	return wrappedUploadCursor{UUID: uuid, Cursor: cursor}.String()
}

func (i *Asset) CreateUpload(
	ctx context.Context,
	inp interfaces.CreateAssetUploadParam,
	op *usecase.Operator,
) (*interfaces.AssetUpload, error) {
	if op.AcOperator.User == nil && op.Integration == nil {
		return nil, interfaces.ErrInvalidOperator
	}

	if inp.ContentEncoding == "gzip" {
		inp.Filename = strings.TrimSuffix(inp.Filename, ".gz")
	}

	var param *gateway.IssueUploadAssetParam
	if inp.Cursor == "" {
		if inp.Filename == "" {
			// TODO: Change to the appropriate error
			return nil, interfaces.ErrFileNotIncluded
		}

		const week = 7 * 24 * time.Hour
		expiresAt := time.Now().Add(1 * week)
		param = &gateway.IssueUploadAssetParam{
			UUID:            uuid.New().String(),
			Filename:        inp.Filename,
			ContentLength:   inp.ContentLength,
			ContentType:     inp.ContentType,
			ContentEncoding: inp.ContentEncoding,
			ExpiresAt:       expiresAt,
			Cursor:          "",
		}
	} else {
		wrapped, err := parseWrappedUploadCursor(inp.Cursor)
		if err != nil {
			return nil, fmt.Errorf("parse cursor(%s): %w", inp.Cursor, err)
		}
		au, err := i.repos.AssetUpload.FindByID(ctx, wrapped.UUID)
		if err != nil {
			return nil, fmt.Errorf("find asset upload(uuid=%s): %w", wrapped.UUID, err)
		}
		if inp.ProjectID.Compare(au.Project()) != 0 {
			return nil, fmt.Errorf("unmatched project id(in=%s,db=%s)", inp.ProjectID, au.Project())
		}
		param = &gateway.IssueUploadAssetParam{
			UUID:            wrapped.UUID,
			Filename:        au.FileName(),
			ContentLength:   au.ContentLength(),
			ContentEncoding: au.ContentEncoding(),
			ContentType:     au.ContentType(),
			ExpiresAt:       au.ExpiresAt(),
			Cursor:          wrapped.Cursor,
		}
	}

	prj, err := i.repos.Project.FindByID(ctx, inp.ProjectID)
	if err != nil {
		return nil, err
	}
	if !op.IsWritableWorkspace(prj.Workspace()) {
		return nil, interfaces.ErrOperationDenied
	}

	uploadLink, err := i.gateways.File.IssueUploadAssetLink(ctx, *param)
	if errors.Is(err, gateway.ErrUnsupportedOperation) {
		return nil, rerror.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if inp.Cursor == "" {
		u := asset.NewUpload().
			UUID(param.UUID).
			Project(prj.ID()).
			FileName(param.Filename).
			ExpiresAt(param.ExpiresAt).
			ContentLength(uploadLink.ContentLength).
			ContentType(uploadLink.ContentType).
			ContentEncoding(uploadLink.ContentEncoding).
			Build()
		if err := i.repos.AssetUpload.Save(ctx, u); err != nil {
			return nil, err
		}
	}

	return &interfaces.AssetUpload{
		URL:             uploadLink.URL,
		UUID:            param.UUID,
		ContentType:     uploadLink.ContentType,
		ContentLength:   uploadLink.ContentLength,
		ContentEncoding: uploadLink.ContentEncoding,
		Next:            wrapUploadCursor(param.UUID, uploadLink.Next),
	}, nil
}

func (i *Asset) triggerDecompressEvent(ctx context.Context, a *asset.Asset, f *asset.File) error {
	if i.gateways.TaskRunner == nil {
		log.Infof(
			"asset: decompression of asset %s was skipped because task runner is not configured",
			a.ID(),
		)
		return nil
	}

	taskPayload := task.DecompressAssetPayload{
		AssetID: a.ID().String(),
		Path:    f.RootPath(a.UUID()),
	}
	if err := i.gateways.TaskRunner.Run(ctx, taskPayload.Payload()); err != nil {
		return err
	}

	a.UpdateArchiveExtractionStatus(lo.ToPtr(asset.ArchiveExtractionStatusInProgress))
	if err := i.repos.Asset.Save(ctx, a); err != nil {
		return err
	}

	return nil
}

func (i *Asset) Update(
	ctx context.Context,
	inp interfaces.UpdateAssetParam,
	operator *usecase.Operator,
) (result *asset.Asset, err error) {
	if operator.AcOperator.User == nil && operator.Integration == nil {
		return nil, interfaces.ErrInvalidOperator
	}

	return Run1(
		ctx, operator, i.repos,
		Usecase().Transaction(),
		func(ctx context.Context) (*asset.Asset, error) {
			a, err := i.repos.Asset.FindByID(ctx, inp.AssetID)
			if err != nil {
				return nil, err
			}

			if a != nil {
				a.SetAccessInfoResolver(i.gateways.File.GetAccessInfoResolver())
			}

			if !operator.CanUpdate(a) {
				return nil, interfaces.ErrOperationDenied
			}

			if inp.PreviewType != nil {
				a.UpdatePreviewType(inp.PreviewType)
			}

			if err := i.repos.Asset.Save(ctx, a); err != nil {
				return nil, err
			}

			return a, nil
		},
	)
}

func (i *Asset) UpdateFiles(
	ctx context.Context,
	aid id.AssetID,
	s *asset.ArchiveExtractionStatus,
	op *usecase.Operator,
) (*asset.Asset, error) {
	if op.AcOperator.User == nil && op.Integration == nil && !op.Machine {
		return nil, interfaces.ErrInvalidOperator
	}

	return Run1(
		ctx, op, i.repos,
		Usecase().Transaction(),
		func(ctx context.Context) (*asset.Asset, error) {
			a, err := i.repos.Asset.FindByID(ctx, aid)
			if err != nil {
				if err == rerror.ErrNotFound {
					return nil, err
				}
				return nil, fmt.Errorf("failed to find an asset: %v", err)
			}

			if a != nil {
				a.SetAccessInfoResolver(i.gateways.File.GetAccessInfoResolver())
			}

			if !op.CanUpdate(a) {
				return nil, interfaces.ErrOperationDenied
			}

			if shouldSkipUpdate(a.ArchiveExtractionStatus(), s) {
				return a, nil
			}

			prj, err := i.repos.Project.FindByID(ctx, a.Project())
			if err != nil {
				return nil, fmt.Errorf("failed to find a project: %v", err)
			}

			srcfile, err := i.repos.AssetFile.FindByID(ctx, aid)
			if err != nil {
				return nil, fmt.Errorf("failed to find an asset file: %v", err)
			}

			files, err := i.gateways.File.GetAssetFiles(ctx, a.UUID())
			if err != nil {
				if err == gateway.ErrFileNotFound {
					return nil, err
				}
				return nil, fmt.Errorf("failed to get asset files: %v", err)
			}

			a.UpdateArchiveExtractionStatus(s)
			if previewType := detectPreviewType(files); previewType != nil {
				a.UpdatePreviewType(previewType)
			}

			if err := i.repos.Asset.Save(ctx, a); err != nil {
				return nil, fmt.Errorf("failed to save an asset: %v", err)
			}

			srcPath := srcfile.Path()
			assetFiles := lo.FilterMap(files, func(f gateway.FileEntry, _ int) (*asset.File, bool) {
				if srcPath == f.Name {
					return nil, false
				}
				return asset.NewFile().
					Name(path.Base(f.Name)).
					Path(f.Name).
					Size(uint64(f.Size)).
					ContentType(f.ContentType).
					GuessContentTypeIfEmpty().
					ContentEncoding(f.ContentEncoding).
					Build(), true
			})

			if err := i.repos.AssetFile.SaveFlat(ctx, a.ID(), srcfile, assetFiles); err != nil {
				return nil, fmt.Errorf("failed to save asset files: %v", err)
			}

			if err := i.event(ctx, Event{
				Project:   prj,
				Workspace: prj.Workspace(),
				Type:      event.AssetDecompress,
				Object:    a,
				Operator:  op.Operator(),
			}); err != nil {
				return nil, fmt.Errorf("failed to create an event: %v", err)
			}

			return a, nil
		},
	)
}

func detectPreviewType(files []gateway.FileEntry) *asset.PreviewType {
	for _, entry := range files {
		if path.Base(entry.Name) == "tileset.json" {
			return lo.ToPtr(asset.PreviewTypeGeo3dTiles)
		}
		if path.Ext(entry.Name) == ".mvt" {
			return lo.ToPtr(asset.PreviewTypeGeoMvt)
		}
	}
	return nil
}

func shouldSkipUpdate(from, to *asset.ArchiveExtractionStatus) bool {
	if from.String() == asset.ArchiveExtractionStatusDone.String() {
		return true
	}
	return from.String() == to.String()
}

func (i *Asset) Delete(
	ctx context.Context,
	aId id.AssetID,
	operator *usecase.Operator,
) (result id.AssetID, err error) {
	if operator.AcOperator.User == nil && operator.Integration == nil {
		return aId, interfaces.ErrInvalidOperator
	}

	return Run1(
		ctx, operator, i.repos,
		Usecase().Transaction(),
		func(ctx context.Context) (id.AssetID, error) {
			a, err := i.repos.Asset.FindByID(ctx, aId)
			if err != nil {
				return aId, err
			}

			if !operator.CanUpdate(a) {
				return aId, interfaces.ErrOperationDenied
			}

			uuid := a.UUID()
			filename := a.FileName()
			if uuid != "" && filename != "" {
				if err := i.gateways.File.DeleteAsset(ctx, uuid, filename); err != nil {
					return aId, err
				}
			}

			err = i.repos.Asset.Delete(ctx, aId)
			if err != nil {
				return aId, err
			}

			p, err := i.repos.Project.FindByID(ctx, a.Project())
			if err != nil {
				return aId, err
			}

			if err := i.event(ctx, Event{
				Project:   p,
				Workspace: p.Workspace(),
				Type:      event.AssetDelete,
				Object:    a,
				Operator:  operator.Operator(),
			}); err != nil {
				return aId, err
			}

			return aId, nil
		},
	)
}

// BatchDelete deletes assets in batch based on multiple asset IDs
func (i *Asset) BatchDelete(
	ctx context.Context,
	assetIDs id.AssetIDList,
	operator *usecase.Operator,
) (result []id.AssetID, err error) {
	if operator.AcOperator.User == nil && operator.Integration == nil {
		return assetIDs, interfaces.ErrInvalidOperator
	}

	if len(assetIDs) == 0 {
		return nil, interfaces.ErrEmptyIDsList
	}

	return Run1(
		ctx, operator, i.repos,
		Usecase().Transaction(),
		func(ctx context.Context) (id.AssetIDList, error) {
			assets, err := i.repos.Asset.FindByIDs(ctx, assetIDs)
			if err != nil {
				return assetIDs, err
			}

			if len(assetIDs) != len(assets) {
				return assetIDs, interfaces.ErrPartialNotFound
			}

			if assets == nil {
				return assetIDs, nil
			}

			UUIDList := lo.FilterMap(assets, func(a *asset.Asset, _ int) (string, bool) {
				if a == nil || a.UUID() == "" || a.FileName() == "" {
					return "", false
				}
				return a.UUID(), true
			})

			// deletes assets' files in
			err = i.gateways.File.DeleteAssets(ctx, UUIDList)
			if err != nil {
				return assetIDs, err
			}

			err = i.repos.Asset.BatchDelete(ctx, assetIDs)
			if err != nil {
				return assetIDs, err
			}

			return assetIDs, nil
		},
	)
}

func (i *Asset) event(ctx context.Context, e Event) error {
	if i.ignoreEvent {
		return nil
	}

	_, err := createEvent(ctx, i.repos, i.gateways, e)
	return err
}

func (i *Asset) RetryDecompression(ctx context.Context, id string) error {
	return i.gateways.TaskRunner.Retry(ctx, id)
}

func (i *Asset) ImportAssetFiles(
	ctx context.Context,
	assets map[string]*zip.File,
	data *[]byte,
	newProject *project.Project,
) (*[]byte, error) {
	var currentHost string
	if i.hostAdapter != nil {
		currentHost = i.hostAdapter.CurrentHost(ctx)
	}

	var d map[string]any
	if err := json.Unmarshal(*data, &d); err != nil {
		return nil, err
	}

	assetNames := make(map[string]string)
	for beforeName, realName := range d["assets"].(map[string]any) {
		if realName, ok := realName.(string); ok {
			assetNames[beforeName] = realName
		}
	}

	for beforeName, zipFile := range assets {
		if zipFile.UncompressedSize64 == 0 {
			continue
		}
		realName := assetNames[beforeName]

		if err := func() error {
			readCloser, err := zipFile.Open()
			if err != nil {
				return err
			}

			defer func() {
				if err := readCloser.Close(); err != nil {
					fmt.Printf("Error closing fileToUpload: %v\n", err)
				}
			}()

			fileToUpload := &file.File{
				Content:     readCloser,
				Path:        realName,
				Size:        int64(zipFile.UncompressedSize64),
				ContentType: http.DetectContentType([]byte(zipFile.Name)),
			}

			uploadURL, size, err := i.gateways.File.UploadAsset(ctx, fileToUpload)
			if err != nil {
				return err
			}

			// Project logo update will be at this time
			if newProject.ImageURL() != nil {
				if path.Base(newProject.ImageURL().Path) == beforeName {
					parsedURL, err := url.Parse(uploadURL)
					if err != nil {
						return err
					}
					newProject.SetImageURL(parsedURL)
					err = i.repos.Project.Save(ctx, newProject)
					if err != nil {
						return err
					}
				}
			}

			systemUserID := accountdomain.NewUserID()

			a, err := asset.New().
				NewID().
				Project(newProject.ID()).
				Workspace(newProject.Workspace()).
				Name(path.Base(realName)).
				Size(uint64(size)).
				UUID(uploadURL).
				CreatedByUser(systemUserID).
				Build()
			if err != nil {
				return err
			}

			if err := i.repos.Asset.Save(ctx, a); err != nil {
				return err
			}

			parsedURL, err := url.Parse(uploadURL)
			if err != nil {
				return err
			}
			afterName := path.Base(parsedURL.Path)

			beforeUrl := fmt.Sprintf("%s/assets/%s", currentHost, beforeName)
			afterUrl := fmt.Sprintf("%s/assets/%s", currentHost, afterName)
			*data = bytes.ReplaceAll(*data, []byte(beforeUrl), []byte(afterUrl))

			return nil
		}(); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (i *Asset) FindByWorkspaceProject(
	ctx context.Context,
	tid accountdomain.WorkspaceID,
	pid *id.ProjectID,
	keyword *string,
	sort *asset.SortType,
	p *usecasex.Pagination,
	operator *usecase.Operator,
) ([]*asset.Asset, *usecasex.PageInfo, error) {
	return Run2(
		ctx, operator, i.repos,
		Usecase().WithReadableWorkspaces(tid),
		func(ctx context.Context) ([]*asset.Asset, *usecasex.PageInfo, error) {
			return i.repos.Asset.FindByWorkspaceProject(ctx, tid, pid, repo.AssetFilter{
				SortType:   sort,
				Keyword:    keyword,
				Pagination: p,
			})
		},
	)
}

func (i *Asset) FindByWorkspace(
	ctx context.Context,
	tid accountdomain.WorkspaceID,
	keyword *string,
	sort *asset.SortType,
	p *interfaces.PaginationParam,
) ([]*asset.Asset, *interfaces.PageBasedInfo, error) {
	var pagination *usecasex.Pagination
	if p != nil && p.Page != nil {
		pagination = usecasex.OffsetPagination{
			Offset: int64((p.Page.Page - 1) * p.Page.PageSize),
			Limit:  int64(p.Page.PageSize),
		}.Wrap()
	}

	return Run2(
		ctx, nil, i.repos,
		Usecase().WithReadableWorkspaces(tid),
		func(ctx context.Context) ([]*asset.Asset, *interfaces.PageBasedInfo, error) {
			return i.repos.Asset.FindByWorkspace(ctx, tid, repo.AssetFilter{
				SortType:   sort,
				Keyword:    keyword,
				Pagination: pagination,
			})
		},
	)
}

func (i *Asset) Fetch(ctx context.Context, assets []id.AssetID) ([]*asset.Asset, error) {
	return i.repos.Asset.FindByIDs(ctx, assets)
}
