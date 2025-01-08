package workspace

import (
	"sort"
	"sync"

	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"github.com/samber/lo"
	"golang.org/x/exp/maps"
)

var (
	ErrUserAlreadyJoined             = rerror.NewE(i18n.T("user already joined"))
	ErrCannotModifyPersonalWorkspace = rerror.NewE(i18n.T("personal workspace cannot be modified"))
	ErrTargetUserNotInTheWorkspace   = rerror.NewE(i18n.T("target user does not exist in the workspace"))
	ErrInvalidWorkspaceName          = rerror.NewE(i18n.T("invalid workspace name"))
	ErrNoSpecifiedUsers              = rerror.NewE(i18n.T("no specified users for removal"))
)

type Member struct {
	Role      Role
	Disabled  bool
	InvitedBy UserID
	Host      string
}

type Members struct {
	users        map[UserID]Member
	integrations map[IntegrationID]Member
	fixed        bool
	mu           sync.Mutex
}

func NewMembers() *Members {
	return &Members{
		users:        map[UserID]Member{},
		integrations: map[IntegrationID]Member{},
	}
}

func NewMembersWith(users map[UserID]Member, integrations map[IntegrationID]Member, fixed bool) *Members {
	m := &Members{
		users:        maps.Clone(users),
		integrations: maps.Clone(integrations),
		fixed:        fixed,
	}
	return m
}

func InitMembers(u UserID) *Members {
	return NewMembersWith(
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
	m.mu.Lock()
	defer m.mu.Unlock()
	return maps.Clone(m.users)
}

func (m *Members) UserIDs() []UserID {
	m.mu.Lock()
	defer m.mu.Unlock()
	users := lo.Keys(m.users)
	sort.SliceStable(users, func(a, b int) bool {
		return users[a].Compare(users[b]) > 0
	})
	return users
}

func (m *Members) Integrations() map[IntegrationID]Member {
	m.mu.Lock()
	defer m.mu.Unlock()
	return maps.Clone(m.integrations)
}

func (m *Members) IntegrationIDs() []IntegrationID {
	m.mu.Lock()
	defer m.mu.Unlock()
	integrations := lo.Keys(m.integrations)
	sort.SliceStable(integrations, func(a, b int) bool {
		return integrations[a].Compare(integrations[b]) > 0
	})
	return integrations
}

func (m *Members) HasUser(u UserID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.users[u]
	return ok
}

func (m *Members) HasIntegration(i IntegrationID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.integrations[i]
	return ok
}

func (m *Members) User(u UserID) *Member {
	m.mu.Lock()
	defer m.mu.Unlock()
	um, ok := m.users[u]
	if ok {
		return &um
	}
	return nil
}

func (m *Members) Integration(i IntegrationID) *Member {
	m.mu.Lock()
	defer m.mu.Unlock()
	im, ok := m.integrations[i]
	if ok {
		return &im
	}
	return nil
}

func (m *Members) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.users)
}

func (m *Members) UserRole(u UserID) Role {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.users[u].Role
}

func (m *Members) IntegrationRole(iId IntegrationID) Role {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.integrations[iId].Role
}

func (m *Members) IsEmpty() bool {
	return m.Count() == 0
}

func (m *Members) Fixed() bool {
	return m != nil && m.fixed
}

func (m *Members) IsOnlyOwner(u UserID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	ownerCount := 0
	for _, member := range m.users {
		if member.Role == RoleOwner {
			ownerCount++
		}
	}
	return ownerCount == 1 && m.users[u].Role == RoleOwner
}

func (m *Members) IsOwnerOrMaintainer(u UserID) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.users[u].Role == RoleOwner || m.users[u].Role == RoleMaintainer
}

func (m *Members) UpdateUserRole(u UserID, role Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

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

func (m *Members) Join(u *user.User, role Role, i UserID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.fixed {
		return ErrCannotModifyPersonalWorkspace
	}
	if _, ok := m.users[u.ID()]; ok {
		return ErrUserAlreadyJoined
	}
	if role == Role("") {
		role = RoleReader
	}
	if m.users == nil {
		m.users = map[UserID]Member{}
	}
	m.users[u.ID()] = Member{
		Role:      role,
		Disabled:  false,
		InvitedBy: i,
		Host:      u.Host(),
	}
	return nil
}

func (m *Members) AddIntegration(iid IntegrationID, role Role, i UserID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.integrations[iid]; ok {
		delete(m.integrations, iid)
	} else {
		return ErrTargetUserNotInTheWorkspace
	}
	return nil
}

func (m *Members) UsersByRole(role Role) []UserID {
	m.mu.Lock()
	defer m.mu.Unlock()

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
