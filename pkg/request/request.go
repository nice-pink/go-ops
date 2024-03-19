package request

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nice-pink/go-ops/pkg/metric"
)

var (
	PublishMetrics bool = false
)

func SetupMetrics(port int) {
	fmt.Println("Publish metrics about auto-healing.")
	go func() {
		metric.Listen(port)
	}()
}

// get

func Get(url string, headers string, verbose bool, noRedirect bool) (*http.Response, error) {
	// request, track duration
	start := time.Now()

	// request
	request, err := getNewRequest("GET", url, headers, "")
	if err != nil {
		return nil, err
	}

	return doRequest(request, url, "get", noRedirect, start, verbose)
}

func RepeatedGet(url string, headers string, repititions int, delay int, verbose bool, noRedirect bool) {
	for i := 0; i < repititions; i++ {
		Get(url, headers, verbose, noRedirect)

		// delay
		if i < repititions-1 {
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}
}

// post

func Post(url string, body string, headers string, verbose bool, noRedirect bool) (*http.Response, error) {
	// request, track duration
	start := time.Now()

	fmt.Println(headers)
	fmt.Println(body)

	if strings.HasPrefix(body, "@") {
		filename := strings.TrimPrefix(body, "@")
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println("Could not read file.")
			fmt.Println(err)
			return nil, err
		}
		body = string(data)
	}

	// request
	request, err := getNewRequest("POST", url, headers, body)
	if err != nil {
		return nil, err
	}

	return doRequest(request, url, "post", noRedirect, start, verbose)
}

// helper

func getNewRequest(method string, url string, headers string, body string) (*http.Request, error) {
	var request *http.Request
	var err error

	method = strings.ToUpper(method)
	// post request
	if body != "" {
		request, err = http.NewRequest(method, url, strings.NewReader(body))
	} else {
		request, err = http.NewRequest(method, url, nil)
	}

	// headers: "x-api-key=something,other-key=something-else"
	if headers != "" {
		headerList := strings.Split(headers, ",")
		for _, item := range headerList {
			item = strings.TrimSpace(item)
			header := strings.Split(item, "=")
			request.Header.Add(header[0], header[1])
		}
	}

	return request, err
}

func doRequest(request *http.Request, url string, method string, noRedirect bool, start time.Time, verbose bool) (*http.Response, error) {
	// client
	client := &http.Client{}
	if noRedirect {
		// return the error, so client won't attempt redirects
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return errors.New("redirect")
		}
	}

	resp, err := client.Do(request)
	if resp != nil && resp.StatusCode == http.StatusFound { //status code 302
		fmt.Println(resp.Location())
		return resp, nil
	} else if err != nil {
		fmt.Println("❌ Request error", err)
		return nil, err
	}

	// handle response
	if verbose {
		duration := time.Since(start)
		fmt.Print(duration, " - ", url, " ")

		if resp.StatusCode < 300 {
			fmt.Println("✅ Status code:", resp.StatusCode)
		} else if resp.StatusCode < 400 {
			fmt.Println("⤴️ Redirect:", resp.StatusCode)
		} else {
			fmt.Println("🔥 Status code:", resp.StatusCode)
		}

		// print body if post
		if strings.ToLower(method) == "post" {
			body, _ := io.ReadAll(resp.Body)
			fmt.Println(string(body))
		}
	}

	// publish metrics
	if PublishMetrics {
		ResponseMetrics(resp.StatusCode)
	}

	return resp, err
}
