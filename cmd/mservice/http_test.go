package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPBasic(t *testing.T) {
	debug := false

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if debug {
			for i, cookie := range r.Cookies() {
				log.Printf("server handler cookie %d: %#v", i, cookie)
			}
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "likes",
			Value: "tacos",
		})
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	jar, err := NewJar()
	if err != nil {
		t.Fatal(err)
	}

	req := HTTPRequest{
		URL:       ts.URL,
		CookieJar: jar,
	}

	saw := make(chan []byte, 2)

	handler := func(ctx context.Context, r *HTTPResponse) error {
		js, err := json.MarshalIndent(&r, "  ", "  ")
		if err != nil {
			return err
		}
		saw <- js
		return nil
	}

	if err = req.Do(ctx, handler); err != nil {
		t.Fatal(err)
	}

	if err = req.Do(ctx, handler); err != nil {
		t.Fatal(err)
	}

	<-saw
	<-saw
}
