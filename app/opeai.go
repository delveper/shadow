package app

import (
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
	ModelGPT         = "gpt-3.5-turbo"
	ModelWhisper     = "whisper-1"
	ModelDavinci     = "text-davinci-003"
	ModelDavinciCode = "code-davinci-edit-001"
)

const (
	MethodEdits               = "edits"
	MethodCompletions         = "completions"
	MethodChatCompletions     = "chat/completions"
	MethodAudioTranscriptions = "audio/transcriptions"
)

// ChatCompletionRequest https://platform.openai.com/docs/guides/chat/chat-completions-beta
type ChatCompletionRequest struct {
	Model    string           `json:"model"`
	Messages []MessageRequest `json:"messages"`
}

type CompletionRequest struct {
	Model       string
	Promt       string
	Suffix      string
	MaxTokens   int
	Temperature float64
}

type MessageRequest struct {
	Role    string `json:"role"` // system | user | assistant
	Content string `json:"content"`
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

func NewOpenAI() *OpenAI {
	token := os.Getenv("OPENAI_TOKEN")

	endpoint := Endpoint{
		URL: &url.URL{
			Scheme: "https",
			Host:   apiOpenAIHost,
			Path:   path.Join(apiOpenAIPath, token),
		},
		Values: make(url.Values),
	}

	return &OpenAI{
		Client:   new(http.Client),
		Endpoint: &endpoint,
	}
}
