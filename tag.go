package gofacto

import (
	"strings"

	"github.com/eyo-chen/gofacto/internal/utils"
)

// tag represents the metadata parsed from the custom tag
type tag struct {
	structName   string
	tableName    string
	foreignField string
	omit         bool
}

// parseTag parses the tag string into a tag struct
func parseTag(tagStr string) (tag, error) {
	parts := strings.Split(tagStr, ";")
	if len(parts) == 0 {
		return tag{}, errTagFormat
	}

	var t tag
	for _, part := range parts {
		if part == "omit" {
			t.omit = true
			continue
		}

		subParts := strings.Split(part, ",")
		if subParts[0] != "foreignKey" {
			return tag{}, errTagFormat
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
				return tag{}, errTagFormat
			}
		}
	}

	if t.tableName == "" {
		t.tableName = utils.CamelToSnake(t.structName) + "s"
	}

	return t, nil
}
