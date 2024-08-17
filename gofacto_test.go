package gofacto

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/eyo-chen/gofacto/internal/testutils"
)

var (
	now     = time.Now()
	mockCTX = context.Background()
)

type testStruct struct {
	Int            int
	PtrInt         *int
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
		return testStruct{
			Int:            i * 2,
			PtrInt:         &i,
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

				return testStruct{
					Int:            i2,
					PtrInt:         &i1,
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

				return testStruct{
					Int:            i4,
					PtrInt:         &i2,
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
				t.Fatalf(err.Error())
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
				t.Fatalf(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Bool", "PtrBool")...); err != nil {
				t.Fatalf(err.Error())
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
				i2, i4 := 2, 4
				str := "test2"

				return testStruct{
					Int:    i4,
					PtrInt: &i2,
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
				t.Fatalf(err.Error())
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

			if err := testutils.IsNotZeroVal(got, testutils.FilterFields(testStruct{})...); err != nil {
				t.Fatalf(err.Error())
			}
		})
	}
}

func build_NoBluePrintNotSetZeroValues(t *testing.T) {
	f := New(testStruct{}).WithIsSetZeroValue(false)

	tests := []struct {
		desc string
		want func() testStruct
	}{
		{
			desc: "first build",
			want: func() testStruct { return testStruct{} },
		},
		{
			desc: "second build",
			want: func() testStruct { return testStruct{} },
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want()); err != nil {
				t.Fatalf(err.Error())
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
				t.Fatalf(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Fatalf(err.Error())
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
		return testStruct{
			Int:            i * 2,
			PtrInt:         &i,
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

				return []testStruct{
					{
						Int:            i2,
						PtrInt:         &i1,
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

				return []testStruct{
					{
						Int:            i6,
						PtrInt:         &i3,
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
				t.Fatalf(err.Error())
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
				t.Fatalf(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, testutils.FilterFields(testStruct{}, "Int", "PtrStruct", "SlicePtrStruct")...); err != nil {
				t.Fatalf(err.Error())
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

			if err := testutils.CompareVal(got, tt.want()); err != nil {
				t.Fatalf(err.Error())
			}
		})
	}
}

func buildList_NoBluePrintNotSetZeroValues(t *testing.T) {
	f := New(testStruct{})

	tests := []struct {
		desc string
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
			got, err := f.BuildList(mockCTX, 2).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.IsNotZeroVal(got, testutils.FilterFields(testStruct{})...); err != nil {
				t.Fatalf(err.Error())
			}
		})
	}
}

func buildList_NoBluePrint(t *testing.T) {
	f := New(testStruct{}).WithIsSetZeroValue(false)

	tests := []struct {
		desc string
		want func() []testStruct
	}{
		{
			desc: "first build",
			want: func() []testStruct { return []testStruct{{}, {}} },
		},
		{
			desc: "second build",
			want: func() []testStruct { return []testStruct{{}, {}} },
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want()); err != nil {
				t.Fatalf(err.Error())
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
			got, err := f.BuildList(mockCTX, 2).Get()
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.IsZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Fatalf(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Fatalf(err.Error())
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
		"when insert on builder without db, error should be returned":      insert_OnBuilder,
		"when insert on builder list without db, error should be returned": insert_OnBuilderList,
	} {
		t.Run(testutils.GetFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func insert_OnBuilder(t *testing.T) {
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

func insert_OnBuilderList(t *testing.T) {
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

func TestOverwrite(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when overwrite on builder, overwrite one value":           overwrite_OnBuilder,
		"when overwrite on builder list, overwrite one value":      overwrite_OnBuilderList,
		"when overwrites on builder list, overwrite target values": overwrites_OnBuilderList,
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
				t.Fatalf(err.Error())
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
				t.Fatalf(err.Error())
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
				t.Fatalf(err.Error())
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
			got, err := f.Build(mockCTX).WithTrait(tt.trait).Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want()); err != nil {
					t.Fatalf(err.Error())
				}

				return
			}

			if err := testutils.CompareVal(got, tt.want(), testutils.FilterFields(testStruct{}, "PtrStr", "Time", "Slice")...); err != nil {
				t.Fatalf(err.Error())
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
			got, err := f.BuildList(mockCTX, 2).WithTrait(tt.tait).Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want()); err != nil {
					t.Fatalf(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want(), testutils.FilterFields(testStruct{}, "PtrStr", "Time", "Slice")...); err != nil {
				t.Fatalf(err.Error())
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
			got, err := f.Build(mockCTX).WithTrait(tt.taits[0]).WithTrait(tt.taits[1]).Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want()); err != nil {
					t.Fatalf(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want(), testutils.FilterFields(testStruct{}, "PtrStr", "Time", "Slice")...); err != nil {
				t.Fatalf(err.Error())
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
			got, err := f.BuildList(mockCTX, 2).WithTraits(tt.taits...).Get()

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error should be %v", tt.wantErr)
				}

				if err := testutils.CompareVal(got, tt.want()); err != nil {
					t.Fatalf(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.CompareVal(got, tt.want(), testutils.FilterFields(testStruct{}, "PtrStr", "Time", "Slice")...); err != nil {
				t.Fatalf(err.Error())
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
		return testStruct{
			Int:            i * 2,
			PtrInt:         &i,
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
			setZeroFields:     []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
			wantZeroFields:    []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "privateField"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"),
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
					t.Fatalf(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.IsZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Fatalf(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Fatalf(err.Error())
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
			wantZeroFields:    []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "privateField"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"),
			wantErr:           nil,
		},
		{
			desc:          "set one zero value",
			setZeroFields: []string{"Int"},
			// interface value will default set to nil
			wantZeroFields:    []string{"Int", "Interface", "privateField"},
			wantNonZeroFields: testutils.FilterFields(testStruct{}, "Int", "Interface"),
			wantErr:           nil,
		},
		{
			desc:          "set no zero value",
			setZeroFields: []string{},
			// interface value will default set to nil
			wantZeroFields:    []string{"Interface", "privateField"},
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
					t.Fatalf(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.IsZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Fatalf(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Fatalf(err.Error())
			}
		})
	}
}

func setZero_OnBuilderListWithBluePrint(t *testing.T) {
	blueprint := func(i int) testStruct {
		str := fmt.Sprintf("test%d", i)
		f := 1.1
		return testStruct{
			Int:            i * 2,
			PtrInt:         &i,
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
			setZeroFields: []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
			wantZeroFields: [][]string{
				{"privateField"},
				{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "privateField"},
			},
			wantNonZeroFields: [][]string{
				{"Int", "PtrInt", "Str", "PtrStr", "Bool", "PtrBool", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
				testutils.FilterFields(testStruct{}, "Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"),
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
					t.Fatalf(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			for i, g := range got {
				if err := testutils.IsZeroVal(g, tt.wantNonZeroFields[i]...); err != nil {
					t.Fatalf(err.Error())
				}

				if err := testutils.IsNotZeroVal(g, tt.wantZeroFields[i]...); err != nil {
					t.Fatalf(err.Error())
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
				{"Interface", "privateField"},
				{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Struct", "PtrStruct", "Interface", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct", "privateField"},
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
					t.Fatalf(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			for i, g := range got {
				if err := testutils.IsZeroVal(g, tt.wantNonZeroFields[i]...); err != nil {
					t.Fatalf(err.Error())
				}

				if err := testutils.IsNotZeroVal(g, tt.wantZeroFields[i]...); err != nil {
					t.Fatalf(err.Error())
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
			wantZeroFields:    []string{"Int", "Slice", "Interface", "privateField"},
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
			wantZeroFields:    []string{"Int", "Slice", "Struct", "Interface", "privateField"},
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
					t.Fatalf(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if err := testutils.IsZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Fatalf(err.Error())
			}

			if err := testutils.IsNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Fatalf(err.Error())
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
				{"Int", "Interface", "privateField", "privateField"},
				{"PtrSlice", "SlicePtrStruct", "Interface", "privateField", "privateField"},
				{"Interface", "privateField", "privateField"},
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
					t.Fatalf(err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			for i, g := range got {
				if err := testutils.IsZeroVal(g, tt.wantNonZeroFields[i]...); err != nil {
					t.Fatalf(err.Error())
				}

				if err := testutils.IsNotZeroVal(g, tt.wantZeroFields[i]...); err != nil {
					t.Fatalf(err.Error())
				}
			}
		})
	}
}
