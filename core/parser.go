package core

import (
	"fmt"
	"reflect"
)

func ParseMap(raw interface{}) (result map[string]interface{}) {
	result = map[string]interface{}{}

	if !IsKindOf(raw, reflect.Map) {
		return
	}

	rawMap := reflect.ValueOf(raw)
	for _, key := range rawMap.MapKeys() {
		result[fmt.Sprintf("%v", key)] = rawMap.MapIndex(key).Interface()
	}

	return
}
