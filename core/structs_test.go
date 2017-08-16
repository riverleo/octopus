package core

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	var s = struct {
		Id   int
		Name string
	}{1, "Leo"}

	assert.Equal(t, Get(s, "Id"), 1)
	assert.Equal(t, Get(s, "Name"), "Leo")
	assert.Nil(t, Get(s, "Anonymous"))
}

func TestGetFromList(t *testing.T) {
	var s = []struct {
		Id   int
		Name string
	}{
		{1, "Leo"},
		{2, "Wanted"},
		{3, "WeWork"},
		{4, "Google"},
		{5, "Facebook"},
	}

	assert.Equal(t, GetFromList(s, "Id"), []interface{}{1, 2, 3, 4, 5})
	assert.Equal(t, GetFromList(s, "Name"), []interface{}{"Leo", "Wanted", "WeWork", "Google", "Facebook"})
	assert.Equal(t, GetFromList(s, "Anonymous"), []interface{}{nil, nil, nil, nil, nil})
	assert.Equal(t, GetFromList(nil, "Id"), []interface{}{})
}

func TestIsKindOf(t *testing.T) {
	intValue := 3
	floatValue := 3.14
	strValue := "foo"
	sliceValue := []string{"bar", "baz"}

	assert.True(t, IsKindOf(intValue, reflect.Int))
	assert.False(t, IsKindOf(intValue, reflect.Int32))
	assert.False(t, IsKindOf(intValue, reflect.Int64))
	assert.False(t, IsKindOf(floatValue, reflect.Float32))
	assert.True(t, IsKindOf(floatValue, reflect.Float64))
	assert.True(t, IsKindOf(strValue, reflect.String))
	assert.True(t, IsKindOf(sliceValue, reflect.Slice))
}

func TestGetByName(t *testing.T) {
	var s = []struct {
		Id   int
		Name string
	}{
		{1, "Leo"},
		{2, "Wanted"},
		{3, "WeWork"},
		{4, "Google"},
		{5, "Facebook"},
	}
	extracted := GetByName(s, "Id", "Name", "Anonymous")
	assert.Equal(t, extracted["Id"], []interface{}{1, 2, 3, 4, 5})
	assert.Equal(t, extracted["Name"], []interface{}{"Leo", "Wanted", "WeWork", "Google", "Facebook"})
	assert.Equal(t, extracted["Anonymous"], []interface{}{nil, nil, nil, nil, nil})
}

func TestSortAsValues(t *testing.T) {
	structs := []struct {
		Id   int64
		Name string
	}{
		{1, "Leo"},
		{2, "WeWork"},
		{3, "Wanted"},
		{4, "Google"},
		{5, "Facebook"},
	}

	sorted, err := SortAsValues(structs, []interface{}{4, 3, nil, 1}, nil, "Id")

	assert.Nil(t, err)
	assert.Len(t, sorted, 4)
	assert.Equal(t, reflect.ValueOf(sorted[0]).FieldByName("Id").Interface(), int64(4))
	assert.Equal(t, reflect.ValueOf(sorted[1]).FieldByName("Id").Interface(), int64(3))
	assert.Equal(t, sorted[2], nil)
	assert.Equal(t, reflect.ValueOf(sorted[3]).FieldByName("Id").Interface(), int64(1))

	assert.Equal(t, reflect.ValueOf(sorted[0]).FieldByName("Name").Interface(), "Google")
	assert.Equal(t, reflect.ValueOf(sorted[1]).FieldByName("Name").Interface(), "Wanted")
	assert.Equal(t, sorted[2], nil)
	assert.Equal(t, reflect.ValueOf(sorted[3]).FieldByName("Name").Interface(), "Leo")
}

func TestSortAsValuesByMap(t *testing.T) {
	structs := []map[string]interface{}{
		{"Id": 1, "Name": "Leo"},
		{"Id": 2, "Name": "WeWork"},
		{"Id": 3, "Name": "Wanted"},
		{"Id": 4, "Name": "Google"},
		{"Id": 5, "Name": "Facebook"},
	}

	sorted, err := SortAsValues(structs, []interface{}{4, 3, nil, 1}, nil, "Id")

	assert.Nil(t, err)
	assert.Len(t, sorted, 4)
	assert.Equal(t, sorted[0].(map[string]interface{})["Id"], 4)
	assert.Equal(t, sorted[1].(map[string]interface{})["Id"], 3)
	assert.Equal(t, sorted[2], nil)
	assert.Equal(t, sorted[3].(map[string]interface{})["Id"], 1)

	assert.Equal(t, sorted[0].(map[string]interface{})["Name"], "Google")
	assert.Equal(t, sorted[1].(map[string]interface{})["Name"], "Wanted")
	assert.Equal(t, sorted[2], nil)
	assert.Equal(t, sorted[3].(map[string]interface{})["Name"], "Leo")
}

func TestSortAsValuesByFieldName(t *testing.T) {
	structs := []struct {
		Id   int64
		Name string
	}{
		{1, "Leo"},
		{2, "WeWork"},
		{3, "Wanted"},
		{4, "Google"},
		{5, "Facebook"},
	}

	sorted, err := SortAsValues(structs, []interface{}{4, 3, nil, 1}, nil, "Id", "Name")

	assert.Nil(t, err)
	assert.Len(t, sorted, 4)
	assert.Equal(t, sorted[0], "Google")
	assert.Equal(t, sorted[1], "Wanted")
	assert.Equal(t, sorted[2], nil)
	assert.Equal(t, sorted[3], "Leo")
}

func TestSortAsValuesByMapByFieldName(t *testing.T) {
	structs := []map[string]interface{}{
		{"Id": 1, "Name": "Leo"},
		{"Id": 2, "Name": "WeWork"},
		{"Id": 3, "Name": "Wanted"},
		{"Id": 4, "Name": "Google"},
		{"Id": 5, "Name": "Facebook"},
	}

	sorted, err := SortAsValues(structs, []interface{}{4, 3, nil, 1}, nil, "Id", "Name")

	assert.Nil(t, err)
	assert.Len(t, sorted, 4)
	assert.Equal(t, sorted[0], "Google")
	assert.Equal(t, sorted[1], "Wanted")
	assert.Equal(t, sorted[2], nil)
	assert.Equal(t, sorted[3], "Leo")
}

func TestSortAsValuesByMapByDuplicated(t *testing.T) {
	structs := []map[string]interface{}{
		{"Id": 1, "Name": "Leo"},
		{"Id": 2, "Name": "WeWork"},
		{"Id": 3, "Name": "Wanted"},
		{"Id": 4, "Name": "Google"},
		{"Id": 5, "Name": "Facebook"},
	}

	sorted, err := SortAsValues(structs, []interface{}{4, 3, 3, nil, 1, 1}, nil, "Id", "Name")

	assert.Nil(t, err)
	assert.Len(t, sorted, 6)
	assert.Equal(t, sorted[0], "Google")
	assert.Equal(t, sorted[1], "Wanted")
	assert.Equal(t, sorted[2], "Wanted")
	assert.Equal(t, sorted[3], nil)
	assert.Equal(t, sorted[4], "Leo")
	assert.Equal(t, sorted[5], "Leo")
}
