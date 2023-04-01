package app

import (
	"encoding/json"
	"log"
	"net/http"
)

type Webhook struct {
	Telegram *Telegram
}

func NewWebhook(bot *Telegram) *Webhook {
	return &Webhook{bot}
}

func (w *Webhook) Handle(_ http.ResponseWriter, req *http.Request) {
	var upd Update
	if err := json.NewDecoder(req.Body).Decode(&upd); err != nil {
		log.Printf("Could not encode update: %v", err)
		return
	}

	log.Printf("Update: %+v\n", upd)

	if upd.ID == 0 {
		log.Printf("Invalid update: expected id != 0")
		return
	}

	if err := w.Telegram.SendMessage(upd.Message.Chat.ID, upd.Message.Text); err != nil {
		log.Printf("Failed sending reply: %v", err)
		return
	}

	log.Println("Reply sent")
}
