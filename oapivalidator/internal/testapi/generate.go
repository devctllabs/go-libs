// Package testapi contains an oapi-codegen Echo 5 strict-server fixture.
package testapi

//go:generate go tool oapi-codegen -config server.yaml -o server.gen.go api.yaml
