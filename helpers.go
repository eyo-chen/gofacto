package gofacto

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/eyo-chen/gofacto/db"
)

const (
	packageName = "gofacto"
)

// copyValues copys non-zero values from src to dest
func copyValues[T any](dest *T, src T) error {
	destValue := reflect.ValueOf(dest).Elem()
	srcValue := reflect.ValueOf(src)

	if destValue.Kind() != reflect.Struct {
		return errors.New("destination value is not a struct")
	}

	if srcValue.Kind() != reflect.Struct {
		return errors.New("source value is not a struct")
	}

	if destValue.Type() != srcValue.Type() {
		return errors.New("destination and source type is different")
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

// genFinalError generates a final error message from the given errors
func genFinalError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	errorMessages := make([]string, len(errs))
	for i, err := range errs {
		errorMessages[i] = err.Error()
	}

	return fmt.Errorf(strings.Join(errorMessages, "\n"))
}

// setNonZeroValues sets non-zero values to the given struct
// v must be a pointer to a struct
func setNonZeroValues(i int, v interface{}) {
	val := reflect.ValueOf(v).Elem()
	typeOfVal := val.Type()

	for k := 0; k < val.NumField(); k++ {
		curVal := val.Field(k)
		curField := typeOfVal.Field(k)

		// skip non-zero fields, unexported fields, and ID field
		if !curVal.IsZero() || !curVal.CanSet() || curField.Name == "ID" || curField.PkgPath != "" {
			continue
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

		// If the field is a struct, recursively set non-zero values for its fields
		if curField.Type.Kind() == reflect.Struct {
			setNonZeroValues(i, curVal.Addr().Interface())
			continue
		}

		// If the field is a pointer, create a new instance of the pointed-to struct type and set its non-zero values
		if curField.Type.Kind() == reflect.Ptr && curField.Type.Elem().Kind() == reflect.Struct {
			if curVal.IsNil() {
				newInstance := reflect.New(curField.Type.Elem()).Elem()
				setNonZeroValues(i, newInstance.Addr().Interface())
				curVal.Set(newInstance.Addr())
			} else {
				setNonZeroValues(i, curVal.Interface())
			}
			continue
		}

		// If the field is a slice
		if curField.Type.Kind() == reflect.Slice {
			setNonZeroValuesForSlice(i, curVal.Addr().Interface())
			continue
		}

		if curField.Type.Kind() == reflect.Ptr && curField.Type.Elem().Kind() == reflect.Slice {
			if curVal.IsNil() {
				newInstance := reflect.New(curField.Type.Elem()).Elem()
				setNonZeroValuesForSlice(i, newInstance.Addr().Interface())
				curVal.Set(newInstance.Addr())
			} else {
				setNonZeroValuesForSlice(i, curVal.Interface())
			}
			continue
		}

		// For other types, set non-zero values if the field is zero
		v := genNonZeroValue(curField.Type, i)
		curVal.Set(reflect.ValueOf(v))
	}
}

// setNonZeroValuesForSlice sets non-zero values to the given slice
// v must be a pointer to a slice
func setNonZeroValuesForSlice(i int, v interface{}) {
	val := reflect.ValueOf(v).Elem()

	// handle slice
	if val.Type().Elem().Kind() == reflect.Slice {
		e := reflect.New(val.Type().Elem()).Elem()
		setNonZeroValuesForSlice(i, e.Addr().Interface())
		val.Set(reflect.Append(val, e))
		return
	}

	// handle slice of pointers
	if val.Type().Elem().Kind() == reflect.Ptr && val.Type().Elem().Elem().Kind() == reflect.Slice {
		e := reflect.New(val.Type().Elem().Elem()).Elem()
		setNonZeroValuesForSlice(i, e.Addr().Interface())
		val.Set(reflect.Append(val, e.Addr()))
		return
	}

	// handle struct
	if val.Type().Elem().Kind() == reflect.Struct {
		e := reflect.New(val.Type().Elem()).Elem()
		setNonZeroValues(i, e.Addr().Interface())
		val.Set(reflect.Append(val, e))
		return
	}

	// handle pointer to struct
	if val.Type().Elem().Kind() == reflect.Ptr && val.Type().Elem().Elem().Kind() == reflect.Struct {
		e := reflect.New(val.Type().Elem().Elem())
		setNonZeroValues(i, e.Interface())
		val.Set(reflect.Append(val, e))
		return
	}

	// handle other types
	t := val.Type().Elem()
	tv := genNonZeroValue(t, i)
	val.Set(reflect.Append(val, reflect.ValueOf(tv)))
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

// setField sets the value to the name field of the target
func setField(target interface{}, name string, source interface{}, sourceFn string) error {
	targetField := reflect.ValueOf(target).Elem().FieldByName(name)
	if !targetField.IsValid() {
		return fmt.Errorf("%s: field %s is not found", sourceFn, name)
	}

	if !targetField.CanSet() {
		return fmt.Errorf("%s: field %s can not be set", sourceFn, name)
	}

	sourceIDField := reflect.ValueOf(source).Elem().FieldByName("ID")
	if !sourceIDField.IsValid() {
		return fmt.Errorf("%s: source field ID is not found", sourceFn)
	}

	sourceIDKind := sourceIDField.Kind()
	if sourceIDKind != reflect.Int &&
		sourceIDKind != reflect.Int64 &&
		sourceIDKind != reflect.Int32 &&
		sourceIDKind != reflect.Int16 &&
		sourceIDKind != reflect.Int8 &&
		sourceIDKind != reflect.Uint &&
		sourceIDKind != reflect.Uint64 &&
		sourceIDKind != reflect.Uint32 &&
		sourceIDKind != reflect.Uint16 &&
		sourceIDKind != reflect.Uint8 {
		return fmt.Errorf("%s: source field ID is not an integer", sourceFn)
	}

	// TODO: What if targetField is int, but sourceIDField is uint?
	switch sourceIDField.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		targetField.SetInt(sourceIDField.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		targetField.SetUint(sourceIDField.Uint())
	}

	return nil
}

// genAndInsertAss inserts the associations value into the database and returns with the inserted values
func genAndInsertAss(d db.Database, associations map[string][]interface{}, tagToInfo map[string]tagInfo) ([]interface{}, error) {
	if len(tagToInfo) == 0 {
		return nil, errors.New("tagToInfo is not set")
	}

	if len(associations) == 0 {
		return nil, errors.New("inserting associations without any associations")
	}

	result := []interface{}{}
	for name, vals := range associations {
		tableName := tagToInfo[name].tableName

		v, err := d.InsertList(db.InserListParams{StorageName: tableName, Values: vals})
		if err != nil {
			return nil, err
		}

		result = append(result, v...)
	}

	return result, nil
}

// setAssValue sets the value to the associations value
func setAssValue(v interface{}, tagToInfo map[string]tagInfo, index int, sourceFn string) error {
	typeOfV := reflect.TypeOf(v)

	// check if it's a pointer
	if typeOfV.Kind() != reflect.Ptr {
		name := typeOfV.Name()
		return fmt.Errorf("%s: type %s, value %v is not a pointer", sourceFn, name, v)
	}

	name := typeOfV.Elem().Name()
	// check if it's a pointer to a struct
	if typeOfV.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("%s: type %s, value %v is not a pointer to a struct", sourceFn, name, v)
	}

	// check if it's existed in tagToInfo
	if _, ok := tagToInfo[name]; !ok {
		return fmt.Errorf("%s: type %s, value %v is not found at tag", sourceFn, name, v)
	}

	setNonZeroValues(index, v)
	return nil
}

// genTagToInfo generates the map from tag to metadata
func genTagToInfo(dataType reflect.Type) (map[string]tagInfo, error) {
	tagToInfo := map[string]tagInfo{}
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		tag := field.Tag.Get(packageName)
		if tag == "" {
			continue
		}

		parts := strings.Split(tag, ",")
		if len(parts) == 0 {
			return nil, errors.New("tag is in wrong format. It should be gofacto:\"<struct_name>,<table_name>\"")
		}

		structName := parts[0]

		var tableName string
		if len(parts) == 2 {
			tableName = parts[1]
		} else {
			tableName = camelToSnake(structName) + "s"
		}

		tagToInfo[structName] = tagInfo{tableName: tableName, fieldName: field.Name}
	}

	return tagToInfo, nil
}

// camelToSnake converts a camel case string to a snake case string
func camelToSnake(input string) string {
	var buf bytes.Buffer

	for i, r := range input {
		if unicode.IsUpper(r) {
			if i > 0 && unicode.IsLower(rune(input[i-1])) {
				buf.WriteRune('_')
			}
			buf.WriteRune(unicode.ToLower(r))
		} else {
			buf.WriteRune(r)
		}
	}

	return buf.String()
}
