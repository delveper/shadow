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
