package accountproxy

import (
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/util"
)

func UserByIDsResponseTo(r *UserByIDsResponse, err error) ([]*user.Simple, error) {
	if err != nil || r == nil {
		return nil, err
	}
	return util.TryMap(r.Nodes, UserByIDsNodesNodeTo)
}

func MeToUser(me FragmentMe) (*user.User, error) {
	id, err := user.IDFrom(me.Id)
	if err != nil {
		return nil, err
	}
	wid, err := user.WorkspaceIDFrom(me.MyWorkspaceId)
	if err != nil {
		return nil, err
	}

	auths := make([]user.Auth, len(me.Auths))
	for i := range me.Auths {
		auths[i] = user.AuthFrom(me.Auths[i])
	}

	u, err := user.New().ID(id).Name(me.Name).
		Email(me.Email).LangFrom(me.Lang).
		Theme(user.ThemeFrom(me.Theme)).
		Auths(auths).
		Workspace(wid).Build()
	if err != nil {
		return nil, err
	}

	return u, nil
}

func UserByIDsNodesNodeTo(r UserByIDsNodesNode) (*user.Simple, error) {
	if r == nil {
		return nil, nil
	}
	u, ok := r.(*UserByIDsNodesUser)
	if !ok || u == nil {
		return nil, nil
	}
	return UserByIDsNodesUserTo(u)
}

func UserByIDsNodesUserTo(r *UserByIDsNodesUser) (*user.Simple, error) {
	if r == nil {
		return nil, nil
	}
	id, err := user.IDFrom(r.Id)
	if err != nil {
		return nil, err
	}
	return &user.Simple{
		ID:    id,
		Name:  r.Name,
		Email: r.Email,
	}, nil
}
