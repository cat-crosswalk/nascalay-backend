//go:generate go run github.com/google/wire/cmd/wire@latest
//go:build wireinject
// +build wireinject

package infrastructure

import (
	"github.com/21hack02win/nascalay-backend/interfaces/handler"
	"github.com/google/wire"
)

func injectServer() handler.ServerInterface {
	wire.Build(
		handler.NewHandler,
	)

	return nil
}
