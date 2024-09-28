package gofacto

import (
	"context"
	"fmt"
	"reflect"

	"github.com/eyo-chen/gofacto/internal/db"
)

// assocNode is the association node.
// Each node contains it's metadata and the list of foreign key references
type assocNode struct {
	name         string
	vals         []interface{}
	tableName    string
	ignoreFields []string
	dependencies []fkRef
}

// fkRef is the foreign key reference
type fkRef struct {
	vals         []interface{}
	tableName    string
	fieldName    string
	foreignField string
}

// nodeInfo is used to store the information of a node for later reference.
//
// e.g. "User" -> {vals: [User1, User2, User3], tableName: "users"}
type nodeInfo struct {
	vals      []interface{}
	tableName string
}

func (b *builder[T]) insertWithAssoc(ctx context.Context) (T, error) {
	// add factory value into association
	b.f.associations = append(b.f.associations, []interface{}{b.v})

	res, err := b.f.prepareAndInsertAssoc(ctx)
	if err != nil {
		return b.f.empty, err
	}

	v, ok := res[0].(*T)
	if !ok {
		return b.f.empty, errCantCvtToPtr
	}

	return *v, nil
}

func (b *builderList[T]) insertWithAssoc(ctx context.Context) ([]T, error) {
	// add factory value into association
	vals := make([]interface{}, len(b.list))
	for i, v := range b.list {
		vals[i] = v
	}
	b.f.associations = append(b.f.associations, vals)

	res, err := b.f.prepareAndInsertAssoc(ctx)
	if err != nil {
		return nil, err
	}

	ts := make([]T, len(res))
	for i, val := range res {
		v, ok := val.(*T)
		if !ok {
			return nil, errCantCvtToPtr
		}

		ts[i] = *v
	}

	return ts, nil
}

// prepareAndInsertAssoc handles the preparation and insertion of associations
func (f *Factory[T]) prepareAndInsertAssoc(ctx context.Context) ([]interface{}, error) {
	// create node info map
	nodeInfoMap, err := f.genNodeInfoMap()
	if err != nil {
		return nil, err
	}

	// generate deep association nodes
	deepAssoc, err := f.genAssocNodes(nodeInfoMap)
	if err != nil {
		return nil, err
	}

	// insert the deep association nodes into the database
	return f.insertAssocNode(ctx, deepAssoc)
}

// insertAssocNode inserts the association nodes into the database.
// It first sets the foreign key fields for each node, then insert the node into the database.
func (f *Factory[T]) insertAssocNode(ctx context.Context, nodes []assocNode) ([]interface{}, error) {
	var fVal []interface{}

	// each node might have multiple values and dependencies
	// e.g. SubCategory have User and MainCategory
	// for each subCategory, has to set the foreign key fields for User and MainCategory
	// cache is used to handle the case when the dependency is less than the number of values
	// e.g. SubCategory*3, User*2, MainCategory*1
	// for the 1st SubCategory, set the foreign key fields for User1 and MainCategory1
	// for the 2nd SubCategory, set the foreign key fields for User2 and MainCategory1
	// for the 3rd SubCategory, set the foreign key fields for User2 and MainCategory1
	// nodes are guaranteed to have correct oreder
	// 1. user is populated with random values, and insert into db
	// 2. mainCategory is populated with random values, and insert into db
	// 3. subCategory is populated with random values, and insert into db
	for _, node := range nodes {
		cache := map[string]interface{}{}
		for i, v := range node.vals {
			for _, dep := range node.dependencies {
				var d interface{}
				if i >= len(dep.vals) {
					d = cache[dep.fieldName]
				} else {
					d = dep.vals[i]
					cache[dep.fieldName] = d
				}

				if d == nil {
					continue
				}

				// set the foreign key field
				if err := setForeignKey(v, dep.fieldName, d); err != nil {
					return nil, err
				}
				if dep.foreignField != "" {
					if err := setField(v, dep.foreignField, d); err != nil {
						return nil, err
					}
				}
			}

			f.setNonZeroValues(v, node.ignoreFields)
			f.index++
		}

		res, err := f.db.InsertList(ctx, db.InsertListParams{StorageName: node.tableName, Values: node.vals})
		if err != nil {
			return nil, err
		}

		// if the node is the factory value, set the fVal, and return later
		if node.name == reflect.TypeOf(f.empty).Name() {
			fVal = res
		}
	}

	return fVal, nil
}

