package accountdomain

type WorkspaceList []*Workspace

func (l WorkspaceList) FilterByID(ids ...WorkspaceID) WorkspaceList {
	if l == nil {
		return nil
	}

	res := make(WorkspaceList, 0, len(l))
	for _, id := range ids {
		var t2 *Workspace
		for _, t := range l {
			if t.ID() == id {
				t2 = t
				break
			}
		}
		if t2 != nil {
			res = append(res, t2)
		}
	}
	return res
}

func (l WorkspaceList) FilterByUserRole(u UserID, r Role) WorkspaceList {
	if l == nil || u.IsEmpty() || r == "" {
		return nil
	}

	res := make(WorkspaceList, 0, len(l))
	for _, t := range l {
		if m := t.Members().User(u); m != nil && m.Role == r {
			res = append(res, t)
		}
	}
	return res
}

func (l WorkspaceList) FilterByIntegrationRole(i IntegrationID, r Role) WorkspaceList {
	if l == nil || i.IsEmpty() || r == "" {
		return nil
	}

	res := make(WorkspaceList, 0, len(l))
	for _, t := range l {
		if m := t.Members().Integration(i); m != nil && m.Role == r {
			res = append(res, t)
		}
	}
	return res
}

func (l WorkspaceList) FilterByUserRoleIncluding(u UserID, r Role) WorkspaceList {
	if l == nil || u.IsEmpty() || r == "" {
		return nil
	}

	res := make(WorkspaceList, 0, len(l))
	for _, t := range l {
		if m := t.Members().User(u); m != nil && m.Role.Includes(r) {
			res = append(res, t)
		}
	}
	return res
}

func (l WorkspaceList) FilterByIntegrationRoleIncluding(i IntegrationID, r Role) WorkspaceList {
	if l == nil || i.IsEmpty() || r == "" {
		return nil
	}

	res := make(WorkspaceList, 0, len(l))
	for _, t := range l {
		if m := t.Members().Integration(i); m != nil && m.Role.Includes(r) {
			res = append(res, t)
		}
	}
	return res
}

func (l WorkspaceList) IDs() []WorkspaceID {
	if l == nil {
		return nil
	}

	res := make([]WorkspaceID, 0, len(l))
	for _, t := range l {
		res = append(res, t.ID())
	}
	return res
}
