package transformer

import "reflect"

func ConvertSlice[T any](s []T) []any {
	val := reflect.ValueOf(s)
	ret := make([]any, val.Len())

	for i := range val.Len() {
		ret[i] = val.Index(i).Interface()
	}

	return ret
}
