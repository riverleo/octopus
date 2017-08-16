package core

import (
	"fmt"
	"reflect"
)

func IsKindOf(v interface{}, kind reflect.Kind) bool {
	return v != nil && reflect.TypeOf(v).Kind() == kind
}

func Get(obj interface{}, name string) (extracted interface{}) {
	value := reflect.ValueOf(obj)

	if value.Kind() == reflect.Ptr {
		value = reflect.Indirect(value)
	}

	if IsKindOf(value.Interface(), reflect.Slice) {
		return GetFromList(value.Interface(), name)
	}

	var field reflect.Value
	if value.Kind() == reflect.Map {
		field = value.MapIndex(reflect.ValueOf(name))
	} else {
		field = value.FieldByName(Classify(name))
	}

	if field.IsValid() {
		extracted = field.Interface()
	}

	return
}

func GetFromList(objs interface{}, name string) (extracted []interface{}) {
	vs := reflect.ValueOf(objs)

	if !vs.IsValid() || vs.Kind() != reflect.Slice {
		return []interface{}{}
	}

	for i := 0; i < vs.Len(); i++ {
		v := vs.Index(i).FieldByName(Classify(name))

		if v.IsValid() {
			extracted = append(extracted, v.Interface())
		} else {
			extracted = append(extracted, nil)
		}
	}

	return
}

func GetByName(objs interface{}, names ...string) (extracted map[string]interface{}) {
	extracted = map[string]interface{}{}

	for _, name := range names {
		extracted[name] = Get(objs, name)
	}

	return
}

// objs 배열을 fields[0] 값과 비교하여 fieldValues 배열과 동일한 순서로 정렬합니다.
// fieldValues 배열의 nil 값은 순서를 유지하기 위한 의도로 인지하여 반환되는 배열에도 동일하게 유지합니다.
// fields[1] 값이 존재할 경우 반환되는 값에서 해당 필드만을 반환합니다.
func SortAsValues(rawObjs interface{}, fieldValues interface{}, defaultValue interface{}, fields ...string) (sorted []interface{}, err error) {
	if len(fields) == 0 || len(fields) > 2 {
		err = fmt.Errorf("One or two field is required. current %v.", fields)
		return
	}

	objTypes := reflect.TypeOf(rawObjs)
	valueTypes := reflect.TypeOf(fieldValues)

	if objTypes.Kind() == reflect.Ptr {
		objTypes = reflect.Indirect(reflect.ValueOf(rawObjs)).Type()
	}

	if objTypes.Kind() != reflect.Slice && valueTypes.Kind() != reflect.Slice {
		err = fmt.Errorf("rawObjs(%v) and fieldValues(%v) is not a slice.", objTypes.Kind(), valueTypes.Kind())
		return
	}

	objTypeKind := objTypes.Elem().Kind()
	if objTypeKind != reflect.Struct && objTypeKind != reflect.Map {
		err = fmt.Errorf("rawObj (%v) is not a struct or map.", objTypeKind)
		return
	}

	objs := reflect.ValueOf(rawObjs)
	values := reflect.ValueOf(fieldValues)
	objKey := fields[0]
	objFieldName := ""

	if objs.Kind() == reflect.Ptr {
		objs = reflect.Indirect(objs)
	}

	if len(fields) == 2 {
		objFieldName = fields[1]
	}

	for valueIndex := 0; valueIndex < values.Len(); valueIndex++ {
		value := values.Index(valueIndex)

		// nil 이거나 올바르지 않은 값인 경우 nil 값을 설정한다.
		if !value.IsValid() || !value.CanInterface() {
			sorted = append(sorted, defaultValue)
			continue
		}

		var val interface{}
		for objIndex := 0; objIndex < objs.Len(); objIndex++ {
			obj := objs.Index(objIndex)
			var objKeyField reflect.Value

			if obj.Kind() == reflect.Map {
				objKeyField = obj.MapIndex(reflect.ValueOf(objKey))
			} else if obj.Kind() == reflect.Struct {
				objKeyField = obj.FieldByName(Classify(objKey))
			}

			// 적절한 값을 찾은 경우 삽입한다.
			if objKeyField.IsValid() && objKeyField.CanInterface() {
				c1 := fmt.Sprintf("%v", objKeyField.Interface())
				c2 := fmt.Sprintf("%v", value.Interface())

				if c1 == c2 {
					// objFieldName 값이 있는 경우 해당 값만을 추출하여 반환한다.
					if objFieldName != "" {
						val = Get(obj.Interface(), objFieldName)
					} else {
						val = obj.Interface()
					}

					break
				}
			}
		}

		if val != nil {
			sorted = append(sorted, val)
		} else {
			sorted = append(sorted, defaultValue)
		}
	}

	return
}
