package rigel

import "fmt"

// getSchemaFieldsPath constructs the path for a schema based on the provided appName, moduleName and version.
func getSchemaFieldsPath(appName string, moduleName string, version int) string {
	return fmt.Sprintf("%s/%s/%s/%d/fields", rigelPrefix, appName, moduleName, version)
}

// getConfPath constructs the path for a configuration based on the provided appName, moduleName and version.
func getConfPath(appName string, moduleName string, version int, namedConfig string) string {
	return fmt.Sprintf("%s/%s/%s/%d/config/%s", rigelPrefix, appName, moduleName, version, namedConfig)
}

// getConfKeyPath constructs the path for a configuration based on the provided appName, moduleName, version, namedConfig, and confKey.
func getConfKeyPath(appName string, moduleName string, version int, namedConfig string, confKey string) string {
	return fmt.Sprintf("%s/%s/%s/%d/config/%s/%s", rigelPrefix, appName, moduleName, version, namedConfig, confKey)
}

// getSchemaPath constructs the base key for a schema in etcd based on the provided appName, moduleName and version.
func getSchemaPath(appName string, moduleName string, version int) string {
	return fmt.Sprintf("%s/%s/%s/%d/", rigelPrefix, appName, moduleName, version)
}
