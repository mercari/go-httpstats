package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	stats "github.com/mercari/go-httpstats"
)

func main() {
	mw, err := stats.New()
	if err != nil {
		fmt.Println(err)
		return
	}
	handler := mw.WrapHandleFunc(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello world\n"))
		}))
	statsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d, err := mw.Data()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if err := json.NewEncoder(w).Encode(d); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	})
	http.Handle("/", handler)
	http.Handle("/stats", statsHandler)
	if err := http.ListenAndServe(":9999", nil); err != nil {
		fmt.Println(err)
	}

}
