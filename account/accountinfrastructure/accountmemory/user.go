package accountmemory

import (
	"context"
	"net/mail"
	"strings"

	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/account/accountusecase/accountrepo"
	"github.com/reearth/reearthx/rerror"
	"github.com/reearth/reearthx/usecasex"
	"github.com/reearth/reearthx/util"
)

type User struct {
	data *util.SyncMap[user.ID, *user.User]
	err  error
}

func NewUser() *User {
	return &User{
		data: &util.SyncMap[user.ID, *user.User]{},
	}
}

func NewUserWith(users ...*user.User) *User {
	r := NewUser()
	for _, u := range users {
		r.data.Store(u.ID(), u)
	}
	return r
}

func (r *User) FindAll(ctx context.Context) (user.List, error) {
	if r.err != nil {
		return nil, r.err
	}

	res := r.data.FindAll(func(key user.ID, value *user.User) bool {
		return true
	})

	return res, nil
}

func (r *User) FindByIDs(_ context.Context, ids user.IDList) (user.List, error) {
	if r.err != nil {
		return nil, r.err
	}

	res := r.data.FindAll(func(key user.ID, value *user.User) bool {
		return ids.Has(key)
	})

	return res, nil
}

func (r *User) FindByIDsWithPagination(_ context.Context, ids user.IDList, pagination *usecasex.Pagination, nameOrAlias ...string) (user.List, *usecasex.PageInfo, error) {
	if r.err != nil {
		return nil, nil, r.err
	}

	if pagination == nil {
		users := r.data.FindAll(func(key user.ID, value *user.User) bool {
			if !ids.Has(key) {
				return false
			}

			if len(nameOrAlias) > 0 && nameOrAlias[0] != "" {
				searchTerm := strings.ToLower(nameOrAlias[0])
				userName := strings.ToLower(value.Name())
				userAlias := strings.ToLower(value.Alias())
				return strings.Contains(userName, searchTerm) || strings.Contains(userAlias, searchTerm)
			}

			return true
		})
		return users, nil, nil
	}

	allUsers := r.data.FindAll(func(key user.ID, value *user.User) bool {
		if !ids.Has(key) {
			return false
		}

		if len(nameOrAlias) > 0 && nameOrAlias[0] != "" {
			searchTerm := strings.ToLower(nameOrAlias[0])
			userName := strings.ToLower(value.Name())
			userAlias := strings.ToLower(value.Alias())
			return strings.Contains(userName, searchTerm) || strings.Contains(userAlias, searchTerm)
		}

		return true
	})

	totalCount := int64(len(allUsers))

	var offset int64
	var limit int64 = 20

	if pagination.Offset != nil {
		offset = pagination.Offset.Offset
		limit = pagination.Offset.Limit
	} else if pagination.Cursor != nil {
		if pagination.Cursor.First != nil {
			limit = *pagination.Cursor.First
		} else if pagination.Cursor.Last != nil {
			limit = *pagination.Cursor.Last
			// For "last" pagination, we want the last N items
			if totalCount > limit {
				offset = totalCount - limit
			}
		}
	}

	var pagedUsers user.List
	start := offset
	end := offset + limit

	if start < totalCount {
		if end > totalCount {
			end = totalCount
		}
		pagedUsers = allUsers[start:end]
	}

	var hasNextPage, hasPreviousPage bool
	if pagination.Cursor != nil && pagination.Cursor.Last != nil {
		// For "last" pagination: has previous if we're not showing all items from the beginning
		hasPreviousPage = offset > 0
		hasNextPage = false // "last" pagination doesn't have a next page
	} else {
		hasNextPage = end < totalCount
		hasPreviousPage = offset > 0
	}

	pageInfo := usecasex.NewPageInfo(totalCount, nil, nil, hasNextPage, hasPreviousPage)

	return pagedUsers, pageInfo, nil
}

func (r *User) FindByID(_ context.Context, v user.ID) (*user.User, error) {
	if r.err != nil {
		return nil, r.err
	}

	return rerror.ErrIfNil(r.data.Find(func(key user.ID, value *user.User) bool {
		return key == v
	}), rerror.ErrNotFound)
}

func (r *User) FindBySub(_ context.Context, auth0sub string) (*user.User, error) {
	if r.err != nil {
		return nil, r.err
	}

	if auth0sub == "" {
		return nil, rerror.ErrInvalidParams
	}

	return rerror.ErrIfNil(r.data.Find(func(key user.ID, value *user.User) bool {
		return value.ContainAuth(user.AuthFrom(auth0sub))
	}), rerror.ErrNotFound)
}

