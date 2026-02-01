package docs

import _ "embed"

var (
	//go:embed openapi.yaml
	OpenAPIYAML []byte
	//go:embed openapi.json
	OpenAPIJSON []byte
)
