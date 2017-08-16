package request

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestGetAuthority(t *testing.T) {
	body := `default:
  read: hasRole("user")
  write: hasId(.userId)
models:
  book:
    read:
      fields:
        lastIp: hasRole("developer")
        password:
          - hasId(.userId)
          - hasRole("admin")
    write:
      default: hasRole("author")
  article: hasRole("user")`
	var rawAuthority map[interface{}]interface{}
	err := yaml.Unmarshal([]byte(body), &rawAuthority)
	assert.Nil(t, err)
	authority := parseAuthority(rawAuthority)

	assert.True(t, authority.Default.Read.IsHasRole())
	assert.True(t, authority.Default.Write.IsHasId())
	assert.Empty(t, authority.Default.Read.Field)
	assert.Equal(t, authority.Default.Write.Field, "userId")
	assert.Contains(t, authority.Default.Read.Values, "user")
	assert.Empty(t, authority.Default.Write.Values)
}

func TestAuthority_AnalyzeRead_SpecificModel(t *testing.T) {
	r := Request{
		Name:      "anonymous",
		Operation: "query",
		Node: &Node{
			Name: "user",
			Type: "User",
			Fields: map[string]*Node{
				"id": &Node{Name: "id", Type: "Int"},
			},
		},
	}

	r.SetUp()

	authority := parseAuthority(map[string]interface{}{
		"default": "hasRole(\"admin\")",
		"models": map[string]interface{}{
			"user": "hasId(.id)",
		},
	})

	validatorMap, fields := authority.AnalyzeRead(r.Node)
	assert.Len(t, validatorMap, 1)
	assert.Contains(t, validatorMap["id"], Validator{Expression: "hasId", Field: "id"})
	assert.Equal(t, fields, []string{"id"})
}

func TestAuthority_AnalyzeRead_Leaf(t *testing.T) {
	r := Request{
		Name:      "anonymous",
		Operation: "query",
		Node: &Node{
			Name: "user",
			Type: "User",
			Fields: map[string]*Node{
				"id":       &Node{Name: "id", Type: "Int"},
				"about":    &Node{Name: "about", Type: "String"},
				"password": &Node{Name: "password", Type: "String"},
			},
		},
	}

	r.SetUp()

	authority := parseAuthority(map[string]interface{}{
		"default": "hasRole(\"admin\")",
		"models": map[string]interface{}{
			"user": map[string]interface{}{
				"read": map[string]interface{}{
					"default": "hasRole(\"user\")",
					"fields": map[string]interface{}{
						"about": "hasRole(\"headhunter\")",
						"password": []string{
							"hasId(.id)",
							"hasRole(\"admin\")",
						},
					},
				},
			},
		},
	})

	validatorMap, fields := authority.AnalyzeRead(r.Node)
	assert.Contains(t, fields, "id")

	// ID
	assert.Len(t, validatorMap["id"], 1)
	assert.Contains(t, validatorMap["id"], Validator{Expression: "hasRole", Values: []string{"user"}})

	// About
	assert.Len(t, validatorMap["about"], 1)
	assert.Contains(t, validatorMap["about"], Validator{Expression: "hasRole", Values: []string{"headhunter"}})

	// Password
	assert.Len(t, validatorMap["password"], 2)
	assert.Contains(t, validatorMap["password"], Validator{Expression: "hasId", Field: "id"})
	assert.Contains(t, validatorMap["password"], Validator{Expression: "hasRole", Values: []string{"admin"}})
}

func TestAuthority_AnalyzeRead_ChildNode(t *testing.T) {
	r := Request{
		Name:      "anonymous",
		Operation: "query",
		Node: &Node{
			Name: "user",
			Type: "User",
			Fields: map[string]*Node{
				"id":       &Node{Name: "id", Type: "Int"},
				"about":    &Node{Name: "about", Type: "String"},
				"password": &Node{Name: "password", Type: "String"},
				"apply": &Node{
					Name: "apply",
					Type: "Apply",
				},
			},
		},
	}

	r.SetUp()

	authority := parseAuthority(map[string]interface{}{
		"default": "hasRole(\"admin\")",
		"models": map[string]interface{}{
			"user": map[string]interface{}{
				"read": map[string]interface{}{
					"default": "hasRole(\"user\")",
					"fields": map[string]interface{}{
						"about": "hasRole(\"headhunter\")",
						"apply": "hasId(.userId)",
						"password": []string{
							"hasId(.id)",
							"hasRole(\"admin\")",
						},
					},
				},
			},
		},
	})

	validatorMap, fields := authority.Analyze(r.Node)
	assert.Contains(t, fields, "id", "userId")

	// ID
	assert.Len(t, validatorMap["id"], 1)
	assert.Contains(t, validatorMap["id"], Validator{Expression: "hasRole", Values: []string{"user"}})

	// About
	assert.Len(t, validatorMap["about"], 1)
	assert.Contains(t, validatorMap["about"], Validator{Expression: "hasRole", Values: []string{"headhunter"}})

	// Password
	assert.Len(t, validatorMap["password"], 2)
	assert.Contains(t, validatorMap["password"], Validator{Expression: "hasId", Field: "id"})
	assert.Contains(t, validatorMap["password"], Validator{Expression: "hasRole", Values: []string{"admin"}})

	// Apply
	assert.Len(t, validatorMap["apply"], 1)
	assert.Contains(t, validatorMap["apply"], Validator{Expression: "hasId", Field: "userId"})
}

