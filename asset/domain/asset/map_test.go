package asset

import (
	"testing"

	"github.com/reearth/reearthx/account/accountdomain"
	"github.com/stretchr/testify/assert"
)

func TestMap_List(t *testing.T) {
	pid := NewProjectID()
	uid := accountdomain.NewUserID()

	a := New().NewID().
		Project(pid).
		CreatedByUser(uid).
		Size(1000).
		Thread(NewThreadID().Ref()).
		NewUUID().
		MustBuild()

	assert.Equal(t, List{a}, Map{a.ID(): a}.List())
	assert.Equal(t, List{}, Map(nil).List())
}

func TestMap_ListFrom(t *testing.T) {
	pid := NewProjectID()
	uid := accountdomain.NewUserID()

	a := New().NewID().
		Project(pid).
		CreatedByUser(uid).
		Size(1000).
		Thread(NewThreadID().Ref()).
		NewUUID().
		MustBuild()

	assert.Equal(t, List{a}, Map{a.ID(): a}.ListFrom(IDList{a.ID()}))
	assert.Nil(t, Map(nil).ListFrom(nil))
}
