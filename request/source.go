package request

// ------------------------------
// Models Helper
// ------------------------------

var NewFunc func(string, bool) interface{}
var GetFunc func(string) interface{}
var GetAllFunc func() map[string]interface{}

func New(candidate string, isList bool) interface{} {
	return NewFunc(candidate, isList)
}

func GetAll() map[string]interface{} {
	return GetAllFunc()
}

func Get(candidate string) interface{} {
	return GetFunc(candidate)
}

// ------------------------------
// Custom GraphQL
// ------------------------------

var Query interface{}
var Mutation interface{}
