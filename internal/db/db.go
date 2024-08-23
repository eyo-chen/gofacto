package db

// InsertParams is a struct that holds the parameters for the Insert method
type InsertParams struct {
	StorageName string
	Value       interface{}
}

// InsertListParams is a struct that holds the parameters for the InsertList method
type InsertListParams struct {
	StorageName string
	Values      []interface{}
}
