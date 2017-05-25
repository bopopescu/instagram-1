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
	conn, Err = db.ConnectDB()
	sess      = conn.NewSession(nil)
)

const location = "Asia/Tokyo"

//----------
// Handlers
//----------

//	Get

func GetUsers(c echo.Context) error {

	var u []model.UserResponse
	_, Err := sess.Select("*").From("user").Load(&u)
	if Err != nil {
		return c.JSON(http.StatusOK, Err.Error())
	} else {
		return c.JSON(http.StatusOK, u)
	}
}

func GetUser(c echo.Context) error {

	var userId int64
	param := c.Param("id")
	userId, Err = strconv.ParseInt(param, 0, 64)
	var counts = model.CountResponse{Media:0,Follows:0,FollowedBy:0}

	var u = model.UserDetailResponse{}

	// ユーザー情報取得
	_, Err = sess.Select("*").From("user").Where("user_id = ?",userId).Load(&u)

	// 投稿数取得
	_, Err = sess.Select("count(*)").From("media").Where("user_id = ?",userId).Load(&counts.Media)

	// フォロー数取得
	_, Err = sess.Select("count(*)").From("follow_list").Where("my_id = ? AND user_id != ?",userId, userId).Load(&counts.Follows)

	// フォロワー数取得
	_, Err = sess.Select("count(*)").From("follow_list").Where("user_id = ? AND my_id != ?",userId, userId).Load(&counts.FollowedBy)

	u.Counts = counts
	if Err != nil {
		return c.JSON(http.StatusOK, Err.Error())
	} else {
		return c.JSON(http.StatusOK, u)
	}
}

func GetFollowStatus(c echo.Context) error {

	var myId int64
	var opponentId int64
	param := c.Param("id")
	param2 := c.Param("id2")
	myId, Err = strconv.ParseInt(param, 0, 64)
	opponentId, Err = strconv.ParseInt(param2, 0, 64)

	var f = model.FollowStatusResponse{OutgoingStatus: "",IncomingStatus: ""}

	followFlg, Err := sess.Select("*").From("follow_list").Where("my_id = ? AND user_id = ?",myId,opponentId).Load(&f)

	if followFlg > 0 {
		f.OutgoingStatus = "follows"
	} else {
		f.OutgoingStatus = "none"
	}

	followerFlg, Err := sess.Select("*").From("follow_list").Where("my_id = ? AND user_id = ?",opponentId,myId).Load(&f)

	if followerFlg > 0 {
		f.IncomingStatus = "follows"
	} else {
		f.IncomingStatus = "none"
	}

	if Err != nil {
		return c.JSON(http.StatusOK, Err.Error())
	} else {
		return c.JSON(http.StatusOK, f)
	}
}

func GetTimeline(c echo.Context) error {

	var id int64
	param := c.Param("id")
	date := c.Param("date")
	id, Err = strconv.ParseInt(param, 0, 64)

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

		_, Err = sess.Select("u.*").From(dbr.I("user").As("u")).
			Join(dbr.I("media").As("m"), "u.user_id = m.user_id").Where("u.user_id = ? AND m.media_id = ?", value.UserID, value.MediaID).Load(&user)

		likeCount, Err = sess.Select("*").From(dbr.I("media").As("m")).
			Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ?", value.MediaID).Load(&likes)

		isLiked, Err = sess.Select("*").From(dbr.I("media").As("m")).
			Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ? AND l.user_id = ?", value.MediaID, id).Load(&likes)


		value.LikeCounts = likeCount
		value.User = user
		if isLiked > 0 {
			value.IsLiked = true
		}
		timeline[key] = value

	}
	//"u.full_name","u.username"ile_picture" Where("u.user_id = ?", id)
	if Err != nil {
		return c.JSON(http.StatusOK, Err.Error())
	} else {
		return c.JSON(http.StatusOK, timeline)
	}

}

