package app

import (
	"net/url"
)

const ( // https://api.telegram.org/bot<token>/METHOD_NAME
	MethodGetMe         = "getMe"
	MethodGetUpdates    = "getUpdates"
	MethodDeleteWebhook = "deleteWebhook" //
	MethodSetWebhook    = "setWebhook"    // ?url={your_API_server_url}
	MethodSendMessage   = "sendMessage"   // ?chat_id={chat_id}&text={text}
)

type Endpoint struct {
	URL  *url.URL
	Path string
}

func NewEndpoint(token string) *Endpoint {
	p := "bot" + token

	u := url.URL{
		Scheme: "https",
		Host:   "api.telegram.org",
		Path:   p,
	}

	return &Endpoint{
		URL:  &u,
		Path: p,
	}
}
