package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
)

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

const (
	TaskTutor  = "You are a helpful assistant who can help me improve my English. Fix grammar, give possible guidelines to improve clarity."
	TaskAccent = "Render your response using: https://core.telegram.org/bots/api#html-style (do not mention it in response). TEXT TO ANALYZE: "
)

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

	return trans.RoundTrip(req)
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

func buildChatRequest(model, promt, text string) *ChatCompletionRequest {
	task := ChatMessage{Role: RoleAssistant, Content: promt}
	cont := ChatMessage{Role: RoleUser, Content: TaskAccent + text}
	msg := []ChatMessage{task, cont}

	return &ChatCompletionRequest{
		Model:    model,
		Messages: msg,
	}
}

func (o *OpenAI) CreateCompletion(msg string) (*ChatCompletionResponse, error) {
	task := buildChatRequest(ModelGPT, TaskTutor, msg)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(task); err != nil {
		return nil, fmt.Errorf("decoding request body: %w", err)
	}

	u := o.Endpoint.BuildURL(MethodChatCompletions)

	req, err := http.NewRequest(http.MethodPost, u.String(), &body)
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

	defer resp.Body.Close()

	var comp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&comp); err != nil {
		return nil, fmt.Errorf("getting decoding completion: %w", err)
	}

	return &comp, nil
}