func GetUserMedia(c echo.Context) error {

	var userId int64
	param := c.Param("id")
	date := c.Param("date")
	userId, Err = strconv.ParseInt(param, 0, 64)

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

		_, Err = sess.Select("u.*").From(dbr.I("user").As("u")).
			Where("u.user_id = ?", userId).Load(&user)

		likeCount, Err = sess.Select("*").From(dbr.I("media").As("m")).
			Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ?", value.MediaID).Load(&likes)

		isLiked, Err = sess.Select("*").From(dbr.I("media").As("m")).
			Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ? AND l.user_id = ?", value.MediaID, userId).Load(&likes)

		value.User = user
		value.LikeCounts = likeCount
		if isLiked > 0 {
			value.IsLiked = true
		}

		userMedia[key] = value

	}
	if Err != nil {
		fmt.Println(Err.Error())
		return c.JSON(http.StatusOK, Err.Error())
	} else {
		return c.JSON(http.StatusOK, userMedia)
	}
}
func GetMedia(c echo.Context) error {

	id1 := c.Param("media_id")
	id2 := c.Param("user_id")
	mediaId, Err := strconv.ParseInt(id1, 0, 64)
	userId, Err := strconv.ParseInt(id2, 0, 64)

	var userMedia model.TimelineResponse

	count, _ := sess.Select("m.*").From(dbr.I("media").As("m")).
		Where("m.user_id = ? AND m.media_id = ?", userId,mediaId).Load(&userMedia)

	if count == 0 {
		return c.JSON(http.StatusOK, "表示するフォトライブラリがありません")
	}


	var user model.UserResponse
	var likes []model.LikesResponse
	var likeCount = 0
	var isLiked = 0

	_, Err = sess.Select("u.*").From(dbr.I("user").As("u")).
		Where("u.user_id = ?", userId).Load(&user)

	likeCount, Err = sess.Select("*").From(dbr.I("media").As("m")).
		Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ?", userMedia.MediaID).Load(&likes)

	isLiked, Err = sess.Select("*").From(dbr.I("media").As("m")).
		Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ? AND l.user_id = ?", userMedia.MediaID, userId).Load(&likes)

	userMedia.User = user
	userMedia.LikeCounts = likeCount

	if isLiked > 0 {
		userMedia.IsLiked = true
	}
	if Err != nil {
		return c.JSON(http.StatusOK, Err.Error())
	} else {
		return c.JSON(http.StatusOK, userMedia)
	}
}

func GetFollowList(c echo.Context) error {

	var userId int64
	id := c.Param("id")
	userId, Err := strconv.ParseInt(id, 0, 64)

	if Err != nil {
		return c.JSON(http.StatusOK, Err.Error())
	}
	followsId := []int64{}
	var followList []model.FollowsResponse
	count, Err := sess.Select("f.user_id").From(dbr.I("follow_list").As("f")).
		Where("f.my_id = ?", userId).
		OrderDir("f.created_time", false).
		Load(&followsId)
	if count == 1 {
		return c.JSON(http.StatusOK, "誰もフォローしていません。")
	}
	f0 := func(x int64) bool { return x == userId }
	followsId = reject_map(f0, followsId)

	for key := range followsId {
		_, Err := sess.Select("user_id","username","profile_picture","full_name").From(dbr.I("user")).
			Where("user_id = ?", followsId[key]).
			Load(&followList)
		if Err != nil {
			return c.JSON(http.StatusOK, Err.Error())
		}
	}
	return c.JSON(http.StatusOK,followList)

}

func GetFollowerList(c echo.Context) error {

	var userId int64
	id := c.Param("id")
	userId, Err := strconv.ParseInt(id, 0, 64)

	if Err != nil {
		return c.JSON(http.StatusOK, Err.Error())
	}
	followsId := []int64{}
	var followerList []model.FollowsResponse
	count, Err := sess.Select("f.my_id").From(dbr.I("follow_list").As("f")).
		Where("f.user_id = ?", userId).
		OrderDir("f.created_time", false).
		Load(&followsId)
	if count == 1 {
		return c.JSON(http.StatusOK, "誰もフォローしていません。")
	}
	f0 := func(x int64) bool {
		return x == userId
	}
	followsId = reject_map(f0, followsId)

	for key := range followsId {
		_, Err := sess.Select("user_id","username","profile_picture","full_name").From(dbr.I("user")).
			Where("user_id = ?", followsId[key]).
			Load(&followerList)
		if Err != nil {
			return c.JSON(http.StatusOK, Err.Error())
		}
	}
	return c.JSON(http.StatusOK,followerList)

}

// 配列から特定の値を抜き取る
func reject_map(f func(s int64) bool, s []int64) []int64 {
	ans := make([]int64, 0)

	for _, x := range s {
		if f(x) == false {
			ans = append(ans, x)
		}
	}
	return ans
}

//	Post

func PostLikes(c echo.Context) error {
	like := new(model.LikesRequest)
	if Err := c.Bind(like); Err != nil {
		return Err
	}
	_, Err = sess.InsertInto("instagram.like").Columns("media_id", "user_id").Values(like.MediaID, like.UserID).Exec()

	if Err != nil{
		return c.JSON(http.StatusBadRequest,Err)
	}

	return c.JSON(http.StatusCreated,"ok")
}

