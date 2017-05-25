package db

import (
	"fmt"

	"github.com/gocraft/dbr"
	"github.com/spf13/viper"
)

func ConnectDB() (*dbr.Connection, error) {
	viper.SetDefault("db.user", "root")
	viper.SetDefault("db.password", "instagram17")//instagram17
	viper.SetDefault("db.host", "104.198.88.163")//104.198.88.163
	viper.SetDefault("db.port", 3306)
	viper.SetDefault("db.database", "instagram")
	viper.SetDefault("db.charset", "utf8mb4")

	dburl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
		viper.GetString("db.user"), viper.GetString("db.password"), viper.GetString("db.host"), viper.GetInt("db.port"), viper.GetString("db.database"), viper.GetString("db.charset"))
	//dburl := fmt.Sprintf("%s:%s@unix(%s)/%s?charset=%s",
	//	viper.GetString("db.user"), viper.GetString("db.password"), viper.GetString("db.host"), viper.GetString("db.database"), viper.GetString("db.charset"))

	fmt.Println(dburl)
	db, err := dbr.Open("mysql", dburl, nil)

	if err != nil {
		return nil, err
	}

	return db, nil
}
