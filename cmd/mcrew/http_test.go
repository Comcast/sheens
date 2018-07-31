/* Copyright 2018 Comcast Cable Communications Management, LLC
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
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
				Logf("server handler cookie %d: %#v", i, cookie)
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
