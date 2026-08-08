// Package openapi contains versioned API contracts embedded into the application.
package openapi

import _ "embed"

//go:embed v1/openapi.yaml
var v1Specification []byte

func V1() []byte { return v1Specification }
