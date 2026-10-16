// Package docs embeds the OpenAPI specification served by the API.
package docs

import _ "embed"

// OpenAPISpec is the embedded OpenAPI 3 document.
//
//go:embed openapi.yaml
var OpenAPISpec []byte
