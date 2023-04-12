package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
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
	Promt          string `json:"promt"`
	ResponseFormat string `json:"response_format"`
	Temperature    string `json:"temperature"`
	Language       string `json:"language"`
	data           []byte
}

type TranscriptionResponse struct {
	Text string `json:"text"`
}

type OpenAI struct {
	Client   *http.Client
	Endpoint *Endpoint
}

type BearerTransport struct {
	Token string
}

func (bt *BearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+bt.Token)
	req.Header.Set("Content-Type", "application/json")

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
		Values: make(url.Values),
	}

	token := os.Getenv("OPENAI_TOKEN")
	client := http.Client{Transport: &BearerTransport{Token: token}}

	return &OpenAI{
		Client:   &client,
		Endpoint: &endpoint,
	}
}

func newCompletionChatRequest(promt, text string) *ChatCompletionRequest {
	accent := os.Getenv("PROMT_TUTOR_ACCENT")
	task := ChatMessage{Role: RoleAssistant, Content: promt}
	cont := ChatMessage{Role: RoleUser, Content: accent + text}

	return &ChatCompletionRequest{
		Model:    ModelGPT,
		Messages: []ChatMessage{task, cont},
	}
}

func (o *OpenAI) CreateCompletion(text string) (*ChatCompletionResponse, error) {
	promt := os.Getenv("PROMT_TUTOR")
	task := newCompletionChatRequest(text, promt)

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

	log.Printf("Request: %+v\n", req)

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
		File:     "audio.mp3",
		Model:    ModelWhisper,
		Promt:    os.Getenv("PROMT_TRANSCRIPTION"),
		Language: DefaultLanguage,
		data:     data,
	}
}

func newMultipartFormData(tr *TranscriptionRequest) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)

	part, err := w.CreateFormFile("file", tr.File)
	if err != nil {
		return nil, fmt.Errorf("creating form file: %w", err)
	}

	if _, err := io.Copy(part, bytes.NewReader(tr.data)); err != nil {
		return nil, fmt.Errorf("copying file: %w", err)
	}

	if err := w.WriteField("model", tr.Model); err != nil {
		return nil, fmt.Errorf("writing model field: %w", err)
	}

	if err := w.WriteField("promt", tr.Promt); err != nil {
		return nil, fmt.Errorf("writing promt field: %w", err)
	}

	if err := w.WriteField("response_format", tr.ResponseFormat); err != nil {
		return nil, fmt.Errorf("writing response format field: %w", err)
	}

	if err := w.WriteField("temperature", tr.Temperature); err != nil {
		return nil, fmt.Errorf("writing temperature field: %w", err)
	}

	if err := w.WriteField("language", tr.Language); err != nil {
		return nil, fmt.Errorf("writing language field: %w", err)
	}

	return buf, nil
}

func (o *OpenAI) CreateTranscription(data []byte) (*TranscriptionResponse, error) {
	task := newTranscriptionRequest(data)

	body, err := newMultipartFormData(task)
	if err != nil {
		return nil, fmt.Errorf("creating multipart: %w", err)
	}

	u := o.Endpoint.BuildURL(MethodAudioTranscriptions)

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	req.Header.Set("Content-Type", "multipart/form-data")

	log.Printf("Request: %+v\n", req)

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting response: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expecting ok, got: %d", resp.StatusCode)
	}

	var trans TranscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&trans); err != nil {
		return nil, fmt.Errorf("decoding transcription: %w", err)
	}

	return &trans, nil
}
