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
	"time"
	"bytes"
	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

var (
	seq       = 1
	conn, err = db.ConnectDB()
	sess      = conn.NewSession(nil)
)

const location = "Asia/Tokyo"

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
	//"u.full_name","u.username"ile_picture" Where("u.user_id = ?", id)
	if err != nil {
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

	return c.JSON(http.StatusCreated,"ok")
}

func PostUser(c echo.Context) error {

	loc, err := time.LoadLocation(location)

	var u = 0
	var maxId = 0
	user := new(model.UserRequest)
	if err := c.Bind(user); err != nil {
		return err
	}

	if user.Username == "" || user.Password == "" || user.Email == "" || user.FullName == "" {
		return c.JSON(http.StatusBadRequest,"必須項目を入力してください。")
	}

	_, err = sess.Select("count(*)").From("user").Where("username = ?",user.Username).Load(&u)

	if u > 0 {
		return c.JSON(http.StatusBadRequest,"すでに同じusernameが使われています。")
	}

	_, err = sess.Select("MAX(user_id)").From("user").Load(&maxId)

	maxId += 1

	_, err = sess.InsertInto("user").
		Columns("user_id","full_name", "username", "bio", "mailaddress", "profile_picture", "created_time", "private_flg", "password").
		Values(maxId,user.FullName, user.Username, "よろしくお願いします！", user.Email, "http://storage.googleapis.com/instagram_17/man.png",time.Now().In(loc),0,user.Password).Exec()

	if err != nil{
		return c.JSON(http.StatusBadRequest,err)
	}

	_, err = sess.InsertInto("follow_list").Columns("my_id","user_id","created_time").Values(maxId,maxId,time.Now().In(loc)).Exec()
	if err != nil{
		return c.JSON(http.StatusBadRequest,err)
	}
	return c.JSON(http.StatusCreated,"登録完了")
}

func PostLogin(c echo.Context) error {

	login := new(model.LoginRequest)
	user := new(model.UserDetailResponse)
	var counts = model.CountResponse{Media:0,Follows:0,FollowedBy:0}

	if err := c.Bind(login); err != nil {
		return err
	}

	count, err := sess.Select("*").From("user").Where("username = ? AND password = ?",login.Username, login.Password).Load(&user)

	if count == 0 {
		return c.JSON(http.StatusBadRequest,"usernameまたはpasswordが間違えています。")
	}

	// 投稿数取得
	_, err = sess.Select("count(*)").From("media").Where("user_id = ?",user.UserID).Load(&counts.Media)

	// フォロー数取得
	_, err = sess.Select("count(*)").From("follow_list").Where("my_id = ? AND user_id != ?",user.UserID, user.UserID).Load(&counts.Follows)

	// フォロワー数取得
	_, err = sess.Select("count(*)").From("follow_list").Where("user_id = ? AND my_id != ?",user.UserID, user.UserID).Load(&counts.FollowedBy)

	user.Counts = counts

	if err != nil{
		return c.JSON(http.StatusBadRequest,err)
	}

	return c.JSON(http.StatusCreated,user)
}

func PostFollow(c echo.Context) error {
	loc, err := time.LoadLocation(location)
	follow := new(model.FollowRequest)
	if err := c.Bind(follow); err != nil {
		return err
	}
	_, err = sess.InsertInto("follow_list").Columns("my_id", "user_id", "created_time").Values(follow.UserID, follow.RequestedUserID,time.Now().In(loc)).Exec()

	if err != nil{
		return c.JSON(http.StatusBadRequest,"すでにフォロー済みです。")
	}

	return c.JSON(http.StatusCreated,"ok")
}


//	Delete

func DeleteLikes(c echo.Context) error {
	like := new(model.LikesRequest)
	if err := c.Bind(like); err != nil {
		return err
	}
	_, err = sess.DeleteFrom("instagram.like").Where("media_id = ? AND user_id = ?",like.MediaID,like.UserID).Exec()

	if err != nil{
		return c.JSON(http.StatusBadRequest,err)
	}

	return c.JSON(http.StatusNoContent,"ok")
}

func DeleteFollow(c echo.Context) error {
	follow := new(model.FollowRequest)

	if err := c.Bind(follow); err != nil {
		return c.JSON(http.StatusBadRequest,"ビルドエラー")
	}

	_, err = sess.DeleteFrom("follow_list").Where("my_id = ? AND user_id = ?",follow.UserID, follow.RequestedUserID).Exec()

	return c.JSON(http.StatusOK,"ok")
}

//	Put

func PutProfile(c echo.Context) error {
	//image := new(model.ImageRequest)
	userId := c.FormValue("user_id")
	fullName := c.FormValue("full_name")
	email := c.FormValue("email")
	bio := c.FormValue("bio")
	image,_ := c.FormFile("profile_picture")

	if userId == "" || fullName == "" || email == "" || bio == ""{
		return c.JSON(http.StatusBadRequest,"必須項目を入力してください。")
	}

	src, err := image.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open image")
	}
	defer src.Close()

	content := new(bytes.Buffer)
	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			// Readエラー処理
			break
		}

		content.Write(buf)
	}
	err = PutContent("instagram_17","user/user" + userId + ".jpg", content.Bytes())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to upload image")
	}

	url := "http://storage.googleapis.com/instagram_17/user/user" + userId +".jpg"
	attrsMap := map[string]interface{}{"full_name": fullName, "mailaddress": email, "bio": bio, "profile_picture": url}
	_, err = sess.Update("user").
		SetMap(attrsMap).
		Where("user_id = ?", userId).Exec()

	if err != nil{
		return c.JSON(http.StatusBadRequest,"必須項目を入力してください。")
	}

	return c.JSON(http.StatusOK,image)
}

func PutContent(bucket, path string, data []byte) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	w := client.Bucket(bucket).Object(path).NewWriter(ctx)
	defer w.Close()

	if n, err := w.Write(data); err != nil {
		return err
	} else if n != len(data) {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return nil
}