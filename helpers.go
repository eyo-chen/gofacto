package gofacto

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/eyo-chen/gofacto/internal/db"
	"github.com/eyo-chen/gofacto/internal/utils"
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

		// handle custom types
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

// setAssValue sets the value to the associations value
func (f *Factory[T]) setAssValue(v interface{}, ignoreFields []string) error {
	typeOfV := reflect.TypeOf(v)

	// check if it's a pointer
	if typeOfV.Kind() != reflect.Ptr {
		name := typeOfV.Name()
		return fmt.Errorf("%s, %v: %w", name, v, errIsNotPtr)
	}

	name := typeOfV.Elem().Name()
	// check if it's a pointer to a struct
	if typeOfV.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("%s, %v: %w", name, v, errIsNotStructPtr)
	}

	// check if it's existed in tagToInfo
	if _, ok := f.tagToInfo[name]; !ok {
		return fmt.Errorf("type %s, value %v: %w", name, v, errNotFoundAtTag)
	}

	f.setNonZeroValues(v, ignoreFields)
	return nil
}

// genAndInsertAss inserts the associations value into the database
func (f *Factory[T]) insertAss(ctx context.Context) error {
	for name, vals := range f.associations {
		tableName := f.tagToInfo[name].tableName
		if _, err := f.db.InsertList(ctx, db.InsertListParams{StorageName: tableName, Values: vals}); err != nil {
			return err
		}
	}

	return nil
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

// setForeignKey sets the value of the source's ID field to the target's foreign key(name) field
func setForeignKey(target interface{}, name string, source interface{}) error {
	targetField := reflect.ValueOf(target).Elem().FieldByName(name)
	if !targetField.IsValid() {
		return fmt.Errorf("%s: %w", name, errFieldNotFound)
	}

	if !targetField.CanSet() {
		return fmt.Errorf("%s: %w", name, errFieldCantSet)
	}

	sourceIDField := reflect.ValueOf(source).Elem().FieldByName("ID")
	if !sourceIDField.IsValid() {
		return fmt.Errorf("%s: %w", "ID", errFieldNotFound)
	}

	sourceIDKind := sourceIDField.Kind()
	if !isIntType(sourceIDKind) && !isUintType(sourceIDKind) {
		return errNotInt
	}

	setIntValue(targetField, sourceIDField)
	return nil
}

// extractTag generates the map from tag to metadata
func extractTag(dataType reflect.Type) (map[string]tagInfo, []string, error) {
	numField := dataType.NumField()
	var ignoreFields []string
	tagToInfo := make(map[string]tagInfo)

	for i := 0; i < numField; i++ {
		field := dataType.Field(i)
		tag := field.Tag.Get(packageName)
		if tag == "" {
			continue
		}

		parts := strings.Split(tag, ";")
		if len(parts) == 0 {
			return nil, nil, errTagFormat
		}

		var structName, tableName, foreignField string
		for _, part := range parts {
			if part == "omit" {
				ignoreFields = append(ignoreFields, field.Name)
				continue
			}

			subParts := strings.Split(part, ",")
			if subParts[0] != "foreignKey" {
				return nil, nil, errTagFormat
			}

			for _, subPart := range subParts[1:] {
				kv := strings.SplitN(subPart, ":", 2)
				if len(kv) != 2 {
					return nil, nil, errTagFormat
				}

				switch kv[0] {
				case "struct":
					structName = kv[1]
				case "table":
					tableName = kv[1]
				case "field":
					foreignField = kv[1]
				default:
					return nil, nil, errTagFormat
				}
			}
		}

		if tableName == "" {
			tableName = utils.CamelToSnake(structName) + "s"
		}
		tagToInfo[structName] = tagInfo{tableName: tableName, fieldName: field.Name, foreignField: foreignField}
	}

	return tagToInfo, ignoreFields, nil
}

// setIntValue sets the value of the source to the target,
// and it also handles the conversion between int and uint.
// Normally, it's used to set the ID field of the target struct
func setIntValue(target, source reflect.Value) {
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

// isIntType checks if the kind is an integer type
func isIntType(k reflect.Kind) bool {
	return k >= reflect.Int && k <= reflect.Int64
}

// isUintType checks if the kind is an unsigned integer type
func isUintType(k reflect.Kind) bool {
	return k >= reflect.Uint && k <= reflect.Uint64
}

// serField sets the value of the source to the field of the target.
// field value of target might be a pointer or value.
// field and source must be a pointer.
func setField(target interface{}, fieldName string, source interface{}) error {
	structValue := reflect.ValueOf(target).Elem()
	fieldVal := structValue.FieldByName(fieldName)

	if !fieldVal.IsValid() {
		return fmt.Errorf("%s: %w", fieldName, errFieldNotFound)
	}

	if !fieldVal.CanSet() {
		return fmt.Errorf("%s: %w", fieldName, errFieldCantSet)
	}

	sourceVal := reflect.ValueOf(source).Elem()

	// when fieldVal is a pointer
	if fieldVal.Kind() == reflect.Ptr {
		// If fieldVal is nil, create a new instance
		if fieldVal.IsNil() {
			fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
		}

		// check if the type of the pointer is the same as the source
		if fieldVal.Type().Elem() != sourceVal.Type() {
			return fmt.Errorf("type mismatch: field %s is %v, source is %v", fieldName, fieldVal.Type().Elem(), sourceVal.Type())
		}

		// set the value of the pointer to the source
		fieldVal.Elem().Set(sourceVal)
		return nil
	}

	// when fieldVal is a value
	if fieldVal.Type() != sourceVal.Type() {
		return fmt.Errorf("type mismatch: field %s is %v, source is %v", fieldName, fieldVal.Type(), sourceVal.Type())
	}
	fieldVal.Set(sourceVal)

	return nil
}
