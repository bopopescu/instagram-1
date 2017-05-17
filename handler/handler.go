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

//	Get

func GetUsers(c echo.Context) error {

	if err != nil {
		return c.JSON(http.StatusOK,"DB connection error")
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

func GetUser(c echo.Context) error {

	if err != nil {
		return c.JSON(http.StatusOK,"DB connection error")
	}

	var userId int64
	param := c.Param("id")
	userId, err = strconv.ParseInt(param, 0, 64)
	var counts = model.CountResponse{Media:0,Follows:0,FollowedBy:0}

	var u = model.UserDetailResponse{}

	// ユーザー情報取得
	_, err = sess.Select("*").From("user").Where("user_id = ?",userId).Load(&u)

	// 投稿数取得
	_, err = sess.Select("count(*)").From("media").Where("user_id = ?",userId).Load(&counts.Media)

	// フォロー数取得
	_, err = sess.Select("count(*)").From("follow_list").Where("my_id = ? AND user_id != ?",userId, userId).Load(&counts.Follows)

	// フォロワー数取得
	_, err = sess.Select("count(*)").From("follow_list").Where("user_id = ? AND my_id != ?",userId, userId).Load(&counts.FollowedBy)

	u.Counts = counts
	if err != nil {
		fmt.Println(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	} else {
		return c.JSON(http.StatusOK, u)
	}
}

func GetFollowStatus(c echo.Context) error {

	if err != nil {
		return c.JSON(http.StatusOK,"DB connection error")
	}

	var myId int64
	var opponentId int64
	param := c.Param("id")
	param2 := c.Param("id2")
	myId, err = strconv.ParseInt(param, 0, 64)
	opponentId, err = strconv.ParseInt(param2, 0, 64)

	var f = model.FollowStatusResponse{OutgoingStatus: "",IncomingStatus: ""}

	followFlg, err := sess.Select("*").From("follow_list").Where("my_id = ? AND user_id = ?",myId,opponentId).Load(&f)

	if followFlg > 0 {
		f.OutgoingStatus = "follows"
	} else {
		f.OutgoingStatus = "none"
	}

	followerFlg, err := sess.Select("*").From("follow_list").Where("my_id = ? AND user_id = ?",opponentId,myId).Load(&f)

	if followerFlg > 0 {
		f.IncomingStatus = "follows"
	} else {
		f.IncomingStatus = "none"
	}

	if err != nil {
		fmt.Println(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	} else {
		return c.JSON(http.StatusOK, f)
	}
}

func GetTimeline(c echo.Context) error {

	if err != nil {
		return c.JSON(http.StatusOK,"DB connection error")
	}

	var id int64
	param := c.Param("id")
	date := c.Param("date")
	id, err = strconv.ParseInt(param, 0, 64)

	var timeline []model.TimelineResponse

	if date != "" {
		count, _ := sess.Select("m.*").From(dbr.I("follow_list").As("f")).
			Join(dbr.I("media").As("m"), "f.user_id = m.user_id").
			Where("f.my_id = ? AND m.created_time < ?", id, date).
			OrderDir("m.created_time", false).
			Limit(10).Load(&timeline)
		if count == 0 {
			return c.JSON(http.StatusOK, "表示するタイムラインがありません")
		}

	} else {
		count, _ := sess.Select("m.*").From(dbr.I("follow_list").As("f")).
			Join(dbr.I("media").As("m"), "f.user_id = m.user_id").
			Where("f.my_id = ?", id).
			OrderDir("m.created_time", false).
			Limit(10).Load(&timeline)
		if count == 0 {
			return c.JSON(http.StatusOK, "表示するタイムラインがありません")
		}
	}
	for key, value := range timeline {
		var user model.UserResponse
		var likes []model.LikesResponse
		var likeCount = 0
		var isLiked = 0

		_, err = sess.Select("u.*").From(dbr.I("user").As("u")).
			Join(dbr.I("media").As("m"), "u.user_id = m.user_id").Where("u.user_id = ? AND m.media_id = ?", value.UserID, value.MediaID).Load(&user)

		likeCount, err = sess.Select("*").From(dbr.I("media").As("m")).
			Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ?", value.MediaID).Load(&likes)

		isLiked, err = sess.Select("*").From(dbr.I("media").As("m")).
			Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ? AND l.user_id = ?", value.MediaID, id).Load(&likes)


		value.LikeCounts = likeCount
		value.User = user
		if isLiked > 0 {
			value.IsLiked = true
		}
		timeline[key] = value

	}
	fmt.Println("timeline", timeline)
	//"u.full_name","u.username"ile_picture" Where("u.user_id = ?", id)
	if err != nil {
		fmt.Println(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	} else {
		return c.JSON(http.StatusOK, timeline)
	}

}

func GetUserMedia(c echo.Context) error {

	if err != nil {
		return c.JSON(http.StatusOK,"DB connection error")
	}

	var userId int64
	param := c.Param("id")
	date := c.Param("date")
	userId, err = strconv.ParseInt(param, 0, 64)

	var userMedia []model.UserMediaResponse

	if date != "" {

		count, _ := sess.Select("m.media_id", "m.created_time", "m.picture", "m.body").From(dbr.I("media").As("m")).
			Where("m.user_id = ? AND m.created_time < ?", userId, date).
			OrderDir("m.created_time", false).
			Limit(30).Load(&userMedia)
		if count == 0 {
			return c.JSON(http.StatusOK, "表示するフォトライブラリがありません")
		}

	} else {

		count, _ := sess.Select("m.media_id", "m.created_time", "m.picture", "m.body").From(dbr.I("media").As("m")).
			Where("m.user_id = ?", userId).
			OrderDir("m.created_time", false).
			Limit(30).Load(&userMedia)
		if count == 0 {
			return c.JSON(http.StatusOK, "表示するフォトライブラリがありません")
		}

	}

	for key, value := range userMedia {
		var user model.UserResponse
		var likes []model.LikesResponse
		var likeCount = 0
		var isLiked = 0

		_, err = sess.Select("u.*").From(dbr.I("user").As("u")).
			Where("u.user_id = ?", userId).Load(&user)

		likeCount, err = sess.Select("*").From(dbr.I("media").As("m")).
			Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ?", value.MediaID).Load(&likes)

		isLiked, err = sess.Select("*").From(dbr.I("media").As("m")).
			Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ? AND l.user_id = ?", value.MediaID, userId).Load(&likes)

		value.User = user
		value.LikeCounts = likeCount
		if isLiked > 0 {
			value.IsLiked = true
		}

		userMedia[key] = value

	}
	if err != nil {
		fmt.Println(err.Error())
		return c.JSON(http.StatusOK, err.Error())
	} else {
		return c.JSON(http.StatusOK, userMedia)
	}
}

//	Post

func PostLikes(c echo.Context) error {
	like := new(model.LikesRequest)
	if err := c.Bind(like); err != nil {
		return err
	}
	_, err = sess.InsertInto("instagram.like").Columns("media_id", "user_id").Values(like.MediaID, like.UserID).Exec()

	if err != nil{
		return c.JSON(http.StatusBadRequest,err)
	}

	return c.JSON(http.StatusOK,"ok")
}

func DeleteLikes(c echo.Context) error {
	like := new(model.LikesRequest)
	if err := c.Bind(like); err != nil {
		return err
	}
	_, err = sess.DeleteFrom("instagram.like").Where("media_id = ? AND user_id = ?",like.MediaID,like.UserID).Exec()

	if err != nil{
		return c.JSON(http.StatusBadRequest,err)
	}

	return c.JSON(http.StatusOK,like)
}

