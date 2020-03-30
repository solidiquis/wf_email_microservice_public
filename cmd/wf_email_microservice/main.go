package main

import (
	"log"
	"net/http"
	"os"

	wfprotobuf "github.com/Wefunder/email_microservice/pkg/api/v1"
	"github.com/joho/godotenv"
)

type application struct {
	errorLog            *log.Logger
	infoLog             *log.Logger
	webhookBatchChannel chan *wfprotobuf.Webhook
	storageSavedCounter uint64
}

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	infoLog := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)

	port := ":" + os.Getenv("PORT")

	if port == ":" {
		errorLog.Println("$PORT must be set")
		return
	}

	app := &application{
		infoLog:             infoLog,
		errorLog:            errorLog,
		webhookBatchChannel: make(chan *wfprotobuf.Webhook, 100_000),
		storageSavedCounter: 0,
	}

	go app.dequeueWebhookBatchLoop()

	server := &http.Server{
		Addr:     port,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("\nListening on tcp://localhost%s", port)
	err := server.ListenAndServe()
	errorLog.Fatal(err)
}
