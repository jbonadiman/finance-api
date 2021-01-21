package main

import (
	"github.com/jbonadiman/finance-bot/api/src/services"
	"github.com/jbonadiman/finance-bot/api/src/services/auth_service"
	"github.com/jbonadiman/finance-bot/api/src/services/finance_service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"time"
)

func main() {
	e := echo.New()

	client := http.Client{Timeout: 10 * time.Second}

	financeService := finance_service.New(&client)
	authService := auth_service.New(&[]services.Authenticated{financeService})

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", healthCheck)

	e.GET("/login", authService.Login)
	e.GET("/authentication", authService.LoginRedirect)
	e.GET("/tasks", financeService.GetTasks)

	e.Logger.Fatal(e.Start(":8080"))
}

func healthCheck(c echo.Context) error {
	return c.String(http.StatusOK, "")
}
