package main

import (
	"net/http"
	"os"
)

func (app *application) validateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _ := os.LookupEnv("WF_EMAIL_MICROSERVICE_TOKEN")

		if r.Header["X-Auth-Token"][0] != token {
			app.infoLog.Println(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
