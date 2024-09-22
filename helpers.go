package gofacto

import (
	"fmt"
	"reflect"
	"slices"
	"time"
)

const (
	packageName = "gofacto"
)

// setNonZeroValues sets non-zero values to the given struct.
// Parameter v must be a pointer to a struct
func (f *Factory[T]) setNonZeroValues(v interface{}, ignoreFields []string) {
	val := reflect.ValueOf(v).Elem()
	typeOfVal := val.Type()

	for k := 0; k < val.NumField(); k++ {
		curVal := val.Field(k)
		curField := typeOfVal.Field(k)

		// skip ignored fields
		if slices.Contains(ignoreFields, curField.Name) {
			continue
		}

		// skip non-zero fields, unexported fields, and ID field
		if !curVal.IsZero() || !curVal.CanSet() || curField.Name == "ID" || curField.PkgPath != "" {
			continue
		}

		// handle db custom types
		if f.db != nil {
			if customValue, ok := f.db.GenCustomType(curField.Type); ok {
				curVal.Set(reflect.ValueOf(customValue))
				continue
			}
		}

		// handle time.Time
		if curField.Type == reflect.TypeOf(time.Time{}) {
			curVal.Set(reflect.ValueOf(time.Now()))
			continue
		}

		// handle *time.Time
		if curField.Type.Kind() == reflect.Ptr && curField.Type.Elem() == reflect.TypeOf(time.Time{}) {
			timeVal := time.Now()
			curVal.Set(reflect.ValueOf(&timeVal))
			continue
		}

		// handle struct
		if curField.Type.Kind() == reflect.Struct {
			f.setNonZeroValues(curVal.Addr().Interface(), ignoreFields)
			continue
		}

		// handle pointer to struct
		if curField.Type.Kind() == reflect.Ptr && curField.Type.Elem().Kind() == reflect.Struct {
			newInstance := reflect.New(curField.Type.Elem()).Elem()
			f.setNonZeroValues(newInstance.Addr().Interface(), ignoreFields)
			curVal.Set(newInstance.Addr())
			continue
		}

		// handle slice
		if curField.Type.Kind() == reflect.Slice {
			f.setNonZeroSlice(curVal.Addr().Interface(), ignoreFields)
			continue
		}

		// handle pointer to slice
		if curField.Type.Kind() == reflect.Ptr && curField.Type.Elem().Kind() == reflect.Slice {
			newInstance := reflect.New(curField.Type.Elem()).Elem()
			f.setNonZeroSlice(newInstance.Addr().Interface(), ignoreFields)
			curVal.Set(newInstance.Addr())
			continue
		}

		// skip client-defined types
		if curField.Type.PkgPath() != "" {
			continue
		}

		// skip pointer to custom type
		if curField.Type.Kind() == reflect.Ptr &&
			curField.Type.Elem().PkgPath() != "" {
			continue
		}

		// For other types, set non-zero values if the field is zero
		if v := genNonZeroValue(curField.Type, f.index); v != nil {
			curVal.Set(reflect.ValueOf(v))
		}
	}
}

// setNonZeroSlice sets non-zero values to the given slice.
// Parameter v must be a pointer to a slice
func (f *Factory[T]) setNonZeroSlice(v interface{}, ignoreFields []string) {
	val := reflect.ValueOf(v).Elem()

	// handle slice
	if val.Type().Elem().Kind() == reflect.Slice {
		e := reflect.New(val.Type().Elem()).Elem()
		f.setNonZeroSlice(e.Addr().Interface(), ignoreFields)
		val.Set(reflect.Append(val, e))
		return
	}

	// handle slice of pointers
	if val.Type().Elem().Kind() == reflect.Ptr && val.Type().Elem().Elem().Kind() == reflect.Slice {
		e := reflect.New(val.Type().Elem().Elem()).Elem()
		f.setNonZeroSlice(e.Addr().Interface(), ignoreFields)
		val.Set(reflect.Append(val, e.Addr()))
		return
	}

	// handle struct
	if val.Type().Elem().Kind() == reflect.Struct {
		e := reflect.New(val.Type().Elem()).Elem()
		f.setNonZeroValues(e.Addr().Interface(), ignoreFields)
		val.Set(reflect.Append(val, e))
		return
	}

	// handle pointer to struct
	if val.Type().Elem().Kind() == reflect.Ptr && val.Type().Elem().Elem().Kind() == reflect.Struct {
		e := reflect.New(val.Type().Elem().Elem())
		f.setNonZeroValues(e.Interface(), ignoreFields)
		val.Set(reflect.Append(val, e))
		return
	}

	// handle other types
	t := val.Type().Elem()
	if tv := genNonZeroValue(t, f.index); tv != nil {
		val.Set(reflect.Append(val, reflect.ValueOf(tv)))
	}
}

// copyValues copys non-zero values from src to dest
func copyValues[T any](dest *T, src T) error {
	destValue := reflect.ValueOf(dest).Elem()
	srcValue := reflect.ValueOf(src)

	if destValue.Kind() != reflect.Struct {
		return errDestIsNotStruct
	}

	if srcValue.Kind() != reflect.Struct {
		return errSrcIsNotStruct
	}

	if destValue.Type() != srcValue.Type() {
		return fmt.Errorf("%w: %s and %s", errTypeDiff, destValue.Type(), srcValue.Type())
	}

	for i := 0; i < destValue.NumField(); i++ {
		destField := destValue.Field(i)
		srcField := srcValue.FieldByName(destValue.Type().Field(i).Name)

		if srcField.IsValid() && destField.Type() == srcField.Type() && !srcField.IsZero() {
			destField.Set(srcField)
		}
	}

	return nil
}

// genNonZeroValue generates a non-zero value for the given type
func genNonZeroValue(t reflect.Type, i int) interface{} {
	switch t.Kind() {
	case reflect.Int:
		return int(i)
	case reflect.Int8:
		return int8(i)
	case reflect.Int16:
		return int16(i)
	case reflect.Int32:
		return int32(i)
	case reflect.Int64:
		return int64(i)
	case reflect.Uint:
		return uint(i)
	case reflect.Uint8:
		return uint8(i)
	case reflect.Uint16:
		return uint16(i)
	case reflect.Uint32:
		return uint32(i)
	case reflect.Uint64:
		return uint64(i)
	case reflect.Float32:
		return float32(i)
	case reflect.Float64:
		return float64(i)
	case reflect.Bool:
		return true
	case reflect.String:
		return fmt.Sprintf("%s%d", "test", i)
	case reflect.Pointer:
		v := genNonZeroValue(t.Elem(), i)
		ptr := reflect.New(t.Elem())
		ptr.Elem().Set(reflect.ValueOf(v))
		return ptr.Interface()
	// TODO: If it's reflect.Chan, reflect.Func, reflect.Slice, reflect.Map, reflect.Array, reflect.Interface, currently it will return nil
	default:
		return nil
	}
}
