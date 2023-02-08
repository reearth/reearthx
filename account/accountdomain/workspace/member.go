package workspace

import (
	"errors"
	"sort"

	"github.com/samber/lo"
	"golang.org/x/exp/maps"
)

var (
	ErrUserAlreadyJoined             = errors.New("user already joined")
	ErrCannotModifyPersonalWorkspace = errors.New("personal workspace cannot be modified")
	ErrTargetUserNotInTheWorkspace   = errors.New("target user does not exist in the workspace")
	ErrInvalidWorkspaceName          = errors.New("invalid workspace name")
)

type Member struct {
	Role      Role
	Disabled  bool
	InvitedBy UserID
}

type Members struct {
	users        map[UserID]Member
	integrations map[IntegrationID]Member
	fixed        bool
}

func NewMembers(u map[UserID]Member, i map[IntegrationID]Member, fixed bool) *Members {
	return &Members{
		users:        maps.Clone(u),
		integrations: maps.Clone(i),
		fixed:        fixed,
	}
}

func NewMembersWith(users map[UserID]Member) *Members {
	m := &Members{
		users:        maps.Clone(users),
		integrations: map[IntegrationID]Member{},
	}
	return m
}

func InitMembers(u UserID) *Members {
	return NewMembers(
		map[UserID]Member{
			u: {
				Role:      RoleOwner,
				Disabled:  false,
				InvitedBy: u,
			},
		},
		nil,
		true,
	)
}

func (m *Members) Clone() *Members {
	if m == nil {
		return nil
	}

	return &Members{
		users:        maps.Clone(m.users),
		integrations: maps.Clone(m.integrations),
		fixed:        m.fixed,
	}
}

func (m *Members) Users() map[UserID]Member {
	return maps.Clone(m.users)
}

func (m *Members) UserIDs() []UserID {
	users := lo.Keys(m.users)
	sort.SliceStable(users, func(a, b int) bool {
		return users[a].Compare(users[b]) > 0
	})
	return users
}

func (m *Members) Integrations() map[IntegrationID]Member {
	return maps.Clone(m.integrations)
}

func (m *Members) IntegrationIDs() []IntegrationID {
	integrations := lo.Keys(m.integrations)
	sort.SliceStable(integrations, func(a, b int) bool {
		return integrations[a].Compare(integrations[b]) > 0
	})
	return integrations
}

func (m *Members) HasUser(u UserID) bool {
	_, ok := m.users[u]
	return ok
}

func (m *Members) HasIntegration(i IntegrationID) bool {
	_, ok := m.integrations[i]
	return ok
}

func (m *Members) User(u UserID) *Member {
	um, ok := m.users[u]
	if ok {
		return &um
	}
	return nil
}

func (m *Members) Integration(i IntegrationID) *Member {
	im, ok := m.integrations[i]
	if ok {
		return &im
	}
	return nil
}

func (m *Members) Count() int {
	return len(m.users)
}

func (m *Members) IsEmpty() bool {
	return m.Count() == 0
}

func (m *Members) Fixed() bool {
	return m != nil && m.fixed
}

func (m *Members) IsOnlyOwner(u UserID) bool {
	return len(m.UsersByRole(RoleOwner)) == 1 && m.users[u].Role == RoleOwner
}

func (m *Members) UpdateUserRole(u UserID, role Role) error {
	if m.fixed {
		return ErrCannotModifyPersonalWorkspace
	}
	if !role.Valid() {
		return nil
	}
	if _, ok := m.users[u]; !ok {
		return ErrTargetUserNotInTheWorkspace
	}
	mm := m.users[u]
	mm.Role = role
	m.users[u] = mm
	return nil
}

func (m *Members) UpdateIntegrationRole(iId IntegrationID, role Role) error {
	if !role.Valid() {
		return nil
	}
	if _, ok := m.integrations[iId]; !ok {
		return ErrTargetUserNotInTheWorkspace
	}
	mm := m.integrations[iId]
	mm.Role = role
	m.integrations[iId] = mm
	return nil
}

func (m *Members) Join(u UserID, role Role, i UserID) error {
	if m.fixed {
		return ErrCannotModifyPersonalWorkspace
	}
	if _, ok := m.users[u]; ok {
		return ErrUserAlreadyJoined
	}
	if role == Role("") {
		role = RoleReader
	}
	if m.users == nil {
		m.users = map[UserID]Member{}
	}
	m.users[u] = Member{
		Role:      role,
		Disabled:  false,
		InvitedBy: i,
	}
	return nil
}

func (m *Members) AddIntegration(iid IntegrationID, role Role, i UserID) error {
	if _, ok := m.integrations[iid]; ok {
		return ErrUserAlreadyJoined
	}
	if role == Role("") {
		role = RoleReader
	}
	if m.integrations == nil {
		m.integrations = map[IntegrationID]Member{}
	}
	m.integrations[iid] = Member{
		Role:      role,
		Disabled:  false,
		InvitedBy: i,
	}
	return nil
}

func (m *Members) Leave(u UserID) error {
	if m.fixed {
		return ErrCannotModifyPersonalWorkspace
	}
	if _, ok := m.users[u]; ok {
		delete(m.users, u)
	} else {
		return ErrTargetUserNotInTheWorkspace
	}
	return nil
}

func (m *Members) DeleteIntegration(iid IntegrationID) error {
	if _, ok := m.integrations[iid]; ok {
		delete(m.integrations, iid)
	} else {
		return ErrTargetUserNotInTheWorkspace
	}
	return nil
}

func (m *Members) UsersByRole(role Role) []UserID {
	users := make([]UserID, 0, len(m.users))
	for u, m := range m.users {
		if m.Role == role {
			users = append(users, u)
		}
	}

	sort.SliceStable(users, func(a, b int) bool {
		return users[a].Compare(users[b]) > 0
	})

	return users
}
