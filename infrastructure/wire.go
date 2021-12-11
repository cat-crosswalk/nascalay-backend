//go:generate go run github.com/google/wire/cmd/wire@latest
//go:build wireinject
// +build wireinject

package infrastructure

import (
	"github.com/21hack02win/nascalay-backend/interfaces/handler"
	"github.com/21hack02win/nascalay-backend/interfaces/repository"
	"github.com/21hack02win/nascalay-backend/oapi"
	"github.com/21hack02win/nascalay-backend/usecases/service/ws"
	"github.com/google/wire"
)

func injectServer() oapi.ServerInterface {
	wire.Build(
		handler.NewHandler,
		repository.NewRepository,
		ws.NewStreamer,
		ws.NewHub,
	)

	return nil
}
