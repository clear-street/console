package serde

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/twmb/franz-go/pkg/kgo"
)

// SchemaInfo holds extracted schema-related details from Kafka headers
type SchemaInfo struct {
	SchemaID          uint32
	KeyEncoding       string
	ValueEncoding     string
	ProtobufTypeValue string
}

// HeaderConfig represents the structure of headers_config.json
type HeaderConfig struct {
	Headers map[string]struct {
		Type  string `json:"type"`
		Field string `json:"field"`
	} `json:"headers"`
}

// Global variable to hold loaded header mappings
var headerConfig HeaderConfig

// init loads the header configuration from JSON file at startup
func init() {
	configFile, err := os.ReadFile("headers_config.json")
	if err != nil {
		panic(fmt.Sprintf("failed to load header config: %v", err))
	}

	if err := json.Unmarshal(configFile, &headerConfig); err != nil {
		panic(fmt.Sprintf("failed to parse header config: %v", err))
	}
}

// getSchemaInfoFromHeaders extracts schema details from Kafka headers using the config-based mappings
func getSchemaInfoFromHeaders(record *kgo.Record) (SchemaInfo, error) {
	var info SchemaInfo
	var schemaIDFound bool

	// Use reflection to assign values to SchemaInfo fields
	infoValue := reflect.ValueOf(&info).Elem()

	for _, header := range record.Headers {
		if config, exists := headerConfig.Headers[header.Key]; exists {
			field := infoValue.FieldByName(config.Field)
			if !field.IsValid() {
				return info, fmt.Errorf("invalid field name %s for header %s", config.Field, header.Key)
			}

			switch config.Type {
			case "int":
				value, err := strconv.Atoi(string(header.Value))
				if err != nil {
					return info, fmt.Errorf("invalid value for %s: expected int", header.Key)
				}
				field.SetUint(uint64(value))
				if header.Key == "schema_id" {
					schemaIDFound = true
				}

			case "string":
				field.SetString(string(header.Value))

			default:
				return info, fmt.Errorf("unsupported type %s for header %s", config.Type, header.Key)
			}
		}
	}

	if !schemaIDFound {
		return info, errors.New("schema_id header not found")
	}

	return info, nil
}
