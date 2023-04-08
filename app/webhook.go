package app

import (
	"encoding/json"
	"log"
	"net/http"
)

type Webhook struct {
	Telegram *Telegram
	OpenAI   *OpenAI
}

func NewWebhook(bot *Telegram, gpt *OpenAI) *Webhook {
	return &Webhook{
		Telegram: bot,
		OpenAI:   gpt,
	}
}

func (w *Webhook) ServeHTTP(_ http.ResponseWriter, req *http.Request) {
	var upd Update
	if err := json.NewDecoder(req.Body).Decode(&upd); err != nil {
		log.Printf("Could not encode update: %v", err)
		return
	}

	log.Printf("Update: %+v\n", upd)

	if upd.Message == nil {
		log.Printf("Expected not nil message")
		return
	}

	if upd.ID == 0 {
		log.Printf("Invalid update: expected id != 0")
		return
	}

	if upd.Message.Voice != nil {
		log.Printf("Voice message: %+v\n", upd.Message.Voice.FileID)
		/*		_, err := w.OpenAI.CreateTranscription(nil)
				if err != nil {
					log.Println(err)
					return
				}
		*/
	}

	/*
		reply = upd.Message.Text

		comp, err := w.OpenAI.CreateCompletion(reply)

		if err != nil {
			log.Printf("Failed gettitg completion: %v", err)
			return
		}

		if err := w.Telegram.SendMessage(upd.Message.Chat.ID, comp.Choices[0].Message.Content); err != nil {
			log.Printf("Failed sending reply: %v", err)
			return
		}

		log.Println("Reply sent")
	*/
}
