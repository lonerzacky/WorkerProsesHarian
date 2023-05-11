package main

import (
	"Gen2Job"
	"Gen2Job/functions"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type PayloadReceivedPublisher struct {
	Id          int64  `json:"id"`
	TglAwal     string `json:"tgl_awal"`
	TglAkhir    string `json:"tgl_akhir"`
	JenisProduk string `json:"jenis_produk"`
	KodeKantor  string `json:"kode_kantor"`
}

type ProsesIntervalHarian struct {
	TglAwal     string `json:"tgl_awal"`
	TglAkhir    string `json:"tgl_akhir"`
	JenisProduk string `json:"jenis_produk"`
	KodeKantor  string `json:"kode_kantor"`
}

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial(Gen2Job.Config.AMQPConnectionURL)
	handleError(err, "Can't connect to AMQP")
	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			handleError(err, "Error Defer AMQP")
		}
	}(conn)

	amqpChannel, err := conn.Channel()
	handleError(err, "Can't create a amqpChannel")

	defer func(amqpChannel *amqp.Channel) {
		err := amqpChannel.Close()
		if err != nil {
			handleError(err, "Error Defer AMQP Channel")

		}
	}(amqpChannel)

	queue, err := amqpChannel.QueueDeclare("requestQueue", true, false, false, false, nil)
	handleError(err, "Could not declare `requestQueue` queue")

	err = amqpChannel.Qos(1, 0, false)
	handleError(err, "Could not configure QoS")

	messageChannel, err := amqpChannel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	handleError(err, "Could not register consumer")

	stopChan := make(chan bool)

	go func() {
		log.Printf("Consumer ready, PID: %d", os.Getpid())
		conn := functions.ConnectDB()
		defer func(conn *sql.DB) {
			err := conn.Close()
			if err != nil {
				functions.LogInfo("Error Defer conn close")
			}
		}(conn)
		for d := range messageChannel {
			log.Printf("Received a message: %s", d.Body)
			var received PayloadReceivedPublisher
			err = json.Unmarshal(d.Body, &received)
			strJenisProduk := fmt.Sprintf("%v", received.JenisProduk)
			strTglAwal := fmt.Sprintf("%v", received.TglAwal)
			strTglAkhir := fmt.Sprintf("%v", received.TglAkhir)
			strKodeKantor := fmt.Sprintf("%v", received.KodeKantor)
			jsonData := ProsesIntervalHarian{
				TglAwal:     strTglAwal,
				TglAkhir:    strTglAkhir,
				JenisProduk: strJenisProduk,
				KodeKantor:  strKodeKantor,
			}
			request := gorequest.New()
			resp, body, errs := request.Post(""+os.Getenv("IP_SERVICE_HARIAN_AKHIR")+"/"+os.Getenv("APICODE_SERVICE_HARIAN_AKHIR")+"").
				Set("Content-Type", "application/json").
				Send(jsonData).
				End()
			if errs != nil {
				functions.LogInfo("An Error Occured, Service is Not Running or IP Not Listened")
				functions.UpdateLogConsumer("An Error Occured, Service is Not Running or IP Not Listened", 2, received.Id, conn)
			}
			log.Printf(body)
			if resp.StatusCode == http.StatusOK {
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					functions.LogInfo(err.Error())
					functions.UpdateLogConsumer(err.Error(), 2, received.Id, conn)
				}
				bodyString := string(bodyBytes)
				functions.UpdateLogConsumer(bodyString, 1, received.Id, conn)
			} else {
				functions.UpdateLogConsumer("An Error Occured, Service is Not Running or IP Not Listened", 2, received.Id, conn)
			}
			if err := d.Ack(false); err != nil {
				log.Printf("Error acknowledging message : %s", err)
			} else {
				log.Printf("Acknowledged message")
			}
		}
	}()
	// Stop for program termination
	<-stopChan

}
