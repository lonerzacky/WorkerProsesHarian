package controllers

import (
	"Gen2Job"
	"Gen2Job/functions"
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"github.com/vjeantet/jodaTime"
	"math/rand"
	"net/http"
	"time"
)

type IntervalParams struct {
	TglAwal     string `json:"tgl_awal"`
	TglAkhir    string `json:"tgl_akhir"`
	JenisProduk string `json:"jenis_produk"`
	KodeKantor  string `json:"kode_kantor"`
}

type PayloadPublisher struct {
	Id          int64  `json:"id"`
	TglAwal     string `json:"tgl_awal"`
	TglAkhir    string `json:"tgl_akhir"`
	JenisProduk string `json:"jenis_produk"`
	KodeKantor  string `json:"kode_kantor"`
}

func handleError(err error, msg string, context *gin.Context) {
	if err != nil {
		context.JSON(http.StatusOK, functions.GetResponseWithData("01", msg, err.Error()))
	}
}

func RequestInterval(context *gin.Context) {
	var intervalParams IntervalParams
	err := context.BindJSON(&intervalParams)
	if err != nil {
		handleError(err, "Json Parse Error", context)
	}
	conn := functions.ConnectDB()
	defer func(conn *sql.DB) {
		err := conn.Close()
		if err != nil {
			handleError(err, "Error Defer Connection", context)
		}
	}(conn)

	stmt, err := conn.Prepare("INSERT INTO log_consumer(kode_kantor,type,name,payload,tgl_start,jam_start,tgl_awal,tgl_akhir)VALUES(?,?,?,?,?,?,?,?)")
	if err != nil {
		handleError(err, "Koneksi Error", context)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			handleError(err, "Error Defer Statement", context)
		}
	}(stmt)
	typeProses := "Interval"
	tglStart := jodaTime.Format("YYYY-MM-dd", time.Now())
	jamStart := jodaTime.Format("HH:mm:ss", time.Now())
	jsonData, err := json.Marshal(intervalParams)
	if err != nil {
		handleError(err, "Error Marshalling Payload", context)
	}
	cKodeKantor := intervalParams.KodeKantor
	if cKodeKantor == "" {
		cKodeKantor = "semua"
	}
	res, err := stmt.Exec(cKodeKantor, typeProses, intervalParams.JenisProduk, string(jsonData), tglStart, jamStart, intervalParams.TglAwal, intervalParams.TglAkhir)
	if err != nil {
		handleError(err, "Insert Log Error", context)
	}
	lid, err := res.LastInsertId()
	connJob, err := amqp.Dial(Gen2Job.Config.AMQPConnectionURL)
	handleError(err, "Can't connect to AMQP", context)
	defer func(connJob *amqp.Connection) {
		err := connJob.Close()
		if err != nil {
			handleError(err, "Error Defer AMQP", context)
		}
	}(connJob)
	amqpChannel, err := connJob.Channel()
	handleError(err, "Can't create a amqpChannel", context)
	defer func(amqpChannel *amqp.Channel) {
		err := amqpChannel.Close()
		if err != nil {
			handleError(err, "Error Defer AMQP Channel", context)
		}
	}(amqpChannel)
	queue, err := amqpChannel.QueueDeclare("requestQueue", true, false, false, false, nil)
	handleError(err, "Could not declare `requestQueue` queue", context)
	rand.Seed(time.Now().UnixNano())
	payloadPublisher := PayloadPublisher{
		Id:          lid,
		TglAwal:     intervalParams.TglAwal,
		TglAkhir:    intervalParams.TglAkhir,
		JenisProduk: intervalParams.JenisProduk,
		KodeKantor:  intervalParams.KodeKantor,
	}
	bodyBytes := new(bytes.Buffer)
	err = json.NewEncoder(bodyBytes).Encode(payloadPublisher)
	if err != nil {
		handleError(err, "Error Convert Bytes Publisher", context)
	}
	err = amqpChannel.Publish("", queue.Name, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         bodyBytes.Bytes(),
	})
	if err != nil {
		handleError(err, "Error publishing message: %s", context)
	}
	context.JSON(http.StatusOK, functions.GetResponseWithData("00", "Proses Interval Produk "+intervalParams.JenisProduk+" Diterima", lid))
}
