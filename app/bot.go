package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const ( // https://api.telegram.org/bot<token>/<method>?key1={val1}&key2{val2}
	MethodGetMe         = "getMe"
	MethodGetUpdates    = "getUpdates"
	MethodDeleteWebhook = "deleteWebhook" //
	MethodSetWebhook    = "setWebhook"    // ?url={your_API_server_url}
	MethodSendMessage   = "sendMessage"   // ?chat_id={chat_id}&text={text}
)

type Endpoint struct {
	URL    *url.URL
	Values url.Values
}

func (e *Endpoint) BuildURL(method string, args ...string) *url.URL {
	for i := 0; i < len(args); i += 2 {
		k, v := args[i], args[i+1]
		e.Values.Add(k, v)
	}

	u := *e.URL.JoinPath(method)
	u.RawQuery = e.Values.Encode()

	return &u
}

type Bot struct {
	Client   *http.Client
	Endpoint *Endpoint
}

func NewBot() *Bot {
	e := Endpoint{
		URL: &url.URL{
			Scheme: "https",
			Host:   "api.telegram.org",
			Path:   "bot" + os.Getenv("TELEGRAM_TOKEN"),
		},
		Values: url.Values{},
	}

	return &Bot{
		Client:   &http.Client{},
		Endpoint: &e,
	}
}

func (b *Bot) PullUpdates(offset int) ([]Update, error) {
	u := b.Endpoint.BuildURL(MethodGetUpdates, "offset", strconv.Itoa(offset))

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
	msg := SendMessage{
		ChatID: chatID,
		Text:   text,
	}

	var body *bytes.Buffer
	if err := json.NewEncoder(body).Encode(msg); err != nil {
		return fmt.Errorf("decoding body: %w", err)
	}

	u := b.Endpoint.BuildURL(MethodSendMessage)

	resp, err := b.Client.Post(u.String(), "application/json", body)
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %v", resp.Status)
	}

	return nil
}
