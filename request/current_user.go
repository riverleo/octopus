package request

var UserModelName string

type (
	CurrentUser interface {
		HasId(id interface{}) bool
		HasRole(role string) bool
		HasProp(key string, value string) bool
	}

	AnonymousUser struct {
	}
)

// ------------------------------
// Anonymous User
// ------------------------------

func (u AnonymousUser) HasId(id interface{}) bool {
	return false
}

func (u AnonymousUser) HasRole(role string) bool {
	return role == "anonymous"
}

func (u AnonymousUser) HasProp(key string, value string) bool {
	return false
}

func (u AnonymousUser) String() string {
	return "anonymous"
}

var _ CurrentUser = (*AnonymousUser)(nil)

// ------------------------------
// Utils
// ------------------------------

func SetUserModelName(modelName string) {
	UserModelName = modelName
}
