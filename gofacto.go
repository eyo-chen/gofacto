package gofacto

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/eyo-chen/gofacto/db"
)

// Config is the configuration for the factory
type Config[T any] struct {
	// BluePrint is a client-defined function to create a new value.
	// If not provided, non-zero default values is set.
	//
	// BluePrint must follow the signature:
	// type bluePrintFunc[T any] func(i int, last T) T
	BluePrint bluePrintFunc[T]

	// DB is the interface for the Database.
	// It must be provided if want to do database operations.
	//
	// use sqlf for raw sql
	//
	// use gormf for gorm
	//
	// use mongof for mongo
	DB db.Database

	// StorageName is the name specified where the value is stored.
	// It will be table name for sql, and collection name for mongodb.
	// It is optional.
	// If not provided, camel case of the type name will be used.
	StorageName string

	// IgnoreFields is the list of fields to be ignored
	// these fields will not be set to non-zero values
	IgnoreFields []string

	// isSetZeroValue is to determine if the zero value should be set.
	// It is optional.
	// If not provided, it will be default to true.
	IsSetZeroValue *bool
}

// Factory is the gofacto factory to create mock data
type Factory[T any] struct {
	db             db.Database
	bluePrint      bluePrintFunc[T]
	storageName    string
	dataType       reflect.Type
	empty          T
	index          int
	ignoreFields   []string
	isSetZeroValue bool

	// map from name to trait function
	traits map[string]setTraiter[T]

	// map from name to list of associations
	// e.g. "User" -> []*User
	associations map[string][]interface{}

	// map from tag to metadata
	// e.g. "User" -> {tableName: "users", fieldName: "UserID"}
	tagToInfo map[string]tagInfo
}

// bluePrintFunc is a client-defined function to create a new value
type bluePrintFunc[T any] func(i int, last T) T

// setTraiter is a client-defined function to add a trait to mutate the value
type setTraiter[T any] func(v *T)

// tagInfo is the metadata for the tag
type tagInfo struct {
	tableName string
	fieldName string
}

// builder is for building a single value
type builder[T any] struct {
	ctx    context.Context
	v      *T
	errors []error
	f      *Factory[T]
}

// builderList is for building a list of values
type builderList[T any] struct {
	ctx    context.Context
	list   []*T
	errors []error
	f      *Factory[T]
}

// New creates a new gofacto factory
func New[T any](v T) *Factory[T] {
	dataType := reflect.TypeOf(v)

	return &Factory[T]{
		dataType:       dataType,
		empty:          reflect.New(dataType).Elem().Interface().(T),
		associations:   map[string][]interface{}{},
		tagToInfo:      map[string]tagInfo{},
		index:          1,
		isSetZeroValue: true,
	}
}

// SetConfig sets the configuration for the factory
func (f *Factory[T]) SetConfig(c Config[T]) *Factory[T] {
	f.bluePrint = c.BluePrint
	f.db = c.DB
	f.ignoreFields = c.IgnoreFields

	if c.StorageName == "" {
		f.storageName = fmt.Sprintf("%ss", camelToSnake(f.dataType.Name()))
	} else {
		f.storageName = c.StorageName
	}

	if c.IsSetZeroValue != nil {
		f.isSetZeroValue = *c.IsSetZeroValue
	}

	return f
}

// SetTrait adds a trait to the factory
func (f *Factory[T]) SetTrait(name string, tr setTraiter[T]) *Factory[T] {
	if f.traits == nil {
		f.traits = map[string]setTraiter[T]{}
	}

	f.traits[name] = tr
	return f
}

// Reset resets the factory to its initial state
func (f *Factory[T]) Reset() {
	f.index = 1
}

// Build builds a value
func (f *Factory[T]) Build(ctx context.Context) *builder[T] {
	var v T
	if f.bluePrint != nil {
		v = f.bluePrint(f.index, v)
	}

	if f.isSetZeroValue {
		setNonZeroValues(f.index, &v, f.ignoreFields)
	}

	f.index++

	return &builder[T]{
		v:      &v,
		errors: []error{},
		f:      f,
	}
}

// BuildList creates a list of n values
func (f *Factory[T]) BuildList(ctx context.Context, n int) *builderList[T] {
	list := make([]*T, n)
	errs := []error{}
	if n < 1 {
		errs = append(errs, errors.New("BuildList: n must be greater than 0"))
		return &builderList[T]{errors: errs}
	}

	for i := 0; i < n; i++ {
		var v T
		if f.bluePrint != nil {
			v = f.bluePrint(f.index, v)
		}

		if f.isSetZeroValue {
			setNonZeroValues(f.index, &v, f.ignoreFields)
		}

		list[i] = &v
		f.index++
	}

	return &builderList[T]{
		list:   list,
		errors: errs,
		f:      f,
	}
}

// Get returns the value
func (b *builder[T]) Get() (T, error) {
	if len(b.errors) > 0 {
		return b.f.empty, genFinalError(b.errors)
	}

	return *b.v, nil
}

