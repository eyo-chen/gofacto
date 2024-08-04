package utils

import "testing"

type Person struct {
	ID   int
	Name string
}

func TestCvtToAnysWithOW(t *testing.T) {
	tests := []struct {
		desc string
		i    int
		ow   *Person
		want []interface{}
	}{
		{
			desc: "i is 1 and ow is nil",
			i:    1,
			ow:   nil,
			want: []interface{}{&Person{}},
		},
		{
			desc: "i is 1 and ow is not nil",
			i:    1,
			ow:   &Person{ID: 1, Name: "foo"},
			want: []interface{}{&Person{ID: 1, Name: "foo"}},
		},
		{
			desc: "i is 3 and ow is nil",
			i:    3,
			ow:   nil,
			want: []interface{}{&Person{}, &Person{}, &Person{}},
		},
		{
			desc: "i is 3 and ow is not nil",
			i:    3,
			ow:   &Person{ID: 1, Name: "foo"},
			want: []interface{}{&Person{ID: 1, Name: "foo"}, &Person{ID: 1, Name: "foo"}, &Person{ID: 1, Name: "foo"}},
		},
		{
			desc: "i is 0 and ow is nil",
			i:    0,
			ow:   nil,
			want: []interface{}{},
		},
		{
			desc: "i is 0 and ow is not nil",
			i:    0,
			ow:   &Person{ID: 1, Name: "foo"},
			want: []interface{}{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			res := CvtToAnysWithOW(test.i, test.ow)

			if len(res) != len(test.want) {
				t.Errorf("expected length of %d, but got %d", len(test.want), len(res))
			}

			for k, v := range res {
				if *v.(*Person) != *test.want[k].(*Person) {
					t.Errorf("expected %v, but got %v", *test.want[k].(*Person), *v.(*Person))
				}
			}
		})
	}
}

func TestCvtToAnysWithOWs(t *testing.T) {
	tests := []struct {
		desc string
		i    int
		ows  []*Person
		want []interface{}
	}{
		{
			desc: "i is 1 and ows is nil",
			i:    1,
			ows:  nil,
			want: []interface{}{&Person{}},
		},
		{
			desc: "i is 1 and ows is not nil",
			i:    1,
			ows:  []*Person{{ID: 1, Name: "foo"}},
			want: []interface{}{&Person{ID: 1, Name: "foo"}},
		},
		{
			desc: "i is 3 and ows is nil",
			i:    3,
			ows:  nil,
			want: []interface{}{&Person{}, &Person{}, &Person{}},
		},
		{
			desc: "i is 3 and ows is not nil",
			i:    3,
			ows:  []*Person{{ID: 1, Name: "foo"}, {ID: 2, Name: "bar"}, {ID: 3, Name: "baz"}},
			want: []interface{}{&Person{ID: 1, Name: "foo"}, &Person{ID: 2, Name: "bar"}, &Person{ID: 3, Name: "baz"}},
		},
		{
			desc: "i is greater than len(ows)",
			i:    3,
			ows:  []*Person{{ID: 1, Name: "foo"}},
			want: []interface{}{&Person{ID: 1, Name: "foo"}, &Person{}, &Person{}},
		},
		{
			desc: "i is less than len(ows)",
			i:    1,
			ows:  []*Person{{ID: 1, Name: "foo"}, {ID: 2, Name: "bar"}},
			want: []interface{}{&Person{ID: 1, Name: "foo"}},
		},
		{
			desc: "i is 0 and ows is nil",
			i:    0,
			ows:  nil,
			want: []interface{}{},
		},
		{
			desc: "i is 0 and ows is not nil",
			i:    0,
			ows:  []*Person{{ID: 1, Name: "foo"}},
			want: []interface{}{},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			res := CvtToAnysWithOWs(test.i, test.ows...)

			if len(res) != len(test.want) {
				t.Errorf("expected length of %d, but got %d", len(test.want), len(res))
			}

			for k, v := range res {
				if *v.(*Person) != *test.want[k].(*Person) {
					t.Errorf("expected %v, but got %v", *test.want[k].(*Person), *v.(*Person))
				}
			}
		})
	}
}

func TestCvtToT(t *testing.T) {
	tests := []struct {
		desc string
		vals []interface{}
		want []Person
	}{
		{
			desc: "vals is nil",
			vals: nil,
			want: nil,
		},
		{
			desc: "vals is empty",
			vals: []interface{}{},
			want: []Person{},
		},
		{
			desc: "vals is not empty",
			vals: []interface{}{&Person{ID: 1, Name: "foo"}, &Person{ID: 2, Name: "bar"}, &Person{ID: 3, Name: "baz"}},
			want: []Person{{ID: 1, Name: "foo"}, {ID: 2, Name: "bar"}, {ID: 3, Name: "baz"}},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			res := CvtToT[Person](test.vals)

			if len(res) != len(test.want) {
				t.Errorf("expected length of %d, but got %d", len(test.want), len(res))
			}

			for k, v := range res {
				if v != test.want[k] {
					t.Errorf("expected %v, but got %v", test.want[k], v)
				}
			}
		})
	}
}

func TestCvtToPointerT(t *testing.T) {
	tests := []struct {
		desc string
		vals []interface{}
		want []*Person
	}{
		{
			desc: "vals is nil",
			vals: nil,
			want: nil,
		},
		{
			desc: "vals is empty",
			vals: []interface{}{},
			want: []*Person{},
		},
		{
			desc: "vals is not empty",
			vals: []interface{}{&Person{ID: 1, Name: "foo"}, &Person{ID: 2, Name: "bar"}, &Person{ID: 3, Name: "baz"}},
			want: []*Person{{ID: 1, Name: "foo"}, {ID: 2, Name: "bar"}, {ID: 3, Name: "baz"}},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			res := CvtToPointerT[Person](test.vals)

			if len(res) != len(test.want) {
				t.Errorf("expected length of %d, but got %d", len(test.want), len(res))
			}

			for k, v := range res {
				if *v != *test.want[k] {
					t.Errorf("expected %v, but got %v", *test.want[k], *v)
				}
			}
		})
	}
}
