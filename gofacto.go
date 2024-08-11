package gofacto

import (
	"context"
	"fmt"
	"reflect"

	"github.com/eyo-chen/gofacto/db"
	"github.com/eyo-chen/gofacto/internal/types"
	"github.com/eyo-chen/gofacto/internal/utils"
)

// Factory is the gofacto factory to create mock data
type Factory[T any] struct {
	db             db.Database
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

	// map from name to list of associations
	// e.g. "User" -> []*User
	associations map[string][]interface{}

	// map from tag to metadata
	// e.g. "User" -> {tableName: "users", fieldName: "UserID"}
	tagToInfo map[string]tagInfo
}

// blueprintFunc is a client-defined function to create a new value
type blueprintFunc[T any] func(i int) T

// setTraiter is a client-defined function to add a trait to mutate the value
type setTraiter[T any] func(v *T)

// tagInfo is the metadata for the tag
type tagInfo struct {
	tableName    string
	fieldName    string
	foreignField string
}

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

// New creates a new gofacto factory
func New[T any](v T) *Factory[T] {
	dataType := reflect.TypeOf(v)

	ti, ifd, err := extractTag(dataType)

	return &Factory[T]{
		dataType:       dataType,
		empty:          reflect.New(dataType).Elem().Interface().(T),
		associations:   map[string][]interface{}{},
		storageName:    fmt.Sprintf("%ss", utils.CamelToSnake(dataType.Name())),
		tagToInfo:      ti,
		ignoreFields:   ifd,
		index:          1,
		isSetZeroValue: true,
		traits:         map[string]setTraiter[T]{},
		err:            err,
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
func (f *Factory[T]) WithDB(db db.Database) *Factory[T] {
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
}

// Build builds a value
func (f *Factory[T]) Build(ctx context.Context) *builder[T] {
	var v T
	if f.blueprint != nil {
		v = f.blueprint(f.index)
	}

	if f.isSetZeroValue {
		f.setNonZeroValues(&v)
	}

	f.index++

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
			err:  types.ErrBuildListNGreaterThanZero,
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
			f.setNonZeroValues(&v)
		}

		list[i] = &v
		f.index++
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
		return b.f.empty, types.ErrDBIsNotProvided
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
		return b.f.empty, types.ErrCantCvtToPtr
	}

	return *v, nil
}

// Insert inserts the list of values into the database
func (b *builderList[T]) Insert() ([]T, error) {
	if b.err != nil {
		return nil, b.err
	}

	if b.f.db == nil {
		return nil, types.ErrDBIsNotProvided
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
			return nil, types.ErrCantCvtToPtr
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

// WithTrait invokes the traiter based on given name
func (b *builder[T]) WithTrait(name string) *builder[T] {
	if b.err != nil {
		return b
	}

	tr, ok := b.f.traits[name]
	if !ok {
		b.err = fmt.Errorf("%w: %s", types.ErrWithTraitNameNotFound, name)
		return b
	}

	tr(b.v)

	return b
}

// WithTraits invokes the traiter based on given names
func (b *builderList[T]) WithTraits(names ...string) *builderList[T] {
	if b.err != nil {
		return b
	}

	for i := 0; i < len(names) && i < len(b.list); i++ {
		tr, ok := b.f.traits[names[i]]
		if !ok {
			b.err = fmt.Errorf("%w: %s", types.ErrWithTraitNameNotFound, names[i])
			return b
		}

		tr(b.list[i])
	}

	return b
}

// WithTrait invokes the traiter based on given name
func (b *builderList[T]) WithTrait(name string) *builderList[T] {
	if b.err != nil {
		return b
	}

	tr, ok := b.f.traits[name]
	if !ok {
		b.err = fmt.Errorf("%w: %s", types.ErrWithTraitNameNotFound, name)
		return b
	}

	for i := 0; i < len(b.list); i++ {
		tr(b.list[i])
	}

	return b
}

// SetZero sets the fields to zero value
func (b *builder[T]) SetZero(fields ...string) *builder[T] {
	if b.err != nil {
		return b
	}

	for _, field := range fields {
		curField := reflect.ValueOf(b.v).Elem().FieldByName(field)
		if !curField.IsValid() {
			b.err = fmt.Errorf("%w: %s", types.ErrFieldNotFound, field)
			return b
		}

		if !curField.CanSet() {
			b.err = fmt.Errorf("%w: %s", types.ErrFieldCantSet, field)
			return b
		}

		curField.Set(reflect.Zero(curField.Type()))
	}

	return b
}

// SetZero sets the fields to zero value for the given index.
// The paramter i is the index of the list you want to set the zero value
func (b *builderList[T]) SetZero(i int, fields ...string) *builderList[T] {
	if b.err != nil {
		return b
	}

	if i >= len(b.list) || i < 0 {
		b.err = types.ErrIndexIsOutOfRange
		return b
	}

	for _, field := range fields {
		curField := reflect.ValueOf(b.list[i]).Elem().FieldByName(field)
		if !curField.IsValid() {
			b.err = fmt.Errorf("%w: %s", types.ErrFieldNotFound, field)
			return b
		}

		if !curField.CanSet() {
			b.err = fmt.Errorf("%w: %s", types.ErrFieldCantSet, field)
			return b
		}

		curField.Set(reflect.Zero(curField.Type()))
	}

	return b
}

// WihtOne set one association to the factory value
func (b *builder[T]) WithOne(v interface{}, ignoreFields ...string) *builder[T] {
	if b.err != nil {
		return b
	}

	if err := b.f.setAssValue(v); err != nil {
		b.err = err
		return b
	}

	name := reflect.TypeOf(v).Elem().Name()
	b.f.associations[name] = []interface{}{v}
	b.f.index++
	return b
}

// WihtOne set one association to the factory value
func (b *builderList[T]) WithOne(v interface{}, ignoreFields ...string) *builderList[T] {
	if b.err != nil {
		return b
	}

	if err := b.f.setAssValue(v); err != nil {
		b.err = err
		return b
	}

	name := reflect.TypeOf(v).Elem().Name()
	b.f.associations[name] = []interface{}{v}
	b.f.index++
	return b
}

// WithMany set many associations to the factory value
func (b *builderList[T]) WithMany(values []interface{}, ignoreFields ...string) *builderList[T] {
	if b.err != nil {
		return b
	}

	var curValName string
	for _, v := range values {
		if err := b.f.setAssValue(v); err != nil {
			b.err = err
			return b
		}

		// check if the provided values are of the same type
		// because we have to make sure all the value is pointer (setAssValue does that for us)
		// before we can use Elem()
		if curValName != "" && curValName != reflect.TypeOf(v).Elem().Name() {
			b.err = types.ErrValueNotTheSameType
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
