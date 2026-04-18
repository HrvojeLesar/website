package testingutils

import (
	"bytes"
	"io"
	"net/http"
)

type MockHTTPClient struct {
	Response *http.Response
	Err      error
	Called   bool
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	m.Called = true

	return m.Response, m.Err
}

type readerCloser struct {
	*bytes.Reader
}

func (readerCloser) Close() error { return nil }

func (m *MockHTTPClient) MockBody(data []byte) io.ReadCloser {
	return readerCloser{bytes.NewReader(data)}
}

