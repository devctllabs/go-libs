// Package oapivalidator validates Echo requests against an OpenAPI document.
//
// It deliberately keeps generated oapi-codegen handlers transport-only:
// request validation and authentication run before the generated Echo wrapper,
// while application authorization remains in handwritten code.
package oapivalidator
