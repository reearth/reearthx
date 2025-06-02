package interactor

import (
	"context"
	"errors"

	"github.com/reearth/reearthx/asset/domain/workspacesettings"
	"github.com/reearth/reearthx/asset/usecase"
	"github.com/reearth/reearthx/asset/usecase/gateway"
	"github.com/reearth/reearthx/asset/usecase/interfaces"
	"github.com/reearth/reearthx/asset/usecase/repo"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/rerror"
)

type WorkspaceSettings struct {
	repos    *repo.Container
	gateways *gateway.Container
}

func NewWorkspaceSettings(r *repo.Container, g *gateway.Container) interfaces.WorkspaceSettings {
	return &WorkspaceSettings{
		repos:    r,
		gateways: g,
	}
}

func (ws *WorkspaceSettings) Fetch(ctx context.Context, wid accountdomain.WorkspaceIDList, op *usecase.Operator) (result workspacesettings.List, err error) {
	return ws.repos.WorkspaceSettings.FindByIDs(ctx, wid)
}

func (ws *WorkspaceSettings) UpdateOrCreate(ctx context.Context, inp interfaces.UpdateOrCreateWorkspaceSettingsParam, op *usecase.Operator) (result *workspacesettings.WorkspaceSettings, err error) {
	wss, err := ws.repos.WorkspaceSettings.FindByID(ctx, inp.ID)
	if err != nil && !errors.Is(err, rerror.ErrNotFound) {
		return nil, err
	}

	return Run1(ctx, op, ws.repos, Usecase().WithMaintainableWorkspaces(inp.ID).Transaction(),
		func(ctx context.Context) (_ *workspacesettings.WorkspaceSettings, err error) {
			if wss == nil {
				wsb := workspacesettings.New().
					ID(inp.ID)

				wss, err = wsb.Build()
				if err != nil {
					return nil, err
				}
			}
			if inp.Tiles != nil {
				wss.SetTiles(inp.Tiles)
			}
			if inp.Terrains != nil {
				wss.SetTerrains(inp.Terrains)
			}
			if err := ws.repos.WorkspaceSettings.Save(ctx, wss); err != nil {
				return nil, err
			}
			return wss, nil
		})
}

func (ws *WorkspaceSettings) Delete(ctx context.Context, inp interfaces.DeleteWorkspaceSettingsParam, op *usecase.Operator) error {
	return Run0(ctx, op, ws.repos, Usecase().WithMaintainableWorkspaces(inp.ID).Transaction(),
		func(ctx context.Context) error {
			return ws.repos.WorkspaceSettings.Remove(ctx, inp.ID)
		})
}
