package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"word-search-in-files/pkg/searcher"
)

const (
	addr = ":3333"
	path = "examples/simple"
)

func encodeJSON(data any) ([]byte, error) {
	return json.Marshal(data)
}

func searchHandler(w http.ResponseWriter, r *http.Request, srch *searcher.Searcher) {
	if r.Method != http.MethodGet {
		http.Error(w, "Err: only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	word := r.URL.Query().Get("word")
	if word == "" {
		http.Error(w, "Err: word parameter is required", http.StatusBadRequest)
		return
	}

	var jsonData []byte
	var e error
	var status int

	files, errors := srch.Search(word)
	if errors != nil {
		jsonData, e = encodeJSON(errors)
		if e != nil {
			http.Error(w, "Err: encoding JSON", http.StatusInternalServerError)
			return
		}
		status = http.StatusInternalServerError
	} else {
		jsonData, _ = encodeJSON(files)
		if e != nil {
			http.Error(w, "Err: encoding JSON", http.StatusInternalServerError)
			return
		}
		if files == nil {
			status = http.StatusNotFound
		} else {
			status = http.StatusOK
		}
	}

	w.WriteHeader(status)
	_, e = w.Write(jsonData)
	if e != nil {
		fmt.Println("Err: writing response:", e)
	}
}

func scanPeriodically(srch *searcher.Searcher, wg *sync.WaitGroup, ctx context.Context, interval time.Duration) {

	srch.Scan()

	wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			srch.Scan()
		}
	}
}

func main() {

	srch, e := searcher.NewSearcher(path)
	if e != nil {
		log.Println(e)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go scanPeriodically(srch, wg, ctx, time.Hour)
	wg.Wait()

	mux := http.NewServeMux()
	mux.HandleFunc("/files/search", func(w http.ResponseWriter, r *http.Request) {
		searchHandler(w, r, srch)
	})

	e = http.ListenAndServe(addr, mux)

	if errors.Is(e, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if e != nil {
		fmt.Printf("error starting server: %s\n", e)
		os.Exit(1)
	}
}
