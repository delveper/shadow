```go
package main

import (
	"log"

	"github.com/delveper/revalid"
)

func main() {
	user := struct {
		Name string `regex:"^[\p{L}&\s-\\'’.]{2,256}$"`
	}{
		Name: "Fudfe",
	}

	err := revalid.ValidateStruct(user)
	log.Println(err)
}
/********************************END_OF_FILE************************************/
package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

func executeTemplate(rw http.ResponseWriter, path string) error {
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return fmt.Errorf("failed parsing template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := tmpl.Execute(buf, nil); err != nil {
		return fmt.Errorf("failed execute template: %w", err)
	}

	if _, err := fmt.Fprintf(rw, "%s", buf); err != nil {
		http.Error(rw, "Failed writing template", http.StatusInternalServerError)
		return fmt.Errorf("error writing template: %v", err)
	}

	return nil
}
/********************************END_OF_FILE************************************/
package main

import "fmt"

func main() {
	a := 0.1
	b := 0.2
	fmt.Println(a + b)
}
/********************************END_OF_FILE************************************/
package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Load() (err error) {
	env, err := os.Open(".env")
	if err != nil {
		return fmt.Errorf("error during opening environment file: %w", err)
	}

	defer func() {
		if err = env.Close(); err != nil {
			err = fmt.Errorf("error during closing environment file: %w", err)
		}
	}()

	buf := bufio.NewScanner(env)
	buf.Split(bufio.ScanLines)

	for buf.Scan() {
		if keyVal := strings.Split(buf.Text(), "="); len(keyVal) > 1 {
			if err := os.Setenv(keyVal[0], keyVal[1]); err != nil {
				return fmt.Errorf("error during setting environment variable: %w", err)
			}
		}
	}

	return nil
}
/********************************END_OF_FILE************************************/
package chrome

import (
	"context"
	"errors"
	"log"
	"path"
	"time"

	"github.com/chromedp/chromedp"
)

func ParseHTML(uri, dir string) (string, error) {
	const defaultTimeout = 15 * time.Second

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.UserDataDir(path.Join("./tmp/chrome/", dir)),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-gpu", false),
	)

	browser, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	browser, cancel = chromedp.NewContext(browser, chromedp.WithErrorf(log.Printf))
	defer cancel()

	var html string
	if err := chromedp.Run(browser, chromedp.Tasks{
		chromedp.Navigate(uri),
		chromedp.Sleep(defaultTimeout / 3),
		chromedp.InnerHTML("//html", &html, chromedp.BySearch),
	}); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		log.Println(err)
	}

	if html == "" {
		return "", errors.New("empty html")
	}

	return html, nil
}
/********************************END_OF_FILE************************************/
package filesys

import (
	"fmt"
	"os"
)

func AppendTextToFile(fileName, text string) error {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(text); err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}
/********************************END_OF_FILE************************************/
package nlp

import (
	"regexp"
)

func GetLinkedinID(uri string) (id string) {
	re, err := regexp.Compile(`(?:(/pub/)|(/in/)|(/company/))([^/]*)`)
	if err != nil {
		return ""
	}

	if match := re.FindStringSubmatch(uri); len(match) > 0 {
		id = match[len(match)-1]
	}

	return
}
/********************************END_OF_FILE************************************/
package query

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const DefaultSchema = "https"

type Option func(*url.URL)

// Query build ulr for all needs
func Query(phrase string, options ...Option) string {
	// set default for a start
	query := &url.URL{
		Scheme: DefaultSchema,
	}

	phrase = strings.Trim(phrase, " ")

	values := &url.Values{}

	values.Set("q", phrase)
	query.RawQuery = values.Encode()

	for _, o := range options {
		o(query)
	}

	return query.String()
}

/*Func options*/

func WithSchema(schema string) Option {
	return func(u *url.URL) {
		u.Scheme = schema
	}
}

func WithPath(path string) Option {
	return func(u *url.URL) {
		u.Path = path
	}
}

func WithHost(host string) Option {
	return func(u *url.URL) {
		u.Host = host
	}
}

func WithPrefix(prefix string) Option {
	return func(u *url.URL) {
		values := u.Query()
		q := values.Get("q")
		values.Set("q", prefix+" "+q)
		u.RawQuery = values.Encode()
	}
}

func WithPostfix(postfix string) Option {
	return func(u *url.URL) {
		values := u.Query()
		q := values.Get("q")
		values.Set("q", q+" "+postfix)
		u.RawQuery = values.Encode()
	}
}

func WithExact() Option {
	return func(u *url.URL) {
		values := u.Query()
		q := strconv.Quote(values.Get("q"))
		values.Set("q", q)
		u.RawQuery = values.Encode()
	}
}

func WithValue(key string, value any) Option {
	return func(u *url.URL) {
		values := u.Query()
		values.Add(key, fmt.Sprint(value))
		u.RawQuery = values.Encode()
	}
}
/********************************END_OF_FILE************************************/
package tmp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/chromedp/chromedp"
	"github.com/delveper/empeek/cmd/roadshow"
	"github.com/delveper/empeek/pkg/extraction/linkedin"
)

func parseHIMSS(id string) main.Organization {
	const baseURL = "https://himss23.mapyourshow.com/8_0/exhibitor/exhibitor-details.cfm?exhid="
	const defaultTimeout = 1_000 * time.Second

	// basic context
	ctx, cancel := chromedp.NewExecAllocator(context.Background(),
		linkedin.SetSessionOptions("roadshow", false)...,
	)
	defer cancel()

	// init chrome instance
	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// set timeout
	ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var html string
	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(baseURL + id),
		chromedp.WaitVisible(`footer`, chromedp.ByQuery),
		chromedp.ScrollIntoView(`footer`, chromedp.ByQuery),
		chromedp.Sleep(3 * time.Second),
		linkedin.WithGetHTML(&html),
	})

	if err != nil {
		log.Fatalf("Failed parse org: %v", err)
	}

	if html == "" {
		log.Fatal("empty HTML")
	}

	doc, err := htmlquery.Parse(strings.NewReader(html))
	if err != nil {
		log.Fatalf("error parsing HTML: %v", err)
	}

	var org main.Organization
	org.ID = id

	if node := htmlquery.FindOne(doc, `//div[@id="showroomTopContentDiv"]//h1`); node != nil {
		org.OrganizationName = htmlquery.InnerText(node)
	}

	if node := htmlquery.FindOne(doc, `//a[@title="Visit our website"]`); node != nil {
		org.Website = htmlquery.SelectAttr(node, `href`)
	}

	if node := htmlquery.FindOne(doc, `//div[contains(@class, "showcase-social")]/a[contains(@href, "linkedin")]`); node != nil {
		org.Linkedin = htmlquery.SelectAttr(node, `href`)
	}

	if node := htmlquery.FindOne(doc, `//div[contains(@class, "showcase-social")]/a[contains(@href, "twitter")]`); node != nil {
		org.Twitter = htmlquery.SelectAttr(node, `href`)
	}

	if node := htmlquery.FindOne(doc, `//div[contains(@class, "showcase-social")]/a[contains(@href, "instagram")]`); node != nil {
		org.Instagram = htmlquery.SelectAttr(node, `href`)
	}

	if node := htmlquery.FindOne(doc, `//div[contains(@class, "showcase-social")]/a[contains(@href, "facebook")]`); node != nil {
		org.Facebook = htmlquery.SelectAttr(node, `href`)
	}

	if node := htmlquery.FindOne(doc, `//p[@class="js-read-more animated"]`); node != nil {
		org.Description = htmlquery.InnerText(node)
	}

	if node := htmlquery.FindOne(doc, `//p[contains(@class, "showcase-address")]`); node != nil {
		org.Location = strings.TrimSpace(htmlquery.InnerText(node))
	}

	if nodes := htmlquery.Find(doc, `//section[@id="scroll-regions"]//li[@class="o-List_Columns_Item  lh-list"]`); nodes != nil {
		var list []string
		for _, node := range nodes {
			list = append(list, htmlquery.InnerText(node))
		}
		org.Location = strings.Join(list, "; ")
	}

	if nodes := htmlquery.Find(doc, `//section[@id="scroll-products"]//div[@role="listitem"]/h2`); nodes != nil {
		var list []string
		for _, node := range nodes {
			list = append(list, htmlquery.InnerText(node))
		}
		org.Industry = strings.Join(list, "; ")
	}

	return org
}

