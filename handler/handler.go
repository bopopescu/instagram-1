package handler

import (
	"fmt"
	"instagram/db"
	"instagram/model"
	"net/http"

	"github.com/labstack/echo"
	//"github.com/gocraft/dbr"
	_ "github.com/go-sql-driver/mysql"
)

var (
	seq       = 1
	conn, err = db.ConnectDB()
	sess      = conn.NewSession(nil)
)

//----------
// Handlers
//----------

func SelectUsers(c echo.Context) error {

        if err != nil {
            panic(fmt.Errorf("DB connection error: %s \n", err))
        }

	var u []model.User
	_, err = sess.Select("*").From("user").Load(&u)
	if err != nil {
		fmt.Println(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	} else {
		response := new(model.ResponseData)
		response.Users = u
		return c.JSON(http.StatusOK, response)
	}
}

func InsertUser(c echo.Context) error {
	u := new(model.UserinfoJSON)
	if err := c.Bind(u); err != nil {
		return err
	}

	sess.InsertInto("user").Columns("id", "email", "first_name", "last_name").Values(u.ID, u.Email, u.Firstname, u.Lastname).Exec()

	return c.NoContent(http.StatusOK)
}
