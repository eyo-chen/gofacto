package gofacto

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
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
func setNonZeroValues(i int, v interface{}, ignoreFields []string) {
	val := reflect.ValueOf(v).Elem()
	typeOfVal := val.Type()

	for k := 0; k < val.NumField(); k++ {
		curVal := val.Field(k)
		curField := typeOfVal.Field(k)

		// skip ignored fields
		if len(ignoreFields) > 0 && slices.Contains(ignoreFields, curField.Name) {
			continue
		}

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

		// handle struct
		if curField.Type.Kind() == reflect.Struct {
			setNonZeroValues(i, curVal.Addr().Interface(), ignoreFields)
			continue
		}

		// handle pointer to struct
		if curField.Type.Kind() == reflect.Ptr && curField.Type.Elem().Kind() == reflect.Struct {
			newInstance := reflect.New(curField.Type.Elem()).Elem()
			setNonZeroValues(i, newInstance.Addr().Interface(), ignoreFields)
			curVal.Set(newInstance.Addr())
			continue
		}

		// handle slice
		if curField.Type.Kind() == reflect.Slice {
			setNonZeroValuesForSlice(i, curVal.Addr().Interface(), ignoreFields)
			continue
		}

		// handle pointer to slice
		if curField.Type.Kind() == reflect.Ptr && curField.Type.Elem().Kind() == reflect.Slice {
			newInstance := reflect.New(curField.Type.Elem()).Elem()
			setNonZeroValuesForSlice(i, newInstance.Addr().Interface(), ignoreFields)
			curVal.Set(newInstance.Addr())
			continue
		}

		// For other types, set non-zero values if the field is zero
		if v := genNonZeroValue(curField.Type, i); v != nil {
			curVal.Set(reflect.ValueOf(v))
		}
	}
}

// setNonZeroValuesForSlice sets non-zero values to the given slice.
// Parameter v must be a pointer to a slice
func setNonZeroValuesForSlice(i int, v interface{}, ignoreFields []string) {
	val := reflect.ValueOf(v).Elem()

	// handle slice
	if val.Type().Elem().Kind() == reflect.Slice {
		e := reflect.New(val.Type().Elem()).Elem()
		setNonZeroValuesForSlice(i, e.Addr().Interface(), ignoreFields)
		val.Set(reflect.Append(val, e))
		return
	}

	// handle slice of pointers
	if val.Type().Elem().Kind() == reflect.Ptr && val.Type().Elem().Elem().Kind() == reflect.Slice {
		e := reflect.New(val.Type().Elem().Elem()).Elem()
		setNonZeroValuesForSlice(i, e.Addr().Interface(), ignoreFields)
		val.Set(reflect.Append(val, e.Addr()))
		return
	}

	// handle struct
	if val.Type().Elem().Kind() == reflect.Struct {
		e := reflect.New(val.Type().Elem()).Elem()
		setNonZeroValues(i, e.Addr().Interface(), ignoreFields)
		val.Set(reflect.Append(val, e))
		return
	}

	// handle pointer to struct
	if val.Type().Elem().Kind() == reflect.Ptr && val.Type().Elem().Elem().Kind() == reflect.Struct {
		e := reflect.New(val.Type().Elem().Elem())
		setNonZeroValues(i, e.Interface(), ignoreFields)
		val.Set(reflect.Append(val, e))
		return
	}

	// handle other types
	t := val.Type().Elem()
	if tv := genNonZeroValue(t, i); tv != nil {
		val.Set(reflect.Append(val, reflect.ValueOf(tv)))
	}
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
	if !isIntType(sourceIDKind) && !isUintType(sourceIDKind) {
		return fmt.Errorf("%s: source field ID is not an integer", sourceFn)
	}

	setFieldValue(targetField, sourceIDField)
	return nil
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

	setNonZeroValues(index, v, nil)
	return nil
}

// genAndInsertAss inserts the associations value into the database
func insertAss(ctx context.Context, d db.Database, associations map[string][]interface{}, tagToInfo map[string]tagInfo) error {
	if len(tagToInfo) == 0 {
		return errors.New("tagToInfo is not set")
	}

	if len(associations) == 0 {
		return errors.New("inserting associations without any associations")
	}

	for name, vals := range associations {
		tableName := tagToInfo[name].tableName
		if _, err := d.InsertList(ctx, db.InserListParams{StorageName: tableName, Values: vals}); err != nil {
			return err
		}
	}

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

// setFieldValue sets the value of the source to the target,
// and it also handles the conversion between int and uint.
// Normally, it's used to set the ID field of the target struct
func setFieldValue(target, source reflect.Value) {
	targetKind := target.Kind()
	sourceKind := source.Kind()

	if isIntType(targetKind) && isIntType(sourceKind) {
		target.SetInt(source.Int())
		return
	}

	if isUintType(targetKind) && isUintType(sourceKind) {
		target.SetUint(source.Uint())
		return
	}

	if isIntType(targetKind) {
		target.SetInt(int64(source.Uint()))
		return
	}

	target.SetUint(uint64(source.Int()))
}

func isIntType(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Int64
}

func isUintType(k reflect.Kind) bool {
	return k >= reflect.Uint && k <= reflect.Uint64
}
