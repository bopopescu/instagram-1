package handler

import (
	"fmt"
	"instagram/db"
	"instagram/model"
	"net/http"
	"strconv"

	"github.com/labstack/echo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr"
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

	var u []model.UserResponse
	_, err = sess.Select("*").From("user").Load(&u)
	if err != nil {
		fmt.Println(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	} else {
		return c.JSON(http.StatusOK, u)
	}
}

func GetTimeline(c echo.Context) error {

	if err != nil {
		panic(fmt.Errorf("DB connection error: %s \n", err))
	}
	var id int64
	param := c.Param("id")
	date := c.Param("date")
	id, err = strconv.ParseInt(param, 0, 64)

	//var u []model.UserResponse
	//_, err = sess.Select("*").From("user").Where("user_id = ?",id).Load(&u)

	var timeline []model.TimelineResponse
	count, err := sess.Select("m.*").From(dbr.I("follow_list").As("f")).
		Join(dbr.I("media").As("m"), "f.user_id = m.user_id").
		Where("f.my_id = ?", id).
		OrderDir("m.created_time", false).
		Limit(10).Load(&timeline)

	if count == 0 {
		return c.JSON(http.StatusOK, "表示するタイムラインがありません")
	}

	for key, value := range timeline {
		var user []model.UserResponse
		var likes []model.LikesResponse
		var likeCount = 0
		var isLiked = 0

		_, err = sess.Select("u.*").From(dbr.I("media").As("m")).
			Join(dbr.I("user").As("u"), "u.user_id = m.user_id").Where("u.user_id = ?", value.UserID).Load(&user)

		value.User = user

		likeCount, err = sess.Select("*").From(dbr.I("media").As("m")).
			Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ?", value.MediaID).Load(&likes)

		isLiked, err = sess.Select("*").From(dbr.I("media").As("m")).
			Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ? AND l.user_id = ?", value.MediaID, id).Load(&likes)


		value.LikeCounts = likeCount

		if isLiked > 0 {
			value.IsLiked = true
		}

		fmt.Println("count", isLiked)
		fmt.Println("count", isLiked)
		fmt.Println("like", value)
		timeline[key] = value

	}
	//"u.full_name","u.username","u.profile_picture" Where("u.user_id = ?", id)
	if err != nil {
		fmt.Println(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	} else {
		return c.JSON(http.StatusOK, timeline)
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
