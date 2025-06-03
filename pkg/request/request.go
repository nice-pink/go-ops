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

func Get(url, headers string, verbose, noRedirect bool, retries int) (*http.Response, error) {
	// request, track duration
	start := time.Now()

	// request
	request, err := getNewRequest("GET", url, headers, "", verbose)
	if err != nil {
		return nil, err
	}

	return doRequest(request, url, "get", noRedirect, start, verbose, retries)
}

// post

func Post(url, body, headers string, verbose, noRedirect bool, retries int) (*http.Response, error) {
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
	request, err := getNewRequest("POST", url, headers, body, verbose)
	if err != nil {
		return nil, err
	}

	return doRequest(request, url, "post", noRedirect, start, verbose, retries)
}

// helper

func getNewRequest(method string, url string, headers string, body string, verbose bool) (*http.Request, error) {
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

	if verbose {
		printMyIP()
	}

	return request, err
}

func doRequest(request *http.Request, url, method string, noRedirect bool, start time.Time, verbose bool, retries int) (*http.Response, error) {
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
		host, _ := resp.Location()
		fmt.Println(host)
		// return resp, nil
		err = nil
	} else if err != nil {
		fmt.Println("❌ Request error", err)
		return nil, err
	}
	defer resp.Body.Close()

	if retries > 0 {
		if resp.StatusCode >= 400 {
			return doRequest(request, url, method, noRedirect, start, verbose, retries-1)
		}
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

func printMyIP() {
	// http://ifconfig.me/ip
	req, _ := http.NewRequest(http.MethodGet, "http://checkip.amazonaws.com/", nil)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("My ip:", string(body))
}
