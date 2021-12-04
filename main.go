package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/21hack02win/backend/infrastructure"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	infrastructure.Setup(e)

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
