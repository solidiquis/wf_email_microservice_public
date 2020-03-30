package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	wfprotobuf "github.com/Wefunder/email_microservice/pkg/api/v1"
	"github.com/golang/protobuf/proto"
	_ "github.com/heroku/x/hmetrics/onload"
)

func (app *application) index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

func (app *application) sendBatchEmails(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Println("Email batch received.")

	emailList := &wfprotobuf.EmailList{}

	if err := proto.Unmarshal(bytes, emailList); err != nil {
		app.errorLog.Println(err)
		return
	}

	successChannel := make(chan uint32, len(emailList.Users))
	failureChannel := make(chan uint32, len(emailList.Users))
	var wg sync.WaitGroup

	responseMap := map[string][]uint32{
		"success": []uint32{},
		"failure": []uint32{},
	}

	start := time.Now()

	for idx, user := range emailList.Users {
		wg.Add(1)
		go app.sendEmail(idx, user, (*emailList).TemplateId, &wg, successChannel, failureChannel)
	}

	wg.Wait()

	close(successChannel) // Need to close the channels to iterate on them
	close(failureChannel)

	delta := time.Since(start).Seconds()
	msPerEmail := (delta / float64(len(emailList.Users))) * 1000

	app.infoLog.Printf("Sending %v emails took %v seconds (%v ms per email)", len(emailList.Users), delta, msPerEmail)

	for successID := range successChannel {
		responseMap["success"] = append(responseMap["success"], successID)
	}

	for failureID := range failureChannel {
		responseMap["failure"] = append(responseMap["failure"], failureID)
	}

	response, err := proto.Marshal(&wfprotobuf.EmailResponse{
		SuccessfulUsers: responseMap["success"],
		ErroredUsers:    responseMap["failure"],
	})
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	w.Write(response)
}

func (app *application) swuWebhooks(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1_000_000)
	if err != nil {
		err := r.ParseForm()
		if err != nil {
			app.errorLog.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	formGetter := func(postForm url.Values, key string) string {
		value := postForm[key]
		if len(value) > 0 {
			return value[0]
		}
		return ""
	}

	key := formGetter(r.PostForm, "timestamp") + formGetter(r.PostForm, "token")
	secret := os.Getenv("MAILGUN_API_KEY")

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(key))

	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if expectedSignature != formGetter(r.PostForm, "signature") {
		app.infoLog.Printf("Expected signature %s, actual was %s", expectedSignature, formGetter(r.PostForm, "signature"))
		w.WriteHeader(http.StatusForbidden)
		return
	}

	timestamp, _ := strconv.ParseUint(formGetter(r.PostForm, "timestamp"), 10, 64)
	eventTimestamp, _ := strconv.ParseUint(formGetter(r.PostForm, "event-timestamp"), 10, 64)

	if formGetter(r.PostForm, "recipient") == "storage@wefunder.com" {
		app.storageSavedCounter++
		w.WriteHeader(http.StatusOK)
		return
	}

	app.webhookBatchChannel <- &wfprotobuf.Webhook{
		SwuTemplateVersionId: formGetter(r.PostForm, "swu_template_version_id"),
		Timestamp:            timestamp,
		Token:                formGetter(r.PostForm, "token"),
		Signature:            formGetter(r.PostForm, "signature"),
		Domain:               formGetter(r.PostForm, "domain"),
		ReceiptId:            formGetter(r.PostForm, "receipt_id"),
		SwuTemplateId:        formGetter(r.PostForm, "swu_template_id"),
		EventTimestamp:       eventTimestamp,
		MessageId:            formGetter(r.PostForm, "message-id"),
		Recipient:            formGetter(r.PostForm, "recipient"),
		Event:                formGetter(r.PostForm, "event"),
		BodyPlain:            formGetter(r.PostForm, "body-plain"),
	}

	if channelLength := len(app.webhookBatchChannel); channelLength < 100 {
		app.infoLog.Printf("Queue: %d webhooks, just added 1", channelLength)
	} else if channelLength%20 == 0 {
		app.infoLog.Printf("Queue: %d webhooks, just added 20", channelLength)
	}

	w.WriteHeader(http.StatusOK)
}

func (app *application) dequeueWebhookBatchLoop() {
	for {
		time.Sleep(5 * time.Second)

		webhookBatch := wfprotobuf.WebhookBatch{}
		channelLength := len(app.webhookBatchChannel)
		numWebhooksCleared := 0

		if channelLength > 1000 {
			channelLength = 1000
		} else if channelLength == 0 {
			app.infoLog.Println("Queue: 0 webhooks")
			continue
		}

		for i := 0; i < channelLength; i++ {
			webhook := <-app.webhookBatchChannel
			webhookBatch.Webhooks = append(webhookBatch.Webhooks, webhook)
			numWebhooksCleared++
		}

		payload, err := proto.Marshal(&webhookBatch)
		if err != nil {
			app.errorLog.Println(err)
			continue
		}

		client := &http.Client{}

		req, err := http.NewRequest("POST", "https://example.com/example_endpoint", bytes.NewBuffer(payload))
		req.Header.Add("Content-Type", "application/octet-stream")
		req.Header.Add("X-Auth-Token", os.Getenv("WF_EMAIL_MICROSERVICE_TOKEN"))

		res, err := client.Do(req)
		if err != nil {
			app.errorLog.Println(err)
		}
		if res.StatusCode != 200 {
			app.errorLog.Printf("Dequeuer: Got a response from wefunder.com, HTTP %d", res.StatusCode)
		} else {
			app.infoLog.Printf("Dequeuer: %d webhooks; cleared %d; avoided %d", channelLength-numWebhooksCleared, numWebhooksCleared, app.storageSavedCounter)
		}
	}
}
