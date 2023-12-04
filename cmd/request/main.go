package main

import (
	"flag"
	"fmt"
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
	delay := flag.Int("delay", 1, "Delay between repititions in seconds.")
	verbose := flag.Bool("verbose", false, "Verbose.")
	flag.Parse()

	// help

	if *action == "" || strings.ToUpper(*action) == "HELP" {
		flag.Usage()
		help()
		os.Exit(0)
	}

	// actions

	if strings.ToUpper(*action) == "GET" {
		if *url == "" {
			fmt.Println("Specify -url parameter!")
			os.Exit(2)
		}
		request.RepeatedGet(*url, *repititions+1, *delay, *verbose)
		os.Exit(0)
	}

	if strings.ToUpper(*action) == "POST" {
		if *url == "" || *body == "" {
			fmt.Println("Specify -url parameter!")
			os.Exit(2)
		}
		request.Post(*url, *body, *headers, *verbose)
		os.Exit(0)
	}
}
