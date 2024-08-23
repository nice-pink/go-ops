package ack_simple

import (
	"fmt"
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
	TIMEOUT              time.Duration
	ENDPOINTS            []string
	ENDPOINTS_ACKED      []string
	WEBHOOK_SUCCESS      string
	WEBHOOK_SUCCESS_BODY string
	WEBHOOK_FAILED       string
	WEBHOOK_FAILED_BODY  string
)

// helper

func StartTimeoutTimer(timeout int) {
	TIMEOUT = time.Duration(timeout) * time.Second
	if TIMEOUT > 0 {
		timer := time.NewTimer(TIMEOUT)
		go func() {
			<-timer.C
			log.Error("Timeout!")
			triggerWebhook(false)
			os.Exit(2)
		}()
	}
}

func GetEndpoints(endpoints string) bool {
	if endpoints == "" {
		log.Error("Empty endpoint string")
		return false
	}
	log.Info("Endpoints:")
	ENDPOINTS = strings.Split(endpoints, ",")
	sort.Strings(ENDPOINTS)
	log.Info(ENDPOINTS)
	return true
}

// http

func HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ackEndpoint(r.URL.Path)
	io.WriteString(w, "OK")
}

func HandlePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	io.WriteString(w, "OK")
	ackEndpoint(r.URL.Path)
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
		triggerWebhook(true)
		os.Exit(0)
	}
	return true
}

func allAcked() bool {
	return slices.Equal[[]string](ENDPOINTS, ENDPOINTS_ACKED)
}

// webhooks

func triggerWebhook(success bool) {
	var reader io.Reader
	var err error
	var resp *http.Response
	statusCode := -1
	if success && WEBHOOK_SUCCESS != "" {
		fmt.Println("success webhook:", WEBHOOK_SUCCESS, "body:", WEBHOOK_SUCCESS_BODY)
		if WEBHOOK_SUCCESS_BODY != "" {
			reader = strings.NewReader(WEBHOOK_SUCCESS_BODY)
		}
		resp, err = http.Post(WEBHOOK_SUCCESS, "application/json", reader)

	} else if !success && WEBHOOK_FAILED != "" {
		fmt.Println("failed webhook:", WEBHOOK_FAILED, "body:", WEBHOOK_FAILED_BODY)
		if WEBHOOK_FAILED_BODY != "" {
			reader = strings.NewReader(WEBHOOK_FAILED_BODY)
		}
		resp, err = http.Post(WEBHOOK_FAILED, "application/json", reader)
	}

	// error log
	if err != nil {
		fmt.Println("Webhook error:")
		fmt.Println(err)
		return
	}
	if resp != nil && resp.StatusCode != 200 {
		fmt.Println("Error: Status code != 200. ->", statusCode)
	}
}
