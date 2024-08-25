// Package example_test provides examples demonstrating the usage of the gofacto library.
//
// These examples are for demonstration purposes and are not meant to be run directly.
// They serve as reference implementations to illustrate various features and use cases of gofacto.
// In a real-world scenario, you would need to provide actual database connections and adjust the code accordingly.
package example_test

import (
	"context"
	"fmt"

	"github.com/eyo-chen/gofacto"
	"github.com/eyo-chen/gofacto/db/mysqlf"
)

var (
	ctx = context.Background()
)

// Gender is a custom defined type
type gender string

const (
	genderMale   gender = "male"
	genderFemale gender = "female"
)

type customer struct {
	ID     int
	Gender gender
	Name   string
	Age    int
}

// Example_basic demonstrates the most basic usage of gofacto
// when we didn't provide any configuration(except db connection), gofacto will generate non-zero value for all fields
func Example_basic() {
	f := gofacto.New(customer{}).
		WithDB(mysqlf.NewConfig(nil)) // you should pass db connection

	// build and insert one data
	customer, err := f.Build(ctx).Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customer) // {ID: 1, Gender: "", Name: {{non-zero value}}, Age: {{non-zero value}}}
	// Note that because Gender is a custom defined type, gofacto will ignore it, and leave it as zero value.

	// build and insert multiple data
	customers, err := f.BuildList(ctx, 2).Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customers[0]) // {ID: 2, Gender: "", Name: {{non-zero value}}, Age: {{non-zero value}}}
	fmt.Println(customers[1]) // {ID: 3, Gender: "", Name: {{non-zero value}}, Age: {{non-zero value}}}
}
