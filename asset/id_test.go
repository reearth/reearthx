package asset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssetID(t *testing.T) {
	id := NewAssetID()
	assert.NotEqual(t, AssetID{}, id)
	assert.NotEmpty(t, id.String())

	idStr := id.String()
	assert.NotEmpty(t, idStr)

	parsedID, err := AssetIDFrom(idStr)
	assert.NoError(t, err)
	assert.Equal(t, id, parsedID)

	_, err = AssetIDFrom("invalid-id")
	assert.Error(t, err)

	var emptyID AssetID
	assert.True(t, emptyID.IsNil())
	assert.False(t, id.IsNil())
}

func TestGroupID(t *testing.T) {
	id := NewGroupID()
	assert.NotEqual(t, GroupID{}, id)
	assert.NotEmpty(t, id.String())

	idStr := id.String()
	assert.NotEmpty(t, idStr)

	parsedID, err := GroupIDFrom(idStr)
	assert.NoError(t, err)
	assert.Equal(t, id, parsedID)

	_, err = GroupIDFrom("invalid-id")
	assert.Error(t, err)

	var emptyID GroupID
	assert.True(t, emptyID.IsNil())
	assert.False(t, id.IsNil())
}

func TestPolicyID(t *testing.T) {
	id := NewPolicyID()
	assert.NotEqual(t, PolicyID{}, id)
	assert.NotEmpty(t, id.String())

	idStr := id.String()
	assert.NotEmpty(t, idStr)

	parsedID, err := PolicyIDFrom(idStr)
	assert.NoError(t, err)
	assert.Equal(t, id, parsedID)

	_, err = PolicyIDFrom("invalid-id")
	assert.Error(t, err)

	var emptyID PolicyID
	assert.True(t, emptyID.IsNil())
	assert.False(t, id.IsNil())
}

func TestIDEquality(t *testing.T) {
	id1 := NewAssetID()
	id2 := NewAssetID()
	idCopy := id1

	assert.NotEqual(t, id1, id2)
	assert.Equal(t, id1, idCopy)

	gid1 := NewGroupID()
	gid2 := NewGroupID()
	gidCopy := gid1

	assert.NotEqual(t, gid1, gid2)
	assert.Equal(t, gid1, gidCopy)

	pid1 := NewPolicyID()
	pid2 := NewPolicyID()
	pidCopy := pid1

	assert.NotEqual(t, pid1, pid2)
	assert.Equal(t, pid1, pidCopy)
}

func TestIDStringFormat(t *testing.T) {
	id := NewAssetID()
	idStr := id.String()

	assert.Len(t, idStr, 26)

	gid := NewGroupID()
	gidStr := gid.String()

	assert.Len(t, gidStr, 26)

	pid := NewPolicyID()
	pidStr := pid.String()

	assert.Len(t, pidStr, 26)
}

func TestMustIDFunctions(t *testing.T) {
	validID := NewAssetID().String()
	mustID := MustAssetID(validID)
	assert.NotEqual(t, AssetID{}, mustID)

	validGroupID := NewGroupID().String()
	mustGID := MustGroupID(validGroupID)
	assert.NotEqual(t, GroupID{}, mustGID)

	validPolicyID := NewPolicyID().String()
	mustPID := MustPolicyID(validPolicyID)
	assert.NotEqual(t, PolicyID{}, mustPID)

}
