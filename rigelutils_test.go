package rigel

import (
	"testing"
)

func TestGetSchemaFieldsPath(t *testing.T) {
	tests := []struct {
		appName      string
		moduleName   string
		version      int
		expectedPath string
	}{
		{"testApp", "testModule", 1, "/remiges/rigel/testApp/testModule/1/fields"},
	}

	for _, tt := range tests {
		path := getSchemaFieldsPath(tt.appName, tt.moduleName, tt.version)
		if path != tt.expectedPath {
			t.Errorf("Expected %s but got %s", tt.expectedPath, path)
		}
	}
}

func TestGetConfPath(t *testing.T) {
	tests := []struct {
		appName      string
		moduleName   string
		version      int
		namedConfig  string
		expectedPath string
	}{
		{"testApp", "testModule", 1, "testConf", "/remiges/rigel/testApp/testModule/1/config/testConf"},
	}

	for _, tt := range tests {
		path := getConfPath(tt.appName, tt.moduleName, tt.version, tt.namedConfig)
		if path != tt.expectedPath {
			t.Errorf("Expected %s but got %s", tt.expectedPath, path)
		}
	}
}

func TestGetConfKeyPath(t *testing.T) {
	tests := []struct {
		appName      string
		moduleName   string
		version      int
		namedConfig  string
		confKey      string
		expectedPath string
	}{
		{"testApp", "testModule", 1, "testConf", "testKey", "/remiges/rigel/testApp/testModule/1/config/testConf/testKey"},
	}

	for _, tt := range tests {
		path := getConfKeyPath(tt.appName, tt.moduleName, tt.version, tt.namedConfig, tt.confKey)
		if path != tt.expectedPath {
			t.Errorf("Expected %s but got %s", tt.expectedPath, path)
		}
	}
}

func TestGetSchemaPath(t *testing.T) {
	tests := []struct {
		appName      string
		moduleName   string
		version      int
		expectedPath string
	}{
		{"testApp", "testModule", 1, "/remiges/rigel/testApp/testModule/1/"},
	}

	for _, tt := range tests {
		path := getSchemaPath(tt.appName, tt.moduleName, tt.version)
		if path != tt.expectedPath {
			t.Errorf("Expected %s but got %s", tt.expectedPath, path)
		}
	}
}
