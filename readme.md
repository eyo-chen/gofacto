# gofacto

gofacto is a strongly-typed and user-friendly factory library for Go, designed to simplify the creation of mock data. It offers:

- Intuitive and straightforward usage
- Strong typing and type safety
- Support for various databases and ORMs
- Basic association relationship support

&nbsp;

# Installation
```bash
go get github.com/eyo-chen/gofacto
```

&nbsp;

# Quick Start
Let's consider two structs: `Customer` and `Order`. `Order` struct has a foreign key `CustomerID` because a customer can have many orders.
```go
type Customer struct {
    ID      int
    Gender  Gender     // defined as enum
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

// Init a factory for Customer
customerFactory := gofacto.New(Customer{}).
                           WithDB(mysqlf.NewConfig(db)) // Assuming db is a database connection

// Create and insert a customer
customer, err := customerFactory.Build(ctx).Insert()

// Create and insert two female customers
customers, err := customerFactory.BuildList(ctx, 2).
                                  Overwrite(Customer{Gender: Female}).
                                  Insert()
// customers[0].Gender = Female
// customers[1].Gender = Female

// Create a factory for Order
orderFactory := gofacto.New(Order{}).
                        WithDB(mysqlf.NewConfig(db))

// Create and insert an order with one customer
c := Customer{}
order, err := orderFactory.Build(ctx).
                           WithOne(&c).
                           Insert()
// order.CustomerID = c.ID

// Create and insert two orders with two customers
c1, c2 := Customer{}, Customer{}
orders, err := orderFactory.BuildList(ctx, 2).
                            WithMany([]interface{}{&c1, &c2}).
                            Insert()
// orders[0].CustomerID = c1.ID
// orders[1].CustomerID = c2.ID
```
Note: All fields in the returned struct are populated with non-zero values, and `ID` field is auto-incremented by the database.

&nbsp;

# Usage
### Initialize
Use `New` to initialize the factory by passing the struct you want to create mock data for
```go
factory := gofacto.New(Order{})
```

### Build & BuildList
Use `Build` to create a single value, and `BuildList` to create a list of values
```go
order, err := factory.Build(ctx).Get()
orders, err := factory.BuildList(ctx, 2).Get()
```
`Get` method returns the struct(s) without inserting them into the database. All fields are populated with non-zero values.

### Insert
Use `Insert` to insert values into the database
```go
order, err := factory.Build(ctx).Insert()
orders, err := factory.BuildList(ctx, 2).Insert()
```
`Insert` method inserts the struct into the database and returns the struct with `ID` field populated with the auto-incremented value.<br>

### Overwrite
Use `Overwrite` to manually set field values
```go
order, err := factory.Build(ctx).Overwrite(Order{Amount: 100}).Insert()
// order.Amount = 100
```
The fields in the struct will be used to overwrite the fields in the generated struct.

When building a list of values, `Overwrite` is used to overwrite all the list of values, and `Overwrites` is used to overwrite each value in the list of values.
```go
orders, err := factory.BuildList(ctx, 2).Overwrite(Order{Amount: 100}).Insert()
// orders[0].Amount = 100
// orders[1].Amount = 100

orders, err := factory.BuildList(ctx, 2).Overwrites(Order{Amount: 100}, Order{Amount: 200}).Insert()
// orders[0].Amount = 100
// orders[1].Amount = 200
```

Note: Explicit zero values are not overwritten by default. Use `SetZero` for this purpose.
```go
order, err := factory.Build(ctx).Overwrite(Order{Amount: 0}).Insert()
// order.Amount != 0
```


### SetTrait
When initializing the factory, use `WithTrait` method to set the trait functions. Then use `SetTrait` method to apply the trait functions when building the struct.
```go
func setFemale(c *Customer) {
  c.Gender = Female
}
factory := gofacto.New(Order{}).
                   WithTrait("female", setFemale)

customer, err := factory.Build(ctx).SetTrait("female").Insert()
// customer.Gender = Female
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
// customers[0].Gender = Female
// customers[1].Gender = Female

customers, err := factory.BuildList(ctx, 2).SetTraits("female", "male").Insert()
// customers[0].Gender = Female
// customers[1].Gender = Male
```

