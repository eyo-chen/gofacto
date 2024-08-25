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

// Example_settrait demonstrates how to use trait function to customize the data.
// We first define a trait function, then use `WithTrait` to register the trait function with a key,
// finally, use `SetTrait` to invoke the trait function.
func Example_settrait() {
	f := gofacto.New(customer{}).
		WithTrait("toMale", updateToMale).     // set the trait function with key
		WithTrait("toFemale", updateToFemale). // set the trait function with key
		WithTrait("ageZero", updateAgeToZero). // set the trait function with key
		WithDB(mysqlf.NewConfig(nil))          // you should pass db connection

	// build one data, and invoke the trait function with key "toMale"
	customer1, err := f.Build(ctx).SetTrait("toMale").Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customer1) // {ID: 1, Gender: "male", Name: {{non-zero value}, Age: {{non-zero value}}}

	// build two data, and invoke the trait function with key "toMale" and "toFemale" respectively
	customers, err := f.BuildList(ctx, 2).SetTraits("toMale", "toFemale").Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customers[0]) // {ID: 2, Gender: "male", Name: {{non-zero value}, Age: {{non-zero value}}}
	fmt.Println(customers[1]) // {ID: 3, Gender: "female", Name: {{non-zero value}, Age: {{non-zero value}}}

	// build two data, and invoke the trait function with key "toMale" on all the data list
	customers1, err := f.BuildList(ctx, 2).SetTrait("toMale").Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customers1[0]) // {ID: 4, Gender: "male", Name: {{non-zero value}, Age: {{non-zero value}}}
	fmt.Println(customers1[1]) // {ID: 5, Gender: "male", Name: {{non-zero value}, Age: {{non-zero value}}}

	// it's also possible to set zero value
	customer2, err := f.Build(ctx).SetTrait("ageZero").Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(customer2) // {ID: 6, Gender: "", Name: {{non-zero value}}, Age: 0}
}

// define a trait function to update the value
func updateToMale(c *customer) {
	c.Gender = genderMale
	// more update logic can be added here
}

// define a trait function to update the value
func updateToFemale(c *customer) {
	c.Gender = genderFemale
	// more update logic can be added here
}

// define a trait function to update the value
func updateAgeToZero(c *customer) {
	c.Age = 0
	// more update logic can be added here
}
