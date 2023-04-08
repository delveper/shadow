package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
)

const (
	apiTelegramHost = "api.telegram.org"
	apiTelegramPath = "bot"
)

const ( // https://api.telegram.org/bot<token>/<method>?key1={val1}&key2{val2}
	MethodGetMe         = "getMe"
	MethodGetUpdates    = "getUpdates"
	MethodGetFile       = "getFile"
	MethodDeleteWebhook = "deleteWebhook" //
	MethodSetWebhook    = "setWebhook"    // ?url={your_API_server_url}
	MethodSendMessage   = "sendMessage"   // ?chat_id={chat_id}&text={text}
)

const (
	FormatHTML       = "HTML"
	FormatMarkdown   = "Markdown"
	FormatMarkdownV2 = "MarkdownV2"
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

// Audio https://core.telegram.org/bots/api#audio
type Audio struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	Duration int    `json:"duration"`
}

type Voice Audio

// File https://core.telegram.org/bots/api#file
type File struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id,omitempty"`
	FilePath     string `json:"file_path"`
	FileSize     int    `json:"file_size,omitempty"`
}

// Message https://core.telegram.org/bots/api#message
type Message struct {
	ID    int    `json:"message_id"`
	Text  string `json:"text,omitempty"`
	From  *User  `json:"from"`
	Chat  *Chat  `json:"chat"`
	Audio *Audio `json:"audio,omitempty"`
	Voice *Voice `json:"voice,omitempty"`
	Date  int    `json:"date"`
}

// Update https://core.telegram.org/bots/api#message
type Update struct {
	ID      int      `json:"update_id"`
	Message *Message `json:"message"`
}

// SendMessage https://core.telegram.org/bots/api#sendmessage
type SendMessage struct {
	ChatID    int    `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// SendVoice https://core.telegram.org/bots/api#sendvoice
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
			Scheme: DefaultSchema,
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

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request update: %w", err)
	}

	resp, err := b.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting updates from request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	var upd Update
	if err := json.NewDecoder(resp.Body).Decode(&upd); err != nil {
		return nil, fmt.Errorf("getting decoding update: %w", err)
	}

	return &upd, nil
}

func (b *Telegram) SendMessage(chatID int, text string) error {
	msg := SendMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: FormatHTML,
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(msg); err != nil {
		return fmt.Errorf("decoding request body: %w", err)
	}

	u := b.Endpoint.BuildURL(MethodSendMessage)

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &body)
	if err != nil {
		return fmt.Errorf("building request message: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %v", resp.Status)
	}

	return nil
}

func (b *Telegram) GetFileData(id string) (*File, error) {
	b.Endpoint.URL.Path = path.Join("file", b.Endpoint.URL.Path)
	u := b.Endpoint.BuildURL(MethodGetFile, "file_id", id)

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request file: %w", err)
	}

	resp, err := b.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting file from request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	var file File
	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, fmt.Errorf("decoding file: %w", err)
	}

	return &file, nil
}

func (b *Telegram) DownloadFile(file *File) ([]byte, error) {
	b.Endpoint.URL.Path = path.Join("file", b.Endpoint.URL.Path)
	u := b.Endpoint.BuildURL(MethodGetFile, "file_path", file.FilePath)

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request file: %w", err)
	}

	resp, err := b.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting file from request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %v", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	return data, nil
}
