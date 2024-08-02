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
	TIMEOUT              time.Duration
	RUNNING              bool     = false
	ENDPOINTS            []string = nil
	ENDPOINTS_ACKED      []string = nil
	WEBHOOK_SUCCESS      string
	WEBHOOK_SUCCESS_BODY string
	WEBHOOK_FAILED       string
	WEBHOOK_FAILED_BODY  string
)

func main() {
	timeout := flag.Int("timeout", 3600, "Timeout for accept")
	webhookSuccess := flag.String("webhookSuccess", "", "Webhook to trigger on success.")
	webhookSuccessBody := flag.String("webhookSuccessBody", "", "Webhook body to send on success.")
	webhookFailed := flag.String("webhookFailed", "", "Webhook to trigger on success.")
	webhookFailedBody := flag.String("webhookFailedBody", "", "Webhook body to send on success.")
	flag.Parse()

	// setup global values
	TIMEOUT = time.Duration(*timeout) * time.Second
	WEBHOOK_SUCCESS = *webhookSuccess
	WEBHOOK_SUCCESS_BODY = *webhookSuccessBody
	WEBHOOK_FAILED = *webhookFailed
	WEBHOOK_FAILED_BODY = *webhookFailedBody

	// start endpoint
	http.HandleFunc("/endpoints", postEndpoints)
	http.HandleFunc("/endpointsJson", postEndpointsJson)
	http.HandleFunc("/ack", postAck)

	// start server, panic on error
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Err(err, "listen and serve error")
		os.Exit(2)
	}
}

// helper

func startTimeoutTimer() {
	if TIMEOUT > 0 {
		timer := time.NewTimer(TIMEOUT)
		go func() {
			<-timer.C
			if !RUNNING {
				return
			}
			failedAck("Timeout!")
		}()
	}
}

func failedAck(msg string) {
	log.Error(msg)
	log.Info("Requested endpoints:", ENDPOINTS)
	log.Info("Acked endpoints:", ENDPOINTS_ACKED)

	triggerWebhook(WEBHOOK_FAILED, WEBHOOK_FAILED_BODY)
	RUNNING = false
}

func finalizeEndpointsSetup() {
	sort.Strings(ENDPOINTS)
	log.Info("Waiting for items:", ENDPOINTS)
}

func start() {
	RUNNING = true
	startTimeoutTimer()
}

func success() {
	log.Info("All endpoints acked.")
	// webhook
	triggerWebhook(WEBHOOK_SUCCESS, WEBHOOK_SUCCESS_BODY)
	ENDPOINTS = nil
	ENDPOINTS_ACKED = nil
	RUNNING = false
}

// webhook

func triggerWebhook(url string, body string) {
	if url == "" {
		return
	}

	if body == "" {
		// send get request
		r, err := http.Get(url)
		if err != nil || r.StatusCode != 200 {
			log.Err(err, "webhook could not be triggered. status code:", r.StatusCode)
		}
	} else {
		// send post request
		bodyReader := strings.NewReader(body)
		r, err := http.Post(url, "application/json", bodyReader)
		if err != nil || r.StatusCode != 200 {
			log.Err(err, "webhook could not be triggered. status code:", r.StatusCode)
		}
	}
}

// http

func postAck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// parse request
	endpoint, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `endpoint`, http.StatusBadRequest)
		return
	}

	ackEndpoint(string(endpoint))
	io.WriteString(w, "OK")
}

func postEndpoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// parse request
	endpoints, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `cannot read endpointsJson body`, http.StatusBadRequest)
		return
	}

	// body: end1,end2
	ENDPOINTS = strings.Split(string(endpoints), ",")
	finalizeEndpointsSetup()
	start()
}

type EndpointsJson struct {
	Key       string
	Endpoints []map[string]string
}

func postEndpointsJson(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// parse request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `cannot read endpointsJson body`, http.StatusBadRequest)
		return
	}

	// body: {"Key": "endpointKey", "Endpoints": [{"endpointKey":"end1"},{"endpointKey":"end2"}]}
	var endpointsBody EndpointsJson
	err = json.Unmarshal(body, &endpointsBody)
	if err != nil {
		log.Err(err, "cannot unmarshal endpoints body")
	}

	ENDPOINTS = make([]string, 0)
	for _, item := range endpointsBody.Endpoints {
		ENDPOINTS = append(ENDPOINTS, item[endpointsBody.Key])
	}
	finalizeEndpointsSetup()
	start()
}

// ack

func ackEndpoint(e string) bool {
	log.Info("ack", e)
	if ENDPOINTS_ACKED == nil {
		ENDPOINTS_ACKED = make([]string, 0)
	}

	for _, endpoint := range ENDPOINTS_ACKED {
		if e == endpoint {
			return true
		}
	}

	log.Info("received request from:", e)
	ENDPOINTS_ACKED = append(ENDPOINTS_ACKED, e)
	sort.Strings(ENDPOINTS_ACKED)

	if allAcked() {
		success()
	}
	return true
}

func allAcked() bool {
	return slices.Equal[[]string](ENDPOINTS, ENDPOINTS_ACKED)
}
