package app

import (
	"encoding/json"
	"log"
	"net/http"
)

type Webhook struct {
	Telegram *Telegram
	OpenAI   *OpenAI
	Session  ChatSession
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

	if upd.Message == nil || upd.ID == 0 {
		log.Printf("Invalid update: %+v\n", upd)
		return
	}

	log.Printf("UPDATE: %+v\n", upd)

	msg := upd.Message.Text

	if upd.Message.From.IsBot {
		log.Println("Passing message from bot")
	}

	log.Printf("MesageEntities: %#v\n", upd.Message.Entities)
	if ent := upd.Message.Entities; ent != nil && ent[0].IsCommand() {
		switch upd.Message.Text {
		case "/start":
			w.Session.Start()
			log.Println("Starting new session.")

		case "/end":
			w.Session.End()
			msg = "Finishing session."
			log.Println(msg)

			if err := w.Telegram.SendMessage(upd.Message.Chat.ID, msg); err != nil {
				log.Printf("Failed sending msg: %v", err)
			}
			return

		default:
			log.Printf("Unknown command.")
			return
		}
	}

	if w.Session.History == nil {
		w.Session.Start()
		log.Printf("Starting new session.")
	}

	if voice := upd.Message.Voice; voice != nil && voice.MimeType == "audio/ogg" {
		log.Printf("Voice message ID: %+v\n", voice.FileID)

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

		msg = res.Text
		log.Printf("Transcription: %+v\n", msg)
	}

	if msg == "" {
		log.Println("Empty promt message.")
		return
	}

	w.Session.AddUserMessage(msg)

	comp, err := w.OpenAI.CreateCompletion(w.Session)
	if err != nil {
		log.Printf("Failed gettitg completion: %v", err)
		return
	}

	w.Session.AddBotMessage(comp.Choices[0].Message.Content)

	if err := w.Telegram.SendMessage(upd.Message.Chat.ID, comp.Choices[0].Message.Content); err != nil {
		log.Printf("Failed sending msg: %v", err)
		return
	}

	log.Println("Reply sent")
}
