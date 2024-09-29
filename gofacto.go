package gofacto

import (
	"context"
	"fmt"
	"reflect"

	"github.com/eyo-chen/gofacto/internal/db"
	"github.com/eyo-chen/gofacto/internal/utils"
)

// Factory is the gofacto factory to create mock data
type Factory[T any] struct {
	db             database
	blueprint      blueprintFunc[T]
	storageName    string
	dataType       reflect.Type
	empty          T
	index          int
	ignoreFields   []string
	isSetZeroValue bool
	err            error

	// map from name to trait function
	traits map[string]setTraiter[T]

	// associations is a list of associations
	associations [][]interface{}
}

// blueprintFunc is a client-defined function to create a new value
type blueprintFunc[T any] func(i int) T

// setTraiter is a client-defined function to add a trait to mutate the value
type setTraiter[T any] func(v *T)

// builder is for building a single value
type builder[T any] struct {
	ctx context.Context
	v   *T
	err error
	f   *Factory[T]
}

// builderList is for building a list of values
type builderList[T any] struct {
	ctx  context.Context
	list []*T
	err  error
	f    *Factory[T]
}

// New initializes a new factory
func New[T any](v T) *Factory[T] {
	dataType := reflect.TypeOf(v)

	if dataType.Kind() != reflect.Struct {
		return &Factory[T]{
			err: fmt.Errorf("%w: %v", errInvalidType, dataType.Kind()),
		}
	}

	ifd, err := extractTag(dataType)
	if err != nil {
		return &Factory[T]{
			err: err,
		}
	}

	return &Factory[T]{
		dataType:       dataType,
		empty:          reflect.New(dataType).Elem().Interface().(T),
		associations:   [][]interface{}{},
		storageName:    fmt.Sprintf("%ss", utils.CamelToSnake(dataType.Name())),
		ignoreFields:   ifd,
		index:          1,
		isSetZeroValue: true,
		traits:         map[string]setTraiter[T]{},
	}
}

// WithBlueprint sets the blueprint function
func (f *Factory[T]) WithBlueprint(bp blueprintFunc[T]) *Factory[T] {
	f.blueprint = bp
	return f
}

// WithStorageName sets the storage name
//
// table name for SQL, collection name for NoSQL
func (f *Factory[T]) WithStorageName(name string) *Factory[T] {
	f.storageName = name
	return f
}

// WithDB sets the database connection
func (f *Factory[T]) WithDB(db database) *Factory[T] {
	f.db = db
	return f
}

// WithIsSetZeroValue sets whether to set zero value for the fields
func (f *Factory[T]) WithIsSetZeroValue(isSetZeroValue bool) *Factory[T] {
	f.isSetZeroValue = isSetZeroValue
	return f
}

// WithTrait sets the trait function
func (f *Factory[T]) WithTrait(name string, tr setTraiter[T]) *Factory[T] {
	f.traits[name] = tr
	return f
}

// Reset resets the factory to its initial state
func (f *Factory[T]) Reset() {
	f.index = 1
	f.err = nil
	f.associations = [][]interface{}{}
}

// Build builds a value
func (f *Factory[T]) Build(ctx context.Context) *builder[T] {
	var v T
	if f.blueprint != nil {
		v = f.blueprint(f.index)
	}

	if f.isSetZeroValue {
		f.setNonZeroValues(&v, f.ignoreFields)
		f.index++
	}

	return &builder[T]{
		ctx: ctx,
		v:   &v,
		f:   f,
		err: nil,
	}
}

// BuildList creates a list of n values
func (f *Factory[T]) BuildList(ctx context.Context, n int) *builderList[T] {
	if n < 1 {
		return &builderList[T]{
			ctx:  ctx,
			list: nil,
			err:  errBuildListNGreaterThanZero,
			f:    f,
		}
	}

	list := make([]*T, n)
	for i := 0; i < n; i++ {
		var v T
		if f.blueprint != nil {
			v = f.blueprint(f.index)
		}

		if f.isSetZeroValue {
			f.setNonZeroValues(&v, f.ignoreFields)
			f.index++
		}

		list[i] = &v
	}

	return &builderList[T]{
		ctx:  ctx,
		list: list,
		err:  nil,
		f:    f,
	}
}