func parseHIMSSOrganizationList(seedURL string) []main.Organization {
	var organizations []main.Organization
	var html string

	const defaultTimeout = 1_000 * time.Second

	// basic context
	ctx, cancel := chromedp.NewExecAllocator(context.Background(),
		linkedin.SetSessionOptions("roadshow", false)...,
	)
	defer cancel()

	// init chrome instance
	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// set timeout
	ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(seedURL),
		chromedp.WaitVisible(`footer`, chromedp.ByQuery),
		chromedp.ScrollIntoView(`footer`, chromedp.ByQuery),
		chromedp.Sleep(3 * time.Second),
		chromedp.WaitVisible(`//a[@id="exhibitor"]/@href`, chromedp.BySearch),
		chromedp.Click(`//a[@id="exhibitor"]/@href`, chromedp.BySearch),
		chromedp.Click(`//a[@title="View results by list"]/@href`, chromedp.BySearch),
		chromedp.ActionFunc(paginateHIMMS),
		chromedp.Sleep(120 * time.Second),
		linkedin.WithGetHTML(&html),
	})
	if err != nil {
		log.Fatalf("Failed parse orgs: %v", err)
	}

	if html == "" {
		log.Fatal("empty HTML")
	}

	doc, err := htmlquery.Parse(strings.NewReader(html))
	if err != nil {
		log.Fatalf("error parsing HTML: %v", err)
	}

	listID := htmlquery.Find(doc, `//*[@id="exhibitor-results" or @id="featured-results"]//h3/a`)
	if len(listID) == 0 {
		log.Fatal("error parsing organization id")
	}

	log.Printf("Got %d orgs\n", len(listID))

	organizations = make([]main.Organization, 0, len(listID))

	for _, node := range listID {
		org := main.Organization{
			ID:               strings.SplitAfter(htmlquery.SelectAttr(node, "href"), "exhid=")[1],
			OrganizationName: htmlquery.InnerText(node),
		}
		organizations = append(organizations, org)
		log.Printf("%+v\n", org)
	}

	return organizations
}

