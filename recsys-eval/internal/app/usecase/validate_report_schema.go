package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// SchemaValidationError captures JSON Schema validation context.
type SchemaValidationError struct {
	SchemaPath       string
	ReportPath       string
	InstanceLocation string
	Message          string
}

func (e SchemaValidationError) Error() string {
	return fmt.Sprintf("schema=%s report=%s path=%s message=%s", e.SchemaPath, e.ReportPath, e.InstanceLocation, e.Message)
}

// ValidateJSONFileAgainstSchema validates a JSON document on disk against the schema.
func ValidateJSONFileAgainstSchema(schemaPath, reportPath string) error {
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat = true

	// #nosec G304 -- schema path provided by operator
	schemaFile, err := os.Open(schemaPath)
	if err != nil {
		return err
	}
	defer schemaFile.Close()

	if err := compiler.AddResource(schemaPath, schemaFile); err != nil {
		return err
	}
	schema, err := compiler.Compile(schemaPath)
	if err != nil {
		return err
	}

	// #nosec G304 -- report path provided by operator
	reportData, err := os.ReadFile(reportPath)
	if err != nil {
		return err
	}
	var payload any
	if err := json.Unmarshal(reportData, &payload); err != nil {
		return err
	}
	if err := schema.Validate(payload); err != nil {
		var verr *jsonschema.ValidationError
		if errors.As(err, &verr) {
			msg := verr.Message
			if msg == "" {
				msg = verr.Error()
			}
			return SchemaValidationError{
				SchemaPath:       schemaPath,
				ReportPath:       reportPath,
				InstanceLocation: verr.InstanceLocation,
				Message:          msg,
			}
		}
		return fmt.Errorf("schema=%s report=%s error=%w", schemaPath, reportPath, err)
	}
	return nil
}
