package serde

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Compile-time check to ensure HeaderSchemaSerde implements Serde
var _ Serde = (*CLSTProtobufHeaderSchema)(nil)

// HeaderSchemaSerde implements the Serde interface
type CLSTProtobufHeaderSchema struct{}

// Name returns the name of the serde payload encoding.
func (CLSTProtobufHeaderSchema) Name() PayloadEncoding {
	return PayloadEncoding("header-schema")
}

// DeserializePayload reads the schema from headers and deserializes the payload.
func (d CLSTProtobufHeaderSchema) DeserializePayload(ctx context.Context, record *kgo.Record, payloadType PayloadType) (*RecordPayload, error) {
	// Get schema info from headers
	schemaInfo, err := getSchemaInfoFromHeaders(record)
	if err != nil {
		return nil, fmt.Errorf("failed to extract schema info from headers: %w", err)
	}

	// Log extracted header info (optional)
	fmt.Printf("SchemaID: %d, KeyEncoding: %s, ValueEncoding: %s, ProtobufTypeValue: %s\n",
		schemaInfo.SchemaID, schemaInfo.KeyEncoding, schemaInfo.ValueEncoding, schemaInfo.ProtobufTypeValue)

	// Deserialize payload (as JSON for simplicity)
	payload := payloadFromRecord(record, payloadType)

	var jsonData map[string]interface{}
	if err := json.Unmarshal(payload, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to deserialize JSON payload: %w", err)
	}

	return &RecordPayload{
		DeserializedPayload: jsonData,
		NormalizedPayload:   payload,
		Encoding:            PayloadEncoding("header-schema"),
	}, nil
}

// SerializeObject serializes data and adds schema ID to headers.
func (d CLSTProtobufHeaderSchema) SerializeObject(ctx context.Context, obj any, payloadType PayloadType, opts ...SerdeOpt) ([]byte, error) {
	cfg := serdeCfg{}
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize object to JSON: %w", err)
	}

	return jsonBytes, nil
}
