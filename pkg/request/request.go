package request

import (
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

func Get(url string, verbose bool) (*http.Response, error) {
	// request, track duration
	start := time.Now()
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("Request error", err)
		return nil, err
	}

	if verbose {
		duration := time.Now().Sub(start)
		fmt.Print(duration, " - ", url, " ")

		if resp.StatusCode < 300 {
			fmt.Println("✅ Status code:", resp.StatusCode)
		} else if resp.StatusCode < 400 {
			fmt.Println("⤴️ Redirect. Status code:", resp.StatusCode)
		} else {
			fmt.Println("🔥 Error. Status code:", resp.StatusCode)
		}
	}

	if PublishMetrics {
		ResponseMetrics(resp.StatusCode)
	}

	return resp, nil
}

func RepeatedGet(url string, repititions int, delay int, verbose bool) {
	for i := 0; i < repititions; i++ {
		Get(url, verbose)

		// delay
		if i < repititions-1 {
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}
}

// post

func Post(url string, body string, headers string, verbose bool) (*http.Response, error) {
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

	request, err := http.NewRequest("POST", url, strings.NewReader(body))

	// headers: "x-api-key=something,other-key=something-else"
	headerList := strings.Split(headers, ",")
	for _, item := range headerList {
		item = strings.TrimSpace(item)
		header := strings.Split(item, "=")
		request.Header.Add(header[0], header[1])
	}

	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		fmt.Println("Request error", err)
		return nil, err
	}

	if verbose {
		duration := time.Now().Sub(start)
		fmt.Print(duration, " - ", url, " ")

		if resp.StatusCode < 300 {
			fmt.Println("✅ Status code:", resp.StatusCode)
		} else if resp.StatusCode < 400 {
			fmt.Println("⤴️ Redirect:", resp.StatusCode)
		} else {
			fmt.Println("🔥 Status code:", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
	}

	if PublishMetrics {
		ResponseMetrics(resp.StatusCode)
	}

	return resp, nil
}
