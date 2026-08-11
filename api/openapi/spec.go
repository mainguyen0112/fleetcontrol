// Package openapi embeds the OpenAPI specification so it can be served
// directly from the compiled binary, independent of the process's working
// directory at runtime.
package openapi

import _ "embed"

//go:embed openapi.yaml
var Spec []byte
