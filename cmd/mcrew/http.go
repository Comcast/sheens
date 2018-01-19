package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"golang.org/x/net/publicsuffix"
)

type Jar struct {
	*cookiejar.Jar
	Kookies []*http.Cookie `json:"cookies"`
}

func NewJar() (*Jar, error) {
	cookieJar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	return &Jar{Jar: cookieJar}, nil
}

func (j *Jar) AddCookies(cs []*http.Cookie) {
	if j.Kookies == nil {
		j.Kookies = make([]*http.Cookie, 0, 2*len(cs))
	}
	j.Kookies = append(j.Kookies, cs...)
}

// HTTPRequest is something I should quit re-implementing over and
// over.
type HTTPRequest struct {
	Id                string      `json:"id,omitempty"`
	Method            string      `json:"method,omitempty"`
	URL               string      `json:"url"`
	Body              string      `json:"body,omitempty"`
	Headers           http.Header `json:"headers,omitempty"`
	ResponseTimeoutMS int         `json:"timeout,omitempty"`
	CookieJar         *Jar        `json:"jar,omitempty"`

	Debug bool `json:"debug,omitempty"`

	// Various timeouts?  Or rely on Context?

	// TestResponse, if there, will be returned instead of
	// attempting a real HTTP request.
	TestResponse *HTTPResponse
}

type HTTPResponse struct {
	StatusCode  int          `json:"statusCode"`
	Status      string       `json:"status"`
	Error       error        `json:"error,omitempty"`
	Headers     http.Header  `json:"headers,omitempty"`
	Body        string       `json:"body,omitempty"`
	ContentType string       `json:"contentType,omitempty"`
	Request     *HTTPRequest `json:"request,omitempty"`

	// Parsed could be the Body parsed as (say) JSON.
	//
	// This field is not written by this code.  Instead, a request
	// Do handler (for example) could parse the Body and write
	// this field.
	Parsed interface{} `json:"parsed,omitempty"`
}

func (r *HTTPRequest) logf(format string, args ...interface{}) {
	if r.Debug {
		log.Printf(format, args...)
	}
}

// Do is the low-level, synchronous method to make the request and
// call the handler with the result.
func (r *HTTPRequest) Do(ctx context.Context, handler func(context.Context, *HTTPResponse) error) error {
	if r.TestResponse != nil {
		r.TestResponse.Request = r
		return handler(ctx, r.TestResponse)
	}

	url, err := url.Parse(r.URL)
	if err != nil {
		return err
	}

	req := &http.Request{
		Method: r.Method,
		URL:    url,
		Header: r.Headers,
		// ToDo: Context: ctx,
	}

	if r.Body != "" {
		req.Body = ioutil.NopCloser(bytes.NewReader([]byte(r.Body)))
	}

	// http.Request doesn't itself support CookieJars; instead,
	// http.Client does.  http.Client includes cached TCP
	// connections, so we shouldn't create http.Clients for each
	// request.
	//
	// We really don't want to try to manage a cache of
	// http.Clients.
	//
	// So we try to use a CookieJar manually with this request.
	// Yuck.  Scary, too.
	//
	// ToDo: Make more correct and audit and test and audit and
	// ...

	if r.CookieJar != nil {
		if req.Header == nil {
			req.Header = make(http.Header)
		}
		for i, cookie := range r.CookieJar.Cookies(url) {
			r.logf("adding cookie %d: %#v", i, cookie)
			req.AddCookie(cookie)
		}
	}

	req = req.WithContext(ctx)

	result := &HTTPResponse{
		Request: r,
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		r.logf("HTTPRequest.Do Do error %v", err)
		result.Error = err
		return handler(ctx, result)
	}

	result.Headers = resp.Header
	result.Status = resp.Status
	result.StatusCode = resp.StatusCode

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		r.logf("HTTPRequest.Do ReadAll error %v", err)
		result.Error = err
		return handler(ctx, result)
	}
	result.Body = string(body)

	if r.CookieJar != nil {
		r.logf("HTTPRequest.Do updating cookies")
		for _, c := range resp.Cookies() {
			r.logf("updating cookie %#v", c)
		}
		r.CookieJar.SetCookies(url, resp.Cookies())
		r.CookieJar.AddCookies(resp.Cookies())
	}

	if js, err := json.MarshalIndent(&result, "  ", "  "); err != nil {
		r.logf("HTTPResponse %#v", result)
	} else {
		r.logf("HTTPResponse\n%s\n", js)
	}

	return handler(ctx, result)
}
