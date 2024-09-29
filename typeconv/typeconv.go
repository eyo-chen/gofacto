package typeconv

import (
	"reflect"
)

// ToAnysWithOW generates the given number of values to a slice of pointers of given type with the given one overwrite.
func ToAnysWithOW[T any](i int, ow *T) []interface{} {
	res := make([]interface{}, i)
	for k := 0; k < i; k++ {
		var v T
		if ow != nil {
			copyValues(&v, *ow)
			res[k] = &v
			continue
		}

		res[k] = &v
	}

	return res
}

// ToAnysWithOWs generates the given number of values to a slice of pointers to the given type with the given multiple overwrites.
func ToAnysWithOWs[T any](i int, ows ...*T) []interface{} {
	res := make([]interface{}, i)
	for k := 0; k < i; k++ {
		var v T

		if ows != nil && k < len(ows) {
			copyValues(&v, *ows[k])
			res[k] = &v
			continue
		}

		res[k] = &v
	}

	return res
}

// ToAnys converts the given slice of any type to a slice of values of the given type. Note that the given slice must be a slice of pointers.
func ToT[T any](vals []interface{}) []T {
	res := make([]T, len(vals))
	for k, v := range vals {
		res[k] = *v.(*T)
	}

	return res
}

// ToPointerT converts the given slice of any type to a slice of pointers to the given type. Note that the given slice must be a slice of pointers.
func ToPointerT[T any](vals []interface{}) []*T {
	res := make([]*T, len(vals))
	for k, v := range vals {
		res[k] = v.(*T)
	}

	return res
}

func copyValues[T any](dest *T, src T) {
	destValue := reflect.ValueOf(dest).Elem()
	srcValue := reflect.ValueOf(src)

	for i := 0; i < destValue.NumField(); i++ {
		destField := destValue.Field(i)
		srcField := srcValue.FieldByName(destValue.Type().Field(i).Name)

		if srcField.IsValid() && destField.Type() == srcField.Type() && !srcField.IsZero() {
			destField.Set(srcField)
		}
	}
}
