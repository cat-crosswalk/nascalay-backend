package oapi

import (
	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
)

func SetupOapiMiddleware() (echo.MiddlewareFunc, error) {
	spec, err := GetSwagger()
	if err != nil {
		return nil, err
	}

	spec.Servers = nil

	openapi3.DefineStringFormatCallback("uuid", func(uuidStr string) error {
		_, err := uuid.FromString(uuidStr)
		return err
	})

	return middleware.OapiRequestValidator(spec), nil
}
