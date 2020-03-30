package main

import (
	"net/http"

	"github.com/bmizerany/pat"
)

func (app *application) routes() *pat.PatternServeMux {
	mux := pat.New()

	// Routes
	mux.Get("/", http.HandlerFunc(app.index))
	mux.Post("/send_batch", app.validateRequest(http.HandlerFunc(app.sendBatchEmails)))
	mux.Post("/webhooks", http.HandlerFunc(app.swuWebhooks))

	return mux
}
