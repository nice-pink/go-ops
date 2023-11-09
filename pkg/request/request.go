package request

import (
	"fmt"
	"net/http"
	"time"
)

func Get(url string, verbose bool) (*http.Response, error) {
	// request, track duration
	start := time.Now()
	res, err := http.Get(url)

	if err != nil {
		fmt.Println("Request error", err)
		return nil, err
	}

	if verbose {
		duration := time.Now().Sub(start)
		fmt.Print(duration, " - ", url, " ")

		if res.StatusCode == 200 {
			fmt.Println("✅")
		} else if res.StatusCode < 400 {
			fmt.Println("⚠️ Status code != 200:", res.StatusCode)
		} else {
			fmt.Println("🔥 Status code != 200:", res.StatusCode)
		}
	}

	return res, nil
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
