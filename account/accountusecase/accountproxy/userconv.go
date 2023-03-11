package accountproxy

import (
	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/util"
)

func UserByIDsResponseTo(r *UserByIDsResponse, err error) (users []*user.User, wids accountdomain.WorkspaceIDList, _ error) {
	if err != nil || r == nil {
		return nil, nil, err
	}

	for _, n := range r.Nodes {
		r1, r2, err := UserByIDsNodesNodeTo(n)
		if err != nil {
			return nil, nil, err
		}
		users = append(users, r1)
		wids = append(wids, r2...)
	}

	return
}

func UserByIDsNodesNodeTo(r UserByIDsNodesNode) (*user.User, accountdomain.WorkspaceIDList, error) {
	if r == nil {
		return nil, nil, nil
	}
	u, ok := r.(*UserByIDsNodesUser)
	if !ok || u == nil {
		return nil, nil, nil
	}
	wids, err := util.TryMap(u.Workspaces, func(w UserByIDsNodesUserWorkspacesWorkspace) (accountdomain.WorkspaceID, error) {
		return accountdomain.WorkspaceIDFrom(w.Id)
	})
	if err != nil {
		return nil, nil, nil
	}
	users, err := UserByIDsNodesUserTo(u)
	return users, wids, err
}

func UserByIDsNodesUserTo(r *UserByIDsNodesUser) (*user.User, error) {
	if r == nil {
		return nil, nil
	}
	return user.New().ParseID(r.Id).Name(r.Name).Email(r.Email).Build()
}
