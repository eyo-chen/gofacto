package gofacto

import (
	"reflect"
	"strings"

	"github.com/eyo-chen/gofacto/internal/utils"
)

// tag represents the metadata parsed from the custom tag
type tag struct {
	fieldName    string
	structName   string
	tableName    string
	foreignField string
	omit         bool
}

// extractTag extracts the tag metadata from the struct type
func extractTag(dataType reflect.Type) (map[string]tagInfo, []string, error) {
	var ignoreFields []string
	tagToInfo := make(map[string]tagInfo)

	err := processStructFields(dataType, func(t tag, hasTag bool) error {
		if !hasTag {
			return nil
		}

		if t.omit {
			ignoreFields = append(ignoreFields, t.fieldName)
		}

		tagToInfo[t.structName] = tagInfo{tableName: t.tableName, fieldName: t.fieldName, foreignField: t.foreignField}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return tagToInfo, ignoreFields, nil
}

// parseTag parses the tag string into a tag struct
func parseTag(field reflect.StructField) (tag, bool, error) {
	tagStr := field.Tag.Get(packageName)
	if tagStr == "" {
		return tag{}, false, nil
	}

	parts := strings.Split(tagStr, ";")
	if len(parts) == 0 {
		return tag{}, false, errTagFormat
	}

	t := tag{fieldName: field.Name}
	for _, part := range parts {
		if part == "omit" {
			t.omit = true
			continue
		}

		subParts := strings.Split(part, ",")
		if subParts[0] != "foreignKey" {
			return tag{}, false, errTagFormat
		}

		for _, subPart := range subParts[1:] {
			kv := strings.SplitN(subPart, ":", 2)
			switch kv[0] {
			case "struct":
				t.structName = kv[1]
			case "table":
				t.tableName = kv[1]
			case "field":
				t.foreignField = kv[1]
			default:
				return tag{}, false, errTagFormat
			}
		}
	}

	if t.tableName == "" {
		t.tableName = utils.CamelToSnake(t.structName) + "s"
	}

	return t, true, nil
}

// processStructFields applies a given function to each field of a struct type
func processStructFields(typ reflect.Type, fn func(tag tag, hasTag bool) error) error {
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		t, hasTag, err := parseTag(field)
		if err != nil {
			return err
		}
		if err := fn(t, hasTag); err != nil {
			return err
		}
	}
	return nil
}
