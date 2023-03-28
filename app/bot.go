package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type Bot struct {
	Client   *http.Client
	Endpoint *Endpoint
}

func NewBot() *Bot {
	token := os.Getenv("TELEGRAM_TOKEN")
	endpoint := NewEndpoint(token)

	client := new(http.Client)

	return &Bot{
		Client:   client,
		Endpoint: endpoint,
	}
}

func (b *Bot) PullUpdates(offset int) ([]Update, error) {
	u := *b.Endpoint.URL
	u = *u.JoinPath(MethodGetUpdates)

	v := make(url.Values)
	v.Add("offset", strconv.Itoa(offset))

	u.RawPath = v.Encode()

	log.Println(u.String())

	resp, err := b.Client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("getting updates from request: %w", err)
	}

	defer resp.Body.Close()

	var updates struct {
		Result []Update `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&updates); err != nil {
		return nil, fmt.Errorf("getting decoding updates: %w", err)
	}

	return updates.Result, nil
}

func (b *Bot) SendMessage(chatID int, text string) error {
	text, err := url.QueryUnescape(text)
	if err != nil {
		return fmt.Errorf("preparing text: %w", err)
	}

	u := *b.Endpoint.URL
	u = *u.JoinPath(MethodSendMessage)

	v := make(url.Values)
	v.Add("chat_id", strconv.Itoa(chatID))
	v.Add("text", text)

	u.RawQuery = v.Encode()

	resp, err := b.Client.Get(u.String())
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	defer resp.Body.Close()

	return nil
}