func paginateHIMMS(ctx context.Context) error {
	for {
		err := func() error {
			ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			return chromedp.Run(ctx, chromedp.Tasks{
				chromedp.ScrollIntoView(`footer`, chromedp.ByQuery),
				chromedp.Click(`//a[@class="btn-secondary"]`, chromedp.BySearch),
				chromedp.Sleep(15 * time.Second),
			})
		}()
		if errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed making pagination: %w", err)
		}
	}
}
/********************************END_OF_FILE************************************/
package linkedin

func main() {
	// ...
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.UserDataDir(abs+"/user-data/"+credential.Email),
	)
	allocatorCtx, allocatorCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	// it depends how will you manage the lifecycle of the browser, maybe you don't want to call allocatorCancel() here
	defer allocatorCancel()
	browserCtx, browserCancel := chromedp.NewContext(allocatorCtx)
	// it depends how will you manage the lifecycle of the browser, maybe you don't want to call browserCancel() here
	defer browserCancel()

	// now start a browser
	if err := chromedp.Run(browserCtx); err != nil {
		log.Fatal(err)
	}

	// now browserCtx can be used directly
	// I will omit the timeout context here
	if err := chromedp.Run(browserCtx);// tasks
	err != nil {
		log.Fatal(err)
	}
	// or create new tabs in the browser
	tabCtx, cancel := chromedp.NewContext(browserCtx)
	defer cancel()
	// I will omit the timeout context here
	if err := chromedp.Run(tabCtx);// tasks
	err != nil {
		log.Fatal(err)
	}
	// All the tasks above will share the cookies
	// ...
}
/********************************END_OF_FILE************************************/
package app

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
)

