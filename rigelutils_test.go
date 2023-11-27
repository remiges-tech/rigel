package rigel

import (
	"testing"
)

func TestGetSchemaFieldsPath(t *testing.T) {
	schemaName := "testSchema"
	schemaVersion := 1
	expectedPath := "/remiges/rigel/schema/testSchema/1/fields"

	path := getSchemaFieldsPath(schemaName, schemaVersion)

	if path != expectedPath {
		t.Errorf("Expected %s but got %s", expectedPath, path)
	}
}

func TestGetConfKeyPath(t *testing.T) {
	schemaName := "testSchema"
	schemaVersion := 1
	confName := "testConf"
	expectedPath := "/remiges/rigel/conf/testSchema/1/testConf"

	path := getConfKeyPath(schemaName, schemaVersion, confName)

	if path != expectedPath {
		t.Errorf("Expected %s but got %s", expectedPath, path)
	}
}
