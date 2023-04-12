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

	msg := upd.Message.Text

	if upd.Message.Voice != nil {
		log.Printf("Voice message: %+v\n", upd.Message.Voice.FileID)

		data, err := w.Telegram.GetVoice(upd.Message.Voice.FileID)
		if err != nil {
			log.Printf("Failed getting voice: %v", err)
			return
		}

		log.Println("Voice received.")

		data, err = Convert(data)
		if err != nil {
			log.Printf("Failed converting voice: %v", err)
			return
		}

		log.Println("Voice converted.")

		res, err := w.OpenAI.CreateTranscription(data)
		if err != nil {
			log.Println(err)
			return
		}

		msg = res.Text
	}

	comp, err := w.OpenAI.CreateCompletion(msg)

	if err != nil {
		log.Printf("Failed gettitg completion: %v", err)
		return
	}

	if err := w.Telegram.SendMessage(upd.Message.Chat.ID, comp.Choices[0].Message.Content); err != nil {
		log.Printf("Failed sending msg: %v", err)
		return
	}

	log.Println("Reply sent")
}
