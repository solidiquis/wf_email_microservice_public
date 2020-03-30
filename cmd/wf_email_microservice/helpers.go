package main

import (
	"os"
	"strings"
	"sync"

	wfprotobuf "github.com/Wefunder/email_microservice/pkg/api/v1"
	swu "github.com/elbuo8/sendwithus_go"
)

func (app *application) sendEmail(wID int, user *wfprotobuf.User, templateID string, wg *sync.WaitGroup, sChan, fChan chan uint32) {
	defer func() {
		if err := recover(); err != nil {
			app.errorLog.Printf("Failed to send email for %s", user.Email)
			app.errorLog.Println(err)
		}
	}()
	defer wg.Done()

	api := swu.New(os.Getenv("SWU_API_KEY"))

	data := make(map[string]string)

	data["user_first_name"] = user.FirstName
	data["user_auto_login_token"] = user.AutoLoginToken

	bcc := &swu.SWURecipient{
		Address: "storage@wefunder.com",
	}

	email := &swu.SWUEmail{
		ID:         templateID,
		ESPAccount: os.Getenv("NEWSLETTER_ESP_ID"),
		Recipient: &swu.SWURecipient{
			Address: strings.Trim(user.Email, " "),
			Name:    strings.Trim(user.FullName, " "),
		},
		BCC: []*swu.SWURecipient{bcc},
		Sender: &swu.SWUSender{
			SWURecipient: swu.SWURecipient{
				Address: "hello@wefunder.com",
				Name:    "Wefunder",
			},
			ReplyTo: "hello@wefunder.com",
		},
		EmailData: data,
	}

	err := api.Send(email)

	if err != nil {
		app.errorLog.Println(err)
		fChan <- user.Id
	} else {
		sChan <- user.Id
	}
}
