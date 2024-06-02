package main

import (
	"pertamina-pleketiplek/handler"
	"pertamina-pleketiplek/service"

	"github.com/labstack/echo/v4"
)

func main() {
	trSrv := service.NewTransaction()
	trHdl := handler.NewTransaction(trSrv)

	e := echo.New()
	e.GET("/", trHdl.Input)
	e.POST("/perday", trHdl.ProcessPerDay)
	e.Logger.Fatal(e.Start(":8000"))
}