func TestParseAuthority(t *testing.T) {
	raw := map[string]interface{}{
		"default": "hasRole(\"admin\")",
		"models": map[string]interface{}{
			"user": "hasId(.userId)",
		},
	}

	authority := parseAuthority(raw)
	assert.True(t, authority.Default.Read.IsHasRole())
	assert.True(t, authority.Default.Write.IsHasRole())
	assert.NotEmpty(t, authority.Models["User"])
	assert.True(t, authority.Models["User"].Read.Default.IsHasId())
	assert.True(t, authority.Models["User"].Write.Default.IsHasId())

	raw = map[string]interface{}{
		"default": "hasRole(\"admin\")",
		"models": map[string]interface{}{
			"user": map[string]interface{}{
				"read": map[string]interface{}{
					"default": "hasRole(\"user\")",
					"fields": map[string]interface{}{
						"password": "hasId(.id)",
					},
				},
			},
		},
	}

	authority = parseAuthority(raw)
	assert.True(t, authority.Default.Read.IsHasRole())
	assert.Equal(t, authority.Default.Read.Values, []string{"admin"})
	assert.True(t, authority.Default.Write.IsHasRole())
	assert.Equal(t, authority.Default.Write.Values, []string{"admin"})
	assert.NotEmpty(t, authority.Models["User"])
	assert.True(t, authority.Models["User"].Read.Default.IsHasRole())
	assert.Equal(t, authority.Models["User"].Read.Default.Values, []string{"user"})
	assert.True(t, authority.Models["User"].Write.Default.IsHasRole())
	assert.Equal(t, authority.Models["User"].Write.Default.Values, []string{"admin"})
}

func TestParseAuthority_SkipDefault(t *testing.T) {
	raw := map[string]interface{}{
		"models": map[string]interface{}{
			"user": "hasId(.userId)",
		},
	}

	authority := parseAuthority(raw)
	assert.True(t, authority.Default.Read.IsAll())
	assert.True(t, authority.Default.Write.IsAll())
	assert.NotEmpty(t, authority.Models["User"])
	assert.True(t, authority.Models["User"].Read.Default.IsHasId())
	assert.True(t, authority.Models["User"].Write.Default.IsHasId())
}

func TestParseAuthority_SkipModels(t *testing.T) {
	raw := map[string]interface{}{
		"default": "hasRole(\"admin\")",
	}

	authority := parseAuthority(raw)
	assert.True(t, authority.Default.Read.IsHasRole())
	assert.True(t, authority.Default.Write.IsHasRole())
	assert.Empty(t, authority.Models)
}

func TestParseAuthorityModel_Shortcut(t *testing.T) {
	defaults := DefaultAuthority{}
	defaults.Read = parseValidator("hasId(.uniqueValidator)")
	defaults.Write = parseValidator("hasId(.uniqueValidator)")

	am := parseAuthorityModel("hasRole(\"developer\")", defaults)
	assert.NotEqual(t, am.Read.Default, defaults.Read)
	assert.NotEqual(t, am.Write.Default, defaults.Write)
	assert.True(t, am.Read.Default.IsHasRole())
	assert.True(t, am.Write.Default.IsHasRole())
	assert.Len(t, am.Read.Fields, 0)
	assert.Len(t, am.Write.Fields, 0)
}

func TestParseAuthorityModel_SkipWrite(t *testing.T) {
	defaults := DefaultAuthority{}
	defaults.Read = parseValidator("hasId(.uniqueValidator)")
	defaults.Write = parseValidator("hasId(.uniqueValidator)")

	raw := map[string]interface{}{
		"read": "hasRole(\"developer\")",
	}

	am := parseAuthorityModel(raw, defaults)
	assert.NotEqual(t, am.Read.Default, defaults.Read)
	assert.Equal(t, am.Write.Default, defaults.Write)
	assert.True(t, am.Read.Default.IsHasRole())
	assert.True(t, am.Write.Default.IsHasId())
}

