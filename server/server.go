package server

import (
	"fmt"
	"github.com/finwhale/octopus/core"
	"github.com/finwhale/octopus/farmer"
	"github.com/finwhale/octopus/request"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
)

func Run(env string, port string) {
	adapter, dbUrl, schemaName, charset, maxOpenConns, plural, logMode := core.GetSchemaInfo(env, true)
	core.SetDB(adapter, dbUrl, schemaName, charset, maxOpenConns, plural, logMode)

	e := echo.New()
	e.Use(middleware.Recover())

	e.POST("/", func(c echo.Context) error {
		r := new(request.Request)

		c.Bind(&r)
		r.Header = c.Request().Header
		result := farmer.Exec(r)

		return c.JSON(http.StatusOK, result)
	})

	fmt.Printf("[%v] ", env)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%v", port)))
}
