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
	e.GET("/users", handler.GetUsers)
	e.GET("/users/:id", handler.GetUser)
	e.GET("/users/:id/media", handler.GetUserMedia)
	e.GET("/users/:id/media/:date", handler.GetUserMedia)
	e.GET("/users/:id/relationship/:id2", handler.GetFollowStatus)
	e.GET("/timeline/:id", handler.GetTimeline)
	e.GET("/timeline/:id/:date", handler.GetTimeline)

	e.POST("/media/likes", handler.PostLikes)
	e.POST("/users", handler.PostUser)
	e.POST("/login", handler.PostLogin)
	e.POST("/users/relationship/follow", handler.PostFollow)

	e.DELETE("/media/likes", handler.DeleteLikes)

	// Start server
	viper.SetDefault("http.port", 1323)
	port := fmt.Sprintf(":%d", viper.GetInt("http.port"))
	e.Logger.Fatal(e.Start(port))
}
