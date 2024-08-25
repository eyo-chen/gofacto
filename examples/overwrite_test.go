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

// Example_overwrite demonstrates how to overwrite specific fields
func Example_overwrite() {
	f := gofacto.New(customer{}).
		WithDB(mysqlf.NewConfig(nil)) // you should pass db connection

	// build one data, and overwrite the Name field
	customer1, err := f.Build(ctx).Overwrite(customer{Name: "Alice"}).Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customer1) // {ID: 1, Gender: "", Name: "Alice", Age: {{non-zero value}}}
	// Note that because Gender is a custom defined type, gofacto will ignore it, and leave it as zero value.

	// build two data, and overwrite the Gender and Name fields respectively
	customers, err := f.BuildList(ctx, 2).
		Overwrites(customer{Gender: genderMale, Name: "Alice"}, customer{Gender: genderFemale, Name: "Mag"}).
		Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customers[0]) // {ID: 2, Gender: "male", Name: "Alice", Age: {{non-zero value}}}
	fmt.Println(customers[1]) // {ID: 3, Gender: "female", Name: "Mag", Age: {{non-zero value}}}

	// build two data, and overwrite all the data list with the same value
	customers1, err := f.BuildList(ctx, 2).
		Overwrite(customer{Age: 10}).
		Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customers1[0]) // {ID: 4, Gender: "", Name: {{non-zero value}}, Age: 10}
	fmt.Println(customers1[1]) // {ID: 5, Gender: "", Name: {{non-zero value}}, Age: 10}

	// Note that we can't overwrite with zero value
	// use `SetZero` or `SetTrait` to set zero value
	customer2, err := f.Build(ctx).Overwrite(customer{Age: 0}).Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customer2) // {ID: 4, Gender: "", Name: {{non-zero value}}, Age: {{non-zero value}}}
	// Age is still a non-zero value
}
