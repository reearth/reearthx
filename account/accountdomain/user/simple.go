package user

type Simple struct {
	ID    ID
	Name  string
	Email string
}

func SimpleFrom(u *User) *Simple {
	if u == nil {
		return nil
	}
	return &Simple{u.ID(), u.Name(), u.Email()}
}
