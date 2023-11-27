package rigel

import "fmt"

// getSchemaFieldsPath constructs the path for a schema based on the provided schemaName and schemaVersion.
func getSchemaFieldsPath(schemaName string, schemaVersion int) string {
	return fmt.Sprintf("%s/schema/%s/%d/fields", RigelPrefix, schemaName, schemaVersion)
}

// getConfKeyPath constructs the path for a configuration based on the provided schemaName, schemaVersion, and confName.
func getConfKeyPath(schemaName string, schemaVersion int, confName string) string {
	return fmt.Sprintf("%s/conf/%s/%d/%s", RigelPrefix, schemaName, schemaVersion, confName)
}

// getSchemaPath constructs the base key for a schema in etcd based on the provided schemaName and schemaVersion.
func getSchemaPath(schemaName string, schemaVersion int) string {
	return fmt.Sprintf("%s/schema/%s/%d/", RigelPrefix, schemaName, schemaVersion)
}
