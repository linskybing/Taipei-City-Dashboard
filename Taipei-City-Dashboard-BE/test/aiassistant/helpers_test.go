package aiassistant_test

import "testing"

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func assertEnum(t *testing.T, schema interface{}, values ...string) {
	t.Helper()
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		t.Fatalf("schema type = %T, want map", schema)
	}
	rawEnum, ok := schemaMap["enum"].([]string)
	if !ok {
		t.Fatalf("enum type = %T, want []string", schemaMap["enum"])
	}
	for _, value := range values {
		if !contains(rawEnum, value) {
			t.Fatalf("enum %#v missing %q", rawEnum, value)
		}
	}
}
