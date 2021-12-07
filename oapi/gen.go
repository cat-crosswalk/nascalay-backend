//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest --config ./.server.yml ../docs/openapi.yml
//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest --config ./.types.yml ../docs/openapi.yml

package oapi

import "github.com/21hack02win/nascalay-backend/model"

// NOTE: UserIdInQueryをuuid.UUID型にするとBindしてくれないためstring型にしてここで詰め替えをする
func (u UserIdInQuery) Refill() (model.UserId, error) {
	return model.UserIdFromString(string(u))
}
