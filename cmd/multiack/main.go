package main

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/nice-pink/goutil/pkg/log"
)

var (
	TIMEOUT         time.Duration
	ENDPOINTS       []string
	ENDPOINTS_ACKED []string
)

func main() {
	endpoints := flag.String("endpoints", "", "Comma separated endpoints to watch on.")
	endpointsJson := flag.String("endpointsJson", "", `Json array with endpoints to watch on. '[{"key":"value"},{"key":"value"}]'`)
	jsonKey := flag.String("jsonKey", "", "Key if endpointsJson is used.")
	method := flag.String("method", "get", "Accepted method.")
	timeout := flag.Int("timeout", 6, "Timeout for accept")
	flag.Parse()

	// start general server timeout
	startTimeoutTimer(*timeout)

	// extract endpoints
	getEndpoints(*endpoints, *endpointsJson, *jsonKey)

	for _, e := range ENDPOINTS {
		if strings.ToLower(*method) == "get" {
			log.Info("endpoint listening. get:", e)
			http.HandleFunc("/"+e, handleGet)
		} else if strings.ToLower(*method) == "post" {
			log.Info("endpoint listening. post:", e)
			http.HandleFunc("/"+e, handlePost)
		}
	}

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    TIMEOUT,
		WriteTimeout:   TIMEOUT,
		IdleTimeout:    TIMEOUT,
		MaxHeaderBytes: 1 << 20,
	}

	s.SetKeepAlivesEnabled(false)

	// start server, panic on error
	err := s.ListenAndServe()
	if err != nil {
		log.Err(err, "listen and serve error")
		os.Exit(2)
	}

	log.Info("Hooray")
}

// helper

func startTimeoutTimer(timeout int) {
	TIMEOUT = time.Duration(timeout) * time.Second
	if TIMEOUT > 0 {
		timer := time.NewTimer(TIMEOUT)
		go func() {
			<-timer.C
			log.Error("Timeout!")
			os.Exit(2)
		}()
	}
}

func getEndpoints(endpoints, endpointsJson, jsonKey string) {
	if endpoints != "" {
		log.Info("Get endpoints from -endpoints parameter")
		ENDPOINTS = strings.Split(endpoints, ",")
	} else {
		log.Info("Get endpoints from json array -endpointsJson, with key", jsonKey)
		var array []map[string]string
		err := json.Unmarshal([]byte(endpointsJson), &array)
		if err != nil {
			log.Err(err, "parse ")
		}
		for _, item := range array {
			ENDPOINTS = append(ENDPOINTS, item[jsonKey])
		}
	}
	sort.Strings(ENDPOINTS)
	log.Info(ENDPOINTS)
}

// http

func handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ackEndpoint(r.URL.Path)

	io.WriteString(w, "OK")
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	io.WriteString(w, "OK")
}

// ack

func ackEndpoint(e string) bool {
	ack := strings.TrimPrefix(e, "/")
	for _, endpoint := range ENDPOINTS_ACKED {
		if ack == endpoint {
			return true
		}
	}

	log.Info("received request from:", ack)
	ENDPOINTS_ACKED = append(ENDPOINTS_ACKED, ack)
	sort.Strings(ENDPOINTS_ACKED)

	if allAcked() {
		log.Info("All endpoints acked.")
		os.Exit(0)
	}
	return true
}

func allAcked() bool {
	return slices.Equal[[]string](ENDPOINTS, ENDPOINTS_ACKED)
}