type TempFile struct{ *os.File }

func NewTempFile(name string) (*TempFile, error) {
	file, err := os.Create(path.Join("tmp", name))
	if err != nil {
		return nil, fmt.Errorf("creating temp file %q: %w", name, err)
	}

	tmp := TempFile{File: file}

	return &tmp, nil
}

func (t *TempFile) Close() error {
	if err := t.File.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Remove(t.Name()); err != nil {
		return fmt.Errorf("removing temp file: %w", err)
	}

	return nil
}

func Convert(input []byte) ([]byte, error) {
	ogg, err := NewTempFile("voice.oga")
	if err != nil {
		return nil, fmt.Errorf("creating temp oga file: %w", err)
	}

	defer func() { _ = ogg.Close() }()

	mp3, err := NewTempFile("voice.mp3")
	if err != nil {
		return nil, fmt.Errorf("creating temp mp3 file: %w", err)
	}

	defer func() { _ = mp3.Close() }()

	if n, err := ogg.Write(input); err != nil {
		return nil, fmt.Errorf("writing oga file: %w", err)
	} else {
		log.Println("wrote", n, "bytes to", ogg.Name())
	}

	if err := convertOGGtoMP3(ogg.Name(), mp3.Name()); err != nil {
		return nil, fmt.Errorf("converting oga to mp3: %w", err)
	}

	buf := new(bytes.Buffer)

	if n, err := buf.ReadFrom(mp3); err != nil {
		return nil, fmt.Errorf("reading mp3 file: %w", err)
	} else {
		log.Println("read", n, "bytes from", mp3.Name())
	}

	return buf.Bytes(), nil
}

func convertOGGtoMP3(inputPath string, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", inputPath, "-vn", "-ar", "48000", "-ac", "1", "-ab", "36k", "-f", "mp3", "-y", "-loglevel", "quiet", outputPath)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("converting ogg to mp3: %w", err)
	}

	return nil
}
/********************************END_OF_FILE************************************/
package app

import "net/url"

const DefaultSchema = "https"

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
/********************************END_OF_FILE************************************/
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
		ID:      xid.New().String(),
		Date:    time.Now(),
		Model:   ModelGPT,
		History: []ChatMessage{{Role: RoleSystem, Content: os.Getenv("SYSTEM_MESSAGE_TUTOR")}},
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
		Values: make(url.Values),
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
/********************************END_OF_FILE************************************/
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
	FileSize     int    `json:"file_siz"`
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
	log.Printf("UDPATEREQUEST: %s\n", req.URL.String())

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

	log.Println(req.URL.String())

	req.Header.Set("Content-Type", "application/json")

	log.Printf("SEND MESSAGE REQUEST: %s\n", req.URL.String())

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
	log.Printf("FILE DATA REQUEST: %s\n", req.URL.String())

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
	u.Path = path.Join("file", b.Endpoint.URL.Path, file.FilePath)

	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("building request file: %w", err)
	}
	log.Printf("DOWNLOAD REQUEST: %s\n", req.URL.String())

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
		return nil, fmt.Errorf("eedaing body: %w", err)
	}

	return data, nil
}

func (m *MessageEntity) IsCommand() bool {
	log.Printf("TYPE: %v\n", m.Type)
	return m.Type == TypeBotCommand
}
/********************************END_OF_FILE************************************/
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
/********************************END_OF_FILE************************************/
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/delveper/env"
	"github.com/delveper/shadow/app"
)

func main() {
	if err := Run(); err != nil {
		log.Fatalln(err)
	}
}

func Run() error {
	if err := env.LoadVars(); err != nil {
		return fmt.Errorf("load envar: %w", err)
	}

	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	addr := host + ":" + port

	bot := app.NewTelegram()
	gpt := app.NewOpenAI()

	webhook := app.NewWebhook(bot, gpt)

	log.Printf("Starting server on port: %s\n", port)

	if err := http.ListenAndServe(addr, webhook); err != nil {
		return fmt.Errorf("serving: %w", err)
	}

	return nil
}
/********************************END_OF_FILE************************************/
```