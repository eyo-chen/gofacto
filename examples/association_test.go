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

// expense has two foreign key fields `UserID` and `CategoryID` to `User` and `Category` structs
type expense struct {
	ID         int
	UserID     int `gofacto:"foreignKey,struct:User"`
	CategoryID int `gofacto:"foreignKey,struct:Category,table:categories"`
}

// category has a foreign key field `UserID` to `User` struct
type category struct {
	ID     int
	UserID int `gofacto:"foreignKey,struct:User"`
}

type user struct {
	ID int
}

// Example_association_basic demonstrates how to build basic associations value in an easy way.
// First, we need to set the correct tag on the foreign key field to tell gofacto which struct to associate with.
// Then, we can use `WithOne` or `WithMany` to create the association value and set the connection between the two structs.
func Example_association_basic() {
	f := gofacto.New(category{}).
		WithDB(mysqlf.NewConfig(nil)) // you should pass db connection

	// build one category with one user
	user1 := user{}
	category1, err := f.Build(ctx).
		WithOne(&user1). // must pass the struct pointer to WithOne or WithMany
		Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(category1.UserID == user1.ID) // true

	// build two categories with two users
	user2 := user{}
	user3 := user{}
	categories1, err := f.BuildList(ctx, 2).
		WithMany([]interface{}{&user2, &user3}). // must pass the struct pointer to WithOne or WithMany
		Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(categories1[0].UserID == user2.ID) // true
	fmt.Println(categories1[1].UserID == user3.ID) // true

	// build two categories with one user
	user4 := user{}
	categories2, err := f.BuildList(ctx, 2).
		WithOne(&user4). // must pass the struct pointer to WithOne or WithMany
		Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(categories2[0].UserID == user4.ID) // true
	fmt.Println(categories2[1].UserID == user4.ID) // true
}

// Example_association_advanced demonstrates how to build advanced associations value in an easy way.
// In this example, we will build the expense with user and category, and the category is associated with user.
func Example_association_advanced() {
	f := gofacto.New(expense{}).
		WithDB(mysqlf.NewConfig(nil)) // you should pass db connection

	// build one expense with one user and one category
	user1 := user{}
	category1 := category{}
	expense1, err := f.Build(ctx).
		WithOne(&user1).
		WithOne(&category1).
		Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(expense1.UserID == user1.ID)         // true
	fmt.Println(expense1.CategoryID == category1.ID) // true
	fmt.Println(category1.UserID == user1.ID)        // true
	// You can also use .WithOne(&user1, &category1) to pass multiple structs to WithOne

	// build two expenses with two users and two categories
	user2 := user{}
	user3 := user{}
	category2 := category{}
	category3 := category{}
	expenses1, err := f.BuildList(ctx, 2).
		WithMany([]interface{}{&user2, &user3}). // must pass same type of structs to WithMany
		WithMany([]interface{}{&category2, &category3}).
		Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(expenses1[0].UserID == user2.ID)         // true
	fmt.Println(expenses1[1].UserID == user3.ID)         // true
	fmt.Println(expenses1[0].CategoryID == category2.ID) // true
	fmt.Println(expenses1[1].CategoryID == category3.ID) // true
	fmt.Println(category2.UserID == user2.ID)            // true
	fmt.Println(category3.UserID == user3.ID)            // true

	// build two expenses with one user and two categories
	user5 := user{}
	category4 := category{}
	category5 := category{}
	expenses2, err := f.BuildList(ctx, 2).
		WithOne(&user5).
		WithMany([]interface{}{&category4, &category5}).
		Insert()
	if err != nil {
		panic(err)
	}
	fmt.Println(expenses2[0].UserID == user5.ID)         // true
	fmt.Println(expenses2[1].UserID == user5.ID)         // true
	fmt.Println(expenses2[0].CategoryID == category4.ID) // true
	fmt.Println(expenses2[1].CategoryID == category5.ID) // true
	fmt.Println(category4.UserID == user5.ID)            // true
	fmt.Println(category5.UserID == user5.ID)            // true
}

// InsertCategories demonstrates how to use the functionality of `typeconv` package to simplify the code.
// In some cases, we might want to wrap the insert logic into a function.
// In this case, we define the `InsertCategories` function to insert `n` categories with `n` users.
func InsertCategories(ctx context.Context, f *gofacto.Factory[category], n int) ([]category, []user, error) {
	// use `ToAnysWithOW` to generate `n` users with any type
	// The first parameter is the number of users to generate
	// The second parameter is the value to override the default value(we pass nil because we don't want to override the default value)
	usersAny := typeconv.ToAnysWithOW[user](n, nil)

	categories, err := f.BuildList(ctx, n).
		WithMany(usersAny).
		Insert()
	if err != nil {
		return nil, nil, err
	}

	// convert the `[]any` to `[]user` using `ToT`
	users := typeconv.ToT[user](usersAny)

	return categories, users, nil
}

// Without the `typeconv` package, we would need to manually convert the `[]any` to `[]user` using `ToT`
func InsertCategoriesWithoutTypeconv(ctx context.Context, f *gofacto.Factory[category], n int) ([]category, []user, error) {
	// manually create `n` users with any type
	usersAny := make([]interface{}, n)
	for i := 0; i < n; i++ {
		usersAny[i] = &user{}
	}

	categories, err := f.BuildList(ctx, n).
		WithMany(usersAny).
		Insert()
	if err != nil {
		return nil, nil, err
	}

	// manually convert the `[]any` to `[]user`
	users := make([]user, n)
	for i := 0; i < n; i++ {
		users[i] = *usersAny[i].(*user)
	}

	return categories, users, nil
}
