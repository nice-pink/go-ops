package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nice-pink/go-ops/pkg/request"
)

func help() {
	fmt.Println(`
Usage:

Print help:
> ./request help

GET request:
> ./request -a get -url http://example.com

POST request:
> ./request -a post -url http://example.com -headers bla=blub -body @file.json
	`)
}

func main() {
	action := flag.String("a", "", "Action. Type '-action help' if needed!")
	url := flag.String("url", "", "Url to request.")
	repititions := flag.Int("n", 0, "Repititions.")
	body := flag.String("body", "", "Json body as string. Or read from file if starts with @.")
	headers := flag.String("headers", "", `e.g. "x-api-key=something,other-key=something-else"`)
	noRedirect := flag.Bool("noRedirect", false, "Don't follow redirects.")
	validate := flag.String("validate", "", "Validate e.g. [body-contains:hello, redirect-contains:dev, ...]")
	delay := flag.Int("delay", 1, "Delay between repititions in seconds.")
	verbose := flag.Bool("verbose", false, "Verbose.")
	publishMetrics := flag.Bool("publishMetrics", false, "Publish prometheus metrics.")
	flag.Parse()

	// help

	if *action == "" || strings.ToUpper(*action) == "HELP" {
		flag.Usage()
		help()
		os.Exit(0)
	}

	// set request module vars

	request.PublishMetrics = *publishMetrics

	// actions

	var resp *http.Response
	var err error

	// get request
	if strings.ToUpper(*action) == "GET" {
		if *url == "" {
			fmt.Println("Specify -url parameter!")
			os.Exit(2)
		}

		// repeated get
		if *repititions > 0 {
			request.RepeatedGet(*url, *headers, *repititions+1, *delay, *verbose, *noRedirect)
			os.Exit(0)
		}

		// simple get
		resp, err = request.Get(*url, *headers, *verbose, *noRedirect)
		if err != nil {
			os.Exit(2)
		} else if resp != nil && resp.StatusCode >= 400 {
			os.Exit(2)
		}
	}

	// post
	if strings.ToUpper(*action) == "POST" {
		if *url == "" || *body == "" {
			fmt.Println("Specify -url parameter!")
			os.Exit(2)
		}
		resp, err = request.Post(*url, *body, *headers, *verbose, *noRedirect)
		if err != nil {
			os.Exit(2)
		} else if resp != nil && resp.StatusCode >= 400 {
			os.Exit(2)
		}

	}

	if *validate != "" {
		if !request.Validate(resp, *validate) {
			os.Exit(2)
		}
	}

	os.Exit(0)
}
