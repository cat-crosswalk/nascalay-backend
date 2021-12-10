package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/21hack02win/nascalay-backend/infrastructure"
	"github.com/labstack/echo/v4"
)

var baseEndpoint string

func main() {
	flag.StringVar(&baseEndpoint, "b", "", "Custom base endpoint .e.g \"/api\"")
	flag.Parse()

	e := echo.New()
	infrastructure.Setup(e, baseEndpoint)

	e.Logger.Fatal(e.Start(port()))
}

func port() string {
	p := 3000
	if env := os.Getenv("APP_PORT"); len(env) > 0 {
		if port, err := strconv.Atoi(env); err == nil {
			p = port
		}
	}

	return fmt.Sprintf(":%d", p)
}