func TestParseAuthorityModel_SkipRead(t *testing.T) {
	defaults := DefaultAuthority{}
	defaults.Read = parseValidator("hasId(.uniqueValidator)")
	defaults.Write = parseValidator("hasId(.uniqueValidator)")

	raw := map[string]interface{}{
		"write": "hasRole(\"developer\")",
	}

	am := parseAuthorityModel(raw, defaults)
	assert.Equal(t, am.Read.Default, defaults.Read)
	assert.NotEqual(t, am.Write.Default, defaults.Write)
	assert.True(t, am.Read.Default.IsHasId())
	assert.True(t, am.Write.Default.IsHasRole())
	assert.Len(t, am.Read.Fields, 0)
	assert.Len(t, am.Write.Fields, 0)
}

func TestParseAuthorityModel_SkipDefault(t *testing.T) {
	defaults := DefaultAuthority{}
	defaults.Read = parseValidator("hasId(.uniqueValidator)")
	defaults.Write = parseValidator("hasId(.uniqueValidator)")

	raw := map[string]interface{}{
		"read":  "hasRole(\"developer\")",
		"write": "hasRole(\"user\")",
	}

	am := parseAuthorityModel(raw, defaults)
	assert.NotEqual(t, am.Read.Default, defaults.Read)
	assert.NotEqual(t, am.Write.Default, defaults.Write)
	assert.True(t, am.Read.Default.IsHasRole())
	assert.True(t, am.Write.Default.IsHasRole())
}

func TestParsePermission_Shortcut(t *testing.T) {
	raw := "hasRole(\"admin\")"
	defaultValidator := parseValidator("hasId(.uniqueValidator)")
	permission := parsePermission(raw, defaultValidator)

	assert.NotEqual(t, permission.Default, defaultValidator)
	assert.True(t, permission.Default.IsHasRole())
	assert.Contains(t, permission.Default.Values, "admin")
	assert.Len(t, permission.Fields, 0)
}

func TestParsePermission_SkipDefault(t *testing.T) {
	raw := map[string]interface{}{
		"fields": map[string]interface{}{
			"lastIp": "hasRole(\"admin\")",
			"password": []interface{}{
				"hasId(.userId)",
				"hasRole(\"developer\")",
			},
		},
	}
	defaultValidator := parseValidator("hasId(.uniqueValidator)")
	permission := parsePermission(raw, defaultValidator)

	assert.Equal(t, permission.Default, defaultValidator)
	assert.Len(t, permission.Fields, 2)
	assert.Len(t, permission.Fields["password"], 2)
	assert.True(t, permission.Fields["password"][0].IsHasId())
	assert.True(t, permission.Fields["password"][1].IsHasRole())
}

func TestParsePermission_DefaultOnly(t *testing.T) {
	raw := map[string]interface{}{
		"default": "hasId(.userId)",
	}
	defaultValidator := parseValidator("hasId(.uniqueValidator)")
	permission := parsePermission(raw, defaultValidator)

	assert.NotEqual(t, permission.Default, defaultValidator)
	assert.True(t, permission.Default.IsHasId())
	assert.Len(t, permission.Fields, 0)
}

func TestParsePermission(t *testing.T) {
	raw := map[string]interface{}{
		"default": "hasId(.userId)",
		"fields": map[string]interface{}{
			"lastIp":   "hasRole(\"admin\")",
			"password": "hasId(.userId)",
		},
	}
	defaultValidator := parseValidator("hasId(.uniqueValidator)")
	permission := parsePermission(raw, defaultValidator)

	assert.NotEqual(t, permission.Default, defaultValidator)
	assert.Len(t, permission.Fields, 2)
}

func TestParseValidator(t *testing.T) {
	hasId := "hasId(.userId)"
	hasRole := "hasRole(\"admin\")"
	hasRoleMulti := "hasRole(\"admin\", \"user\")"

	matcher := parseValidator(hasId)
	assert.True(t, matcher.IsHasId())
	assert.Equal(t, matcher.Expression, "hasId")
	assert.Equal(t, matcher.Field, "userId")
	assert.Empty(t, matcher.Values)

	matcher = parseValidator(hasRole)
	assert.True(t, matcher.IsHasRole())
	assert.Equal(t, matcher.Expression, "hasRole")
	assert.Empty(t, matcher.Field)
	assert.Contains(t, matcher.Values, "admin")

	matcher = parseValidator(hasRoleMulti)
	assert.True(t, matcher.IsHasRole())
	assert.Equal(t, matcher.Expression, "hasRole")
	assert.Empty(t, matcher.Field)
	assert.Contains(t, matcher.Values, "admin", "user")
}

func TestParseValidator_Invalid(t *testing.T) {
	invalidField := "hasRole(invalidField)"
	invalidValue := "hasRole('invalidValue\")"
	invalidExpress := "invalidExpress(.userId)"

	assert.Panics(t, func() { parseValidator(invalidField) })
	assert.Panics(t, func() { parseValidator(invalidValue) })
	assert.Panics(t, func() { parseValidator(invalidExpress) })
}
