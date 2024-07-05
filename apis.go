package apis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

type params struct {
	method        string
	path          string
	body          io.Reader
	queryPrams    url.Values
	header        http.Header
	headerAppaned map[string]bool
}

type requestOption struct {
	fn  func(*params)
	err error
}

func Path(path string) requestOption {
	return requestOption{
		fn: func(o *params) { o.path = path },
	}
}

func Body(v any) requestOption {
	var buff bytes.Buffer
	var err = json.NewEncoder(&buff).Encode(v)
	if err != nil {
		slog.Error("encoding json data", "error", err)
	}
	return requestOption{
		fn:  func(o *params) { o.body = &buff },
		err: err,
	}
}

func QueryParam(key, value string, append ...bool) requestOption {
	return requestOption{
		fn: func(o *params) {
			if len(append) > 0 && append[0] {
				o.queryPrams.Add(key, value)
				return
			}
			o.queryPrams.Set(key, value)
		},
	}
}

func Header(key, value string, append ...bool) requestOption {
	return requestOption{
		fn: func(o *params) {
			if len(append) > 0 && append[0] {
				o.header.Add(key, value)
				o.headerAppaned[key] = true
				return
			}
			o.header.Set(key, value)
		},
	}
}

type Response interface {
	Scan(v any) error
	BodyReader() io.Reader
	BodyLen() int64
	Err() error
	StatusCode() int
}
type response struct {
	statusCode int
	body       io.Reader
	bytesLen   int64
	err        error
}

func (r *response) Scan(v any) error {
	if r.err != nil {
		return r.err
	}
	return json.NewDecoder(r.body).Decode(&v)
}
func (r *response) BodyReader() io.Reader { return r.body }
func (r *response) BodyLen() int64        { return r.bytesLen }
func (r *response) Err() error            { return r.err }
func (r *response) StatusCode() int       { return r.statusCode }

func MakeRequest(ctx context.Context, apiurl string, requestOptions ...requestOption) Response {
	var options = params{
		method:        http.MethodGet,
		body:          nil,
		queryPrams:    make(url.Values),
		header:        make(http.Header),
		headerAppaned: make(map[string]bool),
	}
	var err error
	for _, ro := range requestOptions {
		if ro.err != nil {
			err = errors.Join(err, ro.err)
		}
		ro.fn(&options)
	}
	if err != nil {
		slog.Error("Failed to parse request options", "error", err, "url", apiurl)
		return &response{err: err}
	}
	endpoint, err := url.JoinPath(apiurl, options.path)
	if err != nil {
		slog.Error(
			"Failed to join base server URL with path",
			"error", err,
			"apiurl", apiurl,
			"path", options.path,
		)
		return &response{err: err}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		slog.Error(
			"Failed to create request",
			"error", err,
			"endpoint", endpoint,
		)
		return &response{err: err}
	}
	for key, values := range options.header {
		if !options.headerAppaned[key] {
			req.Header.Del(key)
		}
		for _, val := range values {
			req.Header.Add(key, val)
		}
	}
	req.URL.RawQuery = options.queryPrams.Encode()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(
			"Request failed",
			"error", err,
			"endpoint", endpoint,
			"method", options.method,
		)
		return &response{err: err}
	}
	defer res.Body.Close()
	var buff bytes.Buffer
	written, err := io.Copy(&buff, res.Body)
	return &response{err: err, statusCode: res.StatusCode, body: &buff, bytesLen: written}
}
