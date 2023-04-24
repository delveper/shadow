package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

const (
	TypeBotCommand = "bot_command"
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
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileName     string `json:"file_name"`
	MimeType     string `json:"mime_type"`
	Duration     int    `json:"duration"`
	FileSize     int    `json:"file_size"`
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
	ID       int             `json:"message_id"`
	Text     string          `json:"text,omitempty"`
	From     *User           `json:"from"`
	Chat     *Chat           `json:"chat"`
	Audio    *Audio          `json:"audio,omitempty"`
	Voice    *Voice          `json:"voice,omitempty"`
	Entities []MessageEntity `json:"entities,omitempty"`
	Date     int             `json:"date"`
}

// MessageEntity https://core.telegram.org/bots/api#messageentity
type MessageEntity struct {
	Type     string `json:"type"`
	Offset   int    `json:"offset"`
	Length   int    `json:"length"`
	URL      string `json:"url,omitempty"`
	User     *User  `json:"user,omitempty"`
	Language string `json:"language,omitempty"`
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
	ParseMode string `json:"parse_mode"`
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
	}

	return &Telegram{
		Client:   new(http.Client),
		Endpoint: &endpoint,
	}
}

func (m *MessageEntity) IsCommand() bool {
	log.Printf("TYPE: %v\n", m.Type)
	return m.Type == TypeBotCommand
}

func (b *Telegram) GetUpdate(offset int) (*Update, error) {
	u := *b.Endpoint.BuildURL(MethodGetUpdates, "offset", strconv.Itoa(offset))

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request update: %w", err)
	}
	log.Printf("Udpate request: %s\n", req.URL.String())

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
		ParseMode: FormatMarkdown,
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

	log.Println(req.URL.String())

	req.Header.Set("Content-Type", "application/json")

	log.Printf("Send message request: %s\n", req.URL.String())

	resp, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	log.Printf("response: %#v", resp)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %v", resp.Status)
	}

	return nil
}

func (b *Telegram) GetVoice(id string) ([]byte, error) {
	file, err := b.getFileData(id)
	if err != nil {
		return nil, fmt.Errorf("getting file data: %w", err)
	}

	log.Printf("file %+v\n", file)

	log.Printf("Downloading file_path: %v", file.FilePath)

	audio, err := b.downloadFile(file)
	if err != nil {
		return nil, fmt.Errorf("getting file: %w", err)
	}

	if audio == nil || len(audio) == 0 {
		return nil, fmt.Errorf("empty stream")
	}

	log.Printf("File downloaded with size: %v", len(audio))

	return audio, nil
}

func (b *Telegram) getFileData(id string) (*File, error) {
	u := b.Endpoint.BuildURL(MethodGetFile, "file_id", id)

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request file: %w", err)
	}
	log.Printf("File data request: %s\n", req.URL.String())

	resp, err := b.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting file from request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	var file struct {
		Ok     bool  `json:"ok"`
		Result *File `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&file); err != nil {
		return nil, fmt.Errorf("decoding file: %w", err)
	}

	if !file.Ok {
		return nil, fmt.Errorf("unexpected status: %v", resp.Status)
	}

	return file.Result, nil
}

func (b *Telegram) downloadFile(file *File) ([]byte, error) {
	u := *b.Endpoint.URL
	u.Path = path.Join("file", u.Path, file.FilePath)

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request file: %w", err)
	}
	log.Printf("Download request: %s\n", req.URL.String())

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
		return nil, fmt.Errorf("redaing body: %w", err)
	}

	return data, nil
}
