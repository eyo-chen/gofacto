// Package example_test provides examples demonstrating the usage of the gofacto library.
//
// These examples are for demonstration purposes and are not meant to be run directly.
// They serve as reference implementations to illustrate various features and use cases of gofacto.
// In a real-world scenario, you would need to provide actual database connections and adjust the code accordingly.
package example_test

import (
	"fmt"

	"github.com/eyo-chen/gofacto"
	"github.com/eyo-chen/gofacto/db/mysqlf"
)

// Example_setzero demonstrates how to explicitly set zero value on the specified fields.
func Example_setzero() {
	f := gofacto.New(customer{}).
		WithDB(mysqlf.NewConfig(nil)) // you should pass db connection

	// build one data, and set zero value on the specified fields
	customer, err := f.Build(ctx).SetZero("Age", "Name").Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customer) // {ID: 1, Gender: "", Name: "", Age: 0}
	// Note that because Gender is a custom defined type, gofacto will ignore it, and leave it as zero value..

	// build two data, and set zero value on the specified fields
	// the first parameter is the index of the list
	// the second parameter is the field name that you want to set zero value on
	customers, err := f.BuildList(ctx, 2).SetZero(0, "Age").SetZero(1, "Age", "Name").Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customers) // [{ID: 2, Gender: "", Name: {{non-zero value}}, Age: 0}, {ID: 3, Gender: "", Name: "", Age: 0}]
}
