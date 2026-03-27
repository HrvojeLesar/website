package eveonline

import (
	"context"
	"io"
	"net/http"
)

type customRequest struct{}

var CustomRequest customRequest

const (
	UserAgentKey        = "User-Agent"
	UserAgentValue      = "https://hrveklesarov.com/ Maintainer: Hrvoje (hrvoje.lesar1@hotmail.com)"
	AcceptEncodingKey   = "Accept-Encoding"
	AcceptEncodingValue = "gzip"
)

func (customRequest *customRequest) New(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	request.Header.Set(UserAgentKey, UserAgentValue)
	request.Header.Set(AcceptEncodingKey, AcceptEncodingValue)

	return request, nil
}
