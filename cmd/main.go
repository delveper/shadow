package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/delveper/env"
	"github.com/delveper/shadow/app"
)

func main() {
	if err := Run(); err != nil {
		log.Fatalln(err)
	}
}

func Run() error {
	if err := env.LoadVars(); err != nil {
		return fmt.Errorf("load envar: %w", err)
	}

	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	addr := host + ":" + port

	bot := app.NewTelegram()
	webhook := app.NewWebhook(bot)
	hdl := http.HandlerFunc(webhook.Handle)

	log.Printf("Starting server on port: %s\n", port)

	if err := http.ListenAndServe(addr, hdl); err != nil {
		return fmt.Errorf("serving: %w", err)
	}

	return nil
}
