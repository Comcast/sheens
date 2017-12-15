package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime/pprof"

	_ "net/http/pprof"
)

func (s *Service) HTTPServer(ctx context.Context, port string) error {
	log.Printf("Service.HTTPServer starting on %s", port)

	complain := func(w http.ResponseWriter, x interface{}, status int) {
		w.WriteHeader(status)
		fmt.Fprintf(w, `{"error":"%s"}`+"\n", x)
	}

	http.Handle("/goroutines", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pprof.Lookup("goroutine").WriteTo(w, 1)
	}))

	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		js, err := ioutil.ReadAll(r.Body)
		if err != nil {
			complain(w, err, http.StatusBadRequest)
			return
		}
		if err := r.Body.Close(); err != nil {
			log.Printf("Service.HTTPServer warning on Body.Close(): %v", err)
		}

		var op SOp
		if err := json.Unmarshal(js, &op); err != nil {
			complain(w, err, http.StatusBadRequest)
			return
		}
		if err = op.Do(ctx, s); err != nil {
			complain(w, err, http.StatusInternalServerError)
			return
		}
		js, err = json.Marshal(&op)
		if err != nil {
			complain(w, err, http.StatusInternalServerError)
		}
		if _, err = w.Write(js); err != nil {
			log.Printf("Service.HTTPServer warning on Write(): %v", err)
		}
	}))

	return http.ListenAndServe(port, nil)

}
