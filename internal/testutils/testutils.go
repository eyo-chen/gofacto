package testutils

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"time"
)

// CompareVal compares two values. IgnoreFields is a list of fields that will be ignored during comparison.
func CompareVal(val1, val2 interface{}, ignoreFields ...string) error {
	// check nil values
	if val1 == nil && val2 == nil {
		return nil
	}
	if val1 == nil || val2 == nil {
		return errors.New("one value is nil, the other is not")
	}

	v1 := reflect.ValueOf(val1)
	v2 := reflect.ValueOf(val2)

	// check if both values are same type
	if v1.Type() != v2.Type() || v1.Kind() != v2.Kind() {
		return errors.New("both values must be the same type")
	}

	// check if both values are pointer
	if v1.Kind() == reflect.Ptr {
		if v1.IsNil() && v2.IsNil() {
			return nil
		}

		if v1.IsNil() || v2.IsNil() {
			return errors.New("one value is nil, the other is not")
		}

		// dereference pointer, and recursively compare
		if err := CompareVal(v1.Elem().Interface(), v2.Elem().Interface(), ignoreFields...); err != nil {
			return err
		}

		return nil
	}

	// check time values
	if v1.Type() == reflect.TypeOf(time.Time{}) {
		if v1.Interface().(time.Time).Equal(v2.Interface().(time.Time)) {
			return nil
		}

		return fmt.Errorf("time values are different, the first one is %v, the second one is %v", v1.Interface().(time.Time), v2.Interface().(time.Time))
	}

	// check slice values
	if v1.Kind() == reflect.Slice {
		if v1.Len() != v2.Len() {
			return fmt.Errorf("both slices must have the same length, the first one is %d, the second one is %d", v1.Len(), v2.Len())
		}

		for i := 0; i < v1.Len(); i++ {
			if err := CompareVal(v1.Index(i).Interface(), v2.Index(i).Interface(), ignoreFields...); err != nil {
				return fmt.Errorf("index %d of slice has error: %v", i, err)
			}
		}

		return nil
	}

	// check struct values
	if v1.Kind() == reflect.Struct {
		return compareStruct(v1, v2, ignoreFields)
	}

	// check interface
	if v1.Kind() == reflect.Interface {
		if err := CompareVal(v1.Interface(), v2.Interface(), ignoreFields...); err != nil {
			return err
		}

		return nil
	}

	// check non-struct values(like int, string, slice, etc.)
	if !reflect.DeepEqual(val1, val2) {
		fmt.Println("error", val1, val2)
		return fmt.Errorf("values are different, the first one is %v, the second one is %v", val1, val2)
	}

	return nil
}

func compareStruct(v1 reflect.Value, v2 reflect.Value, ignoreFields []string) error {
	// Iterate through each field in the struct
	for i := 0; i < v1.NumField(); i++ {
		field1 := v1.Field(i)
		field2 := v2.Field(i)
		fieldName := v1.Type().Field(i).Name

		// ignore target fields
		if len(ignoreFields) > 0 && slices.Contains(ignoreFields, fieldName) {
			continue
		}

		// ignore unexported fields
		if v1.Type().Field(i).PkgPath != "" || v2.Type().Field(i).PkgPath != "" {
			continue
		}

		// check if both fields are same type
		if field1.Type() != field2.Type() || field1.Kind() != field2.Kind() {
			return fmt.Errorf("field %s must be the same type", fieldName)
		}

		if err := CompareVal(field1.Interface(), field2.Interface(), ignoreFields...); err != nil {
			return fmt.Errorf("field %s has error: %v", fieldName, err)
		}
	}

	return nil
}

// IsNotZeroVal checks if the value is not zero value. IgnoreFields is a list of fields that will be ignored during comparison.
func IsNotZeroVal(val interface{}, ignoreFields ...string) error {
	v := reflect.ValueOf(val)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}

		return IsNotZeroVal(v.Elem().Interface(), ignoreFields...)
	}

	// check slice values
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			if err := IsNotZeroVal(v.Index(i).Interface(), ignoreFields...); err != nil {
				return err
			}
		}

		return nil
	}

	if v.Kind() != reflect.Struct {
		return errors.New("value must be a struct or a pointer to a struct")
	}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fieldName := v.Type().Field(i).Name

		if slices.Contains(ignoreFields, fieldName) {
			continue
		}

		if err := checkNonZero(f, fieldName); err != nil {
			return err
		}
	}

	return nil
}

func checkNonZero(v reflect.Value, fieldName string) error {
	if !v.IsValid() {
		return errors.New("field not found")
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fmt.Errorf("field %s is nil", fieldName)
		}

		return nil
	}

	if v.IsZero() {
		return fmt.Errorf("field %s is zero value", fieldName)
	}

	return nil
}

// IsZeroVal checks if the value is zero value. IgnoreFields is a list of fields that will be ignored during comparison.
func IsZeroVal(val interface{}, ignoreFields ...string) error {
	v := reflect.ValueOf(val)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}

		return IsZeroVal(v.Elem().Interface(), ignoreFields...)
	}

	// check slice values
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			if err := IsZeroVal(v.Index(i).Interface(), ignoreFields...); err != nil {
				return err
			}
		}

		return nil
	}

	if v.Kind() != reflect.Struct {
		return errors.New("value must be a struct or a pointer to a struct")
	}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		fieldName := v.Type().Field(i).Name

		if slices.Contains(ignoreFields, fieldName) {
			continue
		}

		if err := checkZero(f, fieldName); err != nil {
			return err
		}
	}

	return nil
}

func checkZero(v reflect.Value, fieldName string) error {
	if !v.IsValid() {
		return errors.New("field not found")
	}

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}

		return fmt.Errorf("field %s is not nil", fieldName)
	}

	if v.IsZero() {
		return nil
	}

	return fmt.Errorf("field %s is not zero value", fieldName)
}

// GetFunName returns the name of the function.
func GetFunName(fn interface{}) string {
	if fn == nil {
		return ""
	}

	fullName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}

// FilterFields returns a list of field names that are not in the fields list.
func FilterFields(val interface{}, fields ...string) []string {
	v := reflect.ValueOf(val)

	// convert fields to hash map
	fieldMap := make(map[string]bool)
	for _, f := range fields {
		fieldMap[f] = true
	}

	var fieldNames []string
	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Name
		if fieldMap[fieldName] {
			continue
		}

		fieldNames = append(fieldNames, fieldName)
	}

	return fieldNames
}
