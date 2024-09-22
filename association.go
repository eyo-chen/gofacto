package gofacto

import (
	"context"
	"fmt"
	"reflect"

	"github.com/eyo-chen/gofacto/internal/db"
)

// setAssValue sets the value to the associations value
func (f *Factory[T]) setAssValue(v interface{}, ignoreFields []string) error {
	if err := checkAssoc(v); err != nil {
		return err
	}

	// check if it's existed in tagToInfo
	name := reflect.TypeOf(v).Elem().Name()
	if _, ok := f.tagToInfo[name]; !ok {
		return fmt.Errorf("type %s, value %v: %w", name, v, errNotFoundAtTag)
	}

	f.setNonZeroValues(v, ignoreFields)
	return nil
}

// insertAss inserts the associations value into the database
func (f *Factory[T]) insertAss(ctx context.Context) error {
	for name, vals := range f.associations {
		tableName := f.tagToInfo[name].tableName
		if _, err := f.db.InsertList(ctx, db.InsertListParams{StorageName: tableName, Values: vals}); err != nil {
			return err
		}
	}

	return nil
}

// setAss sets and inserts the associations
func (b *builder[T]) setAss() error {
	// insert the associations
	if err := b.f.insertAss(b.ctx); err != nil {
		return err
	}

	// set the connection between the factory value and the associations
	for name, vals := range b.f.associations {
		// use vs[0] because we can make sure setAss(builder) only invoke with Build function
		// which means there's only one factory value
		// so that each associations only allow one value
		t := b.f.tagToInfo[name]
		if err := setForeignKey(b.v, t.fieldName, vals[0]); err != nil {
			return err
		}

		// set the foreign field value if it's provided
		if t.foreignField != "" {
			if err := setField(b.v, t.foreignField, vals[0]); err != nil {
				return err
			}
		}
	}

	// clear associations
	b.f.associations = map[string][]interface{}{}

	return nil
}

// setAss sets and inserts the associations
func (b *builderList[T]) setAss() error {
	// insert the associations
	if err := b.f.insertAss(b.ctx); err != nil {
		return err
	}

	// set the connection between the factory value and the associations
	// use cachePrev because multiple values can have one association value
	// e.g. multiple transaction belongs to one user
	cachePrev := map[string]interface{}{}
	for i, l := range b.list {
		for name, vs := range b.f.associations {
			var v interface{}
			if i >= len(vs) {
				v = cachePrev[name]
			} else {
				v = vs[i]
				cachePrev[name] = vs[i]
			}

			t := b.f.tagToInfo[name]
			if err := setForeignKey(l, t.fieldName, v); err != nil {
				return err
			}

			// set the foreign field value if it's provided
			if t.foreignField != "" {
				if err := setField(l, t.foreignField, v); err != nil {
					return err
				}
			}
		}
	}

	// clear associations
	b.f.associations = map[string][]interface{}{}

	return nil
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

// setIntValue sets the value of the source to the target,
// and it also handles the conversion between int and uint.
// Normally, it's used to set the ID field of the target struct
func setIntValue(target, source reflect.Value) {
	if target.Kind() == reflect.Ptr {
		target = target.Elem()
	}

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

// checkAssoc checks if the input association value is valid
func checkAssoc(v interface{}) error {
	typeOfV := reflect.TypeOf(v)

	// check if it's a pointer
	if typeOfV.Kind() != reflect.Ptr {
		name := typeOfV.Name()
		return fmt.Errorf("%s, %v: %w", name, v, errIsNotPtr)
	}

	// check if it's a pointer to a struct
	if typeOfV.Elem().Kind() != reflect.Struct {
		name := typeOfV.Elem().Name()
		return fmt.Errorf("%s, %v: %w", name, v, errIsNotStructPtr)
	}

	return nil
}
