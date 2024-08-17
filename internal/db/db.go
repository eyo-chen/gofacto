package db

// InsertParams is a struct that holds the parameters for the Insert method
type InserParams struct {
	StorageName string
	Value       interface{}
}

// InserListParams is a struct that holds the parameters for the InsertList method
type InserListParams struct {
	StorageName string
	Values      []interface{}
}
