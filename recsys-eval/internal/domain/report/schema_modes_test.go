package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestSchemaModesMatchSupportedModes(t *testing.T) {
	schemaPath := filepath.Join(projectRoot(t), "api", "schemas", "report.v1.json")
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var schema struct {
		Properties map[string]struct {
			Enum []string `json:"enum"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	modeEnum := schema.Properties["mode"].Enum
	if len(modeEnum) == 0 {
		t.Fatalf("schema missing mode enum")
	}
	got := append([]string(nil), modeEnum...)
	want := append([]string(nil), SupportedModes()...)
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("schema modes mismatch\nschema=%v\ncode=%v", got, want)
	}
}

func projectRoot(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve project root: %v", err)
	}
	return dir
}
