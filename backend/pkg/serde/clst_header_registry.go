package serde

// SchemaInfo holds extracted schema-related details from Kafka headers
type SchemaInfo struct {
	KeyEncoding       string
	ValueEncoding     string
	ProtobufTypeValue string

	// Extendable fields for custom headers when needed below
}
