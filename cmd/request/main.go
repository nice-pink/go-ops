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

Get request:
> ./request -a get -url http://example.com
	`)
}

func main() {
	action := flag.String("a", "", "Action. Type '-action help' if needed!")
	url := flag.String("url", "", "Url to request.")
	repititions := flag.Int("n", 0, "Repititions.")
	delay := flag.Int("delay", 1, "Delay between repititions in seconds.")
	verbose := flag.Bool("verbose", false, "Verbose.")
	flag.Parse()

	// help

	if *action == "" || strings.ToUpper(*action) == "HELP" {
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
	}
}
