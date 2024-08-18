package typeconv

// CvtToAnysWithOW converts the given number of values to a slice of pointers of given type with the given one overwrite.
func CvtToAnysWithOW[T any](i int, ow *T) []interface{} {
	res := make([]interface{}, i)
	for k := 0; k < i; k++ {
		if ow != nil {
			res[k] = ow
			continue
		}

		var v T
		res[k] = &v
	}

	return res
}

// CvtToAnysWithOWs converts the given number of values to a slice of pointers to the given type with the given multiple overwrites.
func CvtToAnysWithOWs[T any](i int, ows ...*T) []interface{} {
	res := make([]interface{}, i)
	for k := 0; k < i; k++ {
		var v T
		res[k] = &v

		if ows != nil && k < len(ows) {
			res[k] = ows[k]
		}
	}

	return res
}

// CvtToAnys converts the given slice of any type to a slice of values of the given type. Note that the given slice must be a slice of pointers.
func CvtToT[T any](vals []interface{}) []T {
	res := make([]T, len(vals))
	for k, v := range vals {
		res[k] = *v.(*T)
	}

	return res
}

// CvtToPointerT converts the given slice of any type to a slice of pointers to the given type. Note that the given slice must be a slice of pointers.
func CvtToPointerT[T any](vals []interface{}) []*T {
	res := make([]*T, len(vals))
	for k, v := range vals {
		res[k] = v.(*T)
	}

	return res
}
