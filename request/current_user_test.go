package request

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetUserModel(t *testing.T) {
	assert.Empty(t, UserModelName)

	SetUserModelName("User")

	assert.Equal(t, UserModelName, "User")
}

func TestAnonymousUser_HasId(t *testing.T) {
	u := AnonymousUser{}

	assert.False(t, u.HasId(1))
	assert.False(t, u.HasId("foo"))
	assert.False(t, u.HasId(3.14))
}

func TestAnonymousUser_HasRole(t *testing.T) {
	u := AnonymousUser{}

	assert.False(t, u.HasRole("user"))
	assert.False(t, u.HasRole("admin"))
	assert.True(t, u.HasRole("anonymous"))
}

func TestAnonymousUser_HasProp(t *testing.T) {
	u := AnonymousUser{}

	assert.False(t, u.HasProp("lang", "ko"))
	assert.False(t, u.HasProp("country", "kr"))
	assert.False(t, u.HasProp("company", "WANTED"))
}

func TestAnonymousUser_String(t *testing.T) {
	u := AnonymousUser{}

	assert.Equal(t, u.String(), "anonymous")
}
