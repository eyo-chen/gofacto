[![Go Reference](https://pkg.go.dev/badge/github.com/eyo-chen/gofacto.svg)](https://pkg.go.dev/github.com/eyo-chen/gofacto)
[![Go Report Card](https://goreportcard.com/badge/github.com/eyo-chen/gofacto)](https://goreportcard.com/report/github.com/eyo-chen/gofacto)
[![Coverage Status](https://coveralls.io/repos/github/eyo-chen/gofacto/badge.svg?branch=main)](https://coveralls.io/github/eyo-chen/gofacto?branch=main)

# gofacto

gofacto is a strongly-typed and user-friendly factory library for Go, designed to simplify the creation of mock data. It offers:

- Intuitive and straightforward usage
- Strong typing and type safety
- Flexible data customization
- Support for various databases and ORMs
- Basic and multi-level association relationship support

&nbsp;

# Installation
```bash
go get github.com/eyo-chen/gofacto
```

&nbsp;

# Quick Start
You can find more examples in the [examples](https://github.com/eyo-chen/gofacto/tree/main/examples) folder.

Let's consider two structs: `Customer` and `Order`. `Order` struct has a foreign key `CustomerID` because a customer can have many orders.
```go
type Customer struct {
    ID      int
    Gender  Gender     // custom defined type
    Name    string
    Email   *string
    Phone   string
}

type Order struct {
    ID          int
    CustomerID  int       `gofacto:"foreignKey,struct:Customer"`
    OrderDate   time.Time
    Amount      float64
}

// Init a Customer factory
customerFactory := gofacto.New(Customer{}).
                           WithDB(mysqlf.NewConfig(db)) // Assuming db is a database connection

// Create and insert a customer
customer, err := customerFactory.Build(ctx).Insert()

// Create and insert two female customers
customers, err := customerFactory.BuildList(ctx, 2).
                                  Overwrite(Customer{Gender: Female}).
                                  Insert()
// customers[0].Gender == Female
// customers[1].Gender == Female

// Init a Order factory
orderFactory := gofacto.New(Order{}).
                        WithDB(mysqlf.NewConfig(db))

// Create and insert an order with one customer
c := Customer{}
order, err := orderFactory.Build(ctx).
                           WithOne(&c).
                           Insert()
// order.CustomerID == c.ID

// Create and insert two orders with two customers
c1, c2 := Customer{}, Customer{}
orders, err := orderFactory.BuildList(ctx, 2).
                            WithMany([]interface{}{&c1, &c2}).
                            Insert()
// orders[0].CustomerID == c1.ID
// orders[1].CustomerID == c2.ID
```
Note: All fields in the returned struct are populated with non-zero values, and `ID` field is auto-incremented by the database.

&nbsp;

# Usage
### Initialize
Use `New` to initialize the factory by passing the struct you want to create mock data for.
```go
factory := gofacto.New(Order{})
```

### Build & BuildList
Use `Build` to create a single value, and `BuildList` to create a list of values.
```go
order, err := factory.Build(ctx).Get()
orders, err := factory.BuildList(ctx, 2).Get()
```
`Get` method returns the struct(s) without inserting them into the database. All fields are populated with non-zero values.

### Insert
Use `Insert` to insert values into the database.<br>
`Insert` method inserts the struct into the database and returns the struct with `ID` field populated with the auto-incremented value.<br>
```go
order, err := factory.Build(ctx).Insert()
orders, err := factory.BuildList(ctx, 2).Insert()
```
Find out more [examples](https://github.com/eyo-chen/gofacto/blob/main/examples/basic_test.go).

### Overwrite
Use `Overwrite` to set specific fields.<br>
The fields in the struct will be used to overwrite the fields in the generated struct.
```go
order, err := factory.Build(ctx).Overwrite(Order{Amount: 100}).Insert()
// order.Amount == 100
```

When building a list of values, `Overwrite` is used to overwrite all the list of values, and `Overwrites` is used to overwrite each value in the list of values.
```go
orders, err := factory.BuildList(ctx, 2).Overwrite(Order{Amount: 100}).Insert()
// orders[0].Amount == 100
// orders[1].Amount == 100

orders, err := factory.BuildList(ctx, 2).Overwrites(Order{Amount: 100}, Order{Amount: 200}).Insert()
// orders[0].Amount == 100
// orders[1].Amount == 200
```

Note: Explicit zero values are not overwritten by default. Use `SetZero` or `SetTrait` for this purpose.
```go
order, err := factory.Build(ctx).Overwrite(Order{Amount: 0}).Insert()
// order.Amount != 0
```

Find out more [examples](https://github.com/eyo-chen/gofacto/blob/main/examples/overwrite_test.go).


### SetTrait
When initializing the factory, use `WithTrait` method to set the trait functions and the corresponding keys. Then use `SetTrait` method to apply the trait functions when building the struct.
```go
func setFemale(c *Customer) {
  c.Gender = Female
}
factory := gofacto.New(Order{}).
                   WithTrait("female", setFemale)

customer, err := factory.Build(ctx).SetTrait("female").Insert()
// customer.Gender == Female
```


When building a list of values, `SetTrait` is used to apply one trait function to all the list of values, and `SetTraits` is used to apply multiple trait functions to the list of values.
```go
func setFemale(c *Customer) {
  c.Gender = Female
}
func setMale(c *Customer) {
  c.Gender = Male
}
factory := gofacto.New(Order{}).
                   WithTrait("female", setFemale).
                   WithTrait("male", setMale)

customers, err := factory.BuildList(ctx, 2).SetTrait("female").Insert()
// customers[0].Gender == Female
// customers[1].Gender == Female

customers, err := factory.BuildList(ctx, 2).SetTraits("female", "male").Insert()
// customers[0].Gender == Female
// customers[1].Gender == Male
```

Find out more [examples](https://github.com/eyo-chen/gofacto/blob/main/examples/settrait_test.go).

### SetZero
Use `SetZero` to set specific fields to zero values.<br>
`SetZero` method with `Build` accepts multiple string as the field names, and the fields will be set to zero values when building the struct.<br>
`SetZero` method with `BuildList` accepts an index and multiple string as the field names, and the fields will be set to zero values when building the struct at the index.
```go
customer, err := factory.Build(ctx).SetZero("Email", "Phone").Insert()
// customer.Email == nil
// customer.Phone == ""

customers, err := factory.BuildList(ctx, 2).SetZero(0, "Email", "Phone").Insert()
// customers[0].Email == nil
// customers[0].Phone == ""
// customers[1].Email != nil
// customers[1].Phone != ""
```

Find out more [examples](https://github.com/eyo-chen/gofacto/blob/main/examples/setzero_test.go).

### WithOne & WithMany
When there is the associations relationship between the structs, use `WithOne` and `WithMany` methods to build the associated structs.<br>
Before using `WithOne` and `WithMany` methods, make sure setting the correct tag in the struct.
```go
type Order struct {
  ID          int
  CustomerID  int       `gofacto:"foreignKey,struct:Customer"`
  OrderDate   time.Time
  Amount      float64
}
```
You can find more details about the tag format in [foreignKey tag](#foreignkey-tag).

```go
// build an order with one customer
c := Customer{}
order, err := factory.Build(ctx).WithOne(&c).Insert()
// order.CustomerID == c.ID

// build two orders with two customers
c1 := Customer{}
c2 := Customer{}
orders, err := factory.BuildList(ctx, 2).WithMany([]interface{}{&c1, &c2}).Insert()
// orders[0].CustomerID == c1.ID
// orders[1].CustomerID == c2.ID

// build an order with only one customer
c1 := Customer{}
orders, err := factory.BuildList(ctx, 2).WithOne(&c1).Insert()
// orders[0].CustomerID == c1.ID
// orders[1].CustomerID == c1.ID
```

If there are multiple level association relationships, both `WithOne` and `WithMany` methods can also come in handy.<br>
Suppose we have a following schema:
```go
type Expense struct {
	ID         int
	UserID     int `gofacto:"foreignKey,struct:User"`
	CategoryID int `gofacto:"foreignKey,struct:Category,table:categories"`
}

type Category struct {
	ID     int
	UserID int `gofacto:"foreignKey,struct:User"`
}

type User struct {
	ID int
}
```

We can build the `Expense` struct with the associated `User` and `Category` structs by using `WithOne` and `WithMany` methods.
```go
// build one expense with one user and one category
user := User{}
category := Category{}
expense, err := factory.Build(ctx).WithOne(&user).WithOne(&category).Insert()
// expense.UserID == user.ID
// expense.CategoryID == category.ID
// category.UserID == user.ID

// build two expenses with two users and two categories
user1 := User{}
user2 := User{}
category1 := Category{}
category2 := Category{}
expenses, err := factory.BuildList(ctx, 2).WithMany([]interface{}{&user1, &user2}).WithMany([]interface{}{&category1, &category2}).Insert()
// expenses[0].UserID == user1.ID
// expenses[0].CategoryID == category1.ID
// expenses[1].UserID == user2.ID
// expenses[1].CategoryID == category2.ID
// category1.UserID == user1.ID
// category2.UserID == user2.ID
```

This is one of the most powerful features of gofacto, it helps us easily build the structs with the complex associations relationships as long as setting the correct tags in the struct.<br>

Find out more [examples](https://github.com/eyo-chen/gofacto/blob/main/examples/association_test.go).


<details>
    <summary>Best Practice to use <code>WithOne</code> & <code>WithMany</code></summary>
    <ul>
        <li>Must pass the struct pointer to <code>WithOne</code> or <code>WithMany</code></li>
        <li>Must pass same type of struct pointer to <code>WithMany</code></li>
        <li>Do not pass struct with cyclic dependency</li>
    </ul>

    // Do not do this:
    type A struct {
        B_ID int `gofacto:"foreignKey,struct:B"`
    }
    type B struct {
        A_ID int `gofacto:"foreignKey,struct:A"`
    }
</details>

### Reset
Use `Reset` method to reset the factory.
```go
factory.Reset()
```
`Reset` method is recommended to use when tearing down the test.

&nbsp;

### Set Configurations
### WithBlueprint
Use `WithBlueprint` method to set the blueprint function which is a clients defined function to generate the struct values.
```go
func blueprint(i int) *Order {
  return &Order{
    OrderDate: time.Now(),
    Amount:    100*i,
  }
}
factory := gofacto.New(Order{}).
                   WithBlueprint(blueprint)
```
When configure with blueprint, the blueprint function will be called when building the struct first, then the zero value fields will be set to non-zero values by the factory.
It is useful when the clients want to set the default values for the struct.<br>

The signature of the blueprint function is following:<br>
`type blueprintFunc[T any] func(i int) T`

Find out more [examples](https://github.com/eyo-chen/gofacto/blob/main/examples/blueprint_test.go).

### WithStorageName
Use `WithStorageName` method to set the storage name.
```go
factory := gofacto.New(Order{}).
                   WithStorageName("orders")
```
The storage name will be used when inserting the value into the database. <br>

When using SQL databases, the storage name is the table name. <br>
When using NoSQL databases, the storage name is the collection name. <br>

It is optional, the snake case of the struct name(s) will be used if not provided.<br>

### WithDB
Use `WithDB` method to set the database connection.
```go
factory := gofacto.New(Order{}).
                   WithDB(mysqlf.NewConfig(db))
```
When using raw MySQL, use `mysqlf` package. <br>
When using raw PostgreSQL, use `postgresf` package. <br>
When using MongoDB, use `mongof` package. <br>
When using GORM, use `gormf` package. <br>

### WithIsSetZeroValue
Use `WithIsSetZeroValue` method to set if the zero values are set.
```go
factory := gofacto.New(Order{}).
                   WithIsSetZeroValue(false)
```
The zero values will not be set when building the struct if the flag is set to false. <br>

It is optional, it's true by default.

### foreignKey tag
In order to build the struct with the associated struct, we need to set the correct tag in the struct to tell gofacto how to build the associated struct.

Suppose we have the following structs:<br>
`Project` struct has a foreign key `EmployeeID` to reference to `Employee` struct.
```go
type Project struct {
  ID          int
  EmployeeID  int `gofacto:"foreignKey,struct:Employee,table:employees,field:Employee,refField:OtherID"`
  Employee    Employee
}

type Employee struct {
  ID      int
  OtherID int
  Name    string
}
```

The format of the tag is following:<br>
`gofacto:"foreignKey,struct:{{structName}},table:{{tableName}},field:{{fieldName}},refField:{{referenceFieldName}}"`<br>
- `foreignKey` is the tag name. It is required.
- `struct` specifies the name of the associated struct. It is required. In this case, `struct:Employee` indicates that `EmployeeID` is a foreign key to reference to `Employee` struct.
- `table` specifies the table name of the associated struct. It is optional, the snake case and lower case of the struct name(s) will be used if not provided. In this case, `table:employees` indicates that the table name of `Employee` struct is `employees`. However, we can omit it and gofacto will handle it in this example.
- `field` specifies which struct field contains the associated data. It is optional, and it's typically used with gorm. In this example, `field:Employee` indicates that the `Employee` field in the `Project` struct will hold the related `Employee` data after the relationship is loaded.
- `refField` specifies which field to join on in the referenced struct. By default, it joins on the `ID` field, but you can specify a different field. For example, `refField:OtherID` tells gofacto to match `Project.EmployeeID` with `Employee.OtherID` instead of `Employee.ID`.

Find out more [examples](https://github.com/eyo-chen/gofacto/blob/main/examples/association_test.go).


### omit tag
Use `omit` tag in the struct to ignore the field when building the struct.
```go
type Order struct {
  ID          int
  CustomerID  int       `gofacto:"struct:Customer"`
  OrderDate   time.Time
  Amount      float64
  Ignore      string    `gofacto:"omit"`
}
```
The field `Ignore` will not be set to non-zero values when building the struct.

&nbsp;

# Supported Databases
### MySQL
Using `NewConfig` in `mysqlf` package to configure the database connection.
```go
factory := gofacto.New(Order{}).
                   WithDB(mysqlf.NewConfig(db))
```
`db` is `*sql.DB` connection.

Add `mysqlf` tag in the struct to specify the column name.
```go
type Order struct {
  ID          int       `mysqlf:"id"`
  CustomerID  int       `mysqlf:"customer_id"`
  OrderDate   time.Time `mysqlf:"order_date"`
  Amount      float64   `mysqlf:"amount"`
}
```
It is optional to add `mysqlf` tag, the snake case of the field name will be used if not provided.

### PostgreSQL
Using `NewConfig` in `postgresf` package to configure the database connection.
```go
factory := gofacto.New(Order{}).
                   WithDB(postgresf.NewConfig(db))
```
`db` is `*sql.DB` connection.

Add `postgresf` tag in the struct to specify the column name.
```go
type Order struct {
  ID          int       `postgresf:"id"`
  CustomerID  int       `postgresf:"customer_id"`
  OrderDate   time.Time `postgresf:"order_date"`
  Amount      float64   `postgresf:"amount"`
}
```
It is optional to add `postgresf` tag, the snake case of the field name will be used if not provided.

### MongoDB
Using `NewConfig` in `mongof` package to configure the database connection.
```go
factory := gofacto.New(Order{}).
                   WithDB(mongof.NewConfig(db))
```
`db` is `*mongo.Database` connection.

Add `mongof` tag in the struct to specify the column name.
```go
type Order struct {
  ID          primitive.ObjectID  `mongof:"_id"`
  CustomerID  primitive.ObjectID  `mongof:"customer_id"`
  OrderDate   time.Time           `mongof:"order_date"`
  Amount      float64             `mongof:"amount"`
}
```
It is optional to add `mongof` tag, the snake case of the field name will be used if not provided.

&nbsp;

# Supported ORMs
### GORM
Using `NewConfig` in `gormf` package to configure the database connection.
```go
factory := gofacto.New(Order{}).
                   WithDB(gormf.NewConfig(db))
```
`db` is `*gorm.DB` connection.

When using gorm, we might add the association relationship in the struct.
```go
type Order struct {
  ID          int
  Customer    Customer   `gorm:"foreignKey:CustomerID"`
  CustomerID  int        `gofacto:"foreignKey,struct:Customer,field:Customer"`
  OrderDate   time.Time
  Amount      float64
}
```
It basically tells gofacto that `CustomerID` is the foreign key that references the `ID` field in the `Customer` struct, and the field `Customer` is the associated field.

&nbsp;


# Important Considerations
1. gofacto assumes the `ID` field is the primary key and auto-incremented by the database.

2. gofacto cannot set the custom type values defined by the clients.
```go
type CustomType string

type Order struct {
  ID          int
  CustomType  CustomType
  OrderDate   time.Time
  Amount      float64
}
```
If the struct has a custom type, gofacto will ignore the field, and leave it as zero value.<br>
The clients need to set the values manually by using blueprint or overwrite if they don't want the zero value.<br>

&nbsp;

# Acknowledgements
This library is inspired by the [factory](https://github.com/nauyey/factory), [fixtory](https://github.com/k-yomo/fixtory), [factory-go](https://github.com/bluele/factory-go), and [gogo-factory](https://github.com/vx416/gogo-factory).