// Get returns the value
func (b *builder[T]) Get() (T, error) {
	if b.err != nil {
		return b.f.empty, b.err
	}

	return *b.v, nil
}

// Get returns the list of values
func (b *builderList[T]) Get() ([]T, error) {
	if b.err != nil {
		return nil, b.err
	}

	output := make([]T, len(b.list))
	for i, v := range b.list {
		output[i] = *v
	}

	return output, nil
}

// Insert inserts the value into the database
func (b *builder[T]) Insert() (T, error) {
	if b.err != nil {
		return b.f.empty, b.err
	}

	if b.f.db == nil {
		return b.f.empty, errDBIsNotProvided
	}

	if len(b.f.associations) > 0 {
		return b.insertWithAssoc(b.ctx)
	}

	val, err := b.f.db.Insert(b.ctx, db.InsertParams{StorageName: b.f.storageName, Value: b.v})
	if err != nil {
		return b.f.empty, err
	}

	v, ok := val.(*T)
	if !ok {
		return b.f.empty, errCantCvtToPtr
	}

	return *v, nil
}

// Insert inserts the list of values into the database
func (b *builderList[T]) Insert() ([]T, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.f.db == nil {
		return nil, errDBIsNotProvided
	}

	if len(b.f.associations) > 0 {
		return b.insertWithAssoc(b.ctx)
	}

	// convert to any type
	input := make([]interface{}, len(b.list))
	for i, v := range b.list {
		input[i] = v
	}
	vals, err := b.f.db.InsertList(b.ctx, db.InsertListParams{StorageName: b.f.storageName, Values: input})
	if err != nil {
		return nil, err
	}

	// convert to []T
	output := make([]T, len(vals))
	for i, val := range vals {
		v, ok := val.(*T)
		if !ok {
			return nil, errCantCvtToPtr
		}

		output[i] = *v
	}

	return output, nil
}

// Overwrite overwrites the value with the given value
func (b *builder[T]) Overwrite(ow T) *builder[T] {
	if b.err != nil {
		return b
	}

	if err := copyValues(b.v, ow); err != nil {
		b.err = err
		return b
	}

	return b
}

// Overwrites overwrites the values with the given values
func (b *builderList[T]) Overwrites(ows ...T) *builderList[T] {
	if b.err != nil {
		return b
	}

	for i := 0; i < len(ows) && i < len(b.list); i++ {
		if err := copyValues(b.list[i], ows[i]); err != nil {
			b.err = err
			return b
		}
	}

	return b
}

// Overwrite overwrites the values with the given one value
func (b *builderList[T]) Overwrite(ow T) *builderList[T] {
	if b.err != nil {
		return b
	}

	for i := 0; i < len(b.list); i++ {
		if err := copyValues(b.list[i], ow); err != nil {
			b.err = err
			return b
		}
	}

	return b
}

// SetTrait invokes the trait function based on the given key.
// It returns an error if the key is not found.
func (b *builder[T]) SetTrait(key string) *builder[T] {
	if b.err != nil {
		return b
	}

	tr, ok := b.f.traits[key]
	if !ok {
		b.err = fmt.Errorf("%w: %s", errWithTraitNameNotFound, key)
		return b
	}

	tr(b.v)

	return b
}

// SetTraits invokes the trait functions based on the given keys.
// It returns an error if the key is not found.
func (b *builderList[T]) SetTraits(keys ...string) *builderList[T] {
	if b.err != nil {
		return b
	}

	for i := 0; i < len(keys) && i < len(b.list); i++ {
		tr, ok := b.f.traits[keys[i]]
		if !ok {
			b.err = fmt.Errorf("%w: %s", errWithTraitNameNotFound, keys[i])
			return b
		}

		tr(b.list[i])
	}

	return b
}

// SetTrait invokes the trait function based on the given key.
// It returns an error if the key is not found.
func (b *builderList[T]) SetTrait(key string) *builderList[T] {
	if b.err != nil {
		return b
	}

	tr, ok := b.f.traits[key]
	if !ok {
		b.err = fmt.Errorf("%w: %s", errWithTraitNameNotFound, key)
		return b
	}

	for i := 0; i < len(b.list); i++ {
		tr(b.list[i])
	}

	return b
}

