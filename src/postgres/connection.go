package postgres

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"time"
)

type Record struct {
	SessionId string
	Frequency float64
	Timestamp time.Time
}

var dbase *gorm.DB

func Init() *gorm.DB {
	dsn := "host=localhost user=postgres password=qwe dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if !db.Migrator().HasTable(&Record{}) {
		err = db.AutoMigrate(&Record{})
		if err != nil {
			log.Fatal(err)
		}
	}
	return db
}

func GetDB() *gorm.DB {
	if dbase == nil {
		dbase = Init()
	}
	return dbase
}