// genNodeInfoMap generates the node info map
func (f *Factory[T]) genNodeInfoMap() (map[string]nodeInfo, error) {
	nodeInfoMap := make(map[string]nodeInfo)

	// it's guaranteed that the each element in the 1D slice is same type
	// so we can use the 1st element to get the type
	// along with the iteration, we only cares two things:
	// (1) vals: the vals in each iteration
	// (2) tableName: can only know when processing the fields of the struct
	// note that tableName is only found out in other's struct fields
	// e.g. SubCategory has User, we can only know the tableName of User when processing the fields of SubCategory
	for _, vals := range f.associations {
		val := vals[0]
		typ := reflect.TypeOf(val).Elem()
		name := typ.Name()
		updateNodeInfoMap(nodeInfoMap, vals, name, "") // update the vals field
		err := processStructFields(typ, func(t tag, hasTag bool) error {
			if t.omit || !hasTag {
				return nil
			}

			updateNodeInfoMap(nodeInfoMap, nil, t.structName, t.tableName) // update the tableName field

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// add the factory value into nodeInfoMap
	// the last element is guaranteed to be the factory value
	// it's implemented by the caller to avoid passing unnecessary vals
	// the logic here is tricky but necessary
	// because the factory value is not referenced by other association values
	// e.g. SubCategory*3, User*2, MainCategory*1
	// the factory value(SubCategory) is not referenced by other association values
	// so we have to manually add it's info into the nodeInfoMap
	name := reflect.TypeOf(f.empty).Name()
	nodeInfoMap[name] = nodeInfo{
		tableName: f.storageName,
		vals:      f.associations[len(f.associations)-1],
	}

	return nodeInfoMap, nil
}

// updateNodeInfoMap updates the node info map
func updateNodeInfoMap(nodeInfoMap map[string]nodeInfo, vals []interface{}, name, tableName string) {
	if info, ok := nodeInfoMap[name]; ok {
		if len(vals) > 0 {
			info.vals = vals
		}

		if tableName != "" {
			info.tableName = tableName
		}

		nodeInfoMap[name] = info
	} else {
		nodeInfoMap[name] = nodeInfo{
			vals:      vals,
			tableName: tableName,
		}
	}
}

// genAssocNodes returns the association nodes in topological order
// If there's a cycle dependency, it returns an error
func (f *Factory[T]) genAssocNodes(nodeInfoMap map[string]nodeInfo) ([]assocNode, error) {
	d := newDAG()

	// it's guaranteed that the each element in the 1D slice is same type
	// so we can use the 1st element to get the type
	for _, vals := range f.associations {
		typ := reflect.TypeOf(vals[0]).Elem()
		name := typ.Name()

		deepAssoc := assocNode{
			name:      name,
			vals:      vals,
			tableName: nodeInfoMap[name].tableName,
		}

		// process the fields to find out the dependencies
		err := processStructFields(typ, func(t tag, hasTag bool) error {
			if !hasTag {
				return nil
			}

			if t.omit {
				deepAssoc.ignoreFields = append(deepAssoc.ignoreFields, t.fieldName)
				return nil
			}

			deepAssoc.dependencies = append(deepAssoc.dependencies, fkRef{
				vals:         nodeInfoMap[t.structName].vals,
				tableName:    t.tableName,
				fieldName:    t.fieldName,
				foreignField: t.foreignField,
			})

			// e.g. User(fk) -> SubCategory
			d.addEdge(t.structName, name)
			return nil
		})

		if err != nil {
			return nil, err
		}

		d.addNode(deepAssoc)
	}

	if d.hasCycle() {
		return nil, errCycleDependency
	}

	return d.topologicalSort(), nil
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

// checkAssocs checks if the input association values are valid
func checkAssocs(vals []interface{}) error {
	var name string
	for _, v := range vals {
		if err := checkAssoc(v); err != nil {
			return err
		}

		// check if the type of the value is the same as the previous value
		curValName := reflect.TypeOf(v).Elem().Name()
		if name != "" && name != curValName {
			return errValueNotTheSameType
		}

		name = curValName
	}

	return nil
}