func PostUser(c echo.Context) error {

	loc, Err := time.LoadLocation(location)

	var u = 0
	var maxId = 0
	user := new(model.UserRequest)
	if Err := c.Bind(user); Err != nil {
		return Err
	}

	if user.Username == "" || user.Password == "" || user.Email == "" || user.FullName == "" {
		return c.JSON(http.StatusBadRequest,"必須項目を入力してください。")
	}

	_, Err = sess.Select("count(*)").From("user").Where("username = ?",user.Username).Load(&u)

	if u > 0 {
		return c.JSON(http.StatusBadRequest,"すでに同じusernameが使われています。")
	}

	_, Err = sess.Select("MAX(user_id)").From("user").Load(&maxId)

	maxId += 1

	_, Err = sess.InsertInto("user").
		Columns("user_id","full_name", "username", "bio", "mailaddress", "profile_picture", "created_time", "private_flg", "password").
		Values(maxId,user.FullName, user.Username, "よろしくお願いします！", user.Email, "http://storage.googleapis.com/instagram_17/man.png",time.Now().In(loc),0,user.Password).Exec()

	if Err != nil{
		return c.JSON(http.StatusBadRequest,Err)
	}

	_, Err = sess.InsertInto("follow_list").Columns("my_id","user_id","created_time").Values(maxId,maxId,time.Now().In(loc)).Exec()
	if Err != nil{
		return c.JSON(http.StatusBadRequest,Err)
	}
	return c.JSON(http.StatusCreated,"登録完了")
}

func PostLogin(c echo.Context) error {

	login := new(model.LoginRequest)
	user := new(model.UserDetailResponse)
	var counts = model.CountResponse{Media:0,Follows:0,FollowedBy:0}

	if Err := c.Bind(login); Err != nil {
		return Err
	}

	count, Err := sess.Select("*").From("user").Where("username = ? AND password = ?",login.Username, login.Password).Load(&user)

	if count == 0 {
		return c.JSON(http.StatusBadRequest,"usernameまたはpasswordが間違えています。")
	}

	// 投稿数取得
	_, Err = sess.Select("count(*)").From("media").Where("user_id = ?",user.UserID).Load(&counts.Media)

	// フォロー数取得
	_, Err = sess.Select("count(*)").From("follow_list").Where("my_id = ? AND user_id != ?",user.UserID, user.UserID).Load(&counts.Follows)

	// フォロワー数取得
	_, Err = sess.Select("count(*)").From("follow_list").Where("user_id = ? AND my_id != ?",user.UserID, user.UserID).Load(&counts.FollowedBy)

	user.Counts = counts

	if Err != nil{
		return c.JSON(http.StatusBadRequest,Err)
	}

	return c.JSON(http.StatusCreated,user)
}

func PostFollow(c echo.Context) error {
	loc, Err := time.LoadLocation(location)
	follow := new(model.FollowRequest)
	if Err := c.Bind(follow); Err != nil {
		return Err
	}
	_, Err = sess.InsertInto("follow_list").Columns("my_id", "user_id", "created_time").Values(follow.UserID, follow.RequestedUserID,time.Now().In(loc)).Exec()

	if Err != nil{
		return c.JSON(http.StatusBadRequest,"すでにフォロー済みです。")
	}

	return c.JSON(http.StatusCreated,"ok")
}

