package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	input := flag.String("input", "../docs/reference/api/openapi.yaml", "path to OpenAPI YAML")
	outDir := flag.String("out-dir", "./docs", "output directory for api docs (empty to skip write)")
	validate := flag.Bool("validate", true, "validate OpenAPI spec")
	flag.Parse()

	if *input == "" {
		fatalf("input path is required")
	}

	data, err := os.ReadFile(filepath.Clean(*input))
	if err != nil {
		fatalf("read openapi: %v", err)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	spec, err := loader.LoadFromData(data)
	if err != nil {
		fatalf("parse openapi: %v", err)
	}
	if *validate {
		if err := spec.Validate(context.Background()); err != nil {
			fatalf("validate openapi: %v", err)
		}
	}

	if strings.TrimSpace(*outDir) == "" {
		return
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fatalf("create output dir: %v", err)
	}

	if err := writeFile(filepath.Join(*outDir, "openapi.yaml"), data); err != nil {
		fatalf("write openapi.yaml: %v", err)
	}

	jsonBytes, err := spec.MarshalJSON()
	if err != nil {
		fatalf("marshal openapi: %v", err)
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, jsonBytes, "", "  "); err != nil {
		fatalf("format openapi json: %v", err)
	}
	if err := writeFile(filepath.Join(*outDir, "openapi.json"), pretty.Bytes()); err != nil {
		fatalf("write openapi.json: %v", err)
	}
}

func writeFile(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, bytes.NewReader(data)); err != nil {
		return err
	}
	return f.Sync()
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
