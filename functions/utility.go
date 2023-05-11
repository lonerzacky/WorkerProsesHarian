package functions

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"github.com/vjeantet/jodaTime"
	"log"
	"os"
	"time"
)

func PrettyPrint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func GetResponse(responseCode string, responseMessage string) interface{} {
	type Response struct {
		ResponseCode    string `json:"response_code"`
		ResponseMessage string `json:"response_message"`
	}

	result := Response{
		responseCode,
		responseMessage,
	}
	AddLogResponse(result)
	return result
}

func GetResponseWithData(responseCode string, responseMessage string, responseData interface{}) interface{} {
	type Response struct {
		ResponseCode    string      `json:"response_code"`
		ResponseMessage string      `json:"response_message"`
		ResponseData    interface{} `json:"response_data"`
	}

	result := Response{
		responseCode,
		responseMessage,
		responseData,
	}
	AddLogResponse(result)
	return result
}

func AddLogResponse(result interface{}) {
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println("-------------------------RESPONSE---------------------")
	fmt.Println(string(b))
	fmt.Println("-------------------------END--------------------------")
}

func LogInfo(logstring string) {
	log.Print(logstring)
	var logs = logging.MustGetLogger("main")
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfunc} ► %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)
	logs.Info(logstring)
}

func UpdateLogConsumer(response string, status int64, id int64, conn *sql.DB) {
	tglEnd := jodaTime.Format("YYYY-MM-dd", time.Now())
	jamEnd := jodaTime.Format("HH:mm:ss", time.Now())
	_, _ = conn.Exec("UPDATE log_consumer SET response=?,tgl_end=?,jam_end=?,status=? WHERE id=?", response, tglEnd, jamEnd, status, id)
}