### SetZero
Use `SetZero` to set specific fields to zero values.
```go
customer, err := factory.Build(ctx).SetZero("Email", "Phone").Insert()
// customer.Email = nil
// customer.Phone = ""

customers, err := factory.BuildList(ctx, 2).SetZero(0, "Email", "Phone").Insert()
// customers[0].Email = nil
// customers[0].Phone = ""
// customers[1].Email != nil
// customers[1].Phone != ""
```
`SetZero` method with `Build` accepts multiple string as the field names, and the fields will be set to zero values when building the struct.<br>
`SetZero` method with `BuildList` accepts an index and multiple string as the field names, and the fields will be set to zero values when building the struct at the index.

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
The format of the tag is following:<br>
`gofacto:"foreignKey,struct:{{structName}},table:{{tableName}},field:{{fieldName}}"`<br>
- `struct` is the name of the associated struct. It is required.<br>
- `table` is the name of the table. It is optional, the snake case of the struct name(s) will be used if not provided.<br>
- `field` is the name of the foreign value fields within the struct. It is optional, and only required when using gorm.<br>

```go
// build an order with one customer
c := Customer{}
order, err := factory.Build(ctx).WithOne(&c).Insert()
// order.CustomerID = c.ID

// build two orders with two customers
c1 := Customer{}
c2 := Customer{}
orders, err := factory.BuildList(ctx, 2).WithMany(&c1, &c2).Insert()
// orders[0].CustomerID = c1.ID
// orders[1].CustomerID = c2.ID

// build an order with only one customer
c1 := Customer{}
orders, err := factory.BuildList(ctx, 2).WithOne(&c1).Insert()
// orders[0].CustomerID = c1.ID
// orders[1].CustomerID = c1.ID
```

### Reset
Use `Reset` method to reset the factory.
```go
factory.Reset()
```
`Reset` method is recommended to use when tearing down the test.

&nbsp;

### Set Configurations
#### WithBlueprint
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

#### WithStorageName
Use `WithStorageName` method to set the storage name
```go
factory := gofacto.New(Order{}).
                   WithStorageName("orders")
```
The storage name will be used when inserting the value into the database. <br>

When using SQL databases, the storage name is the table name. <br>
When using NoSQL databases, the storage name is the collection name. <br>

It is optional, the snake case of the struct name(s) will be used if not provided.<br>

#### WithDB
Use `WithDB` method to set the database connection
```go
factory := gofacto.New(Order{}).
                   WithDB(mysqlf.NewConfig(db))
```
When using raw MySQL, use `mysqlf` package. <br>
When using raw PostgreSQL, use `postgresf` package. <br>
When using MongoDB, use `mongof` package. <br>
When using GORM, use `gormf` package. <br>

#### WithIsSetZeroValue
Use `WithIsSetZeroValue` method to set if the zero values are set
```go
factory := gofacto.New(Order{}).
                   WithIsSetZeroValue(false)
```
The zero values will not be set when building the struct if the flag is set to false. <br>

It is optional, it's true by default.

#### omit tag
Use `omit` tag in the struct to ignore the field when building the struct
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
Using `NewConfig` in `mysqlf` package to configure the database connection
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
Using `NewConfig` in `postgresf` package to configure the database connection
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
Using `NewConfig` in `mongof` package to configure the database connection
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

# Supported ORMs
### GORM
Using `NewConfig` in `gormf` package to configure the database connection
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

3. gofacto only supports basic associations relationship.
If your database schema has a complex associations relationship, you might need to set the associations manually.<br>
Suppose you have a following schema:
```go
type Expense struct {
  ID          int
  CategoryID  int
}

type Category struct {
  ID      int
  UserID  int
}

type User struct {
  ID    int
}
```
`Expense` struct has a foreign key `CategoryID` to `Category` struct, and `Category` struct has a foreign key `UserID` to `User` struct.<br>
 When building `Expense` data with `WithOne(&Category{})`, `Category` struct will be built, but `User` struct will not be built. The clients need to build `User` struct manually and set `UserID` to `Category` struct.


&nbsp;

# Acknowledgements
This library is inspired by the [factory](https://github.com/nauyey/factory), [fixtory](https://github.com/k-yomo/fixtory), [factory-go](https://github.com/bluele/factory-go), and [gogo-factory](https://github.com/vx416/gogo-factory).