// SetZero sets the fields to zero value.
// It returns an error if the field is not found.
func (b *builder[T]) SetZero(fields ...string) *builder[T] {
	if b.err != nil {
		return b
	}

	for _, field := range fields {
		curField := reflect.ValueOf(b.v).Elem().FieldByName(field)
		if !curField.IsValid() {
			b.err = fmt.Errorf("%w: %s", errFieldNotFound, field)
			return b
		}

		if !curField.CanSet() {
			b.err = fmt.Errorf("%w: %s", errFieldCantSet, field)
			return b
		}

		curField.Set(reflect.Zero(curField.Type()))
	}

	return b
}

// SetZero sets the fields to zero value for the given index.
// The parameter i is the index of the list you want to set the zero value.
// It returns an error if the index is out of range or the field is not found.
func (b *builderList[T]) SetZero(i int, fields ...string) *builderList[T] {
	if b.err != nil {
		return b
	}

	if i >= len(b.list) || i < 0 {
		b.err = errIndexIsOutOfRange
		return b
	}

	for _, field := range fields {
		curField := reflect.ValueOf(b.list[i]).Elem().FieldByName(field)
		if !curField.IsValid() {
			b.err = fmt.Errorf("%w: %s", errFieldNotFound, field)
			return b
		}

		if !curField.CanSet() {
			b.err = fmt.Errorf("%w: %s", errFieldCantSet, field)
			return b
		}

		curField.Set(reflect.Zero(curField.Type()))
	}

	return b
}

// WithOne sets one or more single-value associations for the factory.
//
// This function supports setting associations for both single-level and multi-level relationships.
// Each argument must be a pointer to a struct representing the associated entity.
//
// Examples:
//
//  1. Single-level association (e.g., Transaction has a User):
//     transactionFactory.WithOne(&User{})
//
//  2. Multi-level association (e.g., Transaction -> Category -> User):
//     transactionFactory.WithOne(&Category{}, &User{})
//
// Note: All arguments must be pointers to structs. Non-pointer or non-struct arguments will result in an error.
func (b *builder[T]) WithOne(vals ...interface{}) *builder[T] {
	if b.err != nil {
		return b
	}

	for _, v := range vals {
		if err := checkAssoc(v); err != nil {
			b.err = err
			return b
		}
		b.f.associations = append(b.f.associations, []interface{}{v})
	}

	return b
}

// WithOne sets one or more single-value associations for the factory.
//
// This function supports setting associations for both single-level and multi-level relationships.
// Each argument must be a pointer to a struct representing the associated entity.
//
// Examples:
//
//  1. Single-level association (e.g., Transaction has a User):
//     transactionFactory.WithOne(&User{})
//
//  2. Multi-level association (e.g., Transaction -> Category -> User):
//     transactionFactory.WithOne(&Category{}, &User{})
//
// Note: All arguments must be pointers to structs. Non-pointer or non-struct arguments will result in an error.
func (b *builderList[T]) WithOne(vals ...interface{}) *builderList[T] {
	if b.err != nil {
		return b
	}

	for _, v := range vals {
		if err := checkAssoc(v); err != nil {
			b.err = err
			return b
		}

		b.f.associations = append(b.f.associations, []interface{}{v})
	}

	return b
}

// WithMany sets multiple associations of the same type for each item in the factory list.
//
// The input must be a slice of interface{}, where each element is a pointer to a struct of the same type.
//
// Example:
//
//  1. Single-level association (e.g., Transactions has multiple Users):
//     transactionFactory.WithMany([]interface{}{&User{}, &User{}})
//
//  2. Multi-level association (e.g., Transaction -> Category -> User):
//     transactionFactory.WithMany([]interface{}{&Category{}, &Category{}}).WithMany([]interface{}{&User{}, &User{}})
//
// Note:
//   - All elements in the input slice must be pointers to structs of the same type.
//   - Non-pointer, non-struct, or mixed-type arguments will result in an error.
func (b *builderList[T]) WithMany(vals []interface{}) *builderList[T] {
	if b.err != nil {
		return b
	}

	if err := checkAssocs(vals); err != nil {
		b.err = err
		return b
	}

	b.f.associations = append(b.f.associations, vals)
	return b
}
