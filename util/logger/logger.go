package logger

import "github.com/labstack/echo/v4"

// TODO: main()で初期化するのでnilチェックはしていない、適当にslogとかに変える
var Echo echo.Logger
