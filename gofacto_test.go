package gofacto

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/eyo-chen/gofacto/utils"
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
}

type subStruct struct {
	SubID   int
	SubName string
}

func TestBuild(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when pass buildPrint with all fields, all fields set by bluePrint":                  build_BluePrintAllFields,
		"when pass buildPrint with some fields, other fields set by gofaco":                  build_BluePrintSomeFields,
		"when pass buildPrint without setting zero values, other fileds remain zero value":   build_BluePrintNotSetZeroValues,
		"when not pass buildPrint, all fields set by gofacto":                                build_NoBluePrint,
		"when not pass buildPrint without setting zero values, all fields remain zero value": build_NoBluePrintNotSetZeroValues,
		"when setting ignore fields, ignore fields should be zero value":                     build_IgnoreFields,
	} {
		t.Run(getFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func build_BluePrintAllFields(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
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
			Struct:         subStruct{SubID: i, SubName: str},
			PtrStruct:      &subStruct{SubID: i, SubName: str},
			Slice:          []int{i, i + 1, i + 2},
			PtrSlice:       []*int{&i, &i, &i},
			SliceStruct:    []subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
			SlicePtrStruct: []*subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
		}
	}
	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint})

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
					Struct:         subStruct{SubID: i1, SubName: str},
					PtrStruct:      &subStruct{SubID: i1, SubName: str},
					Slice:          []int{i1, i2, i3},
					PtrSlice:       []*int{&i1, &i1, &i1},
					SliceStruct:    []subStruct{{SubID: i1, SubName: str}, {SubID: i2, SubName: str}},
					SlicePtrStruct: []*subStruct{{SubID: i1, SubName: str}, {SubID: i2, SubName: str}},
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
					Struct:         subStruct{SubID: i2, SubName: str},
					PtrStruct:      &subStruct{SubID: i2, SubName: str},
					Slice:          []int{i2, i3, i4},
					PtrSlice:       []*int{&i2, &i2, &i2},
					SliceStruct:    []subStruct{{SubID: i2, SubName: str}, {SubID: i3, SubName: str}},
					SlicePtrStruct: []*subStruct{{SubID: i2, SubName: str}, {SubID: i3, SubName: str}},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).Get()
			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want()); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func build_BluePrintSomeFields(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
		b := true
		return testStruct{
			Int:     i * 2,
			PtrInt:  &i,
			Bool:    b,
			PtrBool: &b,
		}
	}
	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint})

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
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want(), filterFields(testStruct{}, "Int", "PtrInt", "Bool", "PtrBool")...); err != nil {
				t.Errorf(err.Error())
			}

			if err := isNotZeroVal(got, filterFields(testStruct{}, "Int", "PtrInt", "Bool", "PtrBool")...); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func build_BluePrintNotSetZeroValues(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			Int:    i * 2,
			PtrInt: &i,
			Str:    str,
			PtrStr: &str,
		}
	}
	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint, IsSetZeroValue: utils.Bool(false)})

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
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want()); err != nil {
				t.Errorf(err.Error())
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
				t.Errorf("error from gofacto: %v", err)
			}

			if err := isNotZeroVal(got, filterFields(testStruct{})...); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func build_NoBluePrintNotSetZeroValues(t *testing.T) {
	f := New(testStruct{}).SetConfig(Config[testStruct]{IsSetZeroValue: utils.Bool(false)})

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
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want()); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

