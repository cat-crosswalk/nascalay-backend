package logger

import "github.com/labstack/echo/v4"

// TODO: main()で設定するためnilチェックをしていない
// 適当にslogとかに変える
var Echo echo.Logger
