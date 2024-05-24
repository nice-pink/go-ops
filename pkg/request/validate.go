package request

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

const (
	RedirectContains    string = "redirect-contains:"
	MultiRedirectsEqual string = "multi-redirects-equal"
	BodyContains        string = "body-contains:"
)

var (
	compareable string = ""
)

func Validate(resp *http.Response, validate string) bool {
	fmt.Print("Validate:", validate, " ")

	isValid := true
	if strings.HasPrefix(validate, RedirectContains) {
		// redirect contains string
		short, _ := strings.CutPrefix(validate, RedirectContains)
		if !isRedirectedTo(resp, short) {
			fmt.Println("❌ Is redirected. Status:", resp.StatusCode, "and does not contain", short)
		}
	} else if strings.HasPrefix(validate, BodyContains) {
		// body contains
		short, _ := strings.CutPrefix(validate, RedirectContains)
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			isValid = false
			if !isValid {
				fmt.Println("❌ Can't read body.")
			}
		}
		body := string(b)
		isValid = strings.Contains(body, short)
		if !isValid {
			fmt.Println("❌ Body does not contain:", short)
		}
	} else if strings.HasPrefix(validate, MultiRedirectsEqual) {
		if !isRedirectedTo(resp, compareable) {
			isValid = false
		} else if compareable == "" {
			// save for next try
			arr := strings.Split(validate, ":")
			location := resp.Header.Get("Location")
			if len(arr) == 1 {
				compareable = location
			} else if len(arr) > 2 {
				regex := regexp.MustCompile(arr[1])
				compareable = regex.ReplaceAllString(location, arr[2])
				fmt.Println()
				fmt.Println()
				fmt.Println("Assign comparator:", compareable)
			} else {
				fmt.Println("Error: Malformed regex. Needs REGEX:REPLACEMENT. E.g. multi-redirects-equal:(http://)(.*):${1}")
				return false
			}

		}
	}

	if isValid {
		fmt.Println(" ✅")
	} else {
		fmt.Println(" 💥")
	}

	return isValid
}

func isRedirectedTo(resp *http.Response, redirectContains string) bool {
	loc, err := resp.Location()
	if loc == nil || err != nil {
		// not redirected
		return false
	}
	if redirectContains == "" {
		return true
	}
	// check if redirect contains substring
	if !strings.Contains(resp.Header.Get("Location"), redirectContains) {
		fmt.Println()
		fmt.Println()
		fmt.Println(resp.Header.Get("Location"), "DOES NOT CONTAIN:", redirectContains)
		fmt.Println()
		return false
	}
	return true
}
