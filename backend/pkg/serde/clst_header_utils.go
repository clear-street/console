package serde

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
)

// getSchemaInfoFromHeaders maps headers to SchemaInfo fields using reflection and type handlers
func getSchemaInfoFromHeaders(record *kgo.Record) (SchemaInfo, error) {
	var info SchemaInfo
	infoValue := reflect.ValueOf(&info).Elem()

	// Map of type handlers (dispatch pattern)
	typeHandlers := map[reflect.Kind]func([]byte, reflect.Value) error{
		reflect.String: func(value []byte, field reflect.Value) error {
			field.SetString(string(value))
			return nil
		},
		reflect.Uint32: func(value []byte, field reflect.Value) error {
			val, err := strconv.ParseUint(string(value), 10, 32)
			if err != nil {
				return err
			}
			field.SetUint(val)
			return nil
		},
	}

	for _, header := range record.Headers {
		fieldName := toCamelCase(header.Key)
		field := infoValue.FieldByName(fieldName)

		if field.IsValid() && field.CanSet() {
			if handler, found := typeHandlers[field.Kind()]; found {
				if err := handler(header.Value, field); err != nil {
					return info, err
				}
			}
		}
	}

	// Validation: Ensure at least one known header is set
	if info.KeyEncoding == "" && info.ValueEncoding == "" && info.ProtobufTypeValue == "" {
		return info, errors.New("no known headers found in the record")
	}

	return info, nil
}

// toCamelCase converts snake_case, kebab-case, or dot.case to CamelCase
func toCamelCase(input string) string {
	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == '_' || r == '.' || r == '-'
	})
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, "")
}
