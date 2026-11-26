package mongodoc

import (
	"time"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/reearth/reearthx/account/accountdomain/user"
	"github.com/reearth/reearthx/mongox"
	"github.com/samber/lo"
)

type PasswordResetDocument struct {
	Token     string
	CreatedAt time.Time
}

type UserDocument struct {
	ID            string
	Name          string
	Alias         string
	Email         string
	Subs          []string
	Workspace     string
	Team          string `bson:",omitempty"`
	Lang          string
	Theme         string
	Password      []byte
	PasswordReset *PasswordResetDocument
	Verification  *UserVerificationDoc
	Metadata      *UserMetadataDoc
	// Fields shared with reearth-accounts. ReplaceOne writes the whole document, so they must round-trip.
	LatestLogoutAt *time.Time `bson:"latestlogoutat,omitempty"`
	CreatedAt      *time.Time `bson:"createdat,omitempty"`
	UpdatedAt      time.Time  `bson:"updatedat"`
	DeletedAt      *time.Time `bson:"deletedat,omitempty"`
}

type UserVerificationDoc struct {
	Code       string
	Expiration time.Time
	Verified   bool
}

type UserMetadataDoc struct {
	Description string
	Website     string
	PhotoURL    string
	Lang        string
	Theme       string
}

func NewUser(user *user.User) (*UserDocument, string) {
	id := user.ID().String()
	auths := user.Auths()
	authsdoc := make([]string, 0, len(auths))
	for _, a := range auths {
		authsdoc = append(authsdoc, a.Sub)
	}
	var v *UserVerificationDoc
	if user.Verification() != nil {
		v = &UserVerificationDoc{
			Code:       user.Verification().Code(),
			Expiration: user.Verification().Expiration(),
			Verified:   user.Verification().IsVerified(),
		}
	}
	pwdReset := user.PasswordReset()

	var pwdResetDoc *PasswordResetDocument
	if pwdReset != nil {
		pwdResetDoc = &PasswordResetDocument{
			Token:     pwdReset.Token,
			CreatedAt: pwdReset.CreatedAt,
		}
	}

	metadataDoc := &UserMetadataDoc{}
	if user.Metadata() != nil {
		metadataDoc = &UserMetadataDoc{
			Description: user.Metadata().Description(),
			Website:     user.Metadata().Website(),
			PhotoURL:    user.Metadata().PhotoURL(),
			Lang:        user.Metadata().Lang().String(),
			Theme:       string(user.Metadata().Theme()),
		}
	}

	return &UserDocument{
		ID:             id,
		Name:           user.Name(),
		Alias:          user.Alias(),
		Email:          user.Email(),
		Subs:           authsdoc,
		Workspace:      user.Workspace().String(),
		Verification:   v,
		Password:       user.Password(),
		PasswordReset:  pwdResetDoc,
		Metadata:       metadataDoc,
		LatestLogoutAt: nonZeroTime(user.LatestLogoutAt()),
		CreatedAt:      createdAt(user.CreatedAt(), user.ID().Timestamp()),
		UpdatedAt:      time.Now(),
		DeletedAt:      user.DeletedAt(),
	}, id
}

func (d *UserDocument) Model() (*user.User, error) {
	uid, err := accountdomain.UserIDFrom(d.ID)
	if err != nil {
		return nil, err
	}

	wid := d.Workspace
	if wid == "" {
		wid = d.Team
	}

	tid, err := accountdomain.WorkspaceIDFrom(wid)
	if err != nil {
		return nil, err
	}

	auths := make([]user.Auth, 0, len(d.Subs))
	for _, s := range d.Subs {
		auths = append(auths, user.AuthFrom(s))
	}

	var v *user.Verification
	if d.Verification != nil {
		v = user.VerificationFrom(d.Verification.Code, d.Verification.Expiration, d.Verification.Verified)
	}

	var metadata *user.Metadata
	if d.Metadata != nil {
		metadata = user.NewMetadata()
		metadata.SetDescription(d.Metadata.Description)
		metadata.SetWebsite(d.Metadata.Website)
		metadata.SetPhotoURL(d.Metadata.PhotoURL)
		metadata.LangFrom(d.Metadata.Lang)
		metadata.SetTheme(user.Theme(d.Metadata.Theme))
	}

	u, err := user.New().
		ID(uid).
		Name(d.Name).
		Email(d.Email).
		Metadata(metadata).
		Alias(d.Alias).
		Auths(auths).
		Workspace(tid).
		Verification(v).
		EncodedPassword(d.Password).
		PasswordReset(d.PasswordReset.Model()).
		LatestLogoutAt(lo.FromPtr(d.LatestLogoutAt)).
		CreatedAt(d.CreatedAt).
		DeletedAt(d.DeletedAt).
		Build()

	if err != nil {
		return nil, err
	}
	return u, nil
}

func (d *PasswordResetDocument) Model() *user.PasswordReset {
	if d == nil {
		return nil
	}
	return &user.PasswordReset{
		Token:     d.Token,
		CreatedAt: d.CreatedAt,
	}
}

type UserConsumer = mongox.SliceFuncConsumer[*UserDocument, *user.User]

func NewUserConsumer(host string) *UserConsumer {
	return mongox.NewSliceFuncConsumer(func(d *UserDocument) (*user.User, error) {
		m, err := d.Model()
		if err != nil {
			return nil, err
		}
		return m.WithHost(host), nil
	})
}

func nonZeroTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// createdAt falls back to the creation time encoded in the entity's ULID when the model has none:
// for new entities that is their creation time, and documents written before createdat existed get an
// accurate value instead of the time of their next save.
func createdAt(t *time.Time, idTime time.Time) *time.Time {
	if t != nil {
		return t
	}
	return nonZeroTime(idTime.UTC())
}
