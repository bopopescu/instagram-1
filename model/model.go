package model

import "time"

type (

	User struct {
		UserID         int64     `db:"user_id"`
		FullName       string    `db:"full_name"`
		UserName       string    `db:"username"`
		Bio            string    `db:"bio"`
		Mailaddress    string    `db:"mailaddress"`
		ProfilePicture string    `db:"profile_picture"`
		CreatedTime    time.Time `db:"create_time"`
		PrivateFlg     int64     `db:"private_flg"`
		Token          string    `db:"token"`
	}

	UserResponse struct {
		UserID         int64     `json:"id"`
		FullName       string    `json:"full_name"`
		Username       string    `json:"username"`
		Bio            string    `json:"bio"`
		Mailaddress    string    `json:"mailaddress"`
		ProfilePicture string    `json:"profile_picture"`
		CreatedTime    string 	 `json:"created_time"`
		PrivateFlg     int64     `json:"private_flg"`
		Password          string    `json:"password"`
	}

	UserDetailResponse struct {
		UserID         int64     `json:"id"`
		FullName       string    `json:"full_name"`
		Username       string    `json:"username"`
		Bio            string    `json:"bio"`
		Mailaddress    string    `json:"mailaddress"`
		ProfilePicture string    `json:"profile_picture"`
		CreatedTime    string 	 `json:"created_time"`
		PrivateFlg     int64     `json:"private_flg"`
		Password          string    `json:"password"`
		Counts CountResponse `json:"counts"`
	}

	CountResponse struct {
		Media int `json:"media"`
		Follows int `json:"follows"`
		FollowedBy int `json:"followed_by"`
	}

	TimelineResponse struct {
		MediaID     int64	`json:"media_id"`
		UserID      int64	`json:"user_id"`
		CreatedTime string	`json:"created_time"`
		Picture     string	`json:"img_path"`
		Body        string	`json:"caption"`
		User UserResponse `json:"user"`
		LikeCounts int `json:"like_counts"`
		IsLiked bool `json:"is_liked"`
	}

	FollowStatusResponse struct {
		OutgoingStatus string `json:"outgoing_status"`
		IncomingStatus string `json:"incoming_status"`
	}

	LikesResponse struct {

	}

	UserMediaResponse struct {
		MediaID     int64	`json:"id"`
		CreatedTime string	`json:"created_time"`
		Picture     string	`json:"img_path"`
		Body        string	`json:"caption"`
		LikeCount int64 `json:"like_count"`
		User UserResponse `json:"user"`
		LikeCounts int `json:"like_counts"`
		IsLiked bool `json:"is_liked"`
	}

	LikesRequest struct {
		MediaID int `json:"media_id"`
		UserID int `json:"user_id"`
	}

	UserRequest struct {
		Username       string    `json:"username"`
		FullName       string    `json:"full_name"`
		Password string `json:"password"`
		Email string `json:"email"`
	}

	LoginRequest struct {
		Username       string    `json:"username"`
		Password string `json:"password"`
	}

	FollowRequest struct {
		UserID         int64     `json:"user_id"`
		RequestedUserID int64 `json:"requested_user_id"`
	}

)
