package app

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/rs/xid"
)

const DefaultLanguage = "en"

const (
	apiOpenAIHost = "api.openai.com"
	apiOpenAIPath = "v1"
)

const (
	ModelGPT             = "gpt-3.5-turbo"
	ModelWhisper         = "whisper-1"
	ModelDavinci         = "text-davinci-003"
	ModelDavinciEdit     = "text-davinci-edit-001"
	ModelDavinciCodeEdit = "code-davinci-edit-001"
)

const (
	MethodEdits               = "edits"
	MethodCompletions         = "completions"
	MethodChatCompletions     = "chat/completions"
	MethodAudioTranscriptions = "audio/transcriptions"
)

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

type ChatSession struct {
	ID      string        `json:"id"`
	History []ChatMessage `json:"history"`
	Model   string        `json:"model"`
	Date    time.Time     `json:"date"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest https://platform.openai.com/docs/guides/chat/chat-completions-beta
type ChatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type CompletionRequest struct {
	Model       string  `json:"model"`
	Promt       string  `json:"promt"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type ChatChoice struct {
	Index   int         `json:"index"`
	Message ChatMessage `json:"message"`
}

type ChatCompletionResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int          `json:"created"`
	Choices []ChatChoice `json:"choices"`
}

// TranscriptionRequest https://platform.openai.com/docs/api-reference/audio/create
type TranscriptionRequest struct {
	File           string `json:"file"`
	Model          string `json:"model"`
	Promt          string `json:"promt,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	Temperature    string `json:"temperature,omitempty"`
	Language       string `json:"language,omitempty"`
	data           []byte
}

type TranscriptionResponse struct {
	Text  string        `json:"text"`
	Error ErrorResponse `json:"error"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type OpenAI struct {
	Client   *http.Client
	Endpoint *Endpoint
}

type BearerTransport struct {
	Token string
}

func (cs *ChatSession) Start() {
	*cs = ChatSession{
		ID:    xid.New().String(),
		Date:  time.Now(),
		Model: ModelGPT,
		History: []ChatMessage{
			{Role: RoleSystem, Content: os.Getenv("SYSTEM_MESSAGE_TUTOR")},
			{Role: RoleSystem, Content: os.Getenv("SYSTEM_MESSAGE_FORMAT")},
		},
	}
}

func (cs *ChatSession) End() {
	*cs = *new(ChatSession)
}

func (cs *ChatSession) AddSystemMessage(cont string) {
	cs.History = append(cs.History, ChatMessage{Role: RoleSystem, Content: cont})
}

func (cs *ChatSession) AddUserMessage(cont string) {
	cs.History = append(cs.History, ChatMessage{Role: RoleUser, Content: cont})
}

func (cs *ChatSession) AddBotMessage(cont string) {
	cs.History = append(cs.History, ChatMessage{Role: RoleAssistant, Content: cont})
}

func (bt *BearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+bt.Token)

	trans := http.DefaultTransport

	resp, err := trans.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("executing transaction: %w", err)
	}

	return resp, nil
}

func NewOpenAI() *OpenAI {
	endpoint := Endpoint{
		URL: &url.URL{
			Scheme: DefaultSchema,
			Host:   apiOpenAIHost,
			Path:   path.Join(apiOpenAIPath),
		},
	}

	token := os.Getenv("OPENAI_TOKEN")
	client := http.Client{Transport: &BearerTransport{Token: token}}

	return &OpenAI{
		Client:   &client,
		Endpoint: &endpoint,
	}
}

func newCompletionChatRequest(sess ChatSession) *ChatCompletionRequest {
	return &ChatCompletionRequest{
		Model:    sess.Model,
		Messages: sess.History,
	}
}

func (o *OpenAI) CreateCompletion(sess ChatSession) (*ChatCompletionResponse, error) {
	task := newCompletionChatRequest(sess)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(task); err != nil {
		return nil, fmt.Errorf("decoding request body: %w", err)
	}

	u := o.Endpoint.BuildURL(MethodChatCompletions)

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), &body)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	log.Printf("COMPLETION REQUEST: %s\n", req.URL.String())

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expecting ok, got: %d", resp.StatusCode)
	}

	defer func() { _ = resp.Body.Close() }()

	var comp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&comp); err != nil {
		return nil, fmt.Errorf("getting decoding completion: %w", err)
	}

	return &comp, nil
}

func newTranscriptionRequest(data []byte) *TranscriptionRequest {
	return &TranscriptionRequest{
		File:     "tmp/voice.mp3",
		Model:    ModelWhisper,
		Language: DefaultLanguage,
		data:     data,
	}
}

func (o *OpenAI) CreateTranscription(data []byte) (*TranscriptionResponse, error) {
	u := o.Endpoint.BuildURL(MethodAudioTranscriptions)

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	task := newTranscriptionRequest(data)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", task.File)
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}

	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, bytes.NewReader(task.data)); err != nil {
		return nil, fmt.Errorf("copying file to the buffer: %w", err)
	}

	if _, err = io.Copy(part, buf); err != nil {
		return nil, fmt.Errorf("copying file to the buffer: %w", err)
	}

	if err := writer.WriteField("model", task.Model); err != nil {
		return nil, fmt.Errorf("writing model: %w", err)
	}

	if err := writer.WriteField("language", task.Language); err != nil {
		return nil, fmt.Errorf("writing model: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("closing writer: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Body = io.NopCloser(body)

	log.Printf("TRANSCRIPTION REQUEST: %s\n", req.URL.String())

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting response: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var trans TranscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&trans); err != nil {
		return nil, fmt.Errorf("decoding transcription: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expecting ok, got: %d: %v", resp.StatusCode, trans.Error)
	}

	return &trans, nil
}