// TODO: make error message more readable
func build_IgnoreFields(t *testing.T) {
	f := New(testStruct{}).SetConfig(Config[testStruct]{IgnoreFields: []string{"Int", "PtrInt", "Str", "PtrStr", "Incorrect field"}})

	tests := []struct {
		desc              string
		wantZeroFields    []string
		wantNonZeroFields []string
	}{
		{
			desc:              "first build",
			wantZeroFields:    []string{"Int", "PtrInt", "Str", "PtrStr", "Interface"},
			wantNonZeroFields: filterFields(testStruct{}, "Int", "PtrInt", "Str", "PtrStr", "Interface"),
		},
		{
			desc:              "second build",
			wantZeroFields:    []string{"Int", "PtrInt", "Str", "PtrStr", "Interface"},
			wantNonZeroFields: filterFields(testStruct{}, "Int", "PtrInt", "Str", "PtrStr", "Interface"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).Get()
			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := isZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Errorf(err.Error())
			}

			if err := isNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func TestBuildList(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when pass buildList with all fields, all fields set by bluePrint":                  buildList_BluePrintAllFields,
		"when pass buildList with some fields, other fields set by gofaco":                  buildList_BluePrintSomeFields,
		"when pass buildList without setting zero values, other fileds remain zero value":   buildList_BluePrintNotSetZeroValues,
		"when not pass buildList, all fields set by gofacto":                                buildList_NoBluePrint,
		"when not pass buildList without setting zero values, all fields remain zero value": buildList_NoBluePrintNotSetZeroValues,
	} {
		t.Run(getFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func buildList_BluePrintAllFields(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
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
			Struct:         subStruct{SubID: i, SubName: str},
			PtrStruct:      &subStruct{SubID: i, SubName: str},
			Slice:          []int{i, i + 1, i + 2},
			PtrSlice:       []*int{&i, &i, &i},
			SliceStruct:    []subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
			SlicePtrStruct: []*subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
		}
	}
	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint})

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
						Struct:         subStruct{SubID: i1, SubName: str1},
						PtrStruct:      &subStruct{SubID: i1, SubName: str1},
						Slice:          []int{i1, i2, i3},
						PtrSlice:       []*int{&i1, &i1, &i1},
						SliceStruct:    []subStruct{{SubID: i1, SubName: str1}, {SubID: i2, SubName: str1}},
						SlicePtrStruct: []*subStruct{{SubID: i1, SubName: str1}, {SubID: i2, SubName: str1}},
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
						Struct:         subStruct{SubID: i2, SubName: str2},
						PtrStruct:      &subStruct{SubID: i2, SubName: str2},
						Slice:          []int{i2, i3, i4},
						PtrSlice:       []*int{&i2, &i2, &i2},
						SliceStruct:    []subStruct{{SubID: i2, SubName: str2}, {SubID: i3, SubName: str2}},
						SlicePtrStruct: []*subStruct{{SubID: i2, SubName: str2}, {SubID: i3, SubName: str2}},
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
						Struct:         subStruct{SubID: i3, SubName: str3},
						PtrStruct:      &subStruct{SubID: i3, SubName: str3},
						Slice:          []int{i3, i4, i5},
						PtrSlice:       []*int{&i3, &i3, &i3},
						SliceStruct:    []subStruct{{SubID: i3, SubName: str3}, {SubID: i4, SubName: str3}},
						SlicePtrStruct: []*subStruct{{SubID: i3, SubName: str3}, {SubID: i4, SubName: str3}},
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
						Struct:         subStruct{SubID: i4, SubName: str4},
						PtrStruct:      &subStruct{SubID: i4, SubName: str4},
						Slice:          []int{i4, i5, i6},
						PtrSlice:       []*int{&i4, &i4, &i4},
						SliceStruct:    []subStruct{{SubID: i4, SubName: str4}, {SubID: i5, SubName: str4}},
						SlicePtrStruct: []*subStruct{{SubID: i4, SubName: str4}, {SubID: i5, SubName: str4}},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Get()
			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want()); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func buildList_BluePrintSomeFields(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			Int:            i * 2,
			PtrStruct:      &subStruct{SubID: i, SubName: str},
			SlicePtrStruct: []*subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
		}
	}
	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint})

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
						PtrStruct:      &subStruct{SubID: i1, SubName: str1},
						SlicePtrStruct: []*subStruct{{SubID: i1, SubName: str1}, {SubID: i2, SubName: str1}},
					},
					{
						Int:            i4,
						PtrStruct:      &subStruct{SubID: i2, SubName: str2},
						SlicePtrStruct: []*subStruct{{SubID: i2, SubName: str2}, {SubID: i3, SubName: str2}},
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
						PtrStruct:      &subStruct{SubID: i3, SubName: str3},
						SlicePtrStruct: []*subStruct{{SubID: i3, SubName: str3}, {SubID: i4, SubName: str3}},
					},
					{
						Int:            i8,
						PtrStruct:      &subStruct{SubID: i4, SubName: str4},
						SlicePtrStruct: []*subStruct{{SubID: i4, SubName: str4}, {SubID: i5, SubName: str4}},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Get()
			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want(), filterFields(testStruct{}, "Int", "PtrStruct", "SlicePtrStruct")...); err != nil {
				t.Errorf(err.Error())
			}

			if err := isNotZeroVal(got, filterFields(testStruct{}, "Int", "PtrStruct", "SlicePtrStruct")...); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func buildList_BluePrintNotSetZeroValues(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			Int:            i * 2,
			PtrStruct:      &subStruct{SubID: i, SubName: str},
			SlicePtrStruct: []*subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
		}
	}
	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint, IsSetZeroValue: utils.Bool(false)})

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
						PtrStruct:      &subStruct{SubID: i1, SubName: str1},
						SlicePtrStruct: []*subStruct{{SubID: i1, SubName: str1}, {SubID: i2, SubName: str1}},
					},
					{
						Int:            i4,
						PtrStruct:      &subStruct{SubID: i2, SubName: str2},
						SlicePtrStruct: []*subStruct{{SubID: i2, SubName: str2}, {SubID: i3, SubName: str2}},
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
						PtrStruct:      &subStruct{SubID: i3, SubName: str3},
						SlicePtrStruct: []*subStruct{{SubID: i3, SubName: str3}, {SubID: i4, SubName: str3}},
					},
					{
						Int:            i8,
						PtrStruct:      &subStruct{SubID: i4, SubName: str4},
						SlicePtrStruct: []*subStruct{{SubID: i4, SubName: str4}, {SubID: i5, SubName: str4}},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Get()
			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want()); err != nil {
				t.Errorf(err.Error())
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
				t.Errorf("error from gofacto: %v", err)
			}

			if err := isNotZeroVal(got, filterFields(testStruct{})...); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func buildList_NoBluePrint(t *testing.T) {
	f := New(testStruct{}).SetConfig(Config[testStruct]{IsSetZeroValue: utils.Bool(false)})

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
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want()); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func TestOverwrite(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when overwrite on builder, overwrite one value":           overwrite_OnBuilder,
		"when overwrites on builder list, overwrite target values": overwrite_OnBuilderList,
		"when overwrite on builder list, overwrite one value":      overwrite_OnBuilderListOneValue,
	} {
		t.Run(getFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func overwrite_OnBuilder(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			Int:            i * 2,
			PtrStruct:      &subStruct{SubID: i, SubName: str},
			SlicePtrStruct: []*subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
		}
	}
	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint})

	tests := []struct {
		desc string
		ow   testStruct
		want testStruct
	}{
		{
			desc: "overwrite with value",
			ow:   testStruct{Int: 10, PtrStruct: &subStruct{SubID: 10, SubName: "test10"}},
			want: testStruct{
				Int:            10,
				PtrStruct:      &subStruct{SubID: 10, SubName: "test10"},
				SlicePtrStruct: []*subStruct{{SubID: 1, SubName: "test1"}, {SubID: 2, SubName: "test1"}},
			},
		},
		{
			desc: "overwrite without value",
			ow:   testStruct{},
			want: testStruct{
				Int:            4,
				PtrStruct:      &subStruct{SubID: 2, SubName: "test2"},
				SlicePtrStruct: []*subStruct{{SubID: 2, SubName: "test2"}, {SubID: 3, SubName: "test2"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).Overwrite(tt.ow).Get()
			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want, filterFields(testStruct{}, "Int", "PtrStruct", "SlicePtrStruct")...); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func overwrite_OnBuilderList(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			Int:            i * 2,
			PtrStruct:      &subStruct{SubID: i, SubName: str},
			SlicePtrStruct: []*subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
		}
	}
	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint})

	tests := []struct {
		desc string
		ow   []testStruct
		want []testStruct
	}{
		{
			desc: "overwrite with same length",
			ow: []testStruct{
				{Int: 10, PtrStruct: &subStruct{SubID: 10, SubName: "test10"}},
				{Int: 20, PtrStruct: &subStruct{SubID: 20, SubName: "test20"}},
			},
			want: []testStruct{
				{
					Int:            10,
					PtrStruct:      &subStruct{SubID: 10, SubName: "test10"},
					SlicePtrStruct: []*subStruct{{SubID: 1, SubName: "test1"}, {SubID: 2, SubName: "test1"}},
				},
				{
					Int:            20,
					PtrStruct:      &subStruct{SubID: 20, SubName: "test20"},
					SlicePtrStruct: []*subStruct{{SubID: 2, SubName: "test2"}, {SubID: 3, SubName: "test2"}},
				},
			},
		},
		{
			desc: "overwrite with longer length",
			ow: []testStruct{
				{Int: 10, PtrStruct: &subStruct{SubID: 10, SubName: "test10"}},
				{Int: 20, PtrStruct: &subStruct{SubID: 20, SubName: "test20"}},
				{Int: 30, PtrStruct: &subStruct{SubID: 30, SubName: "test30"}},
				{Int: 40, PtrStruct: &subStruct{SubID: 40, SubName: "test40"}},
			},
			want: []testStruct{
				{
					Int:            10,
					PtrStruct:      &subStruct{SubID: 10, SubName: "test10"},
					SlicePtrStruct: []*subStruct{{SubID: 3, SubName: "test3"}, {SubID: 4, SubName: "test3"}},
				},
				{
					Int:            20,
					PtrStruct:      &subStruct{SubID: 20, SubName: "test20"},
					SlicePtrStruct: []*subStruct{{SubID: 4, SubName: "test4"}, {SubID: 5, SubName: "test4"}},
				},
			},
		},
		{
			desc: "overwrite with shorter length",
			ow: []testStruct{
				{Int: 10, PtrStruct: &subStruct{SubID: 10, SubName: "test10"}},
			},
			want: []testStruct{
				{
					Int:            10,
					PtrStruct:      &subStruct{SubID: 10, SubName: "test10"},
					SlicePtrStruct: []*subStruct{{SubID: 5, SubName: "test5"}, {SubID: 6, SubName: "test5"}},
				},
				{
					Int:            12,
					PtrStruct:      &subStruct{SubID: 6, SubName: "test6"},
					SlicePtrStruct: []*subStruct{{SubID: 6, SubName: "test6"}, {SubID: 7, SubName: "test6"}},
				},
			},
		},
		{
			desc: "overwrite without value",
			ow:   []testStruct{},
			want: []testStruct{
				{
					Int:            14,
					PtrStruct:      &subStruct{SubID: 7, SubName: "test7"},
					SlicePtrStruct: []*subStruct{{SubID: 7, SubName: "test7"}, {SubID: 8, SubName: "test7"}},
				},
				{
					Int:            16,
					PtrStruct:      &subStruct{SubID: 8, SubName: "test8"},
					SlicePtrStruct: []*subStruct{{SubID: 8, SubName: "test8"}, {SubID: 9, SubName: "test8"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Overwrites(tt.ow...).Get()
			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want, filterFields(testStruct{}, "Int", "PtrStruct", "SlicePtrStruct")...); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func overwrite_OnBuilderListOneValue(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
		str := fmt.Sprintf("test%d", i)
		return testStruct{
			Int:            i * 2,
			PtrStruct:      &subStruct{SubID: i, SubName: str},
			SlicePtrStruct: []*subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
		}
	}
	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint})

	tests := []struct {
		desc string
		ow   testStruct
		want []testStruct
	}{
		{
			desc: "overwrite with value",
			ow:   testStruct{Int: 10, PtrStruct: &subStruct{SubID: 10, SubName: "test10"}},
			want: []testStruct{
				{
					Int:            10,
					PtrStruct:      &subStruct{SubID: 10, SubName: "test10"},
					SlicePtrStruct: []*subStruct{{SubID: 1, SubName: "test1"}, {SubID: 2, SubName: "test1"}},
				},
				{
					Int:            10,
					PtrStruct:      &subStruct{SubID: 10, SubName: "test10"},
					SlicePtrStruct: []*subStruct{{SubID: 2, SubName: "test2"}, {SubID: 3, SubName: "test2"}},
				},
			},
		},
		{
			desc: "overwrite without value",
			ow:   testStruct{},
			want: []testStruct{
				{
					Int:            6,
					PtrStruct:      &subStruct{SubID: 3, SubName: "test3"},
					SlicePtrStruct: []*subStruct{{SubID: 3, SubName: "test3"}, {SubID: 4, SubName: "test3"}},
				},
				{
					Int:            8,
					PtrStruct:      &subStruct{SubID: 4, SubName: "test4"},
					SlicePtrStruct: []*subStruct{{SubID: 4, SubName: "test4"}, {SubID: 5, SubName: "test4"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).Overwrite(tt.ow).Get()
			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want, filterFields(testStruct{}, "Int", "PtrStruct", "SlicePtrStruct")...); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func TestWithTrait(t *testing.T) {
	for _, fn := range map[string]func(*testing.T){
		"when withTrait on builder, overwrite one value":           withTrait_OnBuilder,
		"when withTraits on builder list, overwrite target values": withTrait_OnBuilderList,
		"when withTrait on builder list, overwrite one value":      withTrait_OnBuilderListOneValue,
		"when multiple withTrait on builder, overwrite one value":  withTrait_OnBuilderMultiple,
	} {
		t.Run(getFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func withTrait_OnBuilder(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
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

	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint}).
		SetTrait("trait", setTraiter)

	tests := []struct {
		desc     string
		trait    string
		want     func() testStruct
		hasError bool
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
			hasError: false,
		},
		{
			desc:     "set trait with incorrect value",
			trait:    "incorrect trait",
			want:     func() testStruct { return testStruct{} },
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).WithTrait(tt.trait).Get()

			if tt.hasError {
				if err == nil {
					t.Errorf("error should be occurred")
				}

				if err := compareVal(got, tt.want()); err != nil {
					t.Errorf(err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want(), filterFields(testStruct{}, "PtrStr", "Time", "Slice")...); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func withTrait_OnBuilderList(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
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

	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint}).
		SetTrait("trait", setTraiter)

	tests := []struct {
		desc     string
		taits    []string
		want     func() []testStruct
		hasError bool
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
			hasError: false,
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
			hasError: false,
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
			hasError: false,
		},
		{
			desc:     "set trait with incorrect value",
			taits:    []string{"incorrect trait"},
			want:     func() []testStruct { return []testStruct{} },
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).WithTraits(tt.taits...).Get()

			if tt.hasError {
				if err == nil {
					t.Errorf("error should be occurred")
				}

				if err := compareVal(got, tt.want()); err != nil {
					t.Errorf(err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want(), filterFields(testStruct{}, "PtrStr", "Time", "Slice")...); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func withTrait_OnBuilderListOneValue(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
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

	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint}).
		SetTrait("trait", setTraiter)

	tests := []struct {
		desc     string
		tait     string
		want     func() []testStruct
		hasError bool
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
			hasError: false,
		},
		{
			desc:     "set trait with incorrect value",
			tait:     "incorrect trait",
			want:     func() []testStruct { return []testStruct{} },
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).WithTrait(tt.tait).Get()

			if tt.hasError {
				if err == nil {
					t.Errorf("error should be occurred")
				}

				if err := compareVal(got, tt.want()); err != nil {
					t.Errorf(err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want(), filterFields(testStruct{}, "PtrStr", "Time", "Slice")...); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func withTrait_OnBuilderMultiple(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
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

	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint}).
		SetTrait("trait1", setTraiter1).
		SetTrait("trait2", setTraiter2)

	tests := []struct {
		desc     string
		taits    []string
		want     func() testStruct
		hasError bool
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
			hasError: false,
		},
		{
			desc:     "set one trait with incorrect value",
			taits:    []string{"trait1", "incorrect trait"},
			want:     func() testStruct { return testStruct{} },
			hasError: true,
		},
		{
			desc:     "set two traits with incorrect value",
			taits:    []string{"incorrect trait1", "incorrect trait2"},
			want:     func() testStruct { return testStruct{} },
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).WithTrait(tt.taits[0]).WithTrait(tt.taits[1]).Get()

			if tt.hasError {
				if err == nil {
					t.Errorf("error should be occurred")
				}

				if err := compareVal(got, tt.want()); err != nil {
					t.Errorf(err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := compareVal(got, tt.want(), filterFields(testStruct{}, "PtrStr", "Time", "Slice")...); err != nil {
				t.Errorf(err.Error())
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
		t.Run(getFunName(fn), func(t *testing.T) {
			fn(t)
		})
	}
}

func setZero_OnBuilderWithBluePrint(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
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
			Struct:         subStruct{SubID: i, SubName: str},
			PtrStruct:      &subStruct{SubID: i, SubName: str},
			Slice:          []int{i, i + 1, i + 2},
			PtrSlice:       []*int{&i, &i, &i},
			SliceStruct:    []subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
			SlicePtrStruct: []*subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
		}
	}
	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint})

	tests := []struct {
		desc              string
		setZeroFields     []string
		wantZeroFields    []string
		wantNonZeroFields []string
		hasError          bool
		want              testStruct
	}{
		{
			desc:              "set many zero values",
			setZeroFields:     []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
			wantZeroFields:    []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
			wantNonZeroFields: filterFields(testStruct{}, "Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"),
			hasError:          false,
		},
		{
			desc:              "set one zero value",
			setZeroFields:     []string{"Int"},
			wantZeroFields:    []string{"Int"},
			wantNonZeroFields: filterFields(testStruct{}, "Int"),
			hasError:          false,
		},
		{
			desc:              "set no zero value",
			setZeroFields:     []string{},
			wantZeroFields:    []string{},
			wantNonZeroFields: filterFields(testStruct{}),
			hasError:          false,
		},
		{
			desc:          "set incorrect field",
			setZeroFields: []string{"incorrect field"},
			hasError:      true,
			want:          testStruct{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).SetZero(tt.setZeroFields...).Get()

			if tt.hasError {
				if err == nil {
					t.Errorf("error should be occurred")
				}

				if err := compareVal(got, tt.want); err != nil {
					t.Errorf(err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := isZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Errorf(err.Error())
			}

			if err := isNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Errorf(err.Error())
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
		hasError          bool
		want              testStruct
	}{
		{
			desc:              "set many zero values",
			setZeroFields:     []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
			wantZeroFields:    []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
			wantNonZeroFields: filterFields(testStruct{}, "Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"),
			hasError:          false,
		},
		{
			desc:          "set one zero value",
			setZeroFields: []string{"Int"},
			// interface value will default set to nil
			wantZeroFields:    []string{"Int", "Interface"},
			wantNonZeroFields: filterFields(testStruct{}, "Int", "Interface"),
			hasError:          false,
		},
		{
			desc:          "set no zero value",
			setZeroFields: []string{},
			// interface value will default set to nil
			wantZeroFields:    []string{"Interface"},
			wantNonZeroFields: filterFields(testStruct{}),
			hasError:          false,
		},
		{
			desc:          "set incorrect field",
			setZeroFields: []string{"incorrect field"},
			hasError:      true,
			want:          testStruct{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.Build(mockCTX).SetZero(tt.setZeroFields...).Get()

			if tt.hasError {
				if err == nil {
					t.Errorf("error should be occurred")
				}

				if err := compareVal(got, tt.want); err != nil {
					t.Errorf(err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := isZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Errorf(err.Error())
			}

			if err := isNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func setZero_OnBuilderListWithBluePrint(t *testing.T) {
	bluePrint := func(i int, val testStruct) testStruct {
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
			Struct:         subStruct{SubID: i, SubName: str},
			PtrStruct:      &subStruct{SubID: i, SubName: str},
			Slice:          []int{i, i + 1, i + 2},
			PtrSlice:       []*int{&i, &i, &i},
			SliceStruct:    []subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
			SlicePtrStruct: []*subStruct{{SubID: i, SubName: str}, {SubID: i + 1, SubName: str}},
		}
	}
	f := New(testStruct{}).SetConfig(Config[testStruct]{BluePrint: bluePrint})

	tests := []struct {
		desc              string
		index             int
		setZeroFields     []string
		wantZeroFields    [][]string
		wantNonZeroFields [][]string
		hasErro           bool
		want              []testStruct
	}{
		{
			desc:          "set zero values at valid index",
			index:         1,
			setZeroFields: []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
			wantZeroFields: [][]string{
				{""},
				{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
			},
			wantNonZeroFields: [][]string{
				{"Int", "PtrInt", "Str", "PtrStr", "Bool", "PtrBool", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
				filterFields(testStruct{}, "Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"),
			},
			hasErro: false,
		},
		{
			desc:    "set zero values at negative index",
			index:   -1,
			hasErro: true,
			want:    []testStruct{},
		},
		{
			desc:    "set zero values at invalid index",
			index:   5,
			hasErro: true,
			want:    []testStruct{},
		},
		{
			desc:          "set incorrect field",
			index:         0,
			setZeroFields: []string{"incorrect field"},
			hasErro:       true,
			want:          []testStruct{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).SetZero(tt.index, tt.setZeroFields...).Get()

			if tt.hasErro {
				if err == nil {
					t.Errorf("error should be occurred")
				}

				if err := compareVal(got, tt.want); err != nil {
					t.Errorf(err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			for i, g := range got {
				if err := isZeroVal(g, tt.wantNonZeroFields[i]...); err != nil {
					t.Errorf(err.Error())
				}

				if err := isNotZeroVal(g, tt.wantZeroFields[i]...); err != nil {
					t.Errorf(err.Error())
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
		hasErro           bool
		want              []testStruct
	}{
		{
			desc:          "set zero values at valid index",
			index:         1,
			setZeroFields: []string{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Interface", "Struct", "PtrStruct", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
			wantZeroFields: [][]string{
				{"Interface"},
				{"Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Struct", "PtrStruct", "Interface", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"},
			},
			wantNonZeroFields: [][]string{
				filterFields(testStruct{}, "Interface"),
				filterFields(testStruct{}, "Int", "PtrInt", "Time", "PtrTime", "Float", "PtrFloat", "Struct", "PtrStruct", "Interface", "Slice", "PtrSlice", "SliceStruct", "SlicePtrStruct"),
			},
			hasErro: false,
		},
		{
			desc:    "set zero values at negative index",
			index:   -1,
			hasErro: true,
			want:    []testStruct{},
		},
		{
			desc:    "set zero values at invalid index",
			index:   5,
			hasErro: true,
			want:    []testStruct{},
		},
		{
			desc:          "set incorrect field",
			index:         0,
			setZeroFields: []string{"incorrect field"},
			hasErro:       true,
			want:          []testStruct{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := f.BuildList(mockCTX, 2).SetZero(tt.index, tt.setZeroFields...).Get()

			if tt.hasErro {
				if err == nil {
					t.Errorf("error should be occurred")
				}

				if err := compareVal(got, tt.want); err != nil {
					t.Errorf(err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			for i, g := range got {
				if err := isZeroVal(g, tt.wantNonZeroFields[i]...); err != nil {
					t.Errorf(err.Error())
				}

				if err := isNotZeroVal(g, tt.wantZeroFields[i]...); err != nil {
					t.Errorf(err.Error())
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
		hasError          bool
		want              testStruct
	}{
		{
			desc: "set two zero values",
			setZeroFields: [][]string{
				{"Int"},
				{"Slice"},
			},
			wantZeroFields:    []string{"Int", "Slice", "Interface"},
			wantNonZeroFields: filterFields(testStruct{}, "Int", "Slice", "Interface"),
			hasError:          false,
		},
		{
			desc: "set three zero values",
			setZeroFields: [][]string{
				{"Int"},
				{"Slice"},
				{"Struct"},
			},
			wantZeroFields:    []string{"Int", "Slice", "Struct", "Interface"},
			wantNonZeroFields: filterFields(testStruct{}, "Int", "Slice", "Struct", "Interface"),
			hasError:          false,
		},
		{
			desc:          "set incorrect field",
			setZeroFields: [][]string{{"incorrect field"}},
			hasError:      true,
			want:          testStruct{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			preF := f.Build(mockCTX)
			for _, fields := range tt.setZeroFields {
				preF = preF.SetZero(fields...)
			}

			got, err := preF.Get()

			if tt.hasError {
				if err == nil {
					t.Errorf("error should be occurred")
				}

				if err := compareVal(got, tt.want); err != nil {
					t.Errorf(err.Error())
				}

				return
			}

			// TODO: the error message should be more specific
			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			if err := isZeroVal(got, tt.wantNonZeroFields...); err != nil {
				t.Errorf(err.Error())
			}

			if err := isNotZeroVal(got, tt.wantZeroFields...); err != nil {
				t.Errorf(err.Error())
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
		hasError             bool
	}{
		{
			desc:       "set zero values at valid index",
			buildIndex: 3,
			setZeroFieldsByIndex: map[int][]string{
				0: {"Int"},
				1: {"PtrSlice", "SlicePtrStruct"},
			},
			wantZeroFields: [][]string{
				{"Int", "Interface"},
				{"PtrSlice", "SlicePtrStruct", "Interface"},
				{"Interface"},
			},
			wantNonZeroFields: [][]string{
				filterFields(testStruct{}, "Int", "Interface"),
				filterFields(testStruct{}, "PtrSlice", "SlicePtrStruct", "Interface"),
				filterFields(testStruct{}, "Interface"),
			},
			hasError: false,
		},
		{
			desc:       "set zero values at negative index",
			buildIndex: 3,
			setZeroFieldsByIndex: map[int][]string{
				-1: {"Int"},
				0:  {"PtrSlice", "SlicePtrStruct"},
			},
			want:     []testStruct{},
			hasError: true,
		},
		{
			desc:       "set zero values at invalid index",
			buildIndex: 3,
			setZeroFieldsByIndex: map[int][]string{
				5: {"Int"},
				0: {"PtrSlice", "SlicePtrStruct"},
			},
			want:     []testStruct{},
			hasError: true,
		},
		{
			desc:       "set incorrect field",
			buildIndex: 3,
			setZeroFieldsByIndex: map[int][]string{
				0: {"incorrect field"},
			},
			want:     []testStruct{},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			preF := f.BuildList(mockCTX, tt.buildIndex)
			for i, fields := range tt.setZeroFieldsByIndex {
				preF = preF.SetZero(i, fields...)
			}

			got, err := preF.Get()

			if tt.hasError {
				if err == nil {
					t.Errorf("error should be occurred")
				}

				if err := compareVal(got, tt.want); err != nil {
					t.Errorf(err.Error())
				}

				return
			}

			if err != nil {
				t.Errorf("error from gofacto: %v", err)
			}

			for i, g := range got {
				if err := isZeroVal(g, tt.wantNonZeroFields[i]...); err != nil {
					t.Errorf(err.Error())
				}

				if err := isNotZeroVal(g, tt.wantZeroFields[i]...); err != nil {
					t.Errorf(err.Error())
				}
			}
		})
	}
}
