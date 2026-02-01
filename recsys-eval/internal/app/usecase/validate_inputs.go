package usecase

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// ValidateJSONL validates each JSONL line against a JSON schema.
func ValidateJSONL(ctx context.Context, schemaPath, inputPath string) (int, error) {
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat = true

	// #nosec G304 -- schema path provided by operator
	schemaFile, err := os.Open(schemaPath)
	if err != nil {
		return 0, err
	}
	defer schemaFile.Close()

	if err := compiler.AddResource(schemaPath, schemaFile); err != nil {
		return 0, err
	}

	schema, err := compiler.Compile(schemaPath)
	if err != nil {
		return 0, err
	}

	// #nosec G304 -- input path provided by operator
	file, err := os.Open(inputPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 16*1024*1024)

	invalid := 0
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return invalid, ctx.Err()
		default:
		}
		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			invalid++
			continue
		}
		if err := schema.Validate(v); err != nil {
			invalid++
		}
	}
	if err := scanner.Err(); err != nil {
		return invalid, err
	}
	if invalid > 0 {
		return invalid, fmt.Errorf("validation failed for %d line(s)", invalid)
	}
	return 0, nil
}
