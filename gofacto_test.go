package gofacto

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/eyo-chen/gofacto/internal/db"
	"github.com/eyo-chen/gofacto/internal/testutils"
)

var (
	now     = time.Now()
	mockCTX = context.Background()
)

// mockDB is a mock implementation of the db.DB interface.
type mockDB struct{}

// Insert inserts a single value into the database.
func (m *mockDB) Insert(ctx context.Context, params db.InsertParams) (interface{}, error) {
	val := reflect.ValueOf(params.Value)
	if err := setIDField(val); err != nil {
		return nil, err
	}

	return params.Value, nil
}

// InsertList inserts a list of values into the database.
func (m *mockDB) InsertList(ctx context.Context, params db.InsertListParams) ([]interface{}, error) {
	for _, v := range params.Values {
		val := reflect.ValueOf(v)
		if err := setIDField(val); err != nil {
			return nil, err
		}
	}

	return params.Values, nil
}

// GenCustomType generates a custom type.
func (m *mockDB) GenCustomType(reflect.Type) (interface{}, bool) {
	return nil, false
}

// setIDField sets the ID field of a struct.
// In this mock, it always sets the ID field to 1.
func setIDField(val reflect.Value) error {
	v := val.Elem()
	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return errors.New("ID field not found")
	}
	randomID := rand.Intn(1000) + 1
	idField.SetInt(int64(randomID))

	return nil
}

// testAssocStruct is a struct with a foreign key to test the association functionality.
type testAssocStruct struct {
	ID            int
	ForeignKey    int  `gofacto:"foreignKey,struct:testStructWithID,field:ForeignValue"`
	ForeignKey2   *int `gofacto:"foreignKey,struct:testStructWithID2,field:ForeignValue2,table:test_struct_with_id2s"`
	ForeignValue  testStructWithID
	ForeignValue2 *testStructWithID2
}

// testStructWithID is a struct with an ID field to test the insert functionality.
type testStructWithID struct {
	ID int
}

// testStructWithID2 is a struct with an ID field to test the association functionality.
// This is for testing foreign field with pointer struct.
type testStructWithID2 struct {
	ID           int
	ForeignKey   int `gofacto:"foreignKey,struct:testStructWithID3,field:ForeignValue"`
	ForeignValue testStructWithID3
	Name         string
}

// testStructWithID3 is a struct with an ID field to test the association functionality.
type testStructWithID3 struct {
	ID   int
	Name string
}

type testStructWithCycle struct {
	ID         int
	ForeignKey int `gofacto:"foreignKey,struct:testStructWithCycle2"`
}

type testStructWithCycle2 struct {
	ID         int
	ForeignKey int `gofacto:"foreignKey,struct:testStructWithCycle"`
}

// customType is a custom type to test the custom type functionality.
type customType string

const (
	customType1 customType = "customType1"
	customType2 customType = "customType2"
)

type testStruct struct {
	Int            int
	PtrInt         *int
	Int8           int8
	Int32          int32
	Int64          int64
	Uint           uint
	Uint8          uint8
	Uint32         uint32
	Uint64         uint64
	CustomType     customType
	PtrCustomType  *customType
	Str            string
	PtrStr         *string
	Bool           bool
	PtrBool        *bool
	Time           time.Time
	PtrTime        *time.Time
	Float          float64
	PtrFloat       *float64
	Interface      interface{}
	Struct         subStruct
	PtrStruct      *subStruct
	Slice          []int
	PtrSlice       []*int
	SliceStruct    []subStruct
	SlicePtrStruct []*subStruct
	privateField   string
}

type subStruct struct {
	ID   int
	Name string
}

