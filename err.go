package gofacto

import (
	"errors"
)

var (
	// errBuildListNGreaterThanZero is the error representing that n must be greater than 0
	errBuildListNGreaterThanZero = errors.New("n must be greater than 0")

	// errDBIsNotProvided is the error representing that DB connection is not provided
	errDBIsNotProvided = errors.New("db connection is not provided")

	// errCantCvtToPtr is the error representing that can't convert to pointer
	errCantCvtToPtr = errors.New("can't convert to pointer")

	// errWithTraitNameNotFound is the error representing that trait name is not found
	errWithTraitNameNotFound = errors.New("trait name is not found")

	// errFieldNotFound is the error representing that field not found
	errFieldNotFound = errors.New("field not found")

	// errFieldCantSet is the error representing that field can't be set
	errFieldCantSet = errors.New("field can't be set")

	// errIndexIsOutOfRange is the error representing that index is out of range
	errIndexIsOutOfRange = errors.New("index is out of range")

	// errValueNotTheSameType is the error representing that value is not the same type
	errValueNotTheSameType = errors.New("value is not the same type")

	// errTagFormat is the error representing that tag is in wrong format
	errTagFormat = errors.New("tag is in wrong format. It should be gofacto:\"struct:<structName>,table:<TableName>,foreignField:<ForeignField>\"")

	// errNotFoundAtTag is the error representing that not found at tag
	errNotFoundAtTag = errors.New("not found at tag")

	// errIsNotPtr is the error representing that is not pointer
	errIsNotPtr = errors.New("is not pointer")

	// errIsNotStructPtr is the error representing that is not struct pointer
	errIsNotStructPtr = errors.New("is not struct pointer")

	// errDestIsNotStruct is the error representing that dest is not struct
	errDestIsNotStruct = errors.New("dest is not struct")

	// errSrcIsNotStruct is the error representing that src is not struct
	errSrcIsNotStruct = errors.New("src is not struct")

	// errTypeDiff is the error representing that type is different
	errTypeDiff = errors.New("type is different")

	// errNotInt is the error representing that not an integer
	errNotInt = errors.New("not an integer")
)
