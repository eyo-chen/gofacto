package types

import (
	"errors"
)

var (
	// ErrBuildListNGreaterThanZero is the error representing that n must be greater than 0
	ErrBuildListNGreaterThanZero = errors.New("n must be greater than 0")

	// ErrDBIsNotProvided is the error representing that DB connection is not provided
	ErrDBIsNotProvided = errors.New("db connection is not provided")

	// ErrCantCvtToPtr is the error representing that can't convert to pointer
	ErrCantCvtToPtr = errors.New("can't convert to pointer")

	// ErrWithTraitNameNotFound is the error representing that trait name is not found
	ErrWithTraitNameNotFound = errors.New("trait name is not found")

	// ErrFieldNotFound is the error representing that field not found
	ErrFieldNotFound = errors.New("field not found")

	// ErrFieldCantSet is the error representing that field can't be set
	ErrFieldCantSet = errors.New("field can't be set")

	// ErrIndexIsOutOfRange is the error representing that index is out of range
	ErrIndexIsOutOfRange = errors.New("index is out of range")

	// ErrValueNotTheSameType is the error representing that value is not the same type
	ErrValueNotTheSameType = errors.New("value is not the same type")

	// ErrTagFormat is the error representing that tag is in wrong format
	ErrTagFormat = errors.New("tag is in wrong format. It should be gofacto:\"struct:<structName>,table:<TableName>,foreignField:<ForeignField>\"")

	// ErrNotFoundAtTag is the error representing that not found at tag
	ErrNotFoundAtTag = errors.New("not found at tag")

	// ErrIsNotPtr is the error representing that is not pointer
	ErrIsNotPtr = errors.New("is not pointer")

	// ErrIsNotStructPtr is the error representing that is not struct pointer
	ErrIsNotStructPtr = errors.New("is not struct pointer")

	// ErrDestIsNotStruct is the error representing that dest is not struct
	ErrDestIsNotStruct = errors.New("dest is not struct")

	// ErrSrcIsNotStruct is the error representing that src is not struct
	ErrSrcIsNotStruct = errors.New("src is not struct")

	// ErrTypeDiff is the error representing that type is different
	ErrTypeDiff = errors.New("type is different")

	// ErrNotInt is the error representing that not an integer
	ErrNotInt = errors.New("not an integer")
)
