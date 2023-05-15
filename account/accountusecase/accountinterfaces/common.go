package accountinterfaces

import (
	_ "embed"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/reearth/reearthx/i18n"
	"github.com/reearth/reearthx/rerror"
	"github.com/samber/lo"
)

var (
	ErrOperationDenied error = rerror.NewE(i18n.T("operation denied"))
	ErrInvalidOperator error = rerror.NewE(i18n.T("invalid operator"))
)

//go:embed ja.yml
var translationsJa []byte
var MessagesJa []*i18n.Message

func init() {
	mf := lo.Must(goi18n.ParseMessageFileBytes(translationsJa, "", nil))
	MessagesJa = mf.Messages
}
