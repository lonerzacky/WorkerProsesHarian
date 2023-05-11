package functions

import (
	"database/sql"
	"github.com/joho/godotenv"
	"log"
	"os"
	"regexp"
	"time"
)

const projectDirName = "Gen2Job"

func ConnectDB() *sql.DB {
	projectName := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	currentWorkDirectory, _ := os.Getwd()
	rootPath := projectName.Find([]byte(currentWorkDirectory))
	err := godotenv.Load(string(rootPath) + `/.env`)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db, err := sql.Open(os.Getenv("DB_CONNECTION"), ""+os.Getenv("DB_USERNAME")+":"+os.Getenv("DB_PASSWORD")+"@tcp("+os.Getenv("DB_HOST")+":"+os.Getenv("DB_PORT")+")/"+os.Getenv("DB_DATABASE")+"")
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Nanosecond)
	if err != nil {
		panic(err)
	}
	return db
}