func (r *User) FindByPasswordResetRequest(_ context.Context, token string) (*user.User, error) {
	if r.err != nil {
		return nil, r.err
	}

	if token == "" {
		return nil, rerror.ErrInvalidParams
	}

	return rerror.ErrIfNil(r.data.Find(func(key user.ID, value *user.User) bool {
		return value.PasswordReset() != nil && value.PasswordReset().Token == token
	}), rerror.ErrNotFound)
}

func (r *User) FindByEmail(_ context.Context, email string) (*user.User, error) {
	if r.err != nil {
		return nil, r.err
	}

	if email == "" {
		return nil, rerror.ErrInvalidParams
	}

	return rerror.ErrIfNil(r.data.Find(func(key user.ID, value *user.User) bool {
		return value.Email() == email
	}), rerror.ErrNotFound)
}

func (r *User) FindByName(_ context.Context, name string) (*user.User, error) {
	if r.err != nil {
		return nil, r.err
	}

	if name == "" {
		return nil, rerror.ErrInvalidParams
	}

	return rerror.ErrIfNil(r.data.Find(func(key user.ID, value *user.User) bool {
		return value.Name() == name
	}), rerror.ErrNotFound)
}

func (r *User) FindByNameOrEmail(_ context.Context, nameOrEmail string) (*user.User, error) {
	if r.err != nil {
		return nil, r.err
	}

	if nameOrEmail == "" {
		return nil, rerror.ErrInvalidParams
	}

	return rerror.ErrIfNil(r.data.Find(func(key user.ID, value *user.User) bool {
		return value.Email() == nameOrEmail || value.Name() == nameOrEmail
	}), rerror.ErrNotFound)
}

func (r *User) FindByAlias(_ context.Context, alias string) (*user.User, error) {
	if r.err != nil {
		return nil, r.err
	}

	if alias == "" {
		return nil, rerror.ErrInvalidParams
	}

	return rerror.ErrIfNil(r.data.Find(func(key user.ID, value *user.User) bool {
		return value.Alias() == alias
	}), rerror.ErrNotFound)
}

func (r *User) SearchByKeyword(_ context.Context, keyword string, fields ...string) (user.List, error) {
	if r.err != nil {
		return nil, r.err
	}

	if len(keyword) < 3 {
		return nil, nil
	}

	// Reject email addresses as search keywords
	if isEmailAddress(keyword) {
		return nil, accountrepo.ErrInvalidKeyword
	}

	if len(fields) == 0 {
		fields = []string{"name"}
	}

	keyword = strings.TrimSpace(strings.ToLower(keyword))

	return rerror.ErrIfNil(r.data.FindAll(func(key user.ID, value *user.User) bool {
		for _, field := range fields {
			var fieldValue string
			switch field {
			case "email":
				fieldValue = value.Email()
			case "name":
				fieldValue = value.Name()
			case "alias":
				fieldValue = value.Alias()
			default:
				continue
			}
			if strings.Contains(strings.ToLower(fieldValue), keyword) {
				return true
			}
		}
		return false
	}), rerror.ErrNotFound)
}

func isEmailAddress(s string) bool {
	if s == "" {
		return false
	}
	_, err := mail.ParseAddress(s)
	return err == nil
}

func (r *User) FindByVerification(_ context.Context, code string) (*user.User, error) {
	if r.err != nil {
		return nil, r.err
	}

	if code == "" {
		return nil, rerror.ErrInvalidParams
	}

	return rerror.ErrIfNil(r.data.Find(func(key user.ID, value *user.User) bool {
		return value.Verification() != nil && value.Verification().Code() == code

	}), rerror.ErrNotFound)
}

func (r *User) FindBySubOrCreate(_ context.Context, u *user.User, sub string) (*user.User, error) {
	if r.err != nil {
		return nil, r.err
	}

	u2 := r.data.Find(func(key user.ID, value *user.User) bool {
		return value.ContainAuth(user.AuthFrom(sub))
	})
	if u2 == nil {
		r.data.Store(u.ID(), u)
		return u, nil
	}
	return u2, nil
}

func (r *User) Create(_ context.Context, u *user.User) error {
	if r.err != nil {
		return r.err
	}

	if _, ok := r.data.Load(u.ID()); !ok {
		r.data.Store(u.ID(), u)
	} else {
		return accountrepo.ErrDuplicatedUser
	}

	return nil
}

func (r *User) Save(_ context.Context, u *user.User) error {
	if r.err != nil {
		return r.err
	}

	r.data.Store(u.ID(), u)
	return nil
}

func (r *User) Remove(_ context.Context, user user.ID) error {
	if r.err != nil {
		return r.err
	}

	r.data.Delete(user)
	return nil
}

func SetUserError(r accountrepo.User, err error) {
	r.(*User).err = err
}
