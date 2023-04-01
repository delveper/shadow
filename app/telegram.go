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

const (
	apiTelegramHost = "api.telegram.org"
	apiTelegramPath = "bot"
)

const ( // https://api.telegram.org/bot<token>/<method>?key1={val1}&key2{val2}
	MethodGetMe         = "getMe"
	MethodGetUpdates    = "getUpdates"
	MethodDeleteWebhook = "deleteWebhook" //
	MethodSetWebhook    = "setWebhook"    // ?url={your_API_server_url}
	MethodSendMessage   = "sendMessage"   // ?chat_id={chat_id}&text={text}
)

// User https://core.telegram.org/bots/api#user
type User struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

// Chat https://core.telegram.org/bots/api#message
type Chat struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

type Audio struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
}

type Voice Audio

// Message https://core.telegram.org/bots/api#message
type Message struct {
	ID    int    `json:"message_id"`
	Text  string `json:"text,omitempty"`
	From  *User  `json:"from,omitempty"`
	Chat  *Chat  `json:"chat,omitempty"`
	Audio *Audio `json:"audio"`
	Voice *Voice `json:"voice"`
	Date  int    `json:"date"`
}

// Update https://core.telegram.org/bots/api#message
type Update struct {
	ID      int      `json:"update_id"`
	Message *Message `json:"message"`
}

// SendMessage https://core.telegram.org/bots/api#user
type SendMessage struct {
	ChatID    int    `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// SendVoice https://core.telegram.org/bots/api#user
type SendVoice struct {
	ChatID    int    `json:"chat_id"`
	Voice     *Voice `json:"voice"`
	ParseMode string `json:"parse_mode,omitempty"`
}

type Telegram struct {
	Client   *http.Client
	Endpoint *Endpoint
}

func NewTelegram() *Telegram {
	token := os.Getenv("TELEGRAM_TOKEN")

	endpoint := Endpoint{
		URL: &url.URL{
			Scheme: "https",
			Host:   apiTelegramHost,
			Path:   apiTelegramPath + token,
		},
		Values: make(url.Values),
	}

	return &Telegram{
		Client:   new(http.Client),
		Endpoint: &endpoint,
	}
}

func (b *Telegram) GetUpdate(offset int) (*Update, error) {
	u := b.Endpoint.BuildURL(MethodGetUpdates, "offset", strconv.Itoa(offset))

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)

	resp, err := b.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting updates from request: %w", err)
	}

	defer resp.Body.Close()

	var upd Update
	if err := json.NewDecoder(resp.Body).Decode(&upd); err != nil {
		return nil, fmt.Errorf("getting decoding update: %w", err)
	}

	return &upd, nil
}

func (b *Telegram) SendMessage(chatID int, text string) error {
	msg := SendMessage{
		ChatID: chatID,
		Text:   text,
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(msg); err != nil {
		return fmt.Errorf("decoding body: %w", err)
	}

	u := b.Endpoint.BuildURL(MethodSendMessage)

	req, err := http.NewRequest(http.MethodPost, u.String(), &body)
	if err != nil {
		return fmt.Errorf("building request message: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %v", resp.Status)
	}

	return nil
}