func TestNew(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when not pass any config, default config should be set": new_NoConfig,
		"when pass blueprint, blueprint should be set":           new_WithBlueprint,
		"when pass storage name, storage name should be set":     new_WithStorageName,
		"when pass db, db should be set":                         new_WithDB,
		"when pass isSetZeroValue, isSetZeroValue should be set": new_WithIsSetZeroValue,
		"when pass trait, trait should be set":                   new_WithTrait,
		"when pass correct tag, should be set":                   new_CorrectTagFormat,
		"when pass non struct, should return error":              new_PassNonStruct,
		"when pass wrong tag, should return error":               new_WrongTagKey,
		"when pass wrong tag foreign key, should return error":   new_WrongTagForeignKey,
		"when pass empty tag, should return error":               new_EmptyTag,
		"when pass wrong tag format, should return error":        new_WrongTagFormat,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func new_NoConfig(t *testing.T) {
	got := New(testStruct{})

	want := &Factory[testStruct]{
		dataType:       reflect.TypeOf(testStruct{}),
		storageName:    "test_structs",
		ignoreFields:   []string{},
		index:          1,
		isSetZeroValue: true,
	}

	if err := checkFactory(got, want); err != nil {
		t.Fatal(err.Error())
	}
}

func new_WithBlueprint(t *testing.T) {
	blueprint := func(i int) testStruct {
		return testStruct{
			Int: i * 2,
		}
	}

	got := New(testStruct{}).WithBlueprint(blueprint)

	want := &Factory[testStruct]{
		blueprint:    blueprint,
		dataType:     reflect.TypeOf(testStruct{}),
		empty:        testStruct{},
		associations: [][]interface{}{},
		storageName:  "test_structs",

		ignoreFields:   []string{},
		index:          1,
		isSetZeroValue: true,
	}

	if err := checkFactory(got, want); err != nil {
		t.Fatal(err.Error())
	}
}

func new_WithStorageName(t *testing.T) {
	storageName := "test_storage"
	got := New(testStruct{}).WithStorageName(storageName)

	want := &Factory[testStruct]{
		dataType:       reflect.TypeOf(testStruct{}),
		storageName:    "test_storage",
		ignoreFields:   []string{},
		index:          1,
		isSetZeroValue: true,
	}

	if err := checkFactory(got, want); err != nil {
		t.Fatal(err.Error())
	}
}

func new_WithDB(t *testing.T) {
	db := &mockDB{}
	got := New(testStruct{}).WithDB(db)

	want := &Factory[testStruct]{
		dataType:    reflect.TypeOf(testStruct{}),
		storageName: "test_structs",

		ignoreFields:   []string{},
		index:          1,
		isSetZeroValue: true,
		db:             db,
	}

	if err := checkFactory(got, want); err != nil {
		t.Fatal(err.Error())
	}
}

func new_WithIsSetZeroValue(t *testing.T) {
	got := New(testStruct{}).WithIsSetZeroValue(false)

	want := &Factory[testStruct]{
		dataType:    reflect.TypeOf(testStruct{}),
		storageName: "test_structs",

		ignoreFields:   []string{},
		index:          1,
		isSetZeroValue: false,
	}

	if err := checkFactory(got, want); err != nil {
		t.Fatal(err.Error())
	}
}

func new_WithTrait(t *testing.T) {
	trait := func(*testStruct) {}

	got := New(testStruct{}).WithTrait("test", trait)

	want := &Factory[testStruct]{
		dataType:    reflect.TypeOf(testStruct{}),
		storageName: "test_structs",

		ignoreFields:   []string{},
		index:          1,
		isSetZeroValue: true,
		traits: map[string]setTraiter[testStruct]{
			"test": trait,
		},
	}

	if err := checkFactory(got, want); err != nil {
		t.Fatal(err.Error())
	}
}

func new_CorrectTagFormat(t *testing.T) {
	got := New(testAssocStruct{})

	want := &Factory[testAssocStruct]{
		dataType:       reflect.TypeOf(testAssocStruct{}),
		storageName:    "test_assoc_structs",
		ignoreFields:   []string{},
		index:          1,
		isSetZeroValue: true,
	}

	if err := checkFactory(got, want); err != nil {
		t.Fatal(err.Error())
	}
}

func new_PassNonStruct(t *testing.T) {
	got := New(1)

	want := &Factory[int]{
		err: errInvalidType,
	}

	if err := checkFactory(got, want); err != nil {
		t.Fatal(err.Error())
	}
}

func new_WrongTagKey(t *testing.T) {
	type testStructWithWrongTag struct {
		ID int `gofacto:"foreignKey,struct:testStructWithID,field:ForeignValue,wrongName:test_struct_with_ids"`
	}
	got := New(testStructWithWrongTag{})

	want := &Factory[testStructWithWrongTag]{
		err: errTagFormat,
	}

	if err := checkFactory(got, want); err != nil {
		t.Fatal(err.Error())
	}
}

func new_WrongTagForeignKey(t *testing.T) {
	type testStructWithWrongTagColon struct {
		ID int `gofacto:"wrongforeignKey,struct:testStructWithID,field:ForeignValue,table:test_struct_with_ids"`
	}
	got := New(testStructWithWrongTagColon{})

	want := &Factory[testStructWithWrongTagColon]{
		err: errTagFormat,
	}

	if err := checkFactory(got, want); err != nil {
		t.Fatal(err.Error())
	}
}

func new_EmptyTag(t *testing.T) {
	type testStructWithWrongTagColon struct {
		ID int `gofacto:";"`
	}
	got := New(testStructWithWrongTagColon{})

	want := &Factory[testStructWithWrongTagColon]{
		err: errTagFormat,
	}

	if err := checkFactory(got, want); err != nil {
		t.Fatal(err.Error())
	}
}

func new_WrongTagFormat(t *testing.T) {
	type testStructWithWrongTagColon struct {
		ID int `gofacto:"aa;bb"`
	}
	got := New(testStructWithWrongTagColon{})

	want := &Factory[testStructWithWrongTagColon]{
		err: errTagFormat,
	}

	if err := checkFactory(got, want); err != nil {
		t.Fatal(err.Error())
	}
}

func TestBuild(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when pass buildPrint with all fields, all fields set by blueprint":                  build_BluePrintAllFields,
		"when pass buildPrint with some fields, other fields set by gofaco":                  build_BluePrintSomeFields,
		"when pass buildPrint without setting zero values, other fileds remain zero value":   build_BluePrintNotSetZeroValues,
		"when not pass buildPrint, all fields set by gofacto":                                build_NoBluePrint,
		"when not pass buildPrint without setting zero values, all fields remain zero value": build_NoBluePrintNotSetZeroValues,
		"when setting ignore fields, ignore fields should be zero value":                     build_IgnoreFields,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func build_BluePrintAllFields(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		b := true
		f := 1.1
		c := customType2
		return testStruct{
			Int:            i * 2,
			PtrInt:         &i,
			Int8:           int8(i),
			Int32:          int32(i),
			Int64:          int64(i),
			Uint:           uint(i),
			Uint8:          uint8(i),
			Uint32:         uint32(i),
			Uint64:         uint64(i),
			CustomType:     customType1,
			PtrCustomType:  &c,
			Str:            str,
			PtrStr:         &str,
			Bool:           b,
			PtrBool:        &b,
			Time:           now,
			PtrTime:        &now,
			Float:          f,
			PtrFloat:       &f,
			Interface:      str,
			Struct:         subStruct{ID: i, Name: str},
			PtrStruct:      &subStruct{ID: i, Name: str},
			Slice:          []int{i, i + 1, i + 2},
			PtrSlice:       []*int{&i, &i, &i},
			SliceStruct:    []subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
			SlicePtrStruct: []*subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
			privateField:   str,
		}
	}
	f := New(testStruct{}).WithBlueprint(blueprint)

	tests := []struct {
		desc string
		want func() testStruct
	}{
		{
			desc: "first build",
			want: func() testStruct {
				i1, i2, i3 := 1, 2, 3
				str := "test1"
				b := true
				f := 1.1
				c := customType2

				return testStruct{
					Int:            i2,
					PtrInt:         &i1,
					Int8:           int8(i1),
					Int32:          int32(i1),
					Int64:          int64(i1),
					Uint:           uint(i1),
					Uint8:          uint8(i1),
					Uint32:         uint32(i1),
					Uint64:         uint64(i1),
					CustomType:     customType1,
					PtrCustomType:  &c,
					Str:            str,
					PtrStr:         &str,
					Bool:           b,
					PtrBool:        &b,
					Time:           now,
					PtrTime:        &now,
					Float:          f,
					PtrFloat:       &f,
					Interface:      str,
					Struct:         subStruct{ID: i1, Name: str},
					PtrStruct:      &subStruct{ID: i1, Name: str},
					Slice:          []int{i1, i2, i3},
					PtrSlice:       []*int{&i1, &i1, &i1},
					SliceStruct:    []subStruct{{ID: i1, Name: str}, {ID: i2, Name: str}},
					SlicePtrStruct: []*subStruct{{ID: i1, Name: str}, {ID: i2, Name: str}},
				}
			},
		},
		{
			desc: "second build",
			want: func() testStruct {
				i2, i3, i4 := 2, 3, 4
				str := "test2"
				b := true
				f := 1.1
				c := customType2

				return testStruct{
					Int:            i4,
					PtrInt:         &i2,
					Int8:           int8(i2),
					Int32:          int32(i2),
					Int64:          int64(i2),
					Uint:           uint(i2),
					Uint8:          uint8(i2),
					Uint32:         uint32(i2),
					Uint64:         uint64(i2),
					CustomType:     customType1,
					PtrCustomType:  &c,
					Str:            str,
					PtrStr:         &str,
					Bool:           true,
					PtrBool:        &b,
					Time:           now,
					PtrTime:        &now,
					Float:          f,
					PtrFloat:       &f,
					Interface:      str,
					Struct:         subStruct{ID: i2, Name: str},
					PtrStruct:      &subStruct{ID: i2, Name: str},
					Slice:          []int{i2, i3, i4},
					PtrSlice:       []*int{&i2, &i2, &i2},
					SliceStruct:    []subStruct{{ID: i2, Name: str}, {ID: i3, Name: str}},
					SlicePtrStruct: []*subStruct{{ID: i2, Name: str}, {ID: i3, Name: str}},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want()); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func build_BluePrintSomeFields(t *testing.T) {
	blueprint := func(i int) testStruct {
		b := true
		return testStruct{
			Int:     i * 2,
			PtrInt:  &i,
			Bool:    b,
			PtrBool: &b,
		}
	}
	f := New(testStruct{}).WithBlueprint(blueprint)

	tests := []struct {
		desc string
		want func() testStruct
	}{
		{
			desc: "first build",
			want: func() testStruct {
				i1, i2 := 1, 2
				b := true

				return testStruct{
					Int:     i2,
					PtrInt:  &i1,
					Bool:    true,
					PtrBool: &b,
				}
			},
		},
		{
			desc: "second build",
			want: func() testStruct {
				i2, i4 := 2, 4
				b := true

				return testStruct{
					Int:     i4,
					PtrInt:  &i2,
					Bool:    true,
					PtrBool: &b,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want(), testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Bool", "PtrBool")...); err != nil {
				t.Fatal(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Bool", "PtrBool")...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func build_BluePrintNotSetZeroValues(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			Int:    i * 2,
			PtrInt: &i,
			Str:    str,
			PtrStr: &str,
		}
	}
	f := New(testStruct{}).WithBlueprint(blueprint).WithIsSetZeroValue(false)

	tests := []struct {
		desc string
		want func() testStruct
	}{
		{
			desc: "first build",
			want: func() testStruct {
				i1, i2 := 1, 2
				str := "test1"

				return testStruct{
					Int:    i2,
					PtrInt: &i1,
					Str:    str,
					PtrStr: &str,
				}
			},
		},
		{
			desc: "second build",
			want: func() testStruct {
				i1, i2 := 1, 2
				str := "test1"

				return testStruct{
					Int:    i2,
					PtrInt: &i1,
					Str:    str,
					PtrStr: &str,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want()); err != nil {
				fmt.Println("got", got.Int)
				fmt.Println("want", tt.want().Int)
				t.Fatal(err.Error())
			}
		})
	}
}

func build_NoBluePrint(t *testing.T) {
	f := New(testStruct{})

	tests := []struct {
		desc string
		want func() testStruct
	}{
		{
			desc: "first build",
		},
		{
			desc: "second build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.IsNotZeroVal(got, "CustomType", "PtrCustomType", "Interface", "privateField"); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func build_NoBluePrintNotSetZeroValues(t *testing.T) {
	f := New(testStruct{}).WithIsSetZeroValue(false)

	tests := []struct {
		desc string
		want testStruct
	}{
		{
			desc: "first build",
			want: testStruct{},
		},
		{
			desc: "second build",
			want: testStruct{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func build_IgnoreFields(t *testing.T) {
	type testStruct1 struct {
		Int            int     `gofacto:"omit"`
		PtrInt         *int    `gofacto:"omit"`
		Str            string  `gofacto:"omit"`
		PtrStr         *string `gofacto:"omit"`
		Bool           bool
		PtrBool        *bool
		Time           time.Time
		PtrTime        *time.Time
		Float          float64
		PtrFloat       *float64
		Interface      interface{}
		Struct         subStruct
		PtrStruct      *subStruct
		Slice          []int
		PtrSlice       []*int
		SliceStruct    []subStruct
		SlicePtrStruct []*subStruct
	}

	f := New(testStruct1{})

	tests := []struct {
		desc              string
		wantZeroFields    []string
		wantNonZeroFields []string
	}{
		{
			desc:              "first build",
			wantZeroFields:    []string{"Int", "PtrInt", "Str", "PtrStr", "Interface", "privateField"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Str", "PtrStr", "Interface"),
		},
		{
			desc:              "second build",
			wantZeroFields:    []string{"Int", "PtrInt", "Str", "PtrStr", "Interface", "privateField"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Str", "PtrStr", "Interface"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.IsZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Fatal(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func TestBuildList(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when pass buildList with all fields, all fields set by blueprint":                  buildList_BluePrintAllFields,
		"when pass buildList with some fields, other fields set by gofaco":                  buildList_BluePrintSomeFields,
		"when pass buildList without setting zero values, other fileds remain zero value":   buildList_BluePrintNotSetZeroValues,
		"when not pass buildList, all fields set by gofacto":                                buildList_NoBluePrint,
		"when not pass buildList without setting zero values, all fields remain zero value": buildList_NoBluePrintNotSetZeroValues,
		"when setting ignore fields, ignore fields should be zero value":                    buildList_IgnoreFields,
		"when pass negative number, error should be returned":                               buildlist_PassNegativeNumber,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func buildList_BluePrintAllFields(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		b := true
		f := 1.1
		c := customType2
		return testStruct{
			Int:            i * 2,
			PtrInt:         &i,
			Int8:           int8(i),
			Int32:          int32(i),
			Int64:          int64(i),
			Uint:           uint(i),
			Uint8:          uint8(i),
			Uint32:         uint32(i),
			Uint64:         uint64(i),
			CustomType:     c,
			PtrCustomType:  &c,
			Str:            str,
			PtrStr:         &str,
			Bool:           b,
			PtrBool:        &b,
			Time:           now,
			PtrTime:        &now,
			Float:          f,
			PtrFloat:       &f,
			Interface:      str,
			Struct:         subStruct{ID: i, Name: str},
			PtrStruct:      &subStruct{ID: i, Name: str},
			Slice:          []int{i, i + 1, i + 2},
			PtrSlice:       []*int{&i, &i, &i},
			SliceStruct:    []subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
			SlicePtrStruct: []*subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
		}
	}
	f := New(testStruct{}).WithBlueprint(blueprint)

	tests := []struct {
		desc string
		want func() []testStruct
	}{
		{
			desc: "first build",
			want: func() []testStruct {
				i1, i2, i3, i4 := 1, 2, 3, 4
				str1, str2 := "test1", "test2"
				b := true
				f := 1.1
				c := customType2

				return []testStruct{
					{
						Int:            i2,
						PtrInt:         &i1,
						Int8:           int8(i1),
						Int32:          int32(i1),
						Int64:          int64(i1),
						Uint:           uint(i1),
						Uint8:          uint8(i1),
						Uint32:         uint32(i1),
						Uint64:         uint64(i1),
						CustomType:     c,
						PtrCustomType:  &c,
						Str:            str1,
						PtrStr:         &str1,
						Bool:           b,
						PtrBool:        &b,
						Time:           now,
						PtrTime:        &now,
						Float:          f,
						PtrFloat:       &f,
						Interface:      str1,
						Struct:         subStruct{ID: i1, Name: str1},
						PtrStruct:      &subStruct{ID: i1, Name: str1},
						Slice:          []int{i1, i2, i3},
						PtrSlice:       []*int{&i1, &i1, &i1},
						SliceStruct:    []subStruct{{ID: i1, Name: str1}, {ID: i2, Name: str1}},
						SlicePtrStruct: []*subStruct{{ID: i1, Name: str1}, {ID: i2, Name: str1}},
					},
					{
						Int:            i4,
						PtrInt:         &i2,
						Int8:           int8(i2),
						Int32:          int32(i2),
						Int64:          int64(i2),
						Uint:           uint(i2),
						Uint8:          uint8(i2),
						Uint32:         uint32(i2),
						Uint64:         uint64(i2),
						CustomType:     c,
						PtrCustomType:  &c,
						Str:            str2,
						PtrStr:         &str2,
						Bool:           b,
						PtrBool:        &b,
						Time:           now,
						PtrTime:        &now,
						Float:          f,
						PtrFloat:       &f,
						Interface:      str2,
						Struct:         subStruct{ID: i2, Name: str2},
						PtrStruct:      &subStruct{ID: i2, Name: str2},
						Slice:          []int{i2, i3, i4},
						PtrSlice:       []*int{&i2, &i2, &i2},
						SliceStruct:    []subStruct{{ID: i2, Name: str2}, {ID: i3, Name: str2}},
						SlicePtrStruct: []*subStruct{{ID: i2, Name: str2}, {ID: i3, Name: str2}},
					},
				}
			},
		},
		{
			desc: "second build",
			want: func() []testStruct {
				i3, i4, i5, i6, i8 := 3, 4, 5, 6, 8
				str3, str4 := "test3", "test4"
				b := true
				f := 1.1
				c := customType2

				return []testStruct{
					{
						Int:            i6,
						PtrInt:         &i3,
						Int8:           int8(i3),
						Int32:          int32(i3),
						Int64:          int64(i3),
						Uint:           uint(i3),
						Uint8:          uint8(i3),
						Uint32:         uint32(i3),
						Uint64:         uint64(i3),
						CustomType:     c,
						PtrCustomType:  &c,
						Str:            str3,
						PtrStr:         &str3,
						Bool:           b,
						PtrBool:        &b,
						Time:           now,
						PtrTime:        &now,
						Float:          f,
						PtrFloat:       &f,
						Interface:      str3,
						Struct:         subStruct{ID: i3, Name: str3},
						PtrStruct:      &subStruct{ID: i3, Name: str3},
						Slice:          []int{i3, i4, i5},
						PtrSlice:       []*int{&i3, &i3, &i3},
						SliceStruct:    []subStruct{{ID: i3, Name: str3}, {ID: i4, Name: str3}},
						SlicePtrStruct: []*subStruct{{ID: i3, Name: str3}, {ID: i4, Name: str3}},
					},
					{
						Int:            i8,
						PtrInt:         &i4,
						Int8:           int8(i4),
						Int32:          int32(i4),
						Int64:          int64(i4),
						Uint:           uint(i4),
						Uint8:          uint8(i4),
						Uint32:         uint32(i4),
						Uint64:         uint64(i4),
						CustomType:     c,
						PtrCustomType:  &c,
						Str:            str4,
						PtrStr:         &str4,
						Bool:           b,
						PtrBool:        &b,
						Time:           now,
						PtrTime:        &now,
						Float:          f,
						PtrFloat:       &f,
						Interface:      str4,
						Struct:         subStruct{ID: i4, Name: str4},
						PtrStruct:      &subStruct{ID: i4, Name: str4},
						Slice:          []int{i4, i5, i6},
						PtrSlice:       []*int{&i4, &i4, &i4},
						SliceStruct:    []subStruct{{ID: i4, Name: str4}, {ID: i5, Name: str4}},
						SlicePtrStruct: []*subStruct{{ID: i4, Name: str4}, {ID: i5, Name: str4}},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want()); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func buildList_BluePrintSomeFields(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			Int:            i * 2,
			PtrStruct:      &subStruct{ID: i, Name: str},
			SlicePtrStruct: []*subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
		}
	}
	f := New(testStruct{}).WithBlueprint(blueprint)

	tests := []struct {
		desc string
		want func() []testStruct
	}{
		{
			desc: "first build",
			want: func() []testStruct {
				i1, i2, i3, i4 := 1, 2, 3, 4
				str1, str2 := "test1", "test2"

				return []testStruct{
					{
						Int:            i2,
						PtrStruct:      &subStruct{ID: i1, Name: str1},
						SlicePtrStruct: []*subStruct{{ID: i1, Name: str1}, {ID: i2, Name: str1}},
					},
					{
						Int:            i4,
						PtrStruct:      &subStruct{ID: i2, Name: str2},
						SlicePtrStruct: []*subStruct{{ID: i2, Name: str2}, {ID: i3, Name: str2}},
					},
				}
			},
		},
		{
			desc: "second build",
			want: func() []testStruct {
				i3, i4, i5, i6, i8 := 3, 4, 5, 6, 8
				str3, str4 := "test3", "test4"

				return []testStruct{
					{
						Int:            i6,
						PtrStruct:      &subStruct{ID: i3, Name: str3},
						SlicePtrStruct: []*subStruct{{ID: i3, Name: str3}, {ID: i4, Name: str3}},
					},
					{
						Int:            i8,
						PtrStruct:      &subStruct{ID: i4, Name: str4},
						SlicePtrStruct: []*subStruct{{ID: i4, Name: str4}, {ID: i5, Name: str4}},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want(), testutils.FilterFields(testStruct{}, "Int", "PtrStruct", "SlicePtrStruct")...); err != nil {
				t.Fatal(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, testutils.FilterFields(testStruct{}, "Int", "PtrStruct", "SlicePtrStruct")...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func buildList_BluePrintNotSetZeroValues(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			Int:            i * 2,
			PtrStruct:      &subStruct{ID: i, Name: str},
			SlicePtrStruct: []*subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
		}
	}
	f := New(testStruct{}).WithBlueprint(blueprint).WithIsSetZeroValue(false)

	tests := []struct {
		desc string
		want func() []testStruct
	}{
		{
			desc: "first build",
			want: func() []testStruct {
				i1, i2 := 1, 2
				str1 := "test1"

				return []testStruct{
					{
						Int:            i2,
						PtrStruct:      &subStruct{ID: i1, Name: str1},
						SlicePtrStruct: []*subStruct{{ID: i1, Name: str1}, {ID: i2, Name: str1}},
					},
					{
						Int:            i2,
						PtrStruct:      &subStruct{ID: i1, Name: str1},
						SlicePtrStruct: []*subStruct{{ID: i1, Name: str1}, {ID: i2, Name: str1}},
					},
				}
			},
		},
		{
			desc: "second build",
			want: func() []testStruct {
				i1, i2 := 1, 2
				str1 := "test1"

				return []testStruct{
					{
						Int:            i2,
						PtrStruct:      &subStruct{ID: i1, Name: str1},
						SlicePtrStruct: []*subStruct{{ID: i1, Name: str1}, {ID: i2, Name: str1}},
					},
					{
						Int:            i2,
						PtrStruct:      &subStruct{ID: i1, Name: str1},
						SlicePtrStruct: []*subStruct{{ID: i1, Name: str1}, {ID: i2, Name: str1}},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want()); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func buildList_NoBluePrint(t *testing.T) {
	f := New(testStruct{})

	tests := []struct {
		desc string
		want []testStruct
	}{
		{
			desc: "first build",
			want: []testStruct{{}, {}},
		},
		{
			desc: "second build",
			want: []testStruct{{}, {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.IsNotZeroVal(got, "CustomType", "PtrCustomType", "Interface", "privateField"); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func buildList_NoBluePrintNotSetZeroValues(t *testing.T) {
	f := New(testStruct{}).WithIsSetZeroValue(false)

	tests := []struct {
		desc string
		want []testStruct
	}{
		{
			desc: "first build",
			want: []testStruct{{}, {}},
		},
		{
			desc: "second build",
			want: []testStruct{{}, {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func buildList_IgnoreFields(t *testing.T) {
	type testStruct1 struct {
		Int            int     `gofacto:"omit"`
		PtrInt         *int    `gofacto:"omit"`
		Str            string  `gofacto:"omit"`
		PtrStr         *string `gofacto:"omit"`
		Bool           bool
		PtrBool        *bool
		Time           time.Time
		PtrTime        *time.Time
		Float          float64
		PtrFloat       *float64
		Interface      interface{}
		Struct         subStruct
		PtrStruct      *subStruct
		Slice          []int
		PtrSlice       []*int
		SliceStruct    []subStruct
		SlicePtrStruct []*subStruct
	}
	f := New(testStruct1{})

	tests := []struct {
		desc              string
		wantZeroFields    []string
		wantNonZeroFields []string
	}{
		{
			desc:              "first build",
			wantZeroFields:    []string{"Int", "PtrInt", "Str", "PtrStr", "Interface", "privateField"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Str", "PtrStr", "Interface", "privateField"),
		},
		{
			desc:              "second build",
			wantZeroFields:    []string{"Int", "PtrInt", "Str", "PtrStr", "Interface", "privateField"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Str", "PtrStr", "Interface", "privateField"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.IsZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Fatal(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func buildlist_PassNegativeNumber(t *testing.T) {
	f := New(testStruct{})

	want := []testStruct{}
	wantErr := errBuildListNGreaterThanZero

	got, err := f.BuildList(mockCTX, -1).Get()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if testutils.CompareVal(got, want) != nil {
		t.Fatalf("got: %v, want: %v", got, want)
	}
}

func TestInsert(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when insert on builder with db, insert successfully":              insert_OnBuilderWithDB,
		"when insert on builder without db, error should be returned":      insert_OnBuilderWithoutDB,
		"when insert on builder with error, error should be returned":      insert_OnBuilderWithErr,
		"when insert on builder list with db, insert successfully":         insert_OnBuilderListWithDB,
		"when insert on builder list without db, error should be returned": insert_OnBuilderListWithoutDB,
		"when insert on builder list with error, error should be returned": insert_OnBuilderListWithErr,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func insert_OnBuilderWithDB(t *testing.T) {
	f := New(testStructWithID{}).WithDB(&mockDB{})

	val, err := f.Build(mockCTX).Insert()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if err := testutils.IsNotZeroVal(val); err != nil {
		t.Fatal(err.Error())
	}
}

func insert_OnBuilderWithoutDB(t *testing.T) {
	f := New(testStruct{})

	want := testStruct{}
	wantErr := errDBIsNotProvided

	vals, err := f.Build(mockCTX).Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if testutils.CompareVal(vals, want) != nil {
		t.Fatalf("got: %v, want: %v", vals, want)
	}
}

func insert_OnBuilderWithErr(t *testing.T) {
	f := New(testStructWithID{}).WithDB(&mockDB{})

	want := testStructWithID{}
	wantErr := errFieldNotFound

	val, err := f.Build(mockCTX).SetZero("incorrect field").Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if testutils.CompareVal(val, want) != nil {
		t.Fatalf("got: %v, want: %v", val, want)
	}
}

func insert_OnBuilderListWithDB(t *testing.T) {
	f := New(testStructWithID{}).WithDB(&mockDB{})

	val, err := f.BuildList(mockCTX, 2).Insert()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if err := testutils.IsNotZeroVal(val); err != nil {
		t.Fatal(err.Error())
	}
}

func insert_OnBuilderListWithoutDB(t *testing.T) {
	f := New(testStruct{})

	want := []testStruct{}
	wantErr := errDBIsNotProvided

	vals, err := f.BuildList(mockCTX, 2).Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if testutils.CompareVal(vals, want) != nil {
		t.Fatalf("got: %v, want: %v", vals, want)
	}
}

func insert_OnBuilderListWithErr(t *testing.T) {
	f := New(testStructWithID{}).WithDB(&mockDB{})

	want := []testStructWithID{}
	wantErr := errFieldNotFound

	val, err := f.BuildList(mockCTX, 2).SetZero(1, "incorrect field").Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if testutils.CompareVal(val, want) != nil {
		t.Fatalf("got: %v, want: %v", val, want)
	}
}

func TestOverwrite(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when overwrite on builder, overwrite one value":           overwrite_OnBuilder,
		"when overwrite on builder list, overwrite one value":      overwrite_OnBuilderList,
		"when overwrites on builder list, overwrite target values": overwrites_OnBuilderList,
		"when overwrite, already has error, return error":          overwrite_AlreadyHasError,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func overwrite_OnBuilder(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			Int:            i * 2,
			PtrStruct:      &subStruct{ID: i, Name: str},
			SlicePtrStruct: []*subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
		}
	}
	f := New(testStruct{}).WithBlueprint(blueprint)

	tests := []struct {
		desc string
		ow   testStruct
		want testStruct
	}{
		{
			desc: "overwrite with value",
			ow:   testStruct{Int: 10, PtrStruct: &subStruct{ID: 10, Name: "test10"}},
			want: testStruct{
				Int:            10,
				PtrStruct:      &subStruct{ID: 10, Name: "test10"},
				SlicePtrStruct: []*subStruct{{ID: 1, Name: "test1"}, {ID: 2, Name: "test1"}},
			},
		},
		{
			desc: "overwrite without value",
			ow:   testStruct{},
			want: testStruct{
				Int:            4,
				PtrStruct:      &subStruct{ID: 2, Name: "test2"},
				SlicePtrStruct: []*subStruct{{ID: 2, Name: "test2"}, {ID: 3, Name: "test2"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).Overwrite(tt.ow).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want, testutils.FilterFields(testStruct{}, "Int", "PtrStruct", "SlicePtrStruct")...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func overwrite_OnBuilderList(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			Int:            i * 2,
			PtrStruct:      &subStruct{ID: i, Name: str},
			SlicePtrStruct: []*subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
		}
	}
	f := New(testStruct{}).WithBlueprint(blueprint)

	tests := []struct {
		desc string
		ow   testStruct
		want []testStruct
	}{
		{
			desc: "overwrite with value",
			ow:   testStruct{Int: 10, PtrStruct: &subStruct{ID: 10, Name: "test10"}},
			want: []testStruct{
				{
					Int:            10,
					PtrStruct:      &subStruct{ID: 10, Name: "test10"},
					SlicePtrStruct: []*subStruct{{ID: 1, Name: "test1"}, {ID: 2, Name: "test1"}},
				},
				{
					Int:            10,
					PtrStruct:      &subStruct{ID: 10, Name: "test10"},
					SlicePtrStruct: []*subStruct{{ID: 2, Name: "test2"}, {ID: 3, Name: "test2"}},
				},
			},
		},
		{
			desc: "overwrite without value",
			ow:   testStruct{},
			want: []testStruct{
				{
					Int:            6,
					PtrStruct:      &subStruct{ID: 3, Name: "test3"},
					SlicePtrStruct: []*subStruct{{ID: 3, Name: "test3"}, {ID: 4, Name: "test3"}},
				},
				{
					Int:            8,
					PtrStruct:      &subStruct{ID: 4, Name: "test4"},
					SlicePtrStruct: []*subStruct{{ID: 4, Name: "test4"}, {ID: 5, Name: "test4"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Overwrite(tt.ow).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want, testutils.FilterFields(testStruct{}, "Int", "PtrStruct", "SlicePtrStruct")...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func overwrites_OnBuilderList(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			Int:            i * 2,
			PtrStruct:      &subStruct{ID: i, Name: str},
			SlicePtrStruct: []*subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
		}
	}
	f := New(testStruct{}).WithBlueprint(blueprint)

	tests := []struct {
		desc string
		ow   []testStruct
		want []testStruct
	}{
		{
			desc: "overwrite with same length",
			ow: []testStruct{
				{Int: 10, PtrStruct: &subStruct{ID: 10, Name: "test10"}},
				{Int: 20, PtrStruct: &subStruct{ID: 20, Name: "test20"}},
			},
			want: []testStruct{
				{
					Int:            10,
					PtrStruct:      &subStruct{ID: 10, Name: "test10"},
					SlicePtrStruct: []*subStruct{{ID: 1, Name: "test1"}, {ID: 2, Name: "test1"}},
				},
				{
					Int:            20,
					PtrStruct:      &subStruct{ID: 20, Name: "test20"},
					SlicePtrStruct: []*subStruct{{ID: 2, Name: "test2"}, {ID: 3, Name: "test2"}},
				},
			},
		},
		{
			desc: "overwrite with longer length",
			ow: []testStruct{
				{Int: 10, PtrStruct: &subStruct{ID: 10, Name: "test10"}},
				{Int: 20, PtrStruct: &subStruct{ID: 20, Name: "test20"}},
				{Int: 30, PtrStruct: &subStruct{ID: 30, Name: "test30"}},
				{Int: 40, PtrStruct: &subStruct{ID: 40, Name: "test40"}},
			},
			want: []testStruct{
				{
					Int:            10,
					PtrStruct:      &subStruct{ID: 10, Name: "test10"},
					SlicePtrStruct: []*subStruct{{ID: 3, Name: "test3"}, {ID: 4, Name: "test3"}},
				},
				{
					Int:            20,
					PtrStruct:      &subStruct{ID: 20, Name: "test20"},
					SlicePtrStruct: []*subStruct{{ID: 4, Name: "test4"}, {ID: 5, Name: "test4"}},
				},
			},
		},
		{
			desc: "overwrite with shorter length",
			ow: []testStruct{
				{Int: 10, PtrStruct: &subStruct{ID: 10, Name: "test10"}},
			},
			want: []testStruct{
				{
					Int:            10,
					PtrStruct:      &subStruct{ID: 10, Name: "test10"},
					SlicePtrStruct: []*subStruct{{ID: 5, Name: "test5"}, {ID: 6, Name: "test5"}},
				},
				{
					Int:            12,
					PtrStruct:      &subStruct{ID: 6, Name: "test6"},
					SlicePtrStruct: []*subStruct{{ID: 6, Name: "test6"}, {ID: 7, Name: "test6"}},
				},
			},
		},
		{
			desc: "overwrite without value",
			ow:   []testStruct{},
			want: []testStruct{
				{
					Int:            14,
					PtrStruct:      &subStruct{ID: 7, Name: "test7"},
					SlicePtrStruct: []*subStruct{{ID: 7, Name: "test7"}, {ID: 8, Name: "test7"}},
				},
				{
					Int:            16,
					PtrStruct:      &subStruct{ID: 8, Name: "test8"},
					SlicePtrStruct: []*subStruct{{ID: 8, Name: "test8"}, {ID: 9, Name: "test8"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Overwrites(tt.ow...).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want, testutils.FilterFields(testStruct{}, "Int", "PtrStruct", "SlicePtrStruct")...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func overwrite_AlreadyHasError(t *testing.T) {
	f := New(testStruct{})

	tests := []struct {
		desc    string
		gotRes  func() (interface{}, error)
		want    interface{}
		wantErr error
	}{
		{
			desc: "already has error when overwrite on builder",
			gotRes: func() (interface{}, error) {
				return f.Build(mockCTX).SetZero("incorrect field").Overwrite(testStruct{}).Get()
			},
			want:    testStruct{},
			wantErr: errFieldNotFound,
		},
		{
			desc: "already has error when overwrites on builder list",
			gotRes: func() (interface{}, error) {
				return f.BuildList(mockCTX, 2).SetZero(0, "incorrect field").Overwrites(testStruct{}).Get()
			},
			want:    []testStruct{},
			wantErr: errFieldNotFound,
		},
		{
			desc: "already has error when overwrite on builder list",
			gotRes: func() (interface{}, error) {
				return f.BuildList(mockCTX, 2).SetZero(0, "incorrect field").Overwrite(testStruct{}).Get()
			},
			want:    []testStruct{},
			wantErr: errFieldNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := tt.gotRes()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error should be %v", tt.wantErr)
			}

			if err := testutils.CompareVal(got, tt.want); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func TestWithTrait(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when withTrait on builder, overwrite one value":           withTrait_OnBuilder,
		"when withTrait on builder list, overwrite one value":      withTrait_OnBuilderList,
		"when multiple withTrait on builder, overwrite one value":  withTrait_OnBuilderMultiple,
		"when withTraits on builder list, overwrite target values": withTraits_OnBuilderList,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func withTrait_OnBuilder(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			PtrStr: &str,
			Time:   now,
			Slice:  []int{i, i + 1, i + 2},
		}
	}
	setTraiter := func(val *testStruct) {
		val.Slice = []int{1, 1, 1}
	}

	f := New(testStruct{}).WithBlueprint(blueprint).WithTrait("trait", setTraiter)

	tests := []struct {
		desc    string
		trait   string
		want    func() testStruct
		wantErr error
	}{
		{
			desc:  "set trait with correct value",
			trait: "trait",
			want: func() testStruct {
				str := "test1"
				return testStruct{
					PtrStr: &str,
					Time:   now,
					Slice:  []int{1, 1, 1},
				}
			},
			wantErr: nil,
		},
		{
			desc:    "set trait with incorrect value",
			trait:   "incorrect trait",
			want:    func() testStruct { return testStruct{} },
			wantErr: errWithTraitNameNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).SetTrait(tt.trait).Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want()); err != nil {
					t.Fatal(err.Error())
				}

				return
			}

			if err := testutils.CompareVal(got, tt.want(), testutils.FilterFields(testStruct{}, "PtrStr", "Time", "Slice")...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func withTrait_OnBuilderList(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			PtrStr: &str,
			Time:   now,
			Slice:  []int{i, i + 1, i + 2},
		}
	}
	setTraiter := func(val *testStruct) {
		val.Slice = []int{1, 1, 1}
	}

	f := New(testStruct{}).WithBlueprint(blueprint).WithTrait("trait", setTraiter)

	tests := []struct {
		desc    string
		tait    string
		want    func() []testStruct
		wantErr error
	}{
		{
			desc: "set trait with correct value",
			tait: "trait",
			want: func() []testStruct {
				str1, str2 := "test1", "test2"
				return []testStruct{
					{
						PtrStr: &str1,
						Time:   now,
						Slice:  []int{1, 1, 1},
					},
					{
						PtrStr: &str2,
						Time:   now,
						Slice:  []int{1, 1, 1},
					},
				}
			},
			wantErr: nil,
		},
		{
			desc:    "set trait with incorrect value",
			tait:    "incorrect trait",
			want:    func() []testStruct { return []testStruct{} },
			wantErr: errWithTraitNameNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).SetTrait(tt.tait).Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want()); err != nil {
					t.Fatal(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want(), testutils.FilterFields(testStruct{}, "PtrStr", "Time", "Slice")...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func withTrait_OnBuilderMultiple(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			PtrStr: &str,
			Time:   now,
			Slice:  []int{i, i + 1, i + 2},
		}
	}
	setTraiter1 := func(val *testStruct) {
		val.Slice = []int{1, 1, 1}
	}
	setTraiter2 := func(val *testStruct) {
		val.Slice = []int{2, 2, 2}
	}

	f := New(testStruct{}).WithBlueprint(blueprint).WithTrait("trait1", setTraiter1).WithTrait("trait2", setTraiter2)

	tests := []struct {
		desc    string
		taits   []string
		want    func() testStruct
		wantErr error
	}{
		{
			desc:  "set two traits with correct value",
			taits: []string{"trait1", "trait2"},
			want: func() testStruct {
				str := "test1"
				return testStruct{
					PtrStr: &str,
					Time:   now,
					Slice:  []int{2, 2, 2},
				}
			},
			wantErr: nil,
		},
		{
			desc:    "set one trait with incorrect value",
			taits:   []string{"trait1", "incorrect trait"},
			want:    func() testStruct { return testStruct{} },
			wantErr: errWithTraitNameNotFound,
		},
		{
			desc:    "set two traits with incorrect value",
			taits:   []string{"incorrect trait1", "incorrect trait2"},
			want:    func() testStruct { return testStruct{} },
			wantErr: errWithTraitNameNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).SetTrait(tt.taits[0]).SetTrait(tt.taits[1]).Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want()); err != nil {
					t.Fatal(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want(), testutils.FilterFields(testStruct{}, "PtrStr", "Time", "Slice")...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func withTraits_OnBuilderList(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			PtrStr: &str,
			Time:   now,
			Slice:  []int{i, i + 1, i + 2},
		}
	}
	setTraiter := func(val *testStruct) {
		val.Slice = []int{1, 1, 1}
	}

	f := New(testStruct{}).WithBlueprint(blueprint).
		WithTrait("trait", setTraiter)

	tests := []struct {
		desc    string
		taits   []string
		want    func() []testStruct
		wantErr error
	}{
		{
			desc:  "set trait with same length",
			taits: []string{"trait", "trait"},
			want: func() []testStruct {
				str1, str2 := "test1", "test2"
				return []testStruct{
					{
						PtrStr: &str1,
						Time:   now,
						Slice:  []int{1, 1, 1},
					},
					{
						PtrStr: &str2,
						Time:   now,
						Slice:  []int{1, 1, 1},
					},
				}
			},
			wantErr: nil,
		},
		{
			desc:  "set trait with longer length",
			taits: []string{"trait", "trait", "trait", "trait"},
			want: func() []testStruct {
				str3, str4 := "test3", "test4"
				return []testStruct{
					{
						PtrStr: &str3,
						Time:   now,
						Slice:  []int{1, 1, 1},
					},
					{
						PtrStr: &str4,
						Time:   now,
						Slice:  []int{1, 1, 1},
					},
				}
			},
			wantErr: nil,
		},
		{
			desc:  "set trait with shorter length",
			taits: []string{"trait"},
			want: func() []testStruct {
				str5, str6 := "test5", "test6"
				return []testStruct{
					{
						PtrStr: &str5,
						Time:   now,
						Slice:  []int{1, 1, 1},
					},
					{
						PtrStr: &str6,
						Time:   now,
						Slice:  []int{6, 7, 8},
					},
				}
			},
			wantErr: nil,
		},
		{
			desc:    "set trait with incorrect value",
			taits:   []string{"incorrect trait"},
			want:    func() []testStruct { return []testStruct{} },
			wantErr: errWithTraitNameNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).SetTraits(tt.taits...).Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want()); err != nil {
					t.Fatal(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want(), testutils.FilterFields(testStruct{}, "PtrStr", "Time", "Slice")...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func TestSetZero(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when setZero on builder with blueprint":         setZero_OnBuilderWithBluePrint,
		"when setZero on builder without blueprint":      setZero_OnBuilderWithoutBluePrint,
		"when setZero on builder list with blueprint":    setZero_OnBuilderListWithBluePrint,
		"when setZero on builder list without blueprint": setZero_OnBuilderListWithoutBluePrint,
		"when many setZero on builder":                   setZero_OnBuilderMany,
		"when many setZero on builder list":              setZero_OnBuilderListMany,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func setZero_OnBuilderWithBluePrint(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		f := 1.1
		c := customType2
		return testStruct{
			Int:            i * 2,
			PtrInt:         &i,
			CustomType:     customType1,
			PtrCustomType:  &c,
			Time:           now,
			PtrTime:        &now,
			Float:          f,
			PtrFloat:       &f,
			Interface:      str,
			Struct:         subStruct{ID: i, Name: str},
			PtrStruct:      &subStruct{ID: i, Name: str},
			Slice:          []int{i, i + 1, i + 2},
			PtrSlice:       []*int{&i, &i, &i},
			SliceStruct:    []subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
			SlicePtrStruct: []*subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
		}
	}
	f := New(testStruct{}).WithBlueprint(blueprint)

	tests := []struct {
		desc              string
		setZeroFields     []string
		wantZeroFields    []string
		wantNonZeroFields []string
		wantErr           error
		want              testStruct
	}{
		{
			desc:              "set many zero values",
			setZeroFields:     []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "CustomType", "PtrCustomType"},
			wantZeroFields:    []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "privateField", "CustomType", "PtrCustomType"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "CustomType", "PtrCustomType"),
			wantErr:           nil,
		},
		{
			desc:              "set one zero value",
			setZeroFields:     []string{"Int"},
			wantZeroFields:    []string{"Int", "privateField"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int"),
			wantErr:           nil,
		},
		{
			desc:              "set no zero value",
			setZeroFields:     []string{},
			wantZeroFields:    []string{"privateField"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}),
			wantErr:           nil,
		},
		{
			desc:          "set incorrect field",
			setZeroFields: []string{"incorrect field"},
			want:          testStruct{},
			wantErr:       errFieldNotFound,
		},
		{
			desc:          "set private field",
			setZeroFields: []string{"privateField"},
			want:          testStruct{},
			wantErr:       errFieldCantSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).SetZero(tt.setZeroFields...).Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want); err != nil {
					t.Fatal(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.IsZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Fatal(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func setZero_OnBuilderWithoutBluePrint(t *testing.T) {
	f := New(testStruct{})

	tests := []struct {
		desc              string
		setZeroFields     []string
		wantZeroFields    []string
		wantNonZeroFields []string
		want              testStruct
		wantErr           error
	}{
		{
			desc:              "set many zero values",
			setZeroFields:     []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
			wantZeroFields:    []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "privateField", "CustomType", "PtrCustomType"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"),
			wantErr:           nil,
		},
		{
			desc:          "set one zero value",
			setZeroFields: []string{"Int"},
			// interface value will default set to nil
			wantZeroFields:    []string{"Int", "Interface", "privateField", "CustomType", "PtrCustomType"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int", "Interface"),
			wantErr:           nil,
		},
		{
			desc:          "set no zero value",
			setZeroFields: []string{},
			// interface value will default set to nil
			wantZeroFields:    []string{"Interface", "privateField", "CustomType", "PtrCustomType"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}),
			wantErr:           nil,
		},
		{
			desc:          "set incorrect field",
			setZeroFields: []string{"incorrect field"},
			want:          testStruct{},
			wantErr:       errFieldNotFound,
		},
		{
			desc:          "set private field",
			setZeroFields: []string{"privateField"},
			want:          testStruct{},
			wantErr:       errFieldCantSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).SetZero(tt.setZeroFields...).Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want); err != nil {
					t.Fatal(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.IsZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Fatal(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func setZero_OnBuilderListWithBluePrint(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		f := 1.1
		c := customType2
		return testStruct{
			Int:            i * 2,
			PtrInt:         &i,
			CustomType:     c,
			PtrCustomType:  &c,
			Time:           now,
			PtrTime:        &now,
			Float:          f,
			PtrFloat:       &f,
			Interface:      str,
			Struct:         subStruct{ID: i, Name: str},
			PtrStruct:      &subStruct{ID: i, Name: str},
			Slice:          []int{i, i + 1, i + 2},
			PtrSlice:       []*int{&i, &i, &i},
			SliceStruct:    []subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
			SlicePtrStruct: []*subStruct{{ID: i, Name: str}, {ID: i + 1, Name: str}},
		}
	}
	f := New(testStruct{}).WithBlueprint(blueprint)

	tests := []struct {
		desc              string
		index             int
		setZeroFields     []string
		wantZeroFields    [][]string
		wantNonZeroFields [][]string
		want              []testStruct
		wantErr           error
	}{
		{
			desc:          "set zero values at valid index",
			index:         1,
			setZeroFields: []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "CustomType", "PtrCustomType"},
			wantZeroFields: [][]string{
				{"privateField"},
				{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "privateField", "CustomType", "PtrCustomType"},
			},
			wantNonZeroFields: [][]string{
				{"Int", "PtrInt", "Int8", "Int32", "Int64", "Uint", "Uint8", "Uint32", "Uint64", "Str", "PtrStr", "Bool", "PtrBool", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "CustomType", "PtrCustomType"},
				testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "CustomType", "PtrCustomType"),
			},
			wantErr: nil,
		},
		{
			desc:    "set zero values at negative index",
			index:   -1,
			want:    []testStruct{},
			wantErr: errIndexIsOutOfRange,
		},
		{
			desc:    "set zero values at invalid index",
			index:   5,
			want:    []testStruct{},
			wantErr: errIndexIsOutOfRange,
		},
		{
			desc:          "set incorrect field",
			index:         0,
			setZeroFields: []string{"incorrect field"},
			want:          []testStruct{},
			wantErr:       errFieldNotFound,
		},
		{
			desc:          "set private field",
			index:         0,
			setZeroFields: []string{"privateField"},
			want:          []testStruct{},
			wantErr:       errFieldCantSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).SetZero(tt.index, tt.setZeroFields...).Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want); err != nil {
					t.Fatal(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			for i, g := range got {
				if err := testutils.IsZeroVal(g, tt.wantNonZeroFields[i]...); err != nil {
					t.Fatal(err.Error())
				}

				if err := testutils.IsNotZeroVal(g, tt.wantZeroFields[i]...); err != nil {
					t.Fatal(err.Error())
				}
			}
		})
	}
}

func setZero_OnBuilderListWithoutBluePrint(t *testing.T) {
	f := New(testStruct{})

	tests := []struct {
		desc              string
		index             int
		setZeroFields     []string
		wantZeroFields    [][]string
		wantNonZeroFields [][]string
		want              []testStruct
		wantErr           error
	}{
		{
			desc:          "set zero values at valid index",
			index:         1,
			setZeroFields: []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
			wantZeroFields: [][]string{
				{"Interface", "privateField", "CustomType", "PtrCustomType"},
				{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Struct", "PtrStruct", "Interface", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "privateField", "CustomType", "PtrCustomType"},
			},
			wantNonZeroFields: [][]string{
				testutils.FilterFields(testStruct{}, "Interface"),
				testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Struct", "PtrStruct", "Interface", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"),
			},
			wantErr: nil,
		},
		{
			desc:    "set zero values at negative index",
			index:   -1,
			want:    []testStruct{},
			wantErr: errIndexIsOutOfRange,
		},
		{
			desc:    "set zero values at invalid index",
			index:   5,
			want:    []testStruct{},
			wantErr: errIndexIsOutOfRange,
		},
		{
			desc:          "set incorrect field",
			index:         0,
			setZeroFields: []string{"incorrect field"},
			want:          []testStruct{},
			wantErr:       errFieldNotFound,
		},
		{
			desc:          "set private field",
			index:         0,
			setZeroFields: []string{"privateField"},
			want:          []testStruct{},
			wantErr:       errFieldCantSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).SetZero(tt.index, tt.setZeroFields...).Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want); err != nil {
					t.Fatal(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			for i, g := range got {
				if err := testutils.IsZeroVal(g, tt.wantNonZeroFields[i]...); err != nil {
					t.Fatal(err.Error())
				}

				if err := testutils.IsNotZeroVal(g, tt.wantZeroFields[i]...); err != nil {
					t.Fatal(err.Error())
				}
			}
		})
	}
}

func setZero_OnBuilderMany(t *testing.T) {
	f := New(testStruct{})

	tests := []struct {
		desc              string
		setZeroFields     [][]string
		wantZeroFields    []string
		wantNonZeroFields []string
		want              testStruct
		wantErr           error
	}{
		{
			desc: "set two zero values",
			setZeroFields: [][]string{
				{"Int"},
				{"Slice"},
			},
			wantZeroFields:    []string{"Int", "Slice", "Interface", "privateField", "CustomType", "PtrCustomType"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int", "Slice", "Interface"),
			wantErr:           nil,
		},
		{
			desc: "set three zero values",
			setZeroFields: [][]string{
				{"Int"},
				{"Slice"},
				{"Struct"},
			},
			wantZeroFields:    []string{"Int", "Slice", "Struct", "Interface", "privateField", "CustomType", "PtrCustomType"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int", "Slice", "Struct", "Interface"),
			wantErr:           nil,
		},
		{
			desc:          "set incorrect field",
			setZeroFields: [][]string{{"incorrect field"}},
			want:          testStruct{},
			wantErr:       errFieldNotFound,
		},
		{
			desc:          "set private field",
			setZeroFields: [][]string{{"privateField"}},
			want:          testStruct{},
			wantErr:       errFieldCantSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			preF := f.Build(mockCTX)
			for _, fields := range tt.setZeroFields {
				preF = preF.SetZero(fields...)
			}

			got, err := preF.Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want); err != nil {
					t.Fatal(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.IsZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Fatal(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Fatal(err.Error())
			}
		})
	}
}

func setZero_OnBuilderListMany(t *testing.T) {
	f := New(testStruct{})

	tests := []struct {
		desc                 string
		buildIndex           int
		setZeroFieldsByIndex map[int][]string
		wantZeroFields       [][]string
		wantNonZeroFields    [][]string
		want                 []testStruct
		wantErr              error
	}{
		{
			desc:       "set zero values at valid index",
			buildIndex: 3,
			setZeroFieldsByIndex: map[int][]string{
				0: {"Int"},
				1: {"PtrSlice", "SlicePtrStruct"},
			},
			wantZeroFields: [][]string{
				{"Int", "Interface", "privateField", "privateField", "CustomType", "PtrCustomType"},
				{"PtrSlice", "SlicePtrStruct", "Interface", "privateField", "privateField", "CustomType", "PtrCustomType"},
				{"Interface", "privateField", "privateField", "CustomType", "PtrCustomType"},
			},
			wantNonZeroFields: [][]string{
				testutils.FilterFields(testStruct{}, "Int", "Interface"),
				testutils.FilterFields(testStruct{}, "PtrSlice", "SlicePtrStruct", "Interface"),
				testutils.FilterFields(testStruct{}, "Interface"),
			},
			wantErr: nil,
		},
		{
			desc:       "set zero values at negative index",
			buildIndex: 3,
			setZeroFieldsByIndex: map[int][]string{
				-1: {"Int"},
				0:  {"PtrSlice", "SlicePtrStruct"},
			},
			want:    []testStruct{},
			wantErr: errIndexIsOutOfRange,
		},
		{
			desc:       "set zero values at invalid index",
			buildIndex: 3,
			setZeroFieldsByIndex: map[int][]string{
				5: {"Int"},
				0: {"PtrSlice", "SlicePtrStruct"},
			},
			want:    []testStruct{},
			wantErr: errIndexIsOutOfRange,
		},
		{
			desc:       "set incorrect field",
			buildIndex: 3,
			setZeroFieldsByIndex: map[int][]string{
				0: {"incorrect field"},
			},
			want:    []testStruct{},
			wantErr: errFieldNotFound,
		},
		{
			desc:       "set private field",
			buildIndex: 3,
			setZeroFieldsByIndex: map[int][]string{
				0: {"privateField"},
			},
			want:    []testStruct{},
			wantErr: errFieldCantSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			preF := f.BuildList(mockCTX, tt.buildIndex)
			for i, fields := range tt.setZeroFieldsByIndex {
				preF = preF.SetZero(i, fields...)
			}

			got, err := preF.Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want); err != nil {
					t.Fatal(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			for i, g := range got {
				if err := testutils.IsZeroVal(g, tt.wantNonZeroFields[i]...); err != nil {
					t.Fatal(err.Error())
				}

				if err := testutils.IsNotZeroVal(g, tt.wantZeroFields[i]...); err != nil {
					t.Fatal(err.Error())
				}
			}
		})
	}
}

func TestWithOne(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when on builder, insert successfully":                       withOne_OnBuilder,
		"when on builder with multi level, insert successfully":      withOne_OnBuilderMultiLevel,
		"when on builder not pass ptr, return error":                 withOne_OnBuilderNotPassPtr,
		"when on builder not pass struct, return error":              withOne_OnBuilderNotPassStruct,
		"when on builder with err, return error":                     withOne_OnBuilderWithErr,
		"when on builder with cycle, return error":                   withOne_OnBuilderWithCycle,
		"when on builder list, insert successfully":                  withOne_OnBuilderList,
		"when on builder list with multi level, insert successfully": withOne_OnBuilderListMultiLevel,
		"when on builder list not pass ptr, return error":            withOne_OnBuilderListNotPassPtr,
		"when on builder list not pass struct, return error":         withOne_OnBuilderListNotPassStruct,
		"when on builder list with err, return error":                withOne_OnBuilderListWithErr,
		"when on builder list with cycle, return error":              withOne_OnBuilderListWithCycle,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func withOne_OnBuilder(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	assVal := testStructWithID{}
	assVal2 := testStructWithID2{}
	val, err := f.Build(mockCTX).WithOne(&assVal).WithOne(&assVal2).Insert()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if val.ForeignKey != assVal.ID {
		t.Fatalf("ForeignKey should be %v", assVal.ID)
	}

	if val.ForeignKey2 != nil && *val.ForeignKey2 != assVal2.ID {
		t.Fatalf("ForeignKey2 should be %v", assVal2.ID)
	}

	if err := testutils.CompareVal(val.ForeignValue, assVal); err != nil {
		t.Fatal(err.Error())
	}

	if err := testutils.CompareVal(val.ForeignValue2, &assVal2); err != nil {
		t.Fatal(err.Error())
	}
}

func withOne_OnBuilderMultiLevel(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	assVal := testStructWithID{}
	assVal2 := testStructWithID2{}
	assVal3 := testStructWithID3{}

	val, err := f.Build(mockCTX).
		WithOne(&assVal, &assVal2, &assVal3).
		Insert()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if val.ForeignKey != assVal.ID {
		t.Fatalf("ForeignKey should be %v", assVal.ID)
	}
	if val.ForeignKey2 != nil && *val.ForeignKey2 != assVal2.ID {
		t.Fatalf("ForeignKey2 should be %v", assVal2.ID)
	}

	if err := testutils.CompareVal(val.ForeignValue, assVal); err != nil {
		t.Fatal(err.Error())
	}
	if err := testutils.CompareVal(val.ForeignValue2, &assVal2); err != nil {
		t.Fatal(err.Error())
	}

	if assVal2.ForeignKey != assVal3.ID {
		t.Fatalf("ForeignKey should be %v", assVal3.ID)
	}

	if err := testutils.CompareVal(assVal2.ForeignValue, assVal3); err != nil {
		t.Fatal(err.Error())
	}
}

func withOne_OnBuilderNotPassPtr(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	want := testAssocStruct{}
	wantErr := errIsNotPtr

	val, err := f.Build(mockCTX).WithOne(testStructWithID{}).Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if err := testutils.CompareVal(val, want); err != nil {
		t.Fatal(err.Error())
	}
}

func withOne_OnBuilderNotPassStruct(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	want := testAssocStruct{}
	wantErr := errIsNotStructPtr

	assVal := "not struct type"
	val, err := f.Build(mockCTX).WithOne(&assVal).Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if err := testutils.CompareVal(val, want); err != nil {
		t.Fatal(err.Error())
	}
}

func withOne_OnBuilderWithErr(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	want := testAssocStruct{}
	wantErr := errFieldNotFound

	assVal := testStructWithID{}
	val, err := f.Build(mockCTX).SetZero("incorrect field").WithOne(&assVal).Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if err := testutils.CompareVal(val, want); err != nil {
		t.Fatal(err.Error())
	}
}

func withOne_OnBuilderWithCycle(t *testing.T) {
	f := New(testStructWithCycle{}).WithDB(&mockDB{})

	val, err := f.Build(mockCTX).WithOne(&testStructWithCycle2{}).Insert()
	if !errors.Is(err, errCycleDependency) {
		t.Fatalf("error should be %v", errCycleDependency)
	}
	if err := testutils.CompareVal(val, testStructWithCycle{}); err != nil {
		t.Fatal(err.Error())
	}
}

func withOne_OnBuilderList(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	assVal := testStructWithID{}
	assVal2 := testStructWithID2{}
	vals, err := f.BuildList(mockCTX, 2).
		WithOne(&assVal).
		WithOne(&assVal2).
		Insert()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if vals[0].ForeignKey != assVal.ID {
		t.Fatalf("ForeignKey should be %v", assVal.ID)
	}

	if err := testutils.CompareVal(vals[0].ForeignValue, assVal); err != nil {
		t.Fatal(err.Error())
	}

	if vals[1].ForeignKey != assVal.ID {
		t.Fatalf("ForeignKey should be %v", assVal.ID)
	}

	if err := testutils.CompareVal(vals[1].ForeignValue, assVal); err != nil {
		t.Fatal(err.Error())
	}

	if vals[0].ForeignKey2 != nil && *vals[0].ForeignKey2 != assVal2.ID {
		t.Fatalf("ForeignKey2 should be %v", assVal2.ID)
	}

	if err := testutils.CompareVal(vals[0].ForeignValue2, &assVal2); err != nil {
		t.Fatal(err.Error())
	}

	if vals[1].ForeignKey2 != nil && *vals[1].ForeignKey2 != assVal2.ID {
		t.Fatalf("ForeignKey2 should be %v", assVal2.ID)
	}

	if err := testutils.CompareVal(vals[1].ForeignValue2, &assVal2); err != nil {
		t.Fatal(err.Error())
	}
}

func withOne_OnBuilderListMultiLevel(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	assVal := testStructWithID{}
	assVal2 := testStructWithID2{}
	assVal3 := testStructWithID3{}

	vals, err := f.BuildList(mockCTX, 2).
		WithOne(&assVal, &assVal2, &assVal3).
		Insert()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	for i := 0; i < 2; i++ {
		if vals[i].ForeignKey != assVal.ID {
			t.Fatalf("ForeignKey should be %v", assVal.ID)
		}
		if vals[i].ForeignKey2 != nil && *vals[i].ForeignKey2 != assVal2.ID {
			t.Fatalf("ForeignKey2 should be %v", assVal2.ID)
		}
		if err := testutils.CompareVal(vals[i].ForeignValue, assVal); err != nil {
			t.Fatal(err.Error())
		}
		if err := testutils.CompareVal(vals[i].ForeignValue2, &assVal2); err != nil {
			t.Fatal(err.Error())
		}
	}

	if assVal2.ForeignKey != assVal3.ID {
		t.Fatalf("ForeignKey should be %v", assVal3.ID)
	}
	if err := testutils.CompareVal(assVal2.ForeignValue, assVal3); err != nil {
		t.Fatal(err.Error())
	}
}

func withOne_OnBuilderListNotPassPtr(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	want := []testAssocStruct{}
	wantErr := errIsNotPtr

	vals, err := f.BuildList(mockCTX, 2).WithOne(testStructWithID{}).Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if err := testutils.CompareVal(vals, want); err != nil {
		t.Fatal(err.Error())
	}
}

func withOne_OnBuilderListNotPassStruct(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	want := []testAssocStruct{}
	wantErr := errIsNotStructPtr

	assVal := "not struct type"
	vals, err := f.BuildList(mockCTX, 2).WithOne(&assVal).Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if err := testutils.CompareVal(vals, want); err != nil {
		t.Fatal(err.Error())
	}
}

func withOne_OnBuilderListWithErr(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	want := []testAssocStruct{}
	wantErr := errFieldNotFound

	assVal := testStructWithID{}
	vals, err := f.BuildList(mockCTX, 2).SetZero(0, "incorrect field").WithOne(&assVal).Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if err := testutils.CompareVal(vals, want); err != nil {
		t.Fatal(err.Error())
	}
}

func withOne_OnBuilderListWithCycle(t *testing.T) {
	f := New(testStructWithCycle{}).WithDB(&mockDB{})

	var want []testStructWithCycle
	vals, err := f.BuildList(mockCTX, 2).WithOne(&testStructWithCycle2{}).Insert()
	if !errors.Is(err, errCycleDependency) {
		t.Fatalf("error should be %v", errCycleDependency)
	}
	if err := testutils.CompareVal(vals, want); err != nil {
		t.Fatal(err.Error())
	}
}

func TestWithMany(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when withMany on builder, insert successfully":                  withMany_CorrectCase,
		"when withMany on builder with multi level, insert successfully": withMany_MultiLevel,
		"when withMany on builder not pass ptr, return error":            withMany_NotPassPtr,
		"when withMany on builder not pass struct, return error":         withMany_NotPassStruct,
		"when withMany on builder pass diff struct, return error":        withMany_PassDiffStruct,
		"when withMany on builder with cycle, return error":              withMany_WithCycle,
		"when withMany on builder with err, return error":                withMany_WithErr,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func withMany_CorrectCase(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	assVal1 := testStructWithID{}
	assVal2 := testStructWithID{}
	assVal3 := testStructWithID2{}
	assVal4 := testStructWithID2{}
	vals, err := f.BuildList(mockCTX, 2).WithMany([]interface{}{&assVal1, &assVal2}).WithMany([]interface{}{&assVal3, &assVal4}).Insert()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if vals[0].ForeignKey != assVal1.ID {
		t.Fatalf("ForeignKey should be %v", assVal1.ID)
	}

	if err := testutils.CompareVal(vals[0].ForeignValue, assVal1); err != nil {
		t.Fatal(err.Error())
	}

	if vals[1].ForeignKey != assVal2.ID {
		t.Fatalf("ForeignKey should be %v", assVal2.ID)
	}

	if err := testutils.CompareVal(vals[1].ForeignValue, assVal2); err != nil {
		t.Fatal(err.Error())
	}

	if vals[0].ForeignKey2 != nil && *vals[0].ForeignKey2 != assVal3.ID {
		t.Fatalf("ForeignKey2 should be %v", assVal3.ID)
	}

	if err := testutils.CompareVal(vals[0].ForeignValue2, &assVal3); err != nil {
		t.Fatal(err.Error())
	}

	if vals[1].ForeignKey2 != nil && *vals[1].ForeignKey2 != assVal4.ID {
		t.Fatalf("ForeignKey2 should be %v", assVal4.ID)
	}

	if err := testutils.CompareVal(vals[1].ForeignValue2, &assVal4); err != nil {
		t.Fatal(err.Error())
	}
}

func withMany_MultiLevel(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	assVals1 := []interface{}{&testStructWithID{}, &testStructWithID{}}
	assVals2 := []interface{}{&testStructWithID2{}, &testStructWithID2{}}
	assVals3 := []interface{}{&testStructWithID3{}, &testStructWithID3{}}
	vals, err := f.BuildList(mockCTX, 2).
		WithMany(assVals1).
		WithMany(assVals2).
		WithMany(assVals3).
		Insert()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	for i := 0; i < 2; i++ {
		if vals[i].ForeignKey != assVals1[i].(*testStructWithID).ID {
			t.Fatalf("ForeignKey should be %v", assVals1[i].(*testStructWithID).ID)
		}
		if vals[i].ForeignKey2 != nil && *vals[i].ForeignKey2 != assVals2[i].(*testStructWithID2).ID {
			t.Fatalf("ForeignKey2 should be %v", assVals2[i].(*testStructWithID2).ID)
		}
		if err := testutils.CompareVal(vals[i].ForeignValue, *assVals1[i].(*testStructWithID)); err != nil {
			t.Fatal(err.Error())
		}
		if err := testutils.CompareVal(vals[i].ForeignValue2, assVals2[i].(*testStructWithID2)); err != nil {
			t.Fatal(err.Error())
		}
	}

	if assVals2[0].(*testStructWithID2).ForeignKey != assVals3[0].(*testStructWithID3).ID {
		t.Fatalf("ForeignKey should be %v", assVals3[0].(*testStructWithID3).ID)
	}
	if err := testutils.CompareVal(assVals2[0].(*testStructWithID2).ForeignValue, *assVals3[0].(*testStructWithID3)); err != nil {
		t.Fatal(err.Error())
	}
}

func withMany_NotPassPtr(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	want := []testAssocStruct{}
	wantErr := errIsNotPtr

	vals, err := f.BuildList(mockCTX, 2).WithMany([]interface{}{testStructWithID{}, testStructWithID{}}).Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if err := testutils.CompareVal(vals, want); err != nil {
		t.Fatal(err.Error())
	}
}

func withMany_NotPassStruct(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	want := []testAssocStruct{}
	wantErr := errIsNotStructPtr

	assVal := "not struct type"
	vals, err := f.BuildList(mockCTX, 2).WithMany([]interface{}{&assVal, &assVal}).Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if err := testutils.CompareVal(vals, want); err != nil {
		t.Fatal(err.Error())
	}
}

func withMany_PassDiffStruct(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	assVal := testStructWithID{}
	assVal2 := testStructWithID2{}
	vals, err := f.BuildList(mockCTX, 2).WithMany([]interface{}{&assVal, &assVal2}).Insert()

	var want []testAssocStruct
	if !errors.Is(err, errValueNotTheSameType) {
		t.Fatalf("error should be %v", errValueNotTheSameType)
	}
	if err := testutils.CompareVal(vals, want); err != nil {
		t.Fatal(err.Error())
	}
}

func withMany_WithCycle(t *testing.T) {
	f := New(testStructWithCycle{}).WithDB(&mockDB{})

	vals, err := f.BuildList(mockCTX, 2).
		WithMany([]interface{}{&testStructWithCycle2{}, &testStructWithCycle2{}}).
		Insert()

	var want []testStructWithCycle
	if !errors.Is(err, errCycleDependency) {
		t.Fatalf("error should be %v", errCycleDependency)
	}
	if err := testutils.CompareVal(vals, want); err != nil {
		t.Fatal(err.Error())
	}
}

func withMany_WithErr(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	wantErr := errFieldNotFound
	want := []testAssocStruct{}

	assVal1 := testStructWithID{}
	assVal2 := testStructWithID{}
	vals, err := f.BuildList(mockCTX, 2).SetZero(0, "incorrect field").WithMany([]interface{}{&assVal1, &assVal2}).Insert()
	if !errors.Is(err, wantErr) {
		t.Fatalf("error should be %v", wantErr)
	}

	if err := testutils.CompareVal(vals, want); err != nil {
		t.Fatal(err.Error())
	}
}

func TestReset(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when reset, index should be 0":            reset_Index,
		"when reset, associations should be empty": reset_Associations,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func reset_Index(t *testing.T) {
	f := New(testStruct{})

	_, err := f.BuildList(mockCTX, 5).Get()
	if err != nil {
		t.Fatal(err.Error())
	}

	f.Reset()
	if f.index != 1 {
		t.Fatalf("index should be 1")
	}
}

func reset_Associations(t *testing.T) {
	f := New(testAssocStruct{}).WithDB(&mockDB{})

	// explicitly set association and not insert to make the association not empty
	f.Build(mockCTX).WithOne(&testStructWithID{})

	f.Reset()
	if len(f.associations) != 0 {
		t.Fatalf("associations should be empty")
	}
}

func TestWithStorageName(t *testing.T) {
	f := New(testStruct{}).WithStorageName("test")
	if f.storageName != "test" {
		t.Fatalf("storageName should be %v", "test")
	}
}

func TestWithDB(t *testing.T) {
	f := New(testStruct{}).WithDB(&mockDB{})
	if f.db == nil {
		t.Fatalf("db should not be nil")
	}
}

func checkFactory[T any](got, want *Factory[T]) error {
	if got.storageName != want.storageName {
		return fmt.Errorf("storageName should be %s", want.storageName)
	}

	if got.dataType != want.dataType {
		return fmt.Errorf("dataType should be %v", want.dataType)
	}

	if got.index != want.index {
		return fmt.Errorf("index should be %v", want.index)
	}

	if testutils.CompareVal(got.ignoreFields, want.ignoreFields) != nil {
		return fmt.Errorf("ignoreFields should be %v", want.ignoreFields)
	}

	if got.isSetZeroValue != want.isSetZeroValue {
		return fmt.Errorf("isSetZeroValue should be %v", want.isSetZeroValue)
	}

	if !errors.Is(got.err, want.err) {
		return fmt.Errorf("err should be %v", want.err)
	}

	if len(got.traits) != len(want.traits) {
		return fmt.Errorf("traits should be %v", want.traits)
	}

	if (got.blueprint == nil) != (want.blueprint == nil) {
		return fmt.Errorf("blueprint mismatch: got %v, want %v", got.blueprint != nil, want.blueprint != nil)
	}

	if (got.db == nil) != (want.db == nil) {
		return fmt.Errorf("db mismatch: got %v, want %v", got.db != nil, want.db != nil)
	}

	return nil
}
