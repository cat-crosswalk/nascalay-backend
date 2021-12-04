//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest --config ./.server.yml ../../docs/openapi.yml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest --config ./.types.yml ../../docs/openapi.yml

package oapi
