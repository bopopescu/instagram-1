package main

import (
	"fmt"
	"instagram/handler"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"

	_ "github.com/go-sql-driver/mysql"
	//"github.com/gocraft/dbr/dialect"
	//"github.com/labstack/echo/cookbook/twitter/model"
)

func main() {
	e := echo.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/users", handler.SelectUsers)
	//e.POST("/users", InsertUser)

	// Start server
	viper.SetDefault("http.port", 1323)
	port := fmt.Sprintf(":%d", viper.GetInt("http.port"))
	e.Logger.Fatal(e.Start(port))
}
