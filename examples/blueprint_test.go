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

// Example_blueprint demonstrates how to use blueprint to generate data.
// When blueprint is provided, gofacto will first use the blueprint to generate data,
// then it will generate non-zero value for the rest of the fields which are not specified in the blueprint
func Example_blueprint() {
	f := gofacto.New(customer{}).
		WithBlueprint(blueprint).     // set the blueprint function
		WithDB(mysqlf.NewConfig(nil)) // you should pass db connection

	customer, err := f.Build(ctx).Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customer) // {ID: 1, Gender: "male", Name: "Bob", Age: 10}

	customers, err := f.BuildList(ctx, 2).Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customers[0]) // {ID: 2, Gender: "male", Name: "Bob", Age: 20}
	fmt.Println(customers[1]) // {ID: 3, Gender: "male", Name: "Bob", Age: 30}
}

// Define a blueprint function with the following signature:
// type blueprintFunc[T any] func(i int) T
func blueprint(i int) customer {
	return customer{
		Name:   "Bob",
		Gender: genderMale,
		Age:    i * 10,
	}
}