// Get returns the list of values
func (b *builderList[T]) Get() ([]T, error) {
	if len(b.errors) > 0 {
		return nil, genFinalError(b.errors)
	}

	output := make([]T, len(b.list))
	for i, v := range b.list {
		output[i] = *v
	}

	return output, nil
}

// Insert inserts the value into the database
func (b *builder[T]) Insert() (T, error) {
	if len(b.errors) > 0 {
		return b.f.empty, genFinalError(b.errors)
	}

	if b.f.db == nil {
		return b.f.empty, errors.New("DB connection is not provided")
	}

	if len(b.f.associations) > 0 {
		if err := b.setAss(); err != nil {
			return b.f.empty, err
		}
	}

	val, err := b.f.db.Insert(b.ctx, db.InserParams{StorageName: b.f.storageName, Value: b.v})
	if err != nil {
		return b.f.empty, err
	}

	v, ok := val.(*T)
	if !ok {
		return b.f.empty, errors.New("Insert: can't convert to pointer")
	}

	return *v, nil
}

// Insert inserts the list of values into the database
func (b *builderList[T]) Insert() ([]T, error) {
	if len(b.errors) > 0 {
		return nil, genFinalError(b.errors)
	}

	if b.f.db == nil {
		return nil, errors.New("DB connection is not provided")
	}

	if len(b.f.associations) > 0 {
		if err := b.setAss(); err != nil {
			return nil, err
		}
	}

	// convert to any type
	input := make([]interface{}, len(b.list))
	for i, v := range b.list {
		input[i] = v
	}
	vals, err := b.f.db.InsertList(b.ctx, db.InserListParams{StorageName: b.f.storageName, Values: input})
	if err != nil {
		return nil, err
	}

	// convert to []T
	output := make([]T, len(vals))
	for i, val := range vals {
		v, ok := val.(*T)
		if !ok {
			return nil, errors.New("Insert: can't convert to pointer")
		}

		output[i] = *v
	}

	return output, nil
}

// Overwrite overwrites the value with the given value
func (b *builder[T]) Overwrite(ow T) *builder[T] {
	if len(b.errors) > 0 {
		return b
	}

	if err := copyValues(b.v, ow); err != nil {
		b.errors = append(b.errors, err)
	}

	return b
}

// Overwrites overwrites the values with the given values
func (b *builderList[T]) Overwrites(ows ...T) *builderList[T] {
	if len(b.errors) > 0 {
		return b
	}

	for i := 0; i < len(ows) && i < len(b.list); i++ {
		if err := copyValues(b.list[i], ows[i]); err != nil {
			b.errors = append(b.errors, err)
			return b
		}
	}

	return b
}

// Overwrite overwrites the values with the given one value
func (b *builderList[T]) Overwrite(ow T) *builderList[T] {
	if len(b.errors) > 0 {
		return b
	}

	for i := 0; i < len(b.list); i++ {
		if err := copyValues(b.list[i], ow); err != nil {
			b.errors = append(b.errors, err)
			return b
		}
	}

	return b
}

// WithTrait invokes the traiter based on given name
func (b *builder[T]) WithTrait(name string) *builder[T] {
	if len(b.errors) > 0 {
		return b
	}

	tr, ok := b.f.traits[name]
	if !ok {
		b.errors = append(b.errors, fmt.Errorf("WithTrait: %s is not defiend at SetTrait", name))
		return b
	}

	tr(b.v)

	return b
}

// WithTraits invokes the traiter based on given names
func (b *builderList[T]) WithTraits(names ...string) *builderList[T] {
	if len(b.errors) > 0 {
		return b
	}

	for i := 0; i < len(names) && i < len(b.list); i++ {
		tr, ok := b.f.traits[names[i]]
		if !ok {
			b.errors = append(b.errors, fmt.Errorf("WithTrait: %s is not defiend at SetTrait", names[i]))
			return b
		}

		tr(b.list[i])
	}

	return b
}

// WithTrait invokes the traiter based on given name
func (b *builderList[T]) WithTrait(name string) *builderList[T] {
	if len(b.errors) > 0 {
		return b
	}

	tr, ok := b.f.traits[name]
	if !ok {
		b.errors = append(b.errors, fmt.Errorf("WithTrait: %s is not defiend at SetTrait", name))
		return b
	}

	for i := 0; i < len(b.list); i++ {
		tr(b.list[i])
	}

	return b
}

// SetZero sets the fields to zero value
func (b *builder[T]) SetZero(fields ...string) *builder[T] {
	if len(b.errors) > 0 {
		return b
	}

	for _, field := range fields {
		curField := reflect.ValueOf(b.v).Elem().FieldByName(field)
		if !curField.IsValid() {
			b.errors = append(b.errors, fmt.Errorf("SetZero: field %s is not found", field))
			return b
		}

		if !curField.CanSet() {
			b.errors = append(b.errors, fmt.Errorf("SetZero: field %s can not be set", field))
			return b
		}

		curField.Set(reflect.Zero(curField.Type()))
	}

	return b
}

