package rigel

import "fmt"

// getSchemaFieldsPath constructs the path for a schema based on the provided schemaName and schemaVersion.
func getSchemaFieldsPath(schemaName string, schemaVersion int) string {
	return fmt.Sprintf("%s/schema/%s/%d/fields", rigelPrefix, schemaName, schemaVersion)
}

// getConfPath constructs the path for a configuration based on the provided schemaName, schemaVersion.
func getConfPath(schemaName string, schemaVersion int) string {
	return fmt.Sprintf("%s/conf/%s/%d", rigelPrefix, schemaName, schemaVersion)
}

// getConfKeyPath constructs the path for a configuration based on the provided schemaName, schemaVersion, and confKey.
func getConfKeyPath(schemaName string, schemaVersion int, confKey string) string {
	return fmt.Sprintf("%s/conf/%s/%d/%s", rigelPrefix, schemaName, schemaVersion, confKey)
}

// getSchemaPath constructs the base key for a schema in etcd based on the provided schemaName and schemaVersion.
func getSchemaPath(schemaName string, schemaVersion int) string {
	return fmt.Sprintf("%s/schema/%s/%d/", rigelPrefix, schemaName, schemaVersion)
}
