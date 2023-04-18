package app

import (
	"net/url"
)

const DefaultSchema = "https"

type Endpoint struct {
	URL    *url.URL
	Values url.Values
}

func (e Endpoint) BuildURL(path string, params ...string) *url.URL {
	for i := 0; i < len(params); i += 2 {
		k, v := params[i], params[i+1]
		e.Values.Add(k, v)
	}

	u := *e.URL.JoinPath(path)
	u.RawQuery = e.Values.Encode()

	return &u
}
