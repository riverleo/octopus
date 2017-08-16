package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContains(t *testing.T) {
	str := "foo"
	strArray := []string{"foo", "bar"}

	assert.True(t, Contains(strArray, str))
	assert.False(t, Contains(strArray, "baz"))
}

func TestFlattenMap(t *testing.T) {
	names := map[string][]string{
		"1": []string{"a", "b", "c"},
		"2": []string{"c", "d", "e", "f"},
		"3": []string{"e", "f", "g", "h"},
	}

	assert.Contains(t, FlattenMap(names), "a", "b", "c", "d", "e", "f", "g", "h")
	assert.Len(t, FlattenMap(names), 8)
}

func TestCamelCase(t *testing.T) {
	assert.Equal(t, CamelCase("--camel_case  --"), "camelCase")
	assert.Equal(t, CamelCase("    camel case   "), "camelCase")
	assert.Equal(t, CamelCase(" -_ CAMELCa_se _-_ "), "cAMELCaSe")
	assert.Equal(t, CamelCase("camel-case func  "), "camelCaseFunc")
	assert.Equal(t, CamelCase("camel_case_func  "), "camelCaseFunc")
	assert.Equal(t, CamelCase("wanted_job_detail"), "wantedJobDetail")
	assert.Equal(t, CamelCase("WantedDes"), "wantedDes")
	assert.Equal(t, CamelCase("i18n"), "i18n")
}

func TestClassify(t *testing.T) {
	assert.Equal(t, Classify("  classify case "), "ClassifyCase")
	assert.Equal(t, Classify("class-ify func  "), "ClassIfyFunc")
	assert.Equal(t, Classify("class_ify_func  "), "ClassIfyFunc")
	assert.Equal(t, Classify("wanted_job_detail"), "WantedJobDetail")
	assert.Equal(t, Classify("i18n"), "I18n")
}

func TestSnakeCase(t *testing.T) {
	assert.Equal(t, SnakeCase("  snake case "), "snake_case")
	assert.Equal(t, SnakeCase("--SnakeCase__"), "snake_case")
	assert.Equal(t, SnakeCase("--SNAKECase__"), "snake_case")
}

func TestIsSimilar(t *testing.T) {
	assert.True(t, IsSimilar("WantedDes", "wanted_des"))
	assert.True(t, IsSimilar("wanted_des", "WantedDes"))
	assert.True(t, IsSimilar("wantedDes", "wanted_des"))
	assert.True(t, IsSimilar("wanted_des", "wantedDes"))
	assert.True(t, IsSimilar("WANTED_DES", "wanted_des"))
	assert.True(t, IsSimilar("---WANTED_des---", "wanted_des"))
}

func TestEncapCase(t *testing.T) {
	assert.Equal(t, EncapCase("get", "wanted_des"), "GetWantedDes")
	assert.Equal(t, EncapCase("set", "   wanted_des"), "SetWantedDes")
}

func TestLowerFirst(t *testing.T) {
	assert.Equal(t, LowerFirst("WantedDes"), "wantedDes")
	assert.Equal(t, LowerFirst("wantedDes"), "wantedDes")
}

func TestUpperFirst(t *testing.T) {
	assert.Equal(t, UpperFirst("WantedDes"), "WantedDes")
	assert.Equal(t, UpperFirst("wantedDes"), "WantedDes")
}
