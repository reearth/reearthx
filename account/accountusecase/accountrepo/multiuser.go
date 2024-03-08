package accountrepo

/* TODO
type MultiUser []User

func NewMultiUser(users ...User) MultiUser {
	return MultiUser(users)
}

var _ User = MultiUser{}

func (u MultiUser) FindByID(ctx context.Context, id user.ID) (*user.User, error) {
	for _, user := range u {
		if res, err := user.FindByID(ctx, id); err != nil && !errors.Is(err, rerror.ErrNotFound) {
			return nil, err
		} else if res != nil {
			return res, nil
		}
	}
	return nil, nil
}

func (u MultiUser) FindByIDs(context.Context, user.IDList) (user.List, error) {

}

func (u MultiUser) FindBySub(context.Context, string) (*user.User, error) {

}

func (u MultiUser) FindByEmail(context.Context, string) (*user.User, error) {

}

func (u MultiUser) FindByName(context.Context, string) (*user.User, error) {

}

func (u MultiUser) FindByNameOrEmail(context.Context, string) (*user.User, error) {

}

func (u MultiUser) FindByVerification(context.Context, string) (*user.User, error) {

}

func (u MultiUser) FindByPasswordResetRequest(context.Context, string) (*user.User, error) {

}

func (u MultiUser) FindBySubOrCreate(context.Context, *user.User, string) (*user.User, error) {

}
*/
