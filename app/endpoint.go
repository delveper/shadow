package app

import (
	"net/url"
)

const DefaultSchema = "https"

type Endpoint struct {
	URL *url.URL
}

func (e Endpoint) BuildURL(path string, params ...string) *url.URL {
	values := make(url.Values)
	for i := 0; i < len(params); i += 2 {
		k, v := params[i], params[i+1]
		values.Add(k, v)
	}

	u := *e.URL.JoinPath(path)
	u.RawQuery = values.Encode()

	return &u
}