// SetZero sets the fields to zero value for the given index.
// The paramter i is the index of the list you want to set the zero value
func (b *builderList[T]) SetZero(i int, fields ...string) *builderList[T] {
	if len(b.errors) > 0 {
		return b
	}

	if i >= len(b.list) || i < 0 {
		b.errors = append(b.errors, fmt.Errorf("SetZero: index %d is out of range", i))
		return b
	}

	for _, field := range fields {
		curField := reflect.ValueOf(b.list[i]).Elem().FieldByName(field)
		if !curField.IsValid() {
			b.errors = append(b.errors, fmt.Errorf("SetZero: field %s is not found", field))
			return b
		}

		if !curField.CanSet() {
			b.errors = append(b.errors, fmt.Errorf("SetZero: field %s can not be set", field))
			return b
		}

		curField.Set(reflect.Zero(curField.Type()))
	}

	return b
}

// WihtOne set one association to the factory value
func (b *builder[T]) WithOne(v interface{}) *builder[T] {
	if len(b.errors) > 0 {
		return b
	}

	// set tagToInfo if it's not set
	if len(b.f.tagToInfo) == 0 {
		t, err := genTagToInfo(b.f.dataType)
		if err != nil {
			b.errors = append(b.errors, err)
			return b
		}
		b.f.tagToInfo = t
	}

	if err := setAssValue(v, b.f.tagToInfo, b.f.index, "WithOne"); err != nil {
		b.errors = append(b.errors, err)
		return b
	}

	name := reflect.TypeOf(v).Elem().Name()
	b.f.associations[name] = []interface{}{v}
	b.f.index++
	return b
}

// WihtOne set one association to the factory value
func (b *builderList[T]) WithOne(v interface{}) *builderList[T] {
	if len(b.errors) > 0 {
		return b
	}

	// set tagToInfo if it's not set
	if len(b.f.tagToInfo) == 0 {
		t, err := genTagToInfo(b.f.dataType)
		if err != nil {
			b.errors = append(b.errors, err)
			return b
		}
		b.f.tagToInfo = t
	}

	if err := setAssValue(v, b.f.tagToInfo, b.f.index, "WithOne"); err != nil {
		b.errors = append(b.errors, err)
		return b
	}

	name := reflect.TypeOf(v).Elem().Name()
	b.f.associations[name] = []interface{}{v}
	b.f.index++
	return b
}

// WithMany set many associations to the factory value
func (b *builderList[T]) WithMany(values ...interface{}) *builderList[T] {
	if len(b.errors) > 0 {
		return b
	}

	// set tagToInfo if it's not set
	if len(b.f.tagToInfo) == 0 {
		t, err := genTagToInfo(b.f.dataType)
		if err != nil {
			b.errors = append(b.errors, err)
			return b
		}
		b.f.tagToInfo = t
	}

	var curValName string
	for _, v := range values {
		if err := setAssValue(v, b.f.tagToInfo, b.f.index, "WithMany"); err != nil {
			b.errors = append(b.errors, err)
			return b
		}

		// check if the provided values are of the same type
		// because we have to make sure all the value is pointer (setAssValue does that for us)
		// before we can use Elem()
		if curValName != "" && curValName != reflect.TypeOf(v).Elem().Name() {
			b.errors = append(b.errors, fmt.Errorf("WithMany: provided values are not the same type"))
			return b
		}

		name := reflect.TypeOf(v).Elem().Name()
		b.f.associations[name] = append(b.f.associations[name], v)
		b.f.index++
		curValName = name
	}

	return b
}

// setAss sets and inserts the associations
func (b *builder[T]) setAss() error {
	// insert the associations
	if err := insertAss(b.ctx, b.f.db, b.f.associations, b.f.tagToInfo); err != nil {
		return err
	}

	// set the connection between the factory value and the associations
	for name, vals := range b.f.associations {
		// use vs[0] because we can make sure insertAss only invoke with Build function
		// which means there's only one factory value
		// so that each associations only allow one value
		fieldName := b.f.tagToInfo[name].fieldName
		if err := setField(b.v, fieldName, vals[0], "InsertWithAss"); err != nil {
			return err
		}
	}

	// clear associations
	b.f.associations = map[string][]interface{}{}

	return nil
}

// setAss sets and inserts the associations
func (b *builderList[T]) setAss() error {
	// insert the associations
	if err := insertAss(b.ctx, b.f.db, b.f.associations, b.f.tagToInfo); err != nil {
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

			fieldName := b.f.tagToInfo[name].fieldName
			if err := setField(l, fieldName, v, "InsertWithAss"); err != nil {
				return err
			}
		}
	}

	// clear associations
	b.f.associations = map[string][]interface{}{}

	return nil
}
