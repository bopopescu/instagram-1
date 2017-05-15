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

	Userinfo struct {
		ID        int    `db:"id"`
		Email     string `db:"email"`
		Firstname string `db:"first_name"`
		Lastname  string `db:"last_name"`
	}

	UserinfoJSON struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Firstname string `json:"firstName"`
		Lastname  string `json:"lastName"`
	}

	ResponseData struct {
		//User  []userinfo `json:"users"`
		Users []User `json:"user"`
	}
)
