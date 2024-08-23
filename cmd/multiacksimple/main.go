package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/nice-pink/go-ops/pkg/ack_simple"
	"github.com/nice-pink/goutil/pkg/log"
)

func main() {
	endpoints := flag.String("endpoints", "", "Comma separated endpoints to watch on.")
	endpointsJson := flag.String("endpointsJson", "", `Json array with endpoints to watch on. '[{"key":"value"},{"key":"value"}]'`)
	jsonKey := flag.String("jsonKey", "", "Key if endpointsJson is used.")
	method := flag.String("method", "get", "Accepted method.")
	timeout := flag.Int("timeout", 6, "Timeout for accept")
	webhookSuccess := flag.String("webhookSuccess", "", "Webhook to trigger on success.")
	webhookSuccessBody := flag.String("webhookSuccessBody", "", "Webhook body to send on success.")
	webhookFailed := flag.String("webhookFailed", "", "Webhook to trigger on success.")
	webhookFailedBody := flag.String("webhookFailedBody", "", "Webhook body to send on success.")
	flag.Parse()

	// webhooks
	ack_simple.WEBHOOK_SUCCESS = *webhookSuccess
	ack_simple.WEBHOOK_SUCCESS_BODY = *webhookSuccessBody
	ack_simple.WEBHOOK_FAILED = *webhookFailed
	ack_simple.WEBHOOK_FAILED_BODY = *webhookFailedBody

	// start general server timeout
	ack_simple.StartTimeoutTimer(*timeout)

	// extract endpoints
	getEndpoints(*endpoints, *endpointsJson, *jsonKey)

	for _, e := range ack_simple.ENDPOINTS {
		if strings.ToLower(*method) == "get" {
			log.Info("endpoint listening. get:", e)
			http.HandleFunc("/"+e, ack_simple.HandleGet)
		} else if strings.ToLower(*method) == "post" {
			log.Info("endpoint listening. post:", e)
			http.HandleFunc("/"+e, ack_simple.HandlePost)
		}
	}

	s := &http.Server{
		Addr:           ":8080",
		ReadTimeout:    ack_simple.TIMEOUT,
		WriteTimeout:   ack_simple.TIMEOUT,
		IdleTimeout:    ack_simple.TIMEOUT,
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

func getEndpoints(endpoints, endpointsJson, jsonKey string) {
	if endpoints != "" {
		ack_simple.GetEndpoints(endpoints)
		return
	} else {
		log.Info("Get endpoints from json array -endpointsJson, with key", jsonKey)
		var array []map[string]string
		err := json.Unmarshal([]byte(endpointsJson), &array)
		if err != nil {
			log.Err(err, "parse ")
		}
		for _, item := range array {
			ack_simple.ENDPOINTS = append(ack_simple.ENDPOINTS, item[jsonKey])
		}
	}
	sort.Strings(ack_simple.ENDPOINTS)
	log.Info(ack_simple.ENDPOINTS)
}

// test
// terminal 1:
// nc -l 8888
// terminal 2:
// bin/multiacksimple -timeout 3600 -endpoints hello,bello -webhookSuccess http://localhost:8888/ -webhookSuccessBody '{"key":"value"}'
