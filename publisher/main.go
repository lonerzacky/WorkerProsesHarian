package main

import (
	"Gen2Job/functions"
	"Gen2Job/routers"
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kpango/glg"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

const projectDirName = "Gen2Job"

func main() {
	projectName := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	currentWorkDirectory, _ := os.Getwd()
	rootPath := projectName.Find([]byte(currentWorkDirectory))
	err := godotenv.Load(string(rootPath) + `/.env`)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	_ = glg.Log("Application name : ", os.Getenv("APP_NAME"))
	var clog []string
	clog = append(clog, "Version App : "+os.Getenv("APP_VERSION")+"\n")
	_ = glg.Log("Connecting Database ...")
	err = functions.MysqlConnect(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_DATABASE"))
	if err != nil {
		_ = glg.Error("Connection database error : ", err.Error())
		os.Exit(1)
	}
	_ = glg.Log("Starting Services ...")
	err = Start()
	if err != nil {
		_ = glg.Log(err.Error())
	}
}

func Start() error {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(RequestLogger())
	routers.SetupRouter(router)
	portHttp := fmt.Sprint(":", os.Getenv("APP_PORT"))
	_ = glg.Log("[HTTP] Listening at ", portHttp)
	_ = router.Run(portHttp)
	return nil
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		buf, _ := ioutil.ReadAll(c.Request.Body)
		req, _ := functions.PrettyPrint(buf)
		rdr1 := ioutil.NopCloser(bytes.NewBuffer(req))
		rdr2 := ioutil.NopCloser(bytes.NewBuffer(req))
		fmt.Println("--------------------------REQUEST----------------------")
		fmt.Println(readBody(rdr1))
		c.Request.Body = rdr2
		c.Next()
	}
}

func readBody(reader io.Reader) string {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(reader)
	s := buf.String()
	return s
}
