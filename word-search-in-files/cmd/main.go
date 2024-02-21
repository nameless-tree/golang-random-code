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
	"word-search-in-files/internal/args"
	"word-search-in-files/pkg/searcher"
)

type ErrorResponse struct {
	Errors []error `json:"errors"`
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
		strErrors := make([]string, len(errors))

		for i, err := range errors {
			strErrors[i] = err.Error()
		}

		jsonData, e = json.Marshal(strErrors)
		if e != nil {
			http.Error(w, "Err: encoding JSON", http.StatusInternalServerError)
			return
		}
		status = http.StatusInternalServerError
	} else {
		jsonData, e = json.Marshal(files)
		if e != nil {
			http.Error(w, "Err: encoding JSON", http.StatusInternalServerError)
			return
		}
		status = http.StatusOK
	}

	w.WriteHeader(status)
	_, e = w.Write(jsonData)
	if e != nil {
		fmt.Println("Err: writing response:", e)
	}
}

func main() {
	args := args.ArgsParse()

	srch, e := searcher.NewSearcher(args.Path)
	if e != nil {
		log.Println(e)
		return
	}

	// Not used
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go srch.ScanPeriodically(ctx, wg, time.Hour)
	// Wait for the 'init' scan to complete to do not run other code first (http handler)
	wg.Wait()

	mux := http.NewServeMux()
	mux.HandleFunc("/files/search", func(w http.ResponseWriter, r *http.Request) {
		searchHandler(w, r, srch)
	})

	e = http.ListenAndServe(args.HttpAddr, mux)

	if errors.Is(e, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if e != nil {
		fmt.Printf("error starting server: %s\n", e)
		os.Exit(1)
	}

}
