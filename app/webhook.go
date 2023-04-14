package app

import (
	"encoding/json"
	"log"
	"net/http"
)

type Webhook struct {
	Telegram *Telegram
	OpenAI   *OpenAI
	Session  *ChatSession
}

func NewWebhook(bot *Telegram, gpt *OpenAI) *Webhook {
	return &Webhook{
		Telegram: bot,
		OpenAI:   gpt,
	}
}

func (w *Webhook) ServeHTTP(_ http.ResponseWriter, req *http.Request) {
	var upd Update
	{
		if err := json.NewDecoder(req.Body).Decode(&upd); err != nil {
			log.Printf("Could not encode update: %v", err)
			return
		}

		if upd.Message == nil || upd.ID == 0 {
			log.Printf("Invalid update: %+v\n", upd)
			return
		}

		log.Printf("Update: %+v\n", upd)

		if upd.Message.From.IsBot {
			log.Printf("Passing message from bot")
			return
		}
	}

	if entities := upd.Message.Entities; entities != nil && entities[0].IsCommand() {
		w.Session.Start()
		log.Printf("Starting new session.")
	}

	if voice := upd.Message.Voice; voice != nil {
		log.Printf("Voice message fi: %+v\n", voice.FileID)

		audio, err := w.Telegram.GetVoice(voice.FileID)
		if err != nil {
			log.Printf("Failed getting voice: %v", err)
			return
		}

		log.Println("Voice received.")

		audio, err = Convert(audio)
		if err != nil {
			log.Printf("Failed converting voice: %v", err)
			return
		}

		log.Println("Voice converted.")

		res, err := w.OpenAI.CreateTranscription(audio)
		if err != nil {
			log.Println(err)
			return
		}

		msg := NewUserChatMessage(res.Text)
		w.Session.AddMessage(msg)

		log.Printf("Transcription: %s", msg)
	}

	{
		comp, err := w.OpenAI.CreateCompletion(w.Session)
		if err != nil {
			log.Printf("Failed gettitg completion: %v", err)
			return
		}

		msg := NewAssistantChatMessage(comp.Choices[0].Message.Content)
		w.Session.AddMessage(msg)

		if err := w.Telegram.SendMessage(upd.Message.Chat.ID, msg.Content); err != nil {
			log.Printf("Failed sending msg: %v", err)
			return
		}

		log.Println("Reply sent")
	}
}