func PostMedia(c echo.Context) error {

	loc, Err := time.LoadLocation(location)
	userId := c.FormValue("user_id")
	caption := c.FormValue("caption")
	image,_ := c.FormFile("image")

	if userId == "" || caption == ""{
		return c.JSON(http.StatusBadRequest,"必須項目を入力してください。")
	}

	src, Err := image.Open()
	if Err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open image")
	}
	defer src.Close()

	content := new(bytes.Buffer)
	buf := make([]byte, 1024)
	for {
		n, Err := src.Read(buf)
		if n == 0 {
			break
		}
		if Err != nil {
			// Readエラー処理
			break
		}

		content.Write(buf)
	}

	result, Err := sess.InsertInto("media").
		Columns("user_id","created_time","body").
		Values(userId,time.Now().In(loc),caption).
		Exec()
	fmt.Printf(caption)
	if Err != nil{
		return c.JSON(http.StatusBadRequest,Err)
	}

	var userMedia model.UserMediaResponse
	if mediaId,Err := result.LastInsertId(); Err == nil {
		Err = PutContent("instagram_17","media/" + strconv.FormatInt(mediaId,10) + ".jpg", content.Bytes())
		if Err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to upload image")
		}

		url := "http://storage.googleapis.com/instagram_17/media/" + strconv.FormatInt(mediaId,10) +".jpg"

		attrsMap := map[string]interface{}{"picture": url}
		_, Err = sess.Update("media").
			SetMap(attrsMap).
			Where("media_id = ?", mediaId).Exec()

		count, _ := sess.Select("m.*").From(dbr.I("media").As("m")).
			Where("m.media_id = ?", mediaId).Load(&userMedia)
		if count == 0 {
			return c.JSON(http.StatusOK, "表示する投稿はありません")
		}

	}

	var user model.UserResponse
	var likes []model.LikesResponse
	var likeCount = 0
	var isLiked = 0




	_, Err = sess.Select("u.*").From(dbr.I("user").As("u")).
		Where("u.user_id = ?", userId).Load(&user)

	likeCount, Err = sess.Select("*").From(dbr.I("media").As("m")).
		Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ?", userMedia.MediaID).Load(&likes)

	isLiked, Err = sess.Select("*").From(dbr.I("media").As("m")).
		Join(dbr.I("like").As("l"), "l.media_id = m.media_id").Where("l.media_id = ? AND l.user_id = ?", userMedia.MediaID, userId).Load(&likes)

	userMedia.User = user
	userMedia.LikeCounts = likeCount

	if isLiked > 0 {
		userMedia.IsLiked = true
	}

	if Err != nil{
		return c.JSON(http.StatusBadRequest,Err)
	}
	if Err != nil {
		return c.JSON(http.StatusOK, Err.Error())
	} else {
		return c.JSON(http.StatusOK, userMedia)
	}

}


//	Delete

func DeleteLikes(c echo.Context) error {
	like := new(model.LikesRequest)
	if Err := c.Bind(like); Err != nil {
		return Err
	}
	_, Err = sess.DeleteFrom("instagram.like").Where("media_id = ? AND user_id = ?",like.MediaID,like.UserID).Exec()

	if Err != nil{
		return c.JSON(http.StatusBadRequest,Err)
	}

	return c.JSON(http.StatusNoContent,"ok")
}

func DeleteFollow(c echo.Context) error {
	follow := new(model.FollowRequest)

	if Err := c.Bind(follow); Err != nil {
		return c.JSON(http.StatusBadRequest,"ビルドエラー")
	}

	_, Err = sess.DeleteFrom("follow_list").Where("my_id = ? AND user_id = ?",follow.UserID, follow.RequestedUserID).Exec()

	return c.JSON(http.StatusOK,"ok")
}

//	Put

func PutProfile(c echo.Context) error {
	userId := c.FormValue("user_id")
	fullName := c.FormValue("full_name")
	email := c.FormValue("email")
	bio := c.FormValue("bio")
	image,_ := c.FormFile("profile_picture")

	if userId == "" || fullName == "" || email == "" || bio == ""{
		return c.JSON(http.StatusBadRequest,"必須項目を入力してください。")
	}

	src, Err := image.Open()
	if Err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open image")
	}
	defer src.Close()

	content := new(bytes.Buffer)
	buf := make([]byte, 1024)
	for {
		n, Err := src.Read(buf)
		if n == 0 {
			break
		}
		if Err != nil {
			// Readエラー処理
			break
		}

		content.Write(buf)
	}
	Err = PutContent("instagram_17","user/user" + userId + ".jpg", content.Bytes())
	if Err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to upload image")
	}

	url := "http://storage.googleapis.com/instagram_17/user/user" + userId +".jpg"
	attrsMap := map[string]interface{}{"full_name": fullName, "mailaddress": email, "bio": bio, "profile_picture": url}
	_, Err = sess.Update("user").
		SetMap(attrsMap).
		Where("user_id = ?", userId).Exec()

	if Err != nil{
		return c.JSON(http.StatusBadRequest,"必須項目を入力してください。")
	}

	return c.JSON(http.StatusOK,image)
}

func PutContent(bucket, path string, data []byte) error {
	ctx := context.Background()
	client, Err := storage.NewClient(ctx)
	if Err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, Err)
	}

	w := client.Bucket(bucket).Object(path).NewWriter(ctx)
	defer w.Close()

	if n, Err := w.Write(data); Err != nil {
		return Err
	} else if n != len(data) {
		return Err
	}
	if Err := w.Close(); Err != nil {
		return Err
	}

	return nil
}