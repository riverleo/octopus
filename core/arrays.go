package core

import (
	"reflect"
)

// 배열 안에 중복된 내용이나 nil값을 제거합니다.
func Compact(targetArr interface{}) (arr []interface{}) {
	vs := reflect.ValueOf(targetArr)

	if !vs.IsValid() {
		arr = []interface{}{}
	}

	if vs.Kind() == reflect.Ptr {
		vs = reflect.Indirect(vs)
	}

	if vs.Kind() == reflect.Slice {
		for i := 0; i < vs.Len(); i++ {
			v := vs.Index(i)

			if !v.IsValid() {
				continue
			}

			isContains := false
			n := v.Interface()

			for _, a := range arr {
				if a == n {
					isContains = true
					break
				}
			}

			if !isContains {
				arr = append(arr, n)
			}
		}
	}

	return
}
