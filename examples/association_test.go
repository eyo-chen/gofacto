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
	"github.com/eyo-chen/gofacto/typeconv"
)

// order has a foreign key field `CustomerID` to customer
type order struct {
	ID         int
	CustomerID int `gofacto:"struct:customer"` // set the correct tag
	Amount     int
}

// Example_association demonstrates how to build associations value in an easy way.
// First, we need to set the correct tag on the foreign key field to tell gofacto which struct to associate with.
// Then, we can use `WithOne` or `WithMany` to create the association value and set the connection between the two structs.
func Example_association() {
	// init a oreder factory
	f := gofacto.New(order{}).
		WithDB(mysqlf.NewConfig(nil)) // you should pass db connection

	// build one order with one customer
	customer1 := customer{}
	order1, err := f.Build(ctx).
		WithOne(&customer1). // must pass the struct pointer to WithOne or WithMany
		Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(order1)                            // {ID: 1, CustomerID: 1, Amount: {{non-zero value}}}
	fmt.Println(customer1)                         // {ID: 1, Gender: "", Name: {{non-zero value}}, Age: {{non-zero value}}}
	fmt.Println(order1.CustomerID == customer1.ID) // true

	// build two orders with two customers
	customer2 := customer{}
	customer3 := customer{}
	orders1, err := f.BuildList(ctx, 2).
		WithMany([]interface{}{&customer2, &customer3}). // must pass the struct pointer to WithOne or WithMany
		Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(orders1[0])                            // {ID: 2, CustomerID: 2, Amount: {{non-zero value}}}
	fmt.Println(orders1[1])                            // {ID: 3, CustomerID: 3, Amount: {{non-zero value}}}
	fmt.Println(customer2)                             // {ID: 2, Gender: "", Name: {{non-zero value}}, Age: {{non-zero value}}}
	fmt.Println(customer3)                             // {ID: 3, Gender: "", Name: {{non-zero value}}, Age: {{non-zero value}}}
	fmt.Println(orders1[0].CustomerID == customer2.ID) // true
	fmt.Println(orders1[1].CustomerID == customer3.ID) // true

	// build two orders with one customer
	customer4 := customer{}
	orders2, err := f.BuildList(ctx, 2).
		WithOne(&customer4). // must pass the struct pointer to WithOne or WithMany
		Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(orders2[0])                            // {ID: 4, CustomerID: 4, Amount: {{non-zero value}}}
	fmt.Println(orders2[1])                            // {ID: 5, CustomerID: 4, Amount: {{non-zero value}}}
	fmt.Println(customer4)                             // {ID: 4, Gender: "", Name: {{non-zero value}}, Age: {{non-zero value}}}
	fmt.Println(orders2[0].CustomerID == customer4.ID) // true
	fmt.Println(orders2[1].CustomerID == customer4.ID) // true
}

// InsertOrders demonstrates how to use the functionality of `typeconv` package to simplify the code.
// In some cases, we might want to wrap the insert logic into a function.
// In this case, we define the `InsertOrders` function to insert `n` orders with `n` customers.
func InsertOrders(ctx context.Context, f *gofacto.Factory[order], n int) ([]order, []customer, error) {
	// use `ToAnysWithOW` to generate `n` customers with any type
	// The first parameter is the number of customers to generate
	// The second parameter is the value to override the default value(we pass nil because we don't want to override the default value)
	customersAny := typeconv.ToAnysWithOW[customer](n, nil)

	orders, err := f.BuildList(ctx, n).
		WithMany(customersAny).
		Insert()
	if err != nil {
		return nil, nil, err
	}

	// convert the `[]any` to `[]customer` using `ToT`
	customers := typeconv.ToT[customer](customersAny)

	return orders, customers, nil
}

// Without the `typeconv` package, we would need to manually convert the `[]any` to `[]customer` using `ToT`
func InsertOrdersWithoutTypeconv(ctx context.Context, f *gofacto.Factory[order], n int) ([]order, []customer, error) {
	// manually create `n` customers with any type
	customersAny := make([]interface{}, n)
	for i := 0; i < n; i++ {
		customersAny[i] = &customer{}
	}

	orders, err := f.BuildList(ctx, n).
		WithMany(customersAny).
		Insert()
	if err != nil {
		return nil, nil, err
	}

	// manually convert the `[]any` to `[]customer`
	customers := make([]customer, n)
	for i := 0; i < n; i++ {
		customers[i] = *customersAny[i].(*customer)
	}

	return orders, customers, nil
}
