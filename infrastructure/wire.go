//go:generate go run github.com/google/wire/cmd/wire@latest
//go:build wireinject
// +build wireinject

package infrastructure

import (
	"github.com/21hack02win/nascalay-backend/interfaces/handler"
	"github.com/21hack02win/nascalay-backend/interfaces/repository"
	"github.com/21hack02win/nascalay-backend/usecases/service"
	"github.com/21hack02win/nascalay-backend/usecases/service/ws"
	"github.com/google/wire"
)

func injectServer() handler.ServerInterface {
	wire.Build(
		handler.NewHandler,
		service.NewService,
		repository.NewRepository,
		ws.NewStreamer,
	)

	return nil
}